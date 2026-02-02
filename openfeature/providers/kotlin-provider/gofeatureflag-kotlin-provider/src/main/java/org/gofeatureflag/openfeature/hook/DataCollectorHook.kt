package org.gofeatureflag.openfeature.hook

import dev.openfeature.kotlin.sdk.FlagEvaluationDetails
import dev.openfeature.kotlin.sdk.Hook
import dev.openfeature.kotlin.sdk.HookContext
import org.gofeatureflag.openfeature.controller.DataCollectorManager
import java.util.Date

class DataCollectorHook<T>(private val collectorManager: DataCollectorManager) : Hook<T> {
    override fun after(
        ctx: HookContext<T>,
        details: FlagEvaluationDetails<T>,
        hints: Map<String, Any>
    ) {
        val event = Event(
            contextKind = "user",
            creationDate = Date().time / 1000L,
            key = ctx.flagKey,
            kind = "feature",
            userKey = ctx.ctx?.getTargetingKey(),
            value = details.value,
            default = false,
            variation = details.variant,
            source = "PROVIDER_CACHE"
        )
        collectorManager.addEvent(event)
    }

    override fun error(ctx: HookContext<T>, error: Exception, hints: Map<String, Any>) {
        val event = Event(
            contextKind = "user",
            creationDate = Date().time / 1000L,
            key = ctx.flagKey,
            kind = "feature",
            userKey = ctx.ctx?.getTargetingKey(),
            value = ctx.defaultValue,
            default = true,
            variation = "SdkDefault",
            source = "PROVIDER_CACHE"
        )
        collectorManager.addEvent(event)
    }
}
