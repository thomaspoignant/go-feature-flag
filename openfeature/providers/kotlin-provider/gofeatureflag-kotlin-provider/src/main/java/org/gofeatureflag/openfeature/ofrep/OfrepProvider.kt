package org.gofeatureflag.openfeature.ofrep

import FlagDto
import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.FeatureProvider
import dev.openfeature.sdk.Hook
import dev.openfeature.sdk.ImmutableContext
import dev.openfeature.sdk.ProviderEvaluation
import dev.openfeature.sdk.ProviderMetadata
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.events.EventHandler
import dev.openfeature.sdk.events.OpenFeatureEvents
import dev.openfeature.sdk.exceptions.ErrorCode
import dev.openfeature.sdk.exceptions.OpenFeatureError
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions
import org.gofeatureflag.openfeature.ofrep.bean.OfrepProviderMetadata
import org.gofeatureflag.openfeature.ofrep.controller.OfrepApi
import org.gofeatureflag.openfeature.ofrep.enum.BulkEvaluationStatus
import org.gofeatureflag.openfeature.ofrep.error.OfrepError
import java.text.SimpleDateFormat
import java.util.Calendar
import java.util.Date
import java.util.Locale
import java.util.TimeZone
import kotlin.reflect.KClass

class OfrepProvider(
    private val ofrepOptions: OfrepOptions,
    private val eventHandler: EventHandler = EventHandler(Dispatchers.IO)
) : FeatureProvider {
    private val ofrepApi = OfrepApi(ofrepOptions)
    override val hooks: List<Hook<*>>
        get() = listOf()

    override val metadata: ProviderMetadata
        get() = OfrepProviderMetadata()

    private var evaluationContext: EvaluationContext? = null
    private var inMemoryCache: Map<String, FlagDto> = emptyMap()
    private var retryAfter: Date? = null


    override fun observe(): Flow<OpenFeatureEvents> = eventHandler.observe()

    override fun initialize(initialContext: EvaluationContext?) {
        this.evaluationContext = initialContext
        runBlocking {
            launch {
                try {
                    val postBulkEvaluateFlags =
                        evaluateFlags(initialContext ?: ImmutableContext())
                    eventHandler.publish(OpenFeatureEvents.ProviderReady)
                } catch (e: Exception) {
                    print(e)
                    eventHandler.publish(OpenFeatureEvents.ProviderError(e))
                }
            }
        }
    }


    override fun getBooleanEvaluation(
        key: String,
        defaultValue: Boolean,
        context: EvaluationContext?
    ): ProviderEvaluation<Boolean> {
        return genericEvaluation<Boolean>(key, Boolean::class)
    }

    override fun getDoubleEvaluation(
        key: String,
        defaultValue: Double,
        context: EvaluationContext?
    ): ProviderEvaluation<Double> {
        return genericEvaluation<Double>(key, Double::class)
    }

    override fun getIntegerEvaluation(
        key: String,
        defaultValue: Int,
        context: EvaluationContext?
    ): ProviderEvaluation<Int> {
        val res = genericEvaluation<Int>(key, Int::class)
        return ProviderEvaluation<Int>(
            value = res.value,
            reason = res.reason,
            variant = res.variant,
            errorCode = res.errorCode,
            errorMessage = res.errorMessage
        )
    }

    override fun getObjectEvaluation(
        key: String,
        defaultValue: Value,
        context: EvaluationContext?
    ): ProviderEvaluation<Value> {
        return genericEvaluation<Value>(key, Object::class)
    }

    override fun getStringEvaluation(
        key: String,
        defaultValue: String,
        context: EvaluationContext?
    ): ProviderEvaluation<String> {
        return genericEvaluation<String>(key, String::class)
    }

    override fun getProviderStatus(): OpenFeatureEvents {
        return eventHandler.getProviderStatus()
    }

    override fun onContextSet(oldContext: EvaluationContext?, newContext: EvaluationContext) {
        this.eventHandler.publish(OpenFeatureEvents.ProviderStale)
        this.evaluationContext = newContext

        runBlocking {
            launch {
                try {
                    val postBulkEvaluateFlags = evaluateFlags(newContext)
                    if (postBulkEvaluateFlags == BulkEvaluationStatus.RATE_LIMITED) {
                        // we don't emit event if the evaluation is rate limited because
                        // the provider is still stale
                        return@launch
                    }
                    eventHandler.publish(OpenFeatureEvents.ProviderReady)
                } catch (e: Throwable) {
                    eventHandler.publish(OpenFeatureEvents.ProviderError(e))
                }
            }
        }


    }

    override fun shutdown() {
        // TODO: implement?
    }


    private fun <T : Any> genericEvaluation(
        key: String,
        expectedType: KClass<*>
    ): ProviderEvaluation<T> {
        val flag = this.inMemoryCache[key] ?: throw OpenFeatureError.FlagNotFoundError(key)

        if (flag.isError()) {
            when (flag.errorCode) {
                ErrorCode.FLAG_NOT_FOUND -> throw OpenFeatureError.FlagNotFoundError(key)
                ErrorCode.INVALID_CONTEXT -> throw OpenFeatureError.InvalidContextError()
                ErrorCode.PARSE_ERROR -> throw OpenFeatureError.ParseError(
                    flag.errorDetails ?: "parse error"
                )

                ErrorCode.PROVIDER_NOT_READY -> throw OpenFeatureError.ProviderNotReadyError()
                ErrorCode.TARGETING_KEY_MISSING -> throw OpenFeatureError.TargetingKeyMissingError()
                else -> throw OpenFeatureError.GeneralError(flag.errorDetails ?: "general error")
            }
        }
        return flag.toProviderEvaluation(expectedType)
    }


    /**
     * Evaluate the flags for the given context.
     * It will store the flags in the in-memory cache, if any error occurs it will throw an exception.
     */
    private suspend fun evaluateFlags(context: EvaluationContext): BulkEvaluationStatus {
        if (this.retryAfter != null && this.retryAfter!! > Date()) {
            return BulkEvaluationStatus.RATE_LIMITED
        }

        try {
            val postBulkEvaluateFlags =
                this@OfrepProvider.ofrepApi.postBulkEvaluateFlags(context)
            val ofrepEvalResp = postBulkEvaluateFlags.apiResponse
            val httpResp = postBulkEvaluateFlags.httpResponse

            if (httpResp.code == 304) {
                return BulkEvaluationStatus.SUCCESS_NO_CHANGE
            }

            if (postBulkEvaluateFlags.isError()) {
                when (ofrepEvalResp?.errorCode) {
                    ErrorCode.PROVIDER_NOT_READY -> throw OpenFeatureError.ProviderNotReadyError()
                    ErrorCode.PARSE_ERROR -> throw OpenFeatureError.ParseError(
                        ofrepEvalResp.errorDetails ?: ""
                    )

                    ErrorCode.TARGETING_KEY_MISSING -> throw OpenFeatureError.TargetingKeyMissingError()
                    ErrorCode.INVALID_CONTEXT -> throw OpenFeatureError.InvalidContextError()
                    else -> throw OpenFeatureError.GeneralError(ofrepEvalResp?.errorDetails ?: "")
                }
            }
            val inMemoryCacheNew = ofrepEvalResp?.flags?.associateBy { it.key } ?: emptyMap()
            this.inMemoryCache = inMemoryCacheNew
            return BulkEvaluationStatus.SUCCESS_UPDATED
        } catch (e: OfrepError.ApiTooManyRequestsError) {
            this.retryAfter = calculateRetryDate(e.response.headers["Retry-After"] ?: "")
            return BulkEvaluationStatus.RATE_LIMITED
        } catch (e: Throwable) {
            throw e
        }
    }

    private fun calculateRetryDate(retryAfter: String): Date? {
        val retryDate: Calendar = Calendar.getInstance()
        try {
            // If retryAfter is a number, it represents seconds to wait.
            val delayInSeconds = retryAfter.toInt()
            retryDate.add(Calendar.SECOND, delayInSeconds)
        } catch (e: NumberFormatException) {
            // If retryAfter is not a number, it's an HTTP-date.
            val dateFormat = SimpleDateFormat("EEE, dd MMM yyyy HH:mm:ss z", Locale.US)
            dateFormat.timeZone = TimeZone.getTimeZone("GMT")
            retryDate.time = dateFormat.parse(retryAfter) ?: return null
        }
        return retryDate.time
    }
}