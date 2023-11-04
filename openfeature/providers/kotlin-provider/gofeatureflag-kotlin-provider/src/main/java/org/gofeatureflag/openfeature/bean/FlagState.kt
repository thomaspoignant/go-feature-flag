package org.gofeatureflag.openfeature.bean

data class FlagState(
    val value: Any,
    val timestamp: Long,
    val variationType: String,
    val trackEvents: Boolean,
    val failed: Boolean,
    val errorCode: String,
    val reason: String,
    val metadata: Map<String, Any>?
)