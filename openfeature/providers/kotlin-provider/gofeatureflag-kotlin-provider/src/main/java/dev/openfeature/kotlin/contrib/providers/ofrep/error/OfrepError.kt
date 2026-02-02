package dev.openfeature.kotlin.contrib.providers.ofrep.error

import io.ktor.client.statement.HttpResponse

internal sealed class OfrepError : Exception() {
    class ApiUnauthorizedError(
        val response: HttpResponse,
    ) : OfrepError()

    class ForbiddenError(
        val response: HttpResponse,
    ) : OfrepError()

    class ApiTooManyRequestsError(
        val response: HttpResponse? = null,
    ) : OfrepError()

    class UnexpectedResponseError(
        val response: HttpResponse,
    ) : OfrepError()

    class UnmarshallError(
        val e: Exception,
    ) : OfrepError()

    class InvalidOptionsError(
        override val message: String?,
    ) : OfrepError()
}
