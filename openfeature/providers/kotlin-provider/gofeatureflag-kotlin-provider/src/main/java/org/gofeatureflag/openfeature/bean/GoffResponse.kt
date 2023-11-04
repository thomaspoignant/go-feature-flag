package org.gofeatureflag.openfeature.bean

data class GoffResponse(
    val flags: Map<String, FlagState>,
    val valid: Boolean
)

