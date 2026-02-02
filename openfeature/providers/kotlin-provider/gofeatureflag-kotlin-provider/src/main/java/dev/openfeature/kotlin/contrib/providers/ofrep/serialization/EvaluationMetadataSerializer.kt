package dev.openfeature.kotlin.contrib.providers.ofrep.serialization

import dev.openfeature.kotlin.sdk.EvaluationMetadata
import kotlinx.serialization.KSerializer
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.JsonDecoder
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.intOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

internal class EvaluationMetadataSerializer : KSerializer<EvaluationMetadata> {
    override val descriptor: SerialDescriptor
        get() = TODO("Not yet implemented")

    override fun deserialize(decoder: Decoder): EvaluationMetadata {
        val jsonDecoder = decoder as JsonDecoder

        val jsonObject = jsonDecoder.decodeJsonElement().jsonObject

        // check that inputMap is null or empty
        if (jsonObject.isEmpty()) {
            return EvaluationMetadata.EMPTY
        }

        val metadataBuilder = EvaluationMetadata.builder()
        jsonObject.forEach { entry ->
            val jsonPrimitive = entry.value.jsonPrimitive
            val value =
                jsonPrimitive.run {
                    if (isString) {
                        content
                    } else {
                        booleanOrNull ?: intOrNull ?: doubleOrNull
                            ?: error("Cannot parse value")
                    }
                }
            when (value) {
                is String -> metadataBuilder.putString(entry.key, value)
                is Boolean -> metadataBuilder.putBoolean(entry.key, value)
                is Int -> metadataBuilder.putInt(entry.key, value)
                else -> error("Unsupported type for: $value")
            }
        }

        return metadataBuilder.build()
    }

    override fun serialize(
        encoder: Encoder,
        value: EvaluationMetadata,
    ) {
        TODO("Not yet implemented")
    }
}
