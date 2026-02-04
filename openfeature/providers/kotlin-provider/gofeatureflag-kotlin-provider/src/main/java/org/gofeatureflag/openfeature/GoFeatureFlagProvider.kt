package org.gofeatureflag.openfeature

import dev.openfeature.kotlin.contrib.providers.ofrep.OfrepProvider
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepOptions
import dev.openfeature.kotlin.sdk.EvaluationContext
import dev.openfeature.kotlin.sdk.FeatureProvider
import dev.openfeature.kotlin.sdk.Hook
import dev.openfeature.kotlin.sdk.ProviderEvaluation
import dev.openfeature.kotlin.sdk.ProviderMetadata
import dev.openfeature.kotlin.sdk.TrackingEventDetails
import dev.openfeature.kotlin.sdk.Value
import dev.openfeature.kotlin.sdk.events.OpenFeatureProviderEvents
import kotlinx.coroutines.flow.Flow
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.gofeatureflag.openfeature.bean.TrackingEvent
import org.gofeatureflag.openfeature.controller.DataCollectorManager
import org.gofeatureflag.openfeature.controller.GoFeatureFlagApi
import org.gofeatureflag.openfeature.hook.DataCollectorHook
import org.gofeatureflag.openfeature.utils.EvaluationContextUtil

class GoFeatureFlagProvider(private val options: GoFeatureFlagOptions) : FeatureProvider {
    private val ofrepProvider: OfrepProvider
    private var dataCollectorManager: DataCollectorManager? = null

    override var hooks: List<Hook<*>>

    override val metadata: ProviderMetadata
        get() = GoFeatureFlagMetadata()

    init {
        val headers = buildMap {
            options.apiKey?.let { put("X-API-Key", it) }
            put("Content-Type", "application/json")
        }
        val ofrepOptions = OfrepOptions(
            endpoint = options.endpoint,
            timeout = options.timeout,
            maxIdleConnections = options.maxIdleConnections,
            keepAliveDuration = options.keepAlive,
            headers = headers,
            pollingInterval = options.pollingInterval,
        )
        this.ofrepProvider = OfrepProvider(ofrepOptions)

        val dataCollectorManager = DataCollectorManager(
            goffApi = GoFeatureFlagApi(options),
            flushIntervalMs = options.flushIntervalMs
        )
        hooks = listOf(DataCollectorHook<Any>(dataCollectorManager))
        this.dataCollectorManager = dataCollectorManager
    }


    override fun getBooleanEvaluation(
        key: String,
        defaultValue: Boolean,
        context: EvaluationContext?
    ): ProviderEvaluation<Boolean> {
        return this.ofrepProvider.getBooleanEvaluation(key, defaultValue, context)
    }

    override fun getDoubleEvaluation(
        key: String,
        defaultValue: Double,
        context: EvaluationContext?
    ): ProviderEvaluation<Double> {
        return this.ofrepProvider.getDoubleEvaluation(key, defaultValue, context)
    }

    override fun getIntegerEvaluation(
        key: String,
        defaultValue: Int,
        context: EvaluationContext?
    ): ProviderEvaluation<Int> {
        return this.ofrepProvider.getIntegerEvaluation(key, defaultValue, context)
    }

    override fun getObjectEvaluation(
        key: String,
        defaultValue: Value,
        context: EvaluationContext?
    ): ProviderEvaluation<Value> {
        return this.ofrepProvider.getObjectEvaluation(key, defaultValue, context)
    }

    override fun getStringEvaluation(
        key: String,
        defaultValue: String,
        context: EvaluationContext?
    ): ProviderEvaluation<String> {
        return this.ofrepProvider.getStringEvaluation(key, defaultValue, context)
    }

    override suspend fun initialize(initialContext: EvaluationContext?) {
        if (this.options.flushIntervalMs > 0) {
            this.dataCollectorManager?.start()
        }
        return this.ofrepProvider.initialize(initialContext)
    }

    override fun observe(): Flow<OpenFeatureProviderEvents> {
        return this.ofrepProvider.observe()
    }

    override suspend fun onContextSet(oldContext: EvaluationContext?, newContext: EvaluationContext) {
        return this.ofrepProvider.onContextSet(oldContext, newContext)
    }

    override fun shutdown() {
        this.ofrepProvider.shutdown()
        this.dataCollectorManager?.stop()
    }

    /**
     * Feature provider implementations can opt in for to support Tracking by implementing this method.
     *
     * Performs tracking of a particular action or application state.
     *
     * @param trackingEventName Event name to track
     * @param context   Evaluation context used in flag evaluation (Optional)
     * @param details   Data pertinent to a particular tracking event (Optional)
     */
    override fun track(trackingEventName: String, context: EvaluationContext?, details: TrackingEventDetails?) {
        val trackingEventDetails = details?.asObjectMap()?.toMutableMap()
        trackingEventDetails?.put("value", details.`value`)

        val trackingEvent = TrackingEvent(
            kind = "tracking",
            key = trackingEventName,
            evaluationContext = context?.asObjectMap(),
            userKey = context?.getTargetingKey() ?: "undefined-targetingKey",
            contextKind = if (EvaluationContextUtil.isAnonymousUser(context)) "anonymousUser" else "user",
            trackingEventDetails = trackingEventDetails,
            creationDate = System.currentTimeMillis() / 1000L
        )
        this.dataCollectorManager?.addEvent(trackingEvent)
    }
}