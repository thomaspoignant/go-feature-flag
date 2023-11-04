package org.gofeatureflag.openfeature

import com.google.gson.Gson
import dev.gustavoavila.websocketclient.WebSocketClient
import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.FeatureProvider
import dev.openfeature.sdk.Hook
import dev.openfeature.sdk.OpenFeatureAPI.getEvaluationContext
import dev.openfeature.sdk.ProviderEvaluation
import dev.openfeature.sdk.ProviderMetadata
import dev.openfeature.sdk.Reason
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.exceptions.ErrorCode
import dev.openfeature.sdk.exceptions.OpenFeatureError.FlagNotFoundError
import dev.openfeature.sdk.exceptions.OpenFeatureError.GeneralError
import okhttp3.ConnectionPool
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.OkHttpClient
import okhttp3.RequestBody.Companion.toRequestBody
import org.gofeatureflag.openfeature.bean.FlagState
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.gofeatureflag.openfeature.bean.GoffRequest
import org.gofeatureflag.openfeature.bean.GoffResponse
import org.gofeatureflag.openfeature.bean.ProviderStatus
import org.gofeatureflag.openfeature.exception.InvalidEndpoint
import java.net.HttpURLConnection.HTTP_BAD_REQUEST
import java.net.HttpURLConnection.HTTP_UNAUTHORIZED
import java.net.URI
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.TimeUnit


class GoFeatureFlagProvider(private val options: GoFeatureFlagOptions) : FeatureProvider {
    companion object {
        private val gson = Gson()
    }

    private var httpClient: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(this.options.timeout, TimeUnit.MILLISECONDS)
        .readTimeout(this.options.timeout, TimeUnit.MILLISECONDS)
        .callTimeout(this.options.timeout, TimeUnit.MILLISECONDS)
        .writeTimeout(this.options.timeout, TimeUnit.MILLISECONDS)
        .connectionPool(
            ConnectionPool(
                this.options.maxIdleConnections,
                this.options.keepAliveDuration,
                TimeUnit.MILLISECONDS
            )
        )
        .build()
    private var parsedEndpoint: HttpUrl? = options.endpoint.toHttpUrlOrNull()
    private var flags: ConcurrentHashMap<String, FlagState> = ConcurrentHashMap()
    private var status: ProviderStatus = ProviderStatus.NOT_READY
    private var goffWebsocketClient: WebSocketClient? = null

    init {
        if (this.parsedEndpoint == null) {
            throw InvalidEndpoint()
        }
    }

    override val hooks: List<Hook<*>>
        get() = listOf()

    override val metadata: ProviderMetadata
        get() = GoFeatureFlagMetadata()

    override fun getBooleanEvaluation(
        key: String,
        defaultValue: Boolean,
        context: EvaluationContext?
    ): ProviderEvaluation<Boolean> {
        return this.evaluate<Boolean>(key, listOf("Boolean"), defaultValue)
    }

    override fun getDoubleEvaluation(
        key: String,
        defaultValue: Double,
        context: EvaluationContext?
    ): ProviderEvaluation<Double> {
        return this.evaluate<Double>(key, listOf("Double"), defaultValue)
    }

    override fun getIntegerEvaluation(
        key: String,
        defaultValue: Int,
        context: EvaluationContext?
    ): ProviderEvaluation<Int> {
        return this.evaluate<Int>(key, listOf("Int"), defaultValue)
    }

    override fun getObjectEvaluation(
        key: String,
        defaultValue: Value,
        context: EvaluationContext?
    ): ProviderEvaluation<Value> {
        return this.evaluate<Value>(key, listOf("LinkedTreeMap", "ArrayList"), defaultValue)
    }

    override fun getStringEvaluation(
        key: String,
        defaultValue: String,
        context: EvaluationContext?
    ): ProviderEvaluation<String> {
        return this.evaluate<String>(key, listOf("String"), defaultValue)
    }

    override fun initialize(initialContext: EvaluationContext?) {
        try {
            fetchAllFlags(initialContext)
            this.goffWebsocketClient =
                object : WebSocketClient(buildWebsocketURI(this.options)) {
                    init {
                        this.enableAutomaticReconnection(options.retryDelay)
                        this.setReadTimeout(options.timeout.toInt())
                        this.setConnectTimeout(options.timeout.toInt())
                    }

                    override fun onOpen() {
                        status = ProviderStatus.READY
                    }

                    override fun onTextReceived(message: String?) {
                        try {
                            status = ProviderStatus.STALE
                            fetchAllFlags(getEvaluationContext())
                            status = ProviderStatus.READY
                        } catch (e: Exception) {
                            status = ProviderStatus.ERROR
                        }
                    }

                    override fun onException(e: java.lang.Exception?) {
                        status = ProviderStatus.ERROR
                    }

                    override fun onBinaryReceived(data: ByteArray?) {}
                    override fun onPingReceived(data: ByteArray?) {}
                    override fun onPongReceived(data: ByteArray?) {}
                    override fun onCloseReceived(reason: Int, description: String?) {}
                }
            this.goffWebsocketClient?.connect()
        } catch (e: Exception) {
            this.status = ProviderStatus.ERROR
        }
    }

    override fun onContextSet(oldContext: EvaluationContext?, newContext: EvaluationContext) {
        this.status = ProviderStatus.STALE
        try {
            fetchAllFlags(newContext)
            this.status = ProviderStatus.READY
        } catch (e: Exception) {
            this.status = ProviderStatus.ERROR
        }
    }

    override fun shutdown() {
        this.goffWebsocketClient?.close(0, 1000, "stop provider")
    }

    fun status(): ProviderStatus {
        return this.status
    }

    /**
     * Evaluate is the function call by all types to fetch the flag from the cache
     * and return the ProviderEvaluation object.
     *
     * @param flagKey the key of the flag to fetch
     * @param expectedTypes the types of the flag to fetch
     * @param defaultValue the default value to return if there is an error
     *
     * @return ProviderEvaluation<T> the evaluation of the flag
     */
    private fun <T> evaluate(
        flagKey: String,
        expectedTypes: List<String>,
        defaultValue: T
    ): ProviderEvaluation<T> {
        if (this.status == ProviderStatus.NOT_READY) {
            return ProviderEvaluation<T>(
                value = defaultValue,
                variant = null,
                reason = Reason.ERROR.toString(),
                errorCode = ErrorCode.PROVIDER_NOT_READY
            )
        }

        val flag = this.flags[flagKey] ?: throw FlagNotFoundError(flagKey)

        if (!expectedTypes.contains(flag.value::class.simpleName)) {
            // TODO: throw here when the SDK with this PR https://github.com/open-feature/kotlin-sdk/pull/64/files is released
            // throw TypeMismatchError()
            return ProviderEvaluation<T>(
                value = defaultValue,
                variant = flag.variationType,
                reason = Reason.ERROR.toString(),
                errorCode = ErrorCode.TYPE_MISMATCH
            )
        }

        val errorCode: ErrorCode? = try {
            ErrorCode.valueOf(flag.errorCode)
        } catch (e: IllegalArgumentException) {
            null
        }

        return ProviderEvaluation<T>(
            value = flag.value as T,
            variant = flag.variationType,
            reason = flag.reason,
            errorCode = errorCode
        )
    }

    /** fetchAllFlags is the function called to fetch all flags from the relay proxy
     * and store them in the cache.
     *
     * @param context the context to use to fetch the flags
     */
    private fun fetchAllFlags(context: EvaluationContext?) {
        if (context == null) {
            return
        }
        val goffctx = GoffRequest(context)
        val urlBuilder = parsedEndpoint!!.newBuilder()
            .addEncodedPathSegment("v1")
            .addEncodedPathSegment("allflags")

        if (this.options.apiKey != null && this.options.apiKey.trim().isNotEmpty()) {
            urlBuilder.addQueryParameter("apiKey", this.options.apiKey)
        }

        // call an API endpoint to fetch all flags
        val mediaType = "application/json".toMediaTypeOrNull()
        val requestBody = gson.toJson(goffctx).toRequestBody(mediaType)
        val reqBuilder = okhttp3.Request.Builder()
            .url(urlBuilder.build())
            .post(requestBody)

        httpClient.newCall(reqBuilder.build()).execute().use { response ->
            if (response.code == HTTP_UNAUTHORIZED) {
                throw GeneralError("invalid token used to contact GO Feature Flag relay proxy instance")
            }
            if (response.code >= HTTP_BAD_REQUEST) {
                throw GeneralError("impossible to contact GO Feature Flag relay proxy instance")
            }

            val t = response.body?.string()
            val parsedResp = gson.fromJson(t, GoffResponse::class.java)
            this.flags = ConcurrentHashMap(parsedResp.flags)
        }
    }

    /** buildWebsocketURI is the function called to build the websocket URI
     * to connect to the relay proxy.
     *
     * @param options the options to use to build the URI
     * @return URI the URI to connect to
     */
    private fun buildWebsocketURI(options: GoFeatureFlagOptions): URI {
        // take the endpoint and replace http:// or https:// by ws:// or wss://
        val wsEndpoint = options.endpoint
            .replaceFirst("http", "ws")
            .replaceFirst("https", "wss")
        return URI("$wsEndpoint/ws/v1/flag/change")
    }
}