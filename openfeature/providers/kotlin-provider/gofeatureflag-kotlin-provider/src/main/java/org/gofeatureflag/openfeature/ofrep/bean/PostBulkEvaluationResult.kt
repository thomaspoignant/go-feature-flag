package org.gofeatureflag.openfeature.ofrep.bean

import OfrepApiResponse

data class PostBulkEvaluationResult(
    val apiResponse: OfrepApiResponse?,
    val httpResponse: okhttp3.Response
) {
    fun isError(): Boolean {
        return apiResponse?.errorCode != null
    }
}
