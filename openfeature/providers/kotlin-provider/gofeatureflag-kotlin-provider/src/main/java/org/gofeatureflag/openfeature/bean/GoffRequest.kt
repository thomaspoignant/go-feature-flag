package org.gofeatureflag.openfeature.bean

import dev.openfeature.sdk.EvaluationContext

data class GoffRequest(@Transient val ctx: EvaluationContext) {
    private val evaluationContext: GoffEvaluationContext =
        GoffEvaluationContext(ctx.getTargetingKey(), ctx.asObjectMap())
}