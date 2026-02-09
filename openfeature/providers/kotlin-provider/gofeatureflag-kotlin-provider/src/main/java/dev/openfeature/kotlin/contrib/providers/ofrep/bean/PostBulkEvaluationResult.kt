package dev.openfeature.kotlin.contrib.providers.ofrep.bean

import io.ktor.client.statement.HttpResponse

internal data class PostBulkEvaluationResult(
    val apiResponse: OfrepApiResponse?,
    val httpResponse: HttpResponse,
) {
    fun isError(): Boolean = apiResponse?.errorCode != null
}
