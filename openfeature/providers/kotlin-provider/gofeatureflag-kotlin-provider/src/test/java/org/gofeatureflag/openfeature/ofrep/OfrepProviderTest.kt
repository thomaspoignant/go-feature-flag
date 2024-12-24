package org.gofeatureflag.openfeature.ofrep

import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.EvaluationMetadata
import dev.openfeature.sdk.FlagEvaluationDetails
import dev.openfeature.sdk.ImmutableContext
import dev.openfeature.sdk.OpenFeatureAPI
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.events.OpenFeatureEvents
import dev.openfeature.sdk.events.observe
import dev.openfeature.sdk.exceptions.ErrorCode
import dev.openfeature.sdk.exceptions.OpenFeatureError
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.take
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import okhttp3.Headers
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.gofeatureflag.openfeature.ofrep.bean.OfrepOptions
import org.gofeatureflag.openfeature.ofrep.error.OfrepError
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
        OpenFeatureAPI.shutdown()
        OpenFeatureAPI.clearProvider()
        OpenFeatureAPI.clearHooks()
        mockWebServer?.shutdown()
        mockWebServer = null
    }

    @Test
    fun `should have a provider metadata`() {
        val provider = OfrepProvider(OfrepOptions(endpoint = "http://localhost:1031"))
        assertEquals("OFREP Provider", provider.metadata.name)
    }

    @Test
    fun `should be in Fatal status if 401 error during initialise`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 401)

        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
            }
        }
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
    }

    @Test
    fun `should be in Fatal status if 403 error during initialise`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 403)

        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
            }
        }
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
    }


    @Test
    fun `should be in Error status if 429 error during initialise`(): Unit = runBlocking {
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_api_response.json",
            429,
            Headers.headersOf("Retry-After", "3")
        )

        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false
        var exceptionReceived: Throwable? = null

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
                exceptionReceived = it.error
            }
        }
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
        assert(exceptionReceived is OfrepError.ApiTooManyRequestsError) { "The exception is not of type ApiTooManyRequestsError" }
    }


    @Test
    fun `should be in Error status if error targeting key is empty`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)

        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false
        var exceptionReceived: Throwable? = null

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
                exceptionReceived = it.error
            }
        }
        val evalCtx = ImmutableContext(targetingKey = "")
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, evalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
        assert(exceptionReceived is OpenFeatureError.TargetingKeyMissingError) { "The exception is not of type TargetingKeyMissingError" }
    }

    @Test
    fun `should be in Error status if error targeting key is missing`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)

        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false
        var exceptionReceived: Throwable? = null

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
                exceptionReceived = it.error
            }
        }
        val evalCtx = ImmutableContext()
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, evalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
        assert(exceptionReceived is OpenFeatureError.TargetingKeyMissingError) { "The exception is not of type TargetingKeyMissingError" }
    }


    @Test
    fun `should be in error status if error invalid context`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/invalid_context.json", 400)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false
        var exceptionReceived: Throwable? = null

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
                exceptionReceived = it.error
            }
        }
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
        assert(exceptionReceived is OpenFeatureError.InvalidContextError) { "The exception is not of type InvalidContextError" }
    }

    @Test
    fun `should be in error status if error parse error`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/parse_error.json", 400)

        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        var providerErrorReceived = false
        var exceptionReceived: Throwable? = null

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderError>().take(1).collect {
                providerErrorReceived = true
                exceptionReceived = it.error
            }
        }
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        assert(providerErrorReceived) { "ProviderError event was not received" }
        assert(exceptionReceived is OpenFeatureError.ParseError) { "The exception is not of type ParseError" }
    }

    @Test
    fun `should return a flag not found error if the flag does not exist`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getBooleanDetails("non-existent-flag", false)
        val want = FlagEvaluationDetails<Boolean>(
            flagKey = "non-existent-flag",
            value = false,
            variant = null,
            reason = "ERROR",
            errorCode = ErrorCode.FLAG_NOT_FOUND,
            errorMessage = "Could not find flag named: non-existent-flag"
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return evaluation details if the flag exists`(): Unit = runBlocking {
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_api_short_response.json",
            200
        )
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getStringDetails("title-flag", "default")
        val want = FlagEvaluationDetails<String>(
            flagKey = "title-flag",
            value = "GO Feature Flag",
            variant = "default_title",
            reason = "DEFAULT",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putString("description", "This flag controls the title of the feature flag")
                .putString("title", "Feature Flag Title")
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return parse error if the API returns the error`(): Unit = runBlocking {
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_1_flag_in_parse_error.json",
            200
        )
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getStringDetails("my-other-flag", "default")
        val want = FlagEvaluationDetails<String>(
            flagKey = "my-other-flag",
            value = "default",
            variant = null,
            reason = "ERROR",
            errorCode = ErrorCode.PARSE_ERROR,
            errorMessage = "Error details about PARSE_ERROR",
        )
        assertEquals(want, got)
    }

    @Test
    fun `should send a context changed event if context has changed`(): Unit = runBlocking {
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_api_response.json",
            200
        )
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_api_response_2.json",
            200
        )
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)

        // TODO: should change when we have a way to observe context changes event
        //       check issue https://github.com/open-feature/kotlin-sdk/issues/107
        var providerStaleEventReceived = false
        var providerReadyEventReceived = false

        val coroutineScope = CoroutineScope(Dispatchers.IO)
        coroutineScope.launch {
            provider.observe<OpenFeatureEvents.ProviderStale>().take(1).collect {
                providerStaleEventReceived = true
            }
            provider.observe<OpenFeatureEvents.ProviderReady>().take(1).collect {
                providerReadyEventReceived = true
            }
        }
        Thread.sleep(1000) // waiting to be sure that setEvaluationContext has been processed
        val newEvalCtx = ImmutableContext(targetingKey = UUID.randomUUID().toString())
        OpenFeatureAPI.setEvaluationContext(newEvalCtx)
        Thread.sleep(1000) // waiting to be sure that setEvaluationContext has been processed
        assert(providerStaleEventReceived) { "ProviderStale event was not received" }
        assert(providerReadyEventReceived) { "ProviderReady event was not received" }
    }

    @Test
    fun `should not try to call the API before Retry-After header`(): Unit = runBlocking {
        mockWebServer!!.enqueue(
            MockResponse()
                .setResponseCode(429)
                .setHeader("Retry-After", "3")
        )
        val provider = OfrepProvider(
            OfrepOptions(
                pollingIntervalInMillis = 100,
                endpoint = mockWebServer?.url("/").toString()
            )
        )
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        client.getStringDetails("my-other-flag", "default")
        client.getStringDetails("my-other-flag", "default")
        Thread.sleep(2000) // we wait 2 seconds to let the polling loop run
        assertEquals(1, mockWebServer!!.requestCount)
    }

    @Test
    fun `should return a valid evaluation for Boolean`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getBooleanDetails("bool-flag", false)
        val want = FlagEvaluationDetails<Boolean>(
            flagKey = "bool-flag",
            value = true,
            variant = "variantA",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putBoolean("additionalProp1", true)
                .putString("additionalProp2", "value")
                .putInt("additionalProp3", 123)
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return a valid evaluation for Int`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getIntegerDetails("int-flag", 1)
        val want = FlagEvaluationDetails<Int>(
            flagKey = "int-flag",
            value = 1234,
            variant = "variantA",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putBoolean("additionalProp1", true)
                .putString("additionalProp2", "value")
                .putInt("additionalProp3", 123)
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return a valid evaluation for Double`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getDoubleDetails("double-flag", 1.1)
        val want = FlagEvaluationDetails<Double>(
            flagKey = "double-flag",
            value = 12.34,
            variant = "variantA",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putBoolean("additionalProp1", true)
                .putString("additionalProp2", "value")
                .putInt("additionalProp3", 123)
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return a valid evaluation for String`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getStringDetails("string-flag", "default")
        val want = FlagEvaluationDetails<String>(
            flagKey = "string-flag",
            value = "1234value",
            variant = "variantA",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putBoolean("additionalProp1", true)
                .putString("additionalProp2", "value")
                .putInt("additionalProp3", 123)
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return a valid evaluation for List`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getObjectDetails(
            "array-flag",
            Value.List(MutableList(1) { Value.Integer(1234567890) })
        )

        val want = FlagEvaluationDetails<Value>(
            flagKey = "array-flag",
            value = Value.List(listOf(Value.Integer(1234), Value.Integer(5678))),
            variant = "variantA",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putBoolean("additionalProp1", true)
                .putString("additionalProp2", "value")
                .putInt("additionalProp3", 123)
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return a valid evaluation for Map`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getObjectDetails(
            "object-flag",
            Value.Structure(
                mapOf(
                    "default" to Value.Boolean(true)
                )
            )
        )

        val want = FlagEvaluationDetails<Value>(
            flagKey = "object-flag",
            value = Value.Structure(
                mapOf(
                    "testValue" to Value.Structure(
                        mapOf(
                            "toto" to Value.Integer(1234),
                        )
                    )
                )
            ),
            variant = "variantA",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
            metadata = EvaluationMetadata.builder()
                .putBoolean("additionalProp1", true)
                .putString("additionalProp2", "value")
                .putInt("additionalProp3", 123)
                .build()
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return TypeMismatch Bool`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getBooleanDetails("object-flag", false)
        val want = FlagEvaluationDetails<Boolean>(
            flagKey = "object-flag",
            value = false,
            variant = null,
            reason = "ERROR",
            errorCode = ErrorCode.TYPE_MISMATCH,
            errorMessage = "Type mismatch: expect Boolean - Unsupported type for: {testValue={toto=1234}}",
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return TypeMismatch String`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getStringDetails("object-flag", "default")
        val want = FlagEvaluationDetails<String>(
            flagKey = "object-flag",
            value = "default",
            variant = null,
            reason = "ERROR",
            errorCode = ErrorCode.TYPE_MISMATCH,
            errorMessage = "Type mismatch: expect String - Unsupported type for: {testValue={toto=1234}}",
        )
        assertEquals(want, got)
    }

    @Test
    fun `should return TypeMismatch Double`(): Unit = runBlocking {
        enqueueMockResponse("org.gofeatureflag.openfeature.ofrep/valid_api_response.json", 200)
        val provider = OfrepProvider(OfrepOptions(endpoint = mockWebServer?.url("/").toString()))
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getDoubleDetails("object-flag", 1.233)
        val want = FlagEvaluationDetails<Double>(
            flagKey = "object-flag",
            value = 1.233,
            variant = null,
            reason = "ERROR",
            errorCode = ErrorCode.TYPE_MISMATCH,
            errorMessage = "Type mismatch: expect Double - Unsupported type for: {testValue={toto=1234}}",
        )
        assertEquals(want, got)
    }

    @Test
    fun `should have different result if waiting for next polling interval`(): Unit = runBlocking {
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_api_short_response.json",
            200
        )
        enqueueMockResponse(
            "org.gofeatureflag.openfeature.ofrep/valid_api_response_2.json",
            200
        )

        val provider = OfrepProvider(
            OfrepOptions(
                pollingIntervalInMillis = 100,
                endpoint = mockWebServer?.url("/").toString()
            )
        )
        OpenFeatureAPI.setProviderAndWait(provider, Dispatchers.IO, defaultEvalCtx)
        val client = OpenFeatureAPI.getClient()
        val got = client.getStringDetails("badge-class2", "default")
        val want = FlagEvaluationDetails<String>(
            flagKey = "badge-class2",
            value = "green",
            variant = "nocolor",
            reason = "DEFAULT",
            errorCode = null,
            errorMessage = null,
        )
        assertEquals(want, got)
        Thread.sleep(1000)
        val got2 = client.getStringDetails("badge-class2", "default")
        val want2 = FlagEvaluationDetails<String>(
            flagKey = "badge-class2",
            value = "blue",
            variant = "xxxx",
            reason = "TARGETING_MATCH",
            errorCode = null,
            errorMessage = null,
        )
        assertEquals(want2, got2)
    }

    private fun enqueueMockResponse(
        fileName: String,
        responseCode: Int = 200,
        headers: Headers? = null
    ) {
        val jsonFilePath =
            javaClass.classLoader?.getResource(fileName)?.file
        val jsonString = String(Files.readAllBytes(Paths.get(jsonFilePath)))
        var resp = MockResponse()
            .setBody(jsonString.trimIndent())
            .setResponseCode(responseCode)
        if (headers != null) {
            resp = resp.setHeaders(headers)
        }
        mockWebServer!!.enqueue(resp)
    }
}