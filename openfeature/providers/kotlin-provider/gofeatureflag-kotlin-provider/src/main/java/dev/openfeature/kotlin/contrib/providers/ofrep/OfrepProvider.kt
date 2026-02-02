package dev.openfeature.kotlin.contrib.providers.ofrep

import dev.openfeature.kotlin.contrib.providers.ofrep.bean.FlagDto
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepOptions
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepProviderMetadata
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.toProviderEvaluation
import dev.openfeature.kotlin.contrib.providers.ofrep.controller.OfrepApi
import dev.openfeature.kotlin.contrib.providers.ofrep.enum.BulkEvaluationStatus
import dev.openfeature.kotlin.contrib.providers.ofrep.error.OfrepError
import dev.openfeature.kotlin.sdk.EvaluationContext
import dev.openfeature.kotlin.sdk.FeatureProvider
import dev.openfeature.kotlin.sdk.Hook
import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.ProviderEvaluation
import dev.openfeature.kotlin.sdk.ProviderMetadata
import dev.openfeature.kotlin.sdk.Value
import dev.openfeature.kotlin.sdk.events.OpenFeatureProviderEvents
import dev.openfeature.kotlin.sdk.exceptions.ErrorCode
import dev.openfeature.kotlin.sdk.exceptions.OpenFeatureError
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlin.concurrent.Volatile
import kotlin.time.Clock
import kotlin.time.Duration.Companion.seconds
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

@OptIn(ExperimentalTime::class)
class OfrepProvider(
    private val ofrepOptions: OfrepOptions,
) : FeatureProvider {
    private val ofrepApi = OfrepApi(ofrepOptions)
    override val hooks: List<Hook<*>>
        get() = listOf()

    override val metadata: ProviderMetadata
        get() = OfrepProviderMetadata()

    private var evaluationContext: EvaluationContext? = null

    @Volatile
    private var inMemoryCache: Map<String, FlagDto> = emptyMap()
    private var retryAfter: Instant? = null
    private val pollingScope: CoroutineScope = CoroutineScope(ofrepOptions.pollingDispatcher)
    private var pollingJob: Job? = null

    private val statusFlow = MutableSharedFlow<OpenFeatureProviderEvents>(replay = 1)

    override fun observe(): Flow<OpenFeatureProviderEvents> = statusFlow

    override suspend fun initialize(initialContext: EvaluationContext?) {
        this.evaluationContext = initialContext
        try {
            val bulkEvaluationStatus = evaluateFlags(initialContext ?: ImmutableContext())
            if (bulkEvaluationStatus == BulkEvaluationStatus.RATE_LIMITED) {
                statusFlow.emit(
                    OpenFeatureProviderEvents.ProviderError(
                        OpenFeatureError.GeneralError("Rate limited"),
                    ),
                )
            } else {
                statusFlow.emit(OpenFeatureProviderEvents.ProviderReady)
            }
        } catch (e: OpenFeatureError) {
            statusFlow.emit(OpenFeatureProviderEvents.ProviderError(e))
        } catch (e: Exception) {
            statusFlow.emit(OpenFeatureProviderEvents.ProviderError(OpenFeatureError.GeneralError(e.message ?: "Unknown error")))
        }
        startPolling()
    }

    /**
     * Start polling for flag updates
     */
    private fun startPolling() {
        pollingJob =
            pollingScope.launch {
                while (isActive) {
                    try {
                        delay(ofrepOptions.pollingInterval)
                        val resp =
                            evaluateFlags(evaluationContext!!)

                        when (resp) {
                            BulkEvaluationStatus.RATE_LIMITED, BulkEvaluationStatus.SUCCESS_NO_CHANGE -> {
                                // Nothing to do !
                                //
                                // if rate limited: the provider should already be in stale status and
                                //    we don't need to emit an event or call again the API
                                //
                                // if no change: the provider should already be in ready status and
                                //    we don't need to emit an event if nothing has changed
                            }

                            BulkEvaluationStatus.SUCCESS_UPDATED -> {
                                // TODO: we should migrate to configuration change event when it's available
                                // in the kotlin SDK
                                statusFlow.emit(OpenFeatureProviderEvents.ProviderReady)
                            }
                        }
                    } catch (e: CancellationException) {
                        // expected to happen when the job is cancelled, no need to report it via the
                        // statusFlow
                    } catch (e: OfrepError.ApiTooManyRequestsError) {
                        // in that case the provider is just stale because we were not able to
                        statusFlow.emit(OpenFeatureProviderEvents.ProviderStale)
                    } catch (e: Throwable) {
                        statusFlow.emit(
                            OpenFeatureProviderEvents.ProviderError(
                                OpenFeatureError.GeneralError(
                                    e.message ?: "",
                                ),
                            ),
                        )
                    }
                }
            }
    }

    override fun getBooleanEvaluation(
        key: String,
        defaultValue: Boolean,
        context: EvaluationContext?,
    ): ProviderEvaluation<Boolean> = genericEvaluation(key, defaultValue)

    override fun getDoubleEvaluation(
        key: String,
        defaultValue: Double,
        context: EvaluationContext?,
    ): ProviderEvaluation<Double> = genericEvaluation(key, defaultValue)

    override fun getIntegerEvaluation(
        key: String,
        defaultValue: Int,
        context: EvaluationContext?,
    ): ProviderEvaluation<Int> = genericEvaluation(key, defaultValue)

    override fun getObjectEvaluation(
        key: String,
        defaultValue: Value,
        context: EvaluationContext?,
    ): ProviderEvaluation<Value> = genericEvaluation(key, defaultValue)

    override fun getStringEvaluation(
        key: String,
        defaultValue: String,
        context: EvaluationContext?,
    ): ProviderEvaluation<String> = genericEvaluation(key, defaultValue)

    override suspend fun onContextSet(
        oldContext: EvaluationContext?,
        newContext: EvaluationContext,
    ) {
        this.statusFlow.emit(OpenFeatureProviderEvents.ProviderStale)
        this.evaluationContext = newContext

        try {
            val postBulkEvaluateFlags = evaluateFlags(newContext)
            // we don't emit event if the evaluation is rate limited because
            // the provider is still stale
            if (postBulkEvaluateFlags != BulkEvaluationStatus.RATE_LIMITED) {
                statusFlow.emit(OpenFeatureProviderEvents.ProviderReady)
            }
        } catch (e: Throwable) {
            statusFlow.emit(OpenFeatureProviderEvents.ProviderError(OpenFeatureError.GeneralError(e.message ?: "")))
        }
    }

    override fun shutdown() {
        pollingJob?.cancel()
    }

    private inline fun <reified T> genericEvaluation(
        key: String,
        defaultValue: T,
    ): ProviderEvaluation<T> {
        val flag = this.inMemoryCache[key] ?: throw OpenFeatureError.FlagNotFoundError(key)

        if (flag.isError()) {
            when (flag.errorCode) {
                ErrorCode.FLAG_NOT_FOUND -> throw OpenFeatureError.FlagNotFoundError(key)
                ErrorCode.INVALID_CONTEXT -> throw OpenFeatureError.InvalidContextError()
                ErrorCode.PARSE_ERROR -> throw OpenFeatureError.ParseError(
                    flag.errorDetails ?: "parse error",
                )

                ErrorCode.PROVIDER_NOT_READY -> throw OpenFeatureError.ProviderNotReadyError()
                ErrorCode.TARGETING_KEY_MISSING -> throw OpenFeatureError.TargetingKeyMissingError()
                else -> throw OpenFeatureError.GeneralError(flag.errorDetails ?: "general error")
            }
        }
        return flag.toProviderEvaluation(defaultValue)
    }

    /**
     * Evaluate the flags for the given context.
     * It will store the flags in the in-memory cache, if any error occurs it will throw an exception.
     */
    private suspend fun evaluateFlags(context: EvaluationContext): BulkEvaluationStatus {
        if (this.retryAfter != null && this.retryAfter!! > Clock.System.now()) {
            return BulkEvaluationStatus.RATE_LIMITED
        }

        try {
            val postBulkEvaluateFlags =
                this@OfrepProvider.ofrepApi.postBulkEvaluateFlags(context)
            val ofrepEvalResp = postBulkEvaluateFlags.apiResponse
            val httpResp = postBulkEvaluateFlags.httpResponse

            if (httpResp.status.value == 304) {
                return BulkEvaluationStatus.SUCCESS_NO_CHANGE
            }

            if (postBulkEvaluateFlags.isError()) {
                when (ofrepEvalResp?.errorCode) {
                    ErrorCode.PROVIDER_NOT_READY -> throw OpenFeatureError.ProviderNotReadyError()
                    ErrorCode.PARSE_ERROR -> throw OpenFeatureError.ParseError(
                        ofrepEvalResp.errorDetails ?: "",
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
            this.retryAfter = calculateRetryDate(e.response?.headers?.get("Retry-After") ?: "")
            return BulkEvaluationStatus.RATE_LIMITED
        }
    }

    private fun calculateRetryDate(retryAfter: String): Instant? {
        if (retryAfter.isEmpty()) {
            return null
        }

        return try {
            // If retryAfter is a number, it represents seconds to wait.
            val delayInSeconds = retryAfter.toInt().seconds
            Clock.System.now() + delayInSeconds
        } catch (e: NumberFormatException) {
            // If retryAfter is not a number, it's an HTTP-date.
            Instant.parse(retryAfter)
        }
    }
}
