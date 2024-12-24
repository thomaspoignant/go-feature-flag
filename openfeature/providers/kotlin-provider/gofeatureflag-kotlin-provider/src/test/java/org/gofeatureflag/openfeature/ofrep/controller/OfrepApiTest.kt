package org.gofeatureflag.openfeature.ofrep.controller

import FlagDto
import OfrepApiResponse
import dev.openfeature.sdk.ImmutableContext
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.exceptions.ErrorCode
import dev.openfeature.sdk.exceptions.OpenFeatureError
import junit.framework.TestCase.assertFalse
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions
import org.gofeatureflag.openfeature.ofrep.error.OfrepError
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.nio.file.Files
import java.nio.file.Paths


class OfrepApiTest {
    private var mockWebServer: MockWebServer? = null

    @Before
    fun before() {
        mockWebServer = MockWebServer()
        mockWebServer!!.start(10031)
    }

    @After
    fun after() {
        mockWebServer!!.shutdown()
    }

    @Test
    fun shouldReturnAValidEvaluationResponse() = runBlocking {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_short_response.json")?.file
        val jsonString = String(withContext(Dispatchers.IO) {
            Files.readAllBytes(Paths.get(jsonFilePath))
        })

        mockWebServer!!.enqueue(MockResponse().setBody(jsonString.trimIndent()))

        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )
        val ctx = ImmutableContext(
            targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd",
            attributes = mutableMapOf(
                "email" to Value.String("batman@gofeatureflag.org")
            )
        )
        val res = ofrepApi.postBulkEvaluateFlags(ctx)
        assertEquals(200, res.httpResponse.code)

        val expected = OfrepApiResponse(
            flags = listOf(
                FlagDto(
                    key = "badge-class2",
                    value = "green",
                    reason = "DEFAULT",
                    variant = "nocolor",
                    errorCode = null,
                    errorDetails = null,
                    metadata = null
                ),
                FlagDto(
                    key = "hide-logo",
                    value = false,
                    reason = "STATIC",
                    variant = "var_false",
                    errorCode = null,
                    errorDetails = null,
                    metadata = null
                ),
                FlagDto(
                    key = "title-flag",
                    value = "GO Feature Flag",
                    reason = "DEFAULT",
                    variant = "default_title",
                    errorCode = null,
                    errorDetails = null,
                    metadata = hashMapOf<String, Any>(
                        "description" to "This flag controls the title of the feature flag",
                        "title" to "Feature Flag Title"
                    )
                )
            ), null, null
        )
        assertEquals(expected, res.apiResponse)
    }

    @Test
    fun shouldThrowAnUnauthorizedError(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(401))
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )
        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        assertThrows(OfrepError.ApiUnauthorizedError::class.java) {
            runBlocking {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }
    }

    @Test
    fun shouldThrowAForbiddenError(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(403))
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )
        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        assertThrows(OfrepError.ForbiddenError::class.java) {
            runBlocking {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }
    }

    @Test
    fun shouldThrowTooManyRequest(): Unit = runBlocking {
        mockWebServer!!.enqueue(
            MockResponse()
                .setBody("{}")
                .setResponseCode(429)
                .setHeader("Retry-After", "120")
        )
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )
        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        try {
            ofrepApi.postBulkEvaluateFlags(ctx)
            assertTrue("we exited the try block without throwing an exception", false)
        } catch (e: OfrepError.ApiTooManyRequestsError) {
            assertEquals(429, e.response?.code)
            assertEquals(e.response?.headers?.get("Retry-After"), "120")
        }
    }

    @Test
    fun shouldThrowUnexpectedError(): Unit = runBlocking {
        mockWebServer!!.enqueue(
            MockResponse()
                .setBody("{}")
                .setResponseCode(500)
        )
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        assertThrows(OfrepError.UnexpectedResponseError::class.java) {
            runBlocking {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }
    }

    @Test
    fun shouldReturnAnEvaluationResponseInError(): Unit = runBlocking {
        mockWebServer!!.enqueue(
            MockResponse()
                .setBody(
                    """
                    {"errorCode": "INVALID_CONTEXT", "errorDetails":"explanation of the error"}
                """.trimIndent()
                )
                .setResponseCode(400)
        )
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        val resp = ofrepApi.postBulkEvaluateFlags(ctx)
        assertTrue(resp.isError())
        assertEquals(ErrorCode.INVALID_CONTEXT, resp.apiResponse?.errorCode)
        assertEquals("explanation of the error", resp.apiResponse?.errorDetails)
        assertEquals(400, resp.httpResponse.code)
    }

    @Test
    fun shouldReturnaEvaluationResponseIfWeReceiveA304(): Unit = runBlocking {
        mockWebServer!!.enqueue(
            MockResponse()
                .setBody("{}")
                .setResponseCode(304)
        )
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        val resp = ofrepApi.postBulkEvaluateFlags(ctx)
        assertFalse(resp.isError())
        assertEquals(304, resp.httpResponse.code)
    }

    @Test
    fun shouldThrowTargetingKeyMissingErrorWithNoTargetingKey(): Unit = runBlocking {
        mockWebServer!!.enqueue(
            MockResponse()
                .setBody("{}")
                .setResponseCode(304)
        )
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "")
        assertThrows(OpenFeatureError.TargetingKeyMissingError::class.java) {
            runBlocking {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }
    }

    @Test
    fun shouldThrowUnmarshallErrorWithInvalidJson(): Unit = runBlocking {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/invalid_api_response.json")?.file
        val jsonString = String(withContext(Dispatchers.IO) {
            Files.readAllBytes(Paths.get(jsonFilePath))
        })

        mockWebServer!!.enqueue(
            MockResponse().setBody(jsonString.trimIndent()).setResponseCode(400)
        )
        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        assertThrows(OfrepError.UnmarshallError::class.java) {
            runBlocking {
                ofrepApi.postBulkEvaluateFlags(ctx)
            }
        }
    }

    @Test
    fun shouldThrowWithInvalidOptions(): Unit = runBlocking {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/invalid_api_response.json")?.file
        val jsonString = String(withContext(Dispatchers.IO) {
            Files.readAllBytes(Paths.get(jsonFilePath))
        })

        mockWebServer!!.enqueue(
            MockResponse().setBody(jsonString.trimIndent()).setResponseCode(400)
        )
        assertThrows(OfrepError.InvalidOptionsError::class.java) {
            runBlocking {
                OfrepApi(OfrepOptions(endpoint = "invalid_url"))
            }
        }
    }

    @Test
    fun shouldETagShouldNotMatch(): Unit = runBlocking {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")?.file
        val jsonString = String(withContext(Dispatchers.IO) {
            Files.readAllBytes(Paths.get(jsonFilePath))
        })

        mockWebServer!!.enqueue(
            MockResponse()
                .setBody(jsonString.trimIndent())
                .setResponseCode(200)
                .addHeader("ETag", "123")
        )
        mockWebServer!!.enqueue(
            MockResponse()
                .setResponseCode(304)
                .addHeader("ETag", "123")
        )

        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        val eval1 = ofrepApi.postBulkEvaluateFlags(ctx)
        val eval2 = ofrepApi.postBulkEvaluateFlags(ctx)
        assertEquals(eval1.httpResponse.code, 200)
        assertEquals(eval2.httpResponse.code, 304)
        assertEquals(2, mockWebServer!!.requestCount)
    }

    @Test
    fun shouldHaveIfNoneNullInTheHeaders(): Unit = runBlocking {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")?.file
        val jsonString = String(withContext(Dispatchers.IO) {
            Files.readAllBytes(Paths.get(jsonFilePath))
        })

        mockWebServer!!.enqueue(
            MockResponse()
                .setBody(jsonString.trimIndent())
                .setResponseCode(200)
                .addHeader("ETag", "123")
        )
        mockWebServer!!.enqueue(
            MockResponse()
                .setResponseCode(304)
                .addHeader("ETag", "123")
        )

        val ofrepApi = OfrepApi(
            OfrepOptions(endpoint = mockWebServer!!.url("/").toString())
        )

        val ctx = ImmutableContext(targetingKey = "68cf565d-15cd-4e8b-95a6-9399987164cd")
        val eval1 = ofrepApi.postBulkEvaluateFlags(ctx)
        val eval2 = ofrepApi.postBulkEvaluateFlags(ctx)
        assertEquals(eval1.httpResponse.code, 200)
        assertEquals(eval2.httpResponse.code, 304)
        assertEquals(2, mockWebServer!!.requestCount)
    }
}