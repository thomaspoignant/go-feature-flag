package org.gofeatureflag.openfeature

import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.FeatureProvider
import dev.openfeature.sdk.Hook
import dev.openfeature.sdk.ProviderEvaluation
import dev.openfeature.sdk.ProviderMetadata
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.events.OpenFeatureEvents
import kotlinx.coroutines.flow.Flow
import okhttp3.Headers
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.gofeatureflag.openfeature.controller.DataCollectorManager
import org.gofeatureflag.openfeature.controller.GoFeatureFlagApi
import org.gofeatureflag.openfeature.hook.DataCollectorHook
import org.gofeatureflag.openfeature.ofrep.OfrepProvider
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions

class GoFeatureFlagProvider(private val options: GoFeatureFlagOptions) : FeatureProvider {
    private val ofrepProvider: OfrepProvider
    private var dataCollectorManager: DataCollectorManager? = null

    override var hooks: List<Hook<*>>

    override val metadata: ProviderMetadata
        get() = GoFeatureFlagMetadata()

    init {
        val authorizationHeader = options.apiKey?.let { apiKey ->
            val headers = Headers.Builder()
            headers.add("Authorization", "Bearer $apiKey")
            headers.build()
        }
        val ofrepOptions = OfrepOptions(
            endpoint = options.endpoint,
            timeout = options.timeout,
            maxIdleConnections = options.maxIdleConnections,
            keepAliveDuration = options.keepAliveDuration,
            headers = authorizationHeader,
            pollingIntervalInMillis = options.pollingIntervalInMillis,
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

    override fun getProviderStatus(): OpenFeatureEvents {
        return this.ofrepProvider.getProviderStatus()
    }

    override fun initialize(initialContext: EvaluationContext?) {
        if (this.options.flushIntervalMs > 0) {
            this.dataCollectorManager?.start()
        }
        return this.ofrepProvider.initialize(initialContext)
    }

    override fun observe(): Flow<OpenFeatureEvents> {
        return this.ofrepProvider.observe()
    }

    override fun onContextSet(oldContext: EvaluationContext?, newContext: EvaluationContext) {
        return this.ofrepProvider.onContextSet(oldContext, newContext)
    }

    override fun shutdown() {
        this.ofrepProvider.shutdown()
        this.dataCollectorManager?.stop()
    }
}