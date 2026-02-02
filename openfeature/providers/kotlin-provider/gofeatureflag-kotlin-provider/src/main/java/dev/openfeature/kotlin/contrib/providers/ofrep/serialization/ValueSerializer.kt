
package dev.openfeature.kotlin.contrib.providers.ofrep.serialization

import dev.openfeature.kotlin.sdk.Value
import kotlinx.serialization.DeserializationStrategy
import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.InternalSerializationApi
import kotlinx.serialization.KSerializer
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.builtins.MapSerializer
import kotlinx.serialization.builtins.serializer
import kotlinx.serialization.descriptors.PolymorphicKind
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.descriptors.buildClassSerialDescriptor
import kotlinx.serialization.descriptors.buildSerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonDecoder
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.long
import kotlinx.serialization.json.longOrNull
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

internal object ValueSerializer : KSerializer<Value> {
    @OptIn(InternalSerializationApi::class, ExperimentalSerializationApi::class)
    override val descriptor: SerialDescriptor =
        buildSerialDescriptor("dev.openfeature.kotlin.sdk.Value", PolymorphicKind.SEALED)

    // Serializers for the concrete types
    private object BooleanValueSerializer : KSerializer<Value.Boolean> {
        override val descriptor: SerialDescriptor = buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.Value.Boolean")

        override fun serialize(
            encoder: Encoder,
            value: Value.Boolean,
        ) = encoder.encodeBoolean(value.boolean)

        override fun deserialize(decoder: Decoder): Value.Boolean = Value.Boolean(decoder.decodeBoolean())
    }

    private object DoubleValueSerializer : KSerializer<Value.Double> {
        override val descriptor: SerialDescriptor = buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.Value.Double")

        override fun serialize(
            encoder: Encoder,
            value: Value.Double,
        ) = encoder.encodeDouble(value.double)

        override fun deserialize(decoder: Decoder): Value.Double = Value.Double(decoder.decodeDouble())
    }

    private object IntValueSerializer : KSerializer<Value.Integer> {
        override val descriptor: SerialDescriptor = buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.Value.Integer")

        override fun serialize(
            encoder: Encoder,
            value: Value.Integer,
        ) = encoder.encodeInt(value.integer)

        override fun deserialize(decoder: Decoder): Value.Integer = Value.Integer(decoder.decodeInt())
    }

    @OptIn(ExperimentalTime::class)
    private object InstantValueSerializer : KSerializer<Value.Instant> {
        override val descriptor: SerialDescriptor = buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.Value.Instant")

        override fun serialize(
            encoder: Encoder,
            value: Value.Instant,
        ) = encoder.encodeString(value.instant.toString())

        override fun deserialize(decoder: Decoder): Value.Instant = Value.Instant(Instant.parse(decoder.decodeString()))
    }

    private object StringValueSerializer : KSerializer<Value.String> {
        override val descriptor: SerialDescriptor = buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.Value.String")

        override fun serialize(
            encoder: Encoder,
            value: Value.String,
        ) = encoder.encodeString(value.string)

        override fun deserialize(decoder: Decoder): Value.String = Value.String(decoder.decodeString())
    }

    // For ListValue, we need this ValueSerializer itself for its elements
    private object ListValueSerializer : KSerializer<Value.List> {
        private val actualSerializer = ListSerializer(ValueSerializer) // Recursive use
        override val descriptor: SerialDescriptor = actualSerializer.descriptor

        override fun serialize(
            encoder: Encoder,
            value: Value.List,
        ) = encoder.encodeSerializableValue(actualSerializer, value.list)

        override fun deserialize(decoder: Decoder): Value.List = Value.List(decoder.decodeSerializableValue(actualSerializer))
    }

    // For StructureValue (Map), we need this ValueSerializer for its values
    private object StructureValueSerializer : KSerializer<Value.Structure> {
        private val actualSerializer = MapSerializer(String.serializer(), ValueSerializer) // Recursive use
        override val descriptor: SerialDescriptor = actualSerializer.descriptor

        override fun serialize(
            encoder: Encoder,
            value: Value.Structure,
        ) = encoder.encodeSerializableValue(actualSerializer, value.structure)

        override fun deserialize(decoder: Decoder): Value.Structure = Value.Structure(decoder.decodeSerializableValue(actualSerializer))
    }

    @OptIn(ExperimentalSerializationApi::class)
    private object NullValueSerializer : KSerializer<Value.Null> {
        override val descriptor: SerialDescriptor = buildClassSerialDescriptor("dev.openfeature.kotlin.sdk.Value.Null")

        override fun serialize(
            encoder: Encoder,
            value: Value.Null,
        ) = encoder.encodeNull()

        override fun deserialize(decoder: Decoder): Value.Null {
            decoder.decodeNull() // Consume the null value
            return Value.Null
        }
    }

    override fun serialize(
        encoder: Encoder,
        value: Value,
    ): Unit =
        when (value) {
            is Value.Boolean -> encoder.encodeSerializableValue(BooleanValueSerializer, value)
            is Value.Double -> encoder.encodeSerializableValue(DoubleValueSerializer, value)
            is Value.Integer -> encoder.encodeSerializableValue(IntValueSerializer, value)
            is Value.Instant -> encoder.encodeSerializableValue(InstantValueSerializer, value)
            is Value.List -> encoder.encodeSerializableValue(ListValueSerializer, value)
            is Value.String -> encoder.encodeSerializableValue(StringValueSerializer, value)
            is Value.Structure -> encoder.encodeSerializableValue(StructureValueSerializer, value)
            is Value.Null -> encoder.encodeSerializableValue(NullValueSerializer, value)
        }

    private fun selectDeserializer(element: JsonElement): DeserializationStrategy<Value> =
        when (element) {
            is JsonNull -> NullValueSerializer
            is JsonObject -> StructureValueSerializer // Assumes JsonObject is always a Structure
            is JsonArray -> ListValueSerializer // Assumes JsonArray is always a List
            is JsonPrimitive -> {
                when {
                    // Note: we are not attempting to deserialize any Strings into Instants, because
                    // they might as well be a normal date-looking String
                    element.isString -> StringValueSerializer
                    element.booleanOrNull != null -> BooleanValueSerializer
                    // Order matters here: check for Int before Double to avoid loss of precision
                    // if a number is a whole number but represented as a double (e.g., 5.0)
                    element.longOrNull != null -> {
                        // If it fits in Int, use IntValueSerializer, otherwise could be an issue
                        // or you might need a Value.Long type. For now, assume it fits Int if it's an int.
                        // This part might need refinement based on how you handle large integers.
                        // If Value.Integer only holds Int, then a long might be an error or fallback to Double.
                        // Let's assume for now that if it has no decimal, it could be an Int or a long that
                        // should be treated as Int if it fits, or Double if it's too large for Int but fits Double.
                        val longVal = element.long
                        if (longVal >= Int.MIN_VALUE && longVal <= Int.MAX_VALUE) {
                            IntValueSerializer
                        } else {
                            // Fallback to Double if it's a long that doesn't fit Int
                            // or if your Value.Double can represent whole numbers.
                            DoubleValueSerializer
                        }
                    }
                    element.doubleOrNull != null -> DoubleValueSerializer
                    else -> error("Unknown JsonPrimitive type: $element")
                }
            }
        }

    /**
     * Main deserialize method. It decodes the JsonElement and then uses selectDeserializer
     * to pick the correct strategy for actual object creation.
     */
    override fun deserialize(decoder: Decoder): Value {
        val jsonDecoder =
            decoder as? JsonDecoder
                ?: throw IllegalStateException("This serializer can only be used with JsonInput")

        val jsonElement = jsonDecoder.decodeJsonElement()

        val actualDeserializer = selectDeserializer(jsonElement)

        return jsonDecoder.json.decodeFromJsonElement(actualDeserializer, jsonElement)
    }
}
