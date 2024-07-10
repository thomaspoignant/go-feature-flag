package org.gofeatureflag.openfeature.ofrep

import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.FeatureProvider
import dev.openfeature.sdk.Hook
import dev.openfeature.sdk.ProviderEvaluation
import dev.openfeature.sdk.ProviderMetadata
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.events.EventHandler
import dev.openfeature.sdk.events.OpenFeatureEvents
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions
import org.gofeatureflag.openfeature.ofrep.bean.OfrepProviderMetadata
import org.gofeatureflag.openfeature.ofrep.controller.OfrepApi
import org.gofeatureflag.openfeature.ofrep.error.OfrepError

class OfrepProvider(
    private val ofrepOptions: OfrepOptions,
    private val eventHandler: EventHandler = EventHandler(Dispatchers.IO)
) : FeatureProvider {
    private val ofrepApi = OfrepApi(ofrepOptions)
    override val hooks: List<Hook<*>>
        get() = listOf()

    override val metadata: ProviderMetadata
        get() = OfrepProviderMetadata()

    override fun observe(): Flow<OpenFeatureEvents> = eventHandler.observe()

    override fun initialize(initialContext: EvaluationContext?) {
        runBlocking {
            launch {
                try{
                    val postBulkEvaluateFlags =
                        this@OfrepProvider.ofrepApi.postBulkEvaluateFlags(initialContext)
                    eventHandler.publish(OpenFeatureEvents.ProviderReady)
                }catch (e: OfrepError.)


            }
        }
    }


    override fun getBooleanEvaluation(
        key: String,
        defaultValue: Boolean,
        context: EvaluationContext?
    ): ProviderEvaluation<Boolean> {
        throw NotImplementedError()
    }

    override fun getDoubleEvaluation(
        key: String,
        defaultValue: Double,
        context: EvaluationContext?
    ): ProviderEvaluation<Double> {
        throw NotImplementedError()
    }

    override fun getIntegerEvaluation(
        key: String,
        defaultValue: Int,
        context: EvaluationContext?
    ): ProviderEvaluation<Int> {
        throw NotImplementedError()
    }

    override fun getObjectEvaluation(
        key: String,
        defaultValue: Value,
        context: EvaluationContext?
    ): ProviderEvaluation<Value> {
        throw NotImplementedError()
    }

    override fun getProviderStatus(): OpenFeatureEvents {
        return eventHandler.getProviderStatus()
    }

    override fun getStringEvaluation(
        key: String,
        defaultValue: String,
        context: EvaluationContext?
    ): ProviderEvaluation<String> {
        throw NotImplementedError()
    }

    override fun onContextSet(oldContext: EvaluationContext?, newContext: EvaluationContext) {
        throw NotImplementedError()
    }

    override fun shutdown() {
        throw NotImplementedError()
    }

}