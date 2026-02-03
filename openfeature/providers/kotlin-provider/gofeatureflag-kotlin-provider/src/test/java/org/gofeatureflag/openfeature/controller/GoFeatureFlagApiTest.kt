package org.gofeatureflag.openfeature.controller

import kotlinx.coroutines.runBlocking
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.gofeatureflag.openfeature.error.GoFeatureFlagError
import org.gofeatureflag.openfeature.bean.Event
import org.junit.After
import org.junit.Assert
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import org.skyscreamer.jsonassert.JSONAssert
import java.io.File


class GoFeatureFlagApiTest {
    private var mockWebServer: MockWebServer? = null
    private var defaultEventList: List<Event> = listOf(
        Event(
            contextKind = "contextKind",
            creationDate = 1721650841,
            key = "flag-1",
            kind = "feature",
            userKey = "981f2662-1fb4-4732-ac6d-8399d9205aa9",
            value = true,
            default = false,
            variation = "enabled",
            source = "PROVIDER_CACHE"
        )
    )

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
    fun `should not call the API if list of events empty`() = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(200))
        val api =
            GoFeatureFlagApi(GoFeatureFlagOptions(endpoint = mockWebServer!!.url("/").toString()))
        api.postEventsToDataCollector(emptyList())
        assertEquals(0, mockWebServer!!.requestCount)
    }

    @Test
    fun `should throw an error if 401`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(401))
        val api =
            GoFeatureFlagApi(GoFeatureFlagOptions(endpoint = mockWebServer!!.url("/").toString()))
        Assert.assertThrows(GoFeatureFlagError.ApiUnauthorizedError::class.java) {
            runBlocking {
                api.postEventsToDataCollector(defaultEventList)
            }
        }
    }

    @Test
    fun `should throw an error if 403`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(403))
        val api =
            GoFeatureFlagApi(GoFeatureFlagOptions(endpoint = mockWebServer!!.url("/").toString()))
        Assert.assertThrows(GoFeatureFlagError.ApiUnauthorizedError::class.java) {
            runBlocking {
                api.postEventsToDataCollector(defaultEventList)
            }
        }
    }

    @Test
    fun `should throw an error if 500`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(500))
        val api =
            GoFeatureFlagApi(GoFeatureFlagOptions(endpoint = mockWebServer!!.url("/").toString()))
        Assert.assertThrows(GoFeatureFlagError.UnexpectedResponseError::class.java) {
            runBlocking {
                api.postEventsToDataCollector(defaultEventList)
            }
        }
    }

    @Test
    fun `should throw an error if 400`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(400))
        val api =
            GoFeatureFlagApi(GoFeatureFlagOptions(endpoint = mockWebServer!!.url("/").toString()))
        Assert.assertThrows(GoFeatureFlagError.InvalidRequest::class.java) {
            runBlocking {
                api.postEventsToDataCollector(defaultEventList)
            }
        }
    }

    @Test
    fun `should be ok if using an API Key`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(200))
        val api =
            GoFeatureFlagApi(
                GoFeatureFlagOptions(
                    endpoint = mockWebServer!!.url("/").toString(),
                    apiKey = "my-api-key"
                )
            )
        api.postEventsToDataCollector(defaultEventList)
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()
        assertEquals("Bearer my-api-key", recordedRequest.headers["Authorization"])
    }

    @Test
    fun `should return an error if invalid endpoint`(): Unit = runBlocking {
        Assert.assertThrows(GoFeatureFlagError.InvalidOptionsError::class.java) {
            GoFeatureFlagApi(GoFeatureFlagOptions(endpoint = "mockWebServer!!.url().toString()"))
        }
    }

    @Test
    fun `should have a valid body request`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(200))
        val api =
            GoFeatureFlagApi(
                GoFeatureFlagOptions(
                    endpoint = mockWebServer!!.url("/").toString(),
                    apiKey = "my-api-key"
                )
            )
        api.postEventsToDataCollector(defaultEventList)
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()
        val jsonFilePath =
            javaClass.classLoader?.getResource("org/gofeatureflag/openfeature/hook/valid_result.json")?.file
        val want = File(jsonFilePath.toString()).readText(Charsets.UTF_8)
        val got = recordedRequest.body.readUtf8()
        JSONAssert.assertEquals(want, got, false)
    }

    @Test
    fun `should have a valid body request when using exporter metadata`(): Unit = runBlocking {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setResponseCode(200))
        val api =
            GoFeatureFlagApi(
                GoFeatureFlagOptions(
                    endpoint = mockWebServer!!.url("/").toString(),
                    apiKey = "my-api-key",
                    exporterMetadata = mapOf("appVersion" to "1.0.0", "device" to "Pixel 4")
                )
            )
        api.postEventsToDataCollector(defaultEventList)
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()
        val jsonFilePath =
            javaClass.classLoader?.getResource("org/gofeatureflag/openfeature/hook/valid_result_metadata.json")?.file
        val want = File(jsonFilePath.toString()).readText(Charsets.UTF_8)
        val got = recordedRequest.body.readUtf8()
        JSONAssert.assertEquals(want, got, false)
    }
}
