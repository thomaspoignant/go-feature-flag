package dev.openfeature.kotlin.contrib.providers.ofrep.controller

import dev.openfeature.kotlin.contrib.providers.ofrep.FAKE_ENDPOINT
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.FlagDto
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepApiResponse
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepOptions
import dev.openfeature.kotlin.contrib.providers.ofrep.error.OfrepError
import dev.openfeature.kotlin.contrib.providers.ofrep.mockEngineWithOneResponse
import dev.openfeature.kotlin.contrib.providers.ofrep.mockEngineWithTwoResponses
import dev.openfeature.kotlin.contrib.providers.ofrep.payloads.INVALID_API_RESPONSE_PAYLOAD
import dev.openfeature.kotlin.contrib.providers.ofrep.payloads.VALID_API_RESPONSE_PAYLOAD
import dev.openfeature.kotlin.contrib.providers.ofrep.payloads.VALID_API_SHORT_RESPONSE_PAYLOAD
import dev.openfeature.kotlin.sdk.EvaluationMetadata
import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.Value
import dev.openfeature.kotlin.sdk.exceptions.ErrorCode
import dev.openfeature.kotlin.sdk.exceptions.OpenFeatureError
import io.ktor.client.engine.mock.MockEngine
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlin.test.fail

private fun createOfrepApi(mockEngine: MockEngine) =
    OfrepApi(
        OfrepOptions(endpoint = FAKE_ENDPOINT, httpClientEngine = mockEngine),
    )

class OfrepApiTest {
    @Test
    fun shouldReturnAValidEvaluationResponse() =
        runTest {
            val mockEngine = mockEngineWithOneResponse(content = VALID_API_SHORT_RESPONSE_PAYLOAD)

            val ofrepApi = createOfrepApi(mockEngine)

            val ctx =
                ImmutableContext(
                    targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd",
                    attributes =
                        mutableMapOf(
                            "email" to Value.String("batman@gofeatureflag.org"),
                        ),
                )
            val res = ofrepApi.postBulkEvaluateFlags(ctx)
            assertEquals(200, res.httpResponse.status.value)

            val expected =
                OfrepApiResponse(
                    flags =
                        listOf(
                            FlagDto(
                                key = "badge-class2",
                                value = Value.String("green"),
                                reason = "DEFAULT",
                                variant = "nocolor",
                                errorCode = null,
                                errorDetails = null,
                                metadata = EvaluationMetadata.EMPTY,
                            ),
                            FlagDto(
                                key = "hide-logo",
                                value = Value.Boolean(false),
                                reason = "STATIC",
                                variant = "var_false",
                                errorCode = null,
                                errorDetails = null,
                                metadata = EvaluationMetadata.EMPTY,
                            ),
                            FlagDto(
                                key = "title-flag",
                                value = Value.String("GO Feature Flag"),
                                reason = "DEFAULT",
                                variant = "default_title",
                                errorCode = null,
                                errorDetails = null,
                                metadata =
                                    EvaluationMetadata
                                        .builder()
                                        .putString(
                                            "description",
                                            "This flag controls the title of the feature flag",
                                        ).putString("title", "Feature Flag Title")
                                        .build(),
                            ),
                        ),
                    null,
                    null,
                )
            assertEquals(expected, res.apiResponse)
        }

    @Test
    fun shouldThrowAnUnauthorizedError() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = "{}",
                    status = HttpStatusCode.fromValue(401),
                )
            val ofrepApi = createOfrepApi(mockEngine)
            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            assertFailsWith<OfrepError.ApiUnauthorizedError> {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }

    @Test
    fun shouldThrowAForbiddenError() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = "{}",
                    status = HttpStatusCode.fromValue(403),
                )
            val ofrepApi = createOfrepApi(mockEngine)
            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            assertFailsWith<OfrepError.ForbiddenError> {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }

    @Test
    fun shouldThrowTooManyRequest() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = "{}",
                    status = HttpStatusCode.fromValue(429),
                    additionalHeaders = headersOf(HttpHeaders.RetryAfter, "120"),
                )
            val ofrepApi = createOfrepApi(mockEngine)
            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            try {
                ofrepApi.postBulkEvaluateFlags(ctx)
                fail("we exited the try block without throwing an exception")
            } catch (e: OfrepError.ApiTooManyRequestsError) {
                assertEquals(429, e.response?.status?.value)
                assertEquals(e.response?.headers?.get("Retry-After"), "120")
            }
        }

    @Test
    fun shouldThrowUnexpectedError() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = "{}",
                    status = HttpStatusCode.fromValue(500),
                )
            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            assertFailsWith<OfrepError.UnexpectedResponseError> {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }

    @Test
    fun shouldReturnAnEvaluationResponseInError() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content =
                        """
                        {"errorCode": "INVALID_CONTEXT", "errorDetails":"explanation of the error"}
                        """.trimIndent(),
                    status = HttpStatusCode.fromValue(400),
                )

            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            val resp = ofrepApi.postBulkEvaluateFlags(ctx)
            assertTrue(resp.isError())
            assertEquals(ErrorCode.INVALID_CONTEXT, resp.apiResponse?.errorCode)
            assertEquals("explanation of the error", resp.apiResponse?.errorDetails)
            assertEquals(400, resp.httpResponse.status.value)
        }

    @Test
    fun shouldReturnaEvaluationResponseIfWeReceiveA304() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = "{}",
                    status = HttpStatusCode.fromValue(304),
                )

            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            val resp = ofrepApi.postBulkEvaluateFlags(ctx)
            assertFalse(resp.isError())
            assertEquals(304, resp.httpResponse.status.value)
        }

    @Test
    fun shouldThrowTargetingKeyMissingErrorWithNoTargetingKey() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = "{}",
                    status = HttpStatusCode.fromValue(304),
                )
            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "")
            assertFailsWith<OpenFeatureError.TargetingKeyMissingError> {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }

    @Test
    fun shouldThrowUnmarshallErrorWithInvalidJson() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    content = INVALID_API_RESPONSE_PAYLOAD,
                    status = HttpStatusCode.fromValue(400),
                )
            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            assertFailsWith<OfrepError.UnmarshallError> {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }

    @Test
    fun shouldThrowWithInvalidOptions() =
        runTest {
            assertFailsWith<OfrepError.InvalidOptionsError> {
                OfrepApi(OfrepOptions(endpoint = "invalid_url"))
            }
        }

    @Test
    fun shouldETagShouldNotMatch() =
        runTest {
            val mockEngine =
                mockEngineWithTwoResponses(
                    firstContent = VALID_API_RESPONSE_PAYLOAD,
                    firstStatus = HttpStatusCode.fromValue(200),
                    firstAdditionalHeaders = headersOf(HttpHeaders.ETag, "123"),
                    secondContent = "",
                    secondStatus = HttpStatusCode.fromValue(304),
                    secondAdditionalHeaders = headersOf(HttpHeaders.ETag, "123"),
                )

            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            val eval1 = ofrepApi.postBulkEvaluateFlags(ctx)
            val eval2 = ofrepApi.postBulkEvaluateFlags(ctx)
            assertEquals(eval1.httpResponse.status.value, 200)
            assertEquals(eval2.httpResponse.status.value, 304)
            assertEquals(2, mockEngine.requestHistory.size)
        }

    @Test
    fun shouldHaveIfNoneNullInTheHeaders() =
        runTest {
            val mockEngine =
                mockEngineWithTwoResponses(
                    firstContent = VALID_API_RESPONSE_PAYLOAD,
                    firstStatus = HttpStatusCode.fromValue(200),
                    firstAdditionalHeaders = headersOf(HttpHeaders.ETag, "123"),
                    secondContent = "",
                    secondStatus = HttpStatusCode.fromValue(304),
                    secondAdditionalHeaders = headersOf(HttpHeaders.ETag, "123"),
                )

            val ofrepApi = createOfrepApi(mockEngine)

            val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
            val eval1 = ofrepApi.postBulkEvaluateFlags(ctx)
            val eval2 = ofrepApi.postBulkEvaluateFlags(ctx)
            assertEquals(eval1.httpResponse.status.value, 200)
            assertEquals(eval2.httpResponse.status.value, 304)
            assertEquals(2, mockEngine.requestHistory.size)
        }
}
