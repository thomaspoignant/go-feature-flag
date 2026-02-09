package dev.openfeature.kotlin.contrib.providers.ofrep.bean

import dev.openfeature.kotlin.contrib.providers.ofrep.serialization.EvaluationContextSerializer
import dev.openfeature.kotlin.sdk.EvaluationContext
import dev.openfeature.kotlin.sdk.ImmutableContext
import kotlinx.serialization.Serializable

@Serializable
internal data class OfrepApiRequest(
    @Serializable(with = EvaluationContextSerializer::class)
    val context: EvaluationContext = ImmutableContext(),
)
