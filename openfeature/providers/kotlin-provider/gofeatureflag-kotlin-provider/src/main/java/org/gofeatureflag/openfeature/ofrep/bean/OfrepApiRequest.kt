import dev.openfeature.sdk.EvaluationContext

data class OfrepApiRequest(@Transient val ctx: EvaluationContext) {
    private val context: Map<String, Any?> = ctx.asObjectMap().plus("targetingKey" to ctx.getTargetingKey())
}