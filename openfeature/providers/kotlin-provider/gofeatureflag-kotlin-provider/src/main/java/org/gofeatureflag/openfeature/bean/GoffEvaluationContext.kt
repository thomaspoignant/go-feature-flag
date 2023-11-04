package org.gofeatureflag.openfeature.bean

data class GoffEvaluationContext(val key: String, val custom: Map<String, Any?>) {
}