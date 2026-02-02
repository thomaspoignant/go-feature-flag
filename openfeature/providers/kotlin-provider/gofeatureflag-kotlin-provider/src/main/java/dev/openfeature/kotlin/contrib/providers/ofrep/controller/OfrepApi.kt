package dev.openfeature.kotlin.contrib.providers.ofrep.controller

import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepApiRequest
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepApiResponse
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepOptions
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.PostBulkEvaluationResult
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.createHttpClient
import dev.openfeature.kotlin.contrib.providers.ofrep.error.OfrepError
import dev.openfeature.kotlin.sdk.EvaluationContext
import dev.openfeature.kotlin.sdk.exceptions.OpenFeatureError
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.headers
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.Url
import io.ktor.http.appendPathSegments
import io.ktor.http.contentType
import io.ktor.http.parseUrl
import io.ktor.serialization.JsonConvertException

internal class OfrepApi(
    private val options: OfrepOptions,
) {
    private val httpClient: HttpClient = createHttpClient(options)
    private var parsedEndpoint: Url =
        parseUrl(options.endpoint) ?: throw OfrepError.InvalidOptionsError("invalid endpoint configuration: ${options.endpoint}")
    private var etag: String? = null

    /**
     * Call the OFREP API to evaluate in bulk the flags for the given context.
     */
    internal suspend fun postBulkEvaluateFlags(context: EvaluationContext?): PostBulkEvaluationResult {
        val nonNullContext =
            context ?: throw OpenFeatureError.InvalidContextError("EvaluationContext is null")
        validateContext(nonNullContext)

        val response =
            httpClient.post(parsedEndpoint) {
                url {
                    appendPathSegments("ofrep", "v1", "evaluate", "flags")
                }
                contentType(ContentType.Application.Json)
                headers {
                    options.headers.forEach {
                        append(it.key, it.value)
                    }
                    etag?.let {
                        append(HttpHeaders.IfNoneMatch, it)
                    }
                }
                setBody(OfrepApiRequest(nonNullContext))
            }

        when (response.status.value) {
            401 -> throw OfrepError.ApiUnauthorizedError(response)
            403 -> throw OfrepError.ForbiddenError(response)
            429 -> throw OfrepError.ApiTooManyRequestsError(response)
            304 -> return PostBulkEvaluationResult(null, response)
            in 200..299, 400 -> {
                try {
                    response.headers[HttpHeaders.ETag].let { this.etag = it }
                    val ofrepResp: OfrepApiResponse? = response.body()
                    return PostBulkEvaluationResult(ofrepResp, response)
                } catch (e: JsonConvertException) {
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

    private fun validateContext(context: EvaluationContext) {
        if (context.getTargetingKey().isEmpty()) {
            throw OpenFeatureError.TargetingKeyMissingError()
        }
    }
}
