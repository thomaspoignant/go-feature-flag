package org.gofeatureflag.openfeature.ofrep.error

sealed class OfrepError : Exception() {
    class ApiUnauthorizedError(val response: okhttp3.Response) : OfrepError()
    class ForbiddenError(val response: okhttp3.Response) : OfrepError()
    class ApiTooManyRequestsError(val response: okhttp3.Response? = null) : OfrepError()
    class UnexpectedResponseError(val response: okhttp3.Response) : OfrepError()
    class UnmarshallError(val e: Exception) : OfrepError()
    class InvalidOptionsError(override val message: String?) : OfrepError()
}