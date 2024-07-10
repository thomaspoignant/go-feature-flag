package org.gofeatureflag.openfeature.ofrep

import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.ImmutableContext
import dev.openfeature.sdk.OpenFeatureAPI
import dev.openfeature.sdk.events.OpenFeatureEvents
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext
import okhttp3.Headers
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import java.nio.file.Files
import java.nio.file.Paths
import java.util.UUID

class OfrepProviderTest {
    private var mockWebServer: MockWebServer? = null
    private var defaultEvalCtx: EvaluationContext =
        ImmutableContext(targetingKey = UUID.randomUUID().toString())

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
    fun shouldHaveAProviderMetadata() {
        val provider = OfrepProvider(OfrepOptions(endpoint = "http://localhost:1031"))
        assertEquals("OFREP Provider", provider.metadata.name)
    }

    @Test
    fun shouldBeInFatalStatusIf401ErrorDuringInitialise(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")
        val provider = OfrepProvider(OfrepOptions(endpoint = "http://localhost:1031"))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        OpenFeatureAPI.observe<OpenFeatureEvents.ProviderReady>().collect {
            println(">> ProviderReady event received")
        }
    }

    suspend fun enqueueMockResponse(
        filePath: String,
        statusCode: Int = 200,
        headers: Headers = Headers.Builder().build()
    ) {
        val jsonFilePath =
            javaClass.classLoader?.getResource(filePath)?.file
        val jsonString = String(withContext(Dispatchers.IO) {
            Files.readAllBytes(Paths.get(jsonFilePath))
        })
        mockWebServer!!.enqueue(
            MockResponse()
                .setBody(jsonString.trimIndent())
                .setResponseCode(statusCode)
                .setHeaders(headers)
        )
    }
}
