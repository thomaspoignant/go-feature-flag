package org.gofeatureflag.openfeature.ofrep.controller

import OfrepApiRequest
import OfrepApiResponse
import com.google.gson.GsonBuilder
import com.google.gson.JsonSyntaxException
import com.google.gson.ToNumberPolicy
import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.exceptions.OpenFeatureError
import okhttp3.ConnectionPool
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.OkHttpClient
import okhttp3.RequestBody.Companion.toRequestBody
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions
import org.gofeatureflag.openfeature.ofrep.bean.PostBulkEvaluationResult
import org.gofeatureflag.openfeature.ofrep.error.OfrepError
import java.util.concurrent.TimeUnit

class OfrepApi(private val options: OfrepOptions) {
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
            ?: throw OfrepError.InvalidOptionsError("invalid endpoint configuration: ${options.endpoint}")
    private var etag: String? = null

    /**
     * Call the OFREP API to evaluate in bulk the flags for the given context.
     */
    suspend fun postBulkEvaluateFlags(context: EvaluationContext?): PostBulkEvaluationResult {
        val nonNullContext =
            context ?: throw OpenFeatureError.InvalidContextError("EvaluationContext is null")
        validateContext(nonNullContext)

        val urlBuilder = parsedEndpoint.newBuilder()
            .addEncodedPathSegment("ofrep")
            .addEncodedPathSegment("v1")
            .addEncodedPathSegment("evaluate")
            .addEncodedPathSegment("flags")

        val mediaType = "application/json".toMediaTypeOrNull()
        val requestBody = gson.toJson(OfrepApiRequest(nonNullContext)).toRequestBody(mediaType)
        val reqBuilder = okhttp3.Request.Builder()
            .url(urlBuilder.build())
            .post(requestBody)

        // add all the headers
        options.headers?.let { reqBuilder.headers(it) }
        etag?.let { reqBuilder.addHeader("If-None-Match", it) }
        httpClient.newCall(reqBuilder.build()).execute().use { response ->
            when (response.code) {
                401 -> throw OfrepError.ApiUnauthorizedError(response)
                403 -> throw OfrepError.ForbiddenError(response)
                429 -> throw OfrepError.ApiTooManyRequestsError(response)
                304 -> return PostBulkEvaluationResult(null, response)
                in 200..299, 400 -> {
                    try {
                        response.headers["ETag"].let { this.etag = it }
                        val ofrepResp =
                            gson.fromJson(response.body?.string(), OfrepApiResponse::class.java)
                        return PostBulkEvaluationResult(ofrepResp, response)
                    } catch (e: JsonSyntaxException) {
                        throw OfrepError.UnmarshallError(e)
                    } catch (e: Exception) {
                        println(e)
                        throw OfrepError.UnexpectedResponseError(response)
                    }
                }

                else -> {
                    throw OfrepError.UnexpectedResponseError(response)
                }
            }
        }
    }

    private fun validateContext(context: EvaluationContext) {
        if (context.getTargetingKey().isEmpty()) {
            throw OpenFeatureError.TargetingKeyMissingError()
        }
    }
}