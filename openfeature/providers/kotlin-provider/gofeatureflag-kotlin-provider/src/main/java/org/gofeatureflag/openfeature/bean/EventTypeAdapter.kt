package org.gofeatureflag.openfeature.bean

import com.google.gson.Gson
import com.google.gson.JsonDeserializationContext
import com.google.gson.JsonDeserializer
import com.google.gson.JsonElement
import com.google.gson.JsonParseException
import com.google.gson.JsonSyntaxException
import java.lang.reflect.Type

/**
 * Gson TypeAdapter for Event interface.
 * Deserializes events based on the "kind" field.
 */
class EventTypeAdapter : JsonDeserializer<Event> {
    override fun deserialize(
        json: JsonElement,
        typeOfT: Type,
        context: JsonDeserializationContext
    ): Event {
        val jsonObject = json.asJsonObject
        val kind = jsonObject.get("kind")?.asString

        return try {
            when (kind) {
                "feature", "feature_event" -> context.deserialize(json, FeatureEvent::class.java)
                "tracking" -> context.deserialize(json, TrackingEvent::class.java)
                else -> context.deserialize(json, FeatureEvent::class.java)
            }
        } catch (e: JsonSyntaxException) {
            throw JsonParseException("Failed to deserialize event with kind: $kind", e)
        }
    }
}

/**
 * Extension function to create a Gson instance with Event type adapter.
 */
fun createEventsGson(): Gson = Gson()
    .newBuilder()
    .registerTypeAdapter(Event::class.java, EventTypeAdapter())
    .create()
