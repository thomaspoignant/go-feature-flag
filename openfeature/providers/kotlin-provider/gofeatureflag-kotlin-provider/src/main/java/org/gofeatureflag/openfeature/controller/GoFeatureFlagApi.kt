package org.gofeatureflag.openfeature.controller

import com.google.gson.GsonBuilder
import com.google.gson.ToNumberPolicy
import okhttp3.ConnectionPool
import okhttp3.Headers
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.OkHttpClient
import okhttp3.RequestBody.Companion.toRequestBody
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.gofeatureflag.openfeature.error.GoFeatureFlagError
import org.gofeatureflag.openfeature.hook.Event
import org.gofeatureflag.openfeature.hook.Events
import java.util.concurrent.TimeUnit

class GoFeatureFlagApi(private val options: GoFeatureFlagOptions) {
    companion object {
        private val gson =
            GsonBuilder().setObjectToNumberStrategy(ToNumberPolicy.LONG_OR_DOUBLE).create()
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
    private var parsedEndpoint: HttpUrl =
        options.endpoint.toHttpUrlOrNull()
            ?: throw GoFeatureFlagError.InvalidOptionsError("invalid endpoint configuration: ${options.endpoint}")

    /**
     * Call the GO Feature Flag API to collect the data
     */
    suspend fun postEventsToDataCollector(events: List<Event>) {
        val urlBuilder = parsedEndpoint.newBuilder()
            .addEncodedPathSegment("v1")
            .addEncodedPathSegment("data")
            .addEncodedPathSegment("collector")

        if (events.isEmpty()) {
            return // nothing to send
        }

        val mediaType = "application/json".toMediaTypeOrNull()
        val requestBody = gson.toJson(Events(events)).toRequestBody(mediaType)
        val reqBuilder = okhttp3.Request.Builder()
            .url(urlBuilder.build())
            .post(requestBody)

        val authorizationHeader = options.apiKey?.let { apiKey ->
            val headers = Headers.Builder()
            headers.add("Authorization", "Bearer $apiKey")
            headers.build()
        }
        authorizationHeader?.let { reqBuilder.headers(it) }
        httpClient.newCall(reqBuilder.build()).execute().use { response ->
            when (response.code) {
                200 -> {
                    // SUCCESS - nothing to do here, collection in success
                }

                401, 403 -> throw GoFeatureFlagError.ApiUnauthorizedError(response)
                400 -> throw GoFeatureFlagError.InvalidRequest(response)
                else -> {
                    throw GoFeatureFlagError.UnexpectedResponseError(response)
                }
            }
        }
    }
}