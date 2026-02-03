package dev.openfeature.kotlin.contrib.providers.ofrep.serialization

import dev.openfeature.kotlin.sdk.EvaluationContext
import dev.openfeature.kotlin.sdk.Value
import kotlinx.serialization.KSerializer
import kotlinx.serialization.builtins.MapSerializer
import kotlinx.serialization.builtins.serializer
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.descriptors.buildClassSerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder

internal class EvaluationContextSerializer : KSerializer<EvaluationContext> {
    private val delegateSerializer = MapSerializer(String.serializer(), ValueSerializer)
    override val descriptor: SerialDescriptor =
        buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.EvaluationContext")

    override fun serialize(
        encoder: Encoder,
        value: EvaluationContext,
    ) = delegateSerializer.serialize(encoder, value.asMap() + mapOf("targetingKey" to Value.String(value.getTargetingKey())))

    override fun deserialize(decoder: Decoder): EvaluationContext = error("Not implemented")
}
