package org.gofeatureflag.openfeature.bean

data class Event(
    val contextKind: String? = null,
    val creationDate: Long? = null,
    val key: String? = null,
    val kind: String? = null,
    val userKey: String? = null,
    val value: Any? = null,
    val default: Any? = null, // this is the default value of the flag
    val variation: String? = null,
    val source: String? = null
)