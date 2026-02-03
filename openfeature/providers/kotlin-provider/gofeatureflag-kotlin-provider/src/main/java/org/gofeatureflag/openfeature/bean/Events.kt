package org.gofeatureflag.openfeature.bean


data class Events(
    val events: List<Event>?,
    val meta: Map<String, Any> = emptyMap()
)