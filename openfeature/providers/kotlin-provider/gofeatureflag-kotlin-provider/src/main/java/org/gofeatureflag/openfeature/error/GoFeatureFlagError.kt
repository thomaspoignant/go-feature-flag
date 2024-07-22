package org.gofeatureflag.openfeature.error

sealed class GoFeatureFlagError : Exception() {
    class InvalidOptionsError(override val message: String?) : GoFeatureFlagError()
    class ApiUnauthorizedError(val response: okhttp3.Response) : GoFeatureFlagError()
    class InvalidRequest(val response: okhttp3.Response) : GoFeatureFlagError()
    class UnexpectedResponseError(val response: okhttp3.Response) : GoFeatureFlagError()
}