package dev.openfeature.kotlin.contrib.providers.ofrep

import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.http.ContentType
import io.ktor.http.Headers
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headers
import io.ktor.http.headersOf
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestScope

internal const val FAKE_ENDPOINT = "http://localhost/"

internal fun mockEngineWithOneResponse(
    content: String = "",
    status: HttpStatusCode = HttpStatusCode.OK,
    additionalHeaders: Headers = headersOf(),
) = MockEngine {
    respond(
        content = content,
        status = status,
        headers =
            headers {
                appendAll(additionalHeaders)
                append(
                    HttpHeaders.ContentType,
                    ContentType.Application.Json.toString(),
                )
            },
    )
}

internal fun TestScope.mockEngineWithTwoResponses(
    firstContent: String,
    firstStatus: HttpStatusCode = HttpStatusCode.OK,
    firstAdditionalHeaders: Headers = headersOf(),
    secondContent: String,
    secondStatus: HttpStatusCode = HttpStatusCode.OK,
    secondAdditionalHeaders: Headers = headersOf(),
): MockEngine {
    var counter = 0
    return MockEngine.create {
        dispatcher = StandardTestDispatcher(testScheduler)
        addHandler {
            val (content, status, additionalHeaders) =
                when (counter++) {
                    0 -> arrayOf(firstContent, firstStatus, firstAdditionalHeaders)
                    1 -> arrayOf(secondContent, secondStatus, secondAdditionalHeaders)
                    else -> error("Only two calls expected")
                }
            respond(
                content = content as String,
                status = status as HttpStatusCode,
                headers =
                    headers {
                        appendAll(additionalHeaders as Headers)
                        append(
                            HttpHeaders.ContentType,
                            ContentType.Application.Json.toString(),
                        )
                    },
            )
        }
    } as MockEngine
}
