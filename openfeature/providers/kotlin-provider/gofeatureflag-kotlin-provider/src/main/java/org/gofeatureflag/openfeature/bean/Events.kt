package org.gofeatureflag.openfeature.bean

import com.google.gson.annotations.SerializedName

data class Events(
    @SerializedName("events")
    val events: List<Event>?,
    @SerializedName("meta")
    val meta: Map<String, Any> = emptyMap()
) {
    val featureEvents: List<FeatureEvent>?
        get() = events?.filterIsInstance<FeatureEvent>()

    val trackingEvents: List<TrackingEvent>?
        get() = events?.filterIsInstance<TrackingEvent>()
}
