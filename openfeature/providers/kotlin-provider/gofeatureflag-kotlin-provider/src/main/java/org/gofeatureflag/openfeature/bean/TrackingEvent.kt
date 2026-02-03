package org.gofeatureflag.openfeature.bean

import com.google.gson.annotations.SerializedName

/**
 * Tracking event data.
 *
 * A tracking event will only be generated if the trackEvents attribute of the flag is set to true.
 */
data class TrackingEvent(
    /**
     * Kind for a feature event is "feature".
     */
    @SerializedName("kind")
    val kind: String? = null,

    /**
     * ContextKind is the kind of context which generated an event.
     * This will only be "anonymousUser" for events generated on behalf of an anonymous user
     * or "user" for events generated on behalf of a non-anonymous user.
     */
    @SerializedName("contextKind")
    val contextKind: String? = null,

    /**
     * UserKey The key of the user object used in a feature flag evaluation.
     */
    @SerializedName("userKey")
    val userKey: String? = null,

    /**
     * CreationDate When the feature flag was requested at Unix epoch time in milliseconds.
     */
    @SerializedName("creationDate")
    val creationDate: Long? = null,

    /**
     * Key of the event.
     */
    @SerializedName("key")
    val key: String? = null,

    /**
     * EvaluationContext contains the evaluation context used for the tracking.
     */
    @SerializedName("evaluationContext")
    val evaluationContext: Map<String, Any?>? = null,

    /**
     * TrackingDetails contains the details of the tracking event.
     */
    @SerializedName("trackingEventDetails")
    val trackingEventDetails: Map<String, Any?>? = null
) : Event
