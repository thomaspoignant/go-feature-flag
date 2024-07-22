package org.gofeatureflag.openfeature.hook


data class Events(
    val events: List<Event>?,
    val meta: Map<String, String> = mapOf("provider" to "android", "openfeature" to "true")
)