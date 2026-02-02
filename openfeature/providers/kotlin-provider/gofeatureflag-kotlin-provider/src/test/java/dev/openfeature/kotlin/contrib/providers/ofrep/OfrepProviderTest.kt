@file:OptIn(ExperimentalCoroutinesApi::class)

package dev.openfeature.kotlin.contrib.providers.ofrep

import INVALID_CONTEXT_PAYLOAD
import PARSE_ERROR_PAYLOAD
import VALID_1_FLAG_IN_PARSE_ERROR_PAYLOAD
import VALID_API_RESPONSE2_PAYLOAD
import VALID_API_RESPONSE_PAYLOAD
import VALID_API_SHORT_RESPONSE_PAYLOAD
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepOptions
import dev.openfeature.kotlin.sdk.Client
import dev.openfeature.kotlin.sdk.EvaluationContext
import dev.openfeature.kotlin.sdk.EvaluationMetadata
import dev.openfeature.kotlin.sdk.FeatureProvider
import dev.openfeature.kotlin.sdk.FlagEvaluationDetails
import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.OpenFeatureAPI
import dev.openfeature.kotlin.sdk.Value
import dev.openfeature.kotlin.sdk.events.OpenFeatureProviderEvents
import dev.openfeature.kotlin.sdk.exceptions.ErrorCode
import dev.openfeature.kotlin.sdk.exceptions.OpenFeatureError
import io.ktor.client.engine.mock.MockEngine
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.filterIsInstance
import kotlinx.coroutines.flow.take
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import kotlin.test.AfterTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.uuid.ExperimentalUuidApi
import kotlin.uuid.Uuid

private val POLLING_INTERVAL = 10.minutes

private fun TestScope.createOfrepProvider(mockEngine: MockEngine) =
    OfrepProvider(
        OfrepOptions(
            endpoint = FAKE_ENDPOINT,
            httpClientEngine = mockEngine,
            pollingInterval = POLLING_INTERVAL,
            pollingDispatcher = StandardTestDispatcher(testScheduler),
        ),
    )

private suspend fun withClient(
    provider: FeatureProvider,
    initialContext: EvaluationContext,
    body: (client: Client) -> Unit,
) {
    OpenFeatureAPI.setProviderAndWait(provider, initialContext, StandardTestDispatcher())
    try {
        val client = OpenFeatureAPI.getClient()
        body(client)
    } finally {
        OpenFeatureAPI.shutdown()
    }
}

@OptIn(ExperimentalUuidApi::class)
class OfrepProviderTest {
    private val defaultEvalCtx: EvaluationContext =
        ImmutableContext(targetingKey = Uuid.random().toHexString())

    @AfterTest
    fun after() =
        runTest {
            OpenFeatureAPI.shutdown()
        }

    @Test
    fun `should have a provider metadata`() {
        val provider = OfrepProvider(OfrepOptions(endpoint = "http://localhost:1031"))
        assertEquals("OFREP Provider", provider.metadata.name)
    }

    @Test
    fun `should be in Fatal status if 401 error during initialise`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    VALID_API_RESPONSE_PAYLOAD,
                    status = HttpStatusCode.fromValue(401),
                )

            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                    }
            }
            runCurrent()
            withClient(provider, defaultEvalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
            }
        }

    @Test
    fun `should be in Fatal status if 403 error during initialise`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    VALID_API_RESPONSE_PAYLOAD,
                    status = HttpStatusCode.fromValue(403),
                )

            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                    }
            }
            runCurrent()
            withClient(provider, defaultEvalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
            }
        }

    @Test
    fun `should be in Error status if 429 error during initialise`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    VALID_API_RESPONSE_PAYLOAD,
                    status = HttpStatusCode.fromValue(429),
                    additionalHeaders = headersOf(HttpHeaders.RetryAfter, "3"),
                )

            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false
            var exceptionReceived: Throwable? = null

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                        exceptionReceived = it.error
                    }
            }
            runCurrent()
            withClient(provider, defaultEvalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
                assertIs<OpenFeatureError.GeneralError>(exceptionReceived, "The exception is not of type GeneralError")
                assertEquals(
                    "Rate limited",
                    (exceptionReceived as OpenFeatureError.GeneralError).message,
                    "The exception's message is not correct",
                )
            }
        }

    @Test
    fun `should be in Error status if error targeting key is empty`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)

            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false
            var exceptionReceived: Throwable? = null

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                        exceptionReceived = it.error
                    }
            }
            runCurrent()
            val evalCtx = ImmutableContext(targetingKey = "")
            withClient(provider, evalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
                assertIs<OpenFeatureError.TargetingKeyMissingError>(
                    exceptionReceived,
                    "The exception is not of type TargetingKeyMissingError",
                )
            }
        }

    @Test
    fun `should be in Error status if error targeting key is missing`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)

            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false
            var exceptionReceived: Throwable? = null

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                        exceptionReceived = it.error
                    }
            }
            runCurrent()
            val evalCtx = ImmutableContext()
            withClient(provider, evalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
                assertIs<OpenFeatureError.TargetingKeyMissingError>(
                    exceptionReceived,
                    "The exception is not of type TargetingKeyMissingError",
                )
            }
        }

    @Test
    fun `should be in error status if error invalid context`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    INVALID_CONTEXT_PAYLOAD,
                    status = HttpStatusCode.fromValue(400),
                )
            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false
            var exceptionReceived: Throwable? = null

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                        exceptionReceived = it.error
                    }
            }
            runCurrent()
            withClient(provider, defaultEvalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
                assertIs<OpenFeatureError.InvalidContextError>(exceptionReceived, "The exception is not of type InvalidContextError")
            }
        }

    @Test
    fun `should be in error status if error parse error`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    PARSE_ERROR_PAYLOAD,
                    status = HttpStatusCode.fromValue(400),
                )

            val provider = createOfrepProvider(mockEngine)
            var providerErrorReceived = false
            var exceptionReceived: Throwable? = null

            launch {
                provider
                    .observe()
                    .filterIsInstance<OpenFeatureProviderEvents.ProviderError>()
                    .take(1)
                    .collect {
                        providerErrorReceived = true
                        exceptionReceived = it.error
                    }
            }
            runCurrent()
            withClient(provider, defaultEvalCtx) { client ->
                runCurrent()
                assertTrue(providerErrorReceived, "ProviderError event was not received")
                assertIs<OpenFeatureError.ParseError>(exceptionReceived, "The exception is not of type ParseError")
            }
        }

    @Test
    fun `should return a flag not found error if the flag does not exist`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getBooleanDetails("non-existent-flag", false)
                val want =
                    FlagEvaluationDetails<Boolean>(
                        flagKey = "non-existent-flag",
                        value = false,
                        variant = null,
                        reason = "ERROR",
                        errorCode = ErrorCode.FLAG_NOT_FOUND,
                        errorMessage = "Could not find flag named: non-existent-flag",
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return evaluation details if the flag exists`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    VALID_API_SHORT_RESPONSE_PAYLOAD,
                )
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getStringDetails("title-flag", "default")
                val want =
                    FlagEvaluationDetails<String>(
                        flagKey = "title-flag",
                        value = "GO Feature Flag",
                        variant = "default_title",
                        reason = "DEFAULT",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putString("description", "This flag controls the title of the feature flag")
                                .putString("title", "Feature Flag Title")
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return parse error if the API returns the error`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    VALID_1_FLAG_IN_PARSE_ERROR_PAYLOAD,
                )
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getStringDetails("my-other-flag", "default")
                val want =
                    FlagEvaluationDetails<String>(
                        flagKey = "my-other-flag",
                        value = "default",
                        variant = null,
                        reason = "ERROR",
                        errorCode = ErrorCode.PARSE_ERROR,
                        errorMessage = "Error details about PARSE_ERROR",
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should send a context changed event if context has changed`() =
        runTest {
            val mockEngine =
                mockEngineWithTwoResponses(
                    firstContent = VALID_API_RESPONSE_PAYLOAD,
                    secondContent = VALID_API_RESPONSE2_PAYLOAD,
                )
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->

                // TODO: should change when we have a way to observe context changes event
                //       check issue https://github.com/open-feature/kotlin-sdk/issues/107
                var providerStaleEventReceived = false
                var providerReadyEventReceived = false

                launch {
                    provider
                        .observe()
                        .filterIsInstance<OpenFeatureProviderEvents.ProviderStale>()
                        .take(1)
                        .collect {
                            providerStaleEventReceived = true
                        }
                    provider
                        .observe()
                        .filterIsInstance<OpenFeatureProviderEvents.ProviderReady>()
                        .take(1)
                        .collect {
                            providerReadyEventReceived = true
                        }
                }
                runCurrent()
                val newEvalCtx = ImmutableContext(targetingKey = Uuid.random().toHexString())
                OpenFeatureAPI.setEvaluationContext(
                    newEvalCtx,
                    dispatcher = StandardTestDispatcher(testScheduler),
                )
                runCurrent()
                assertTrue(providerStaleEventReceived, "ProviderStale event was not received")
                assertTrue(providerReadyEventReceived, "ProviderReady event was not received")
            }
        }

    @Test
    fun `should not try to call the API before Retry-After header`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(
                    status = HttpStatusCode.fromValue(429),
                    additionalHeaders = headersOf("Retry-After", "3"),
                )
            val provider =
                createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                client.getStringDetails("my-other-flag", "default")
                client.getStringDetails("my-other-flag", "default")
                advanceTimeBy(POLLING_INTERVAL)
                assertEquals(1, mockEngine.requestHistory.size)
            }
        }

    @Test
    fun `should return a valid evaluation for Boolean`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getBooleanDetails("bool-flag", false)
                val want =
                    FlagEvaluationDetails<Boolean>(
                        flagKey = "bool-flag",
                        value = true,
                        variant = "variantA",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putBoolean("additionalProp1", true)
                                .putString("additionalProp2", "value")
                                .putInt("additionalProp3", 123)
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return a valid evaluation for Int`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getIntegerDetails("int-flag", 1)
                val want =
                    FlagEvaluationDetails<Int>(
                        flagKey = "int-flag",
                        value = 1234,
                        variant = "variantA",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putBoolean("additionalProp1", true)
                                .putString("additionalProp2", "value")
                                .putInt("additionalProp3", 123)
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return a valid evaluation for Double`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getDoubleDetails("double-flag", 1.1)
                val want =
                    FlagEvaluationDetails<Double>(
                        flagKey = "double-flag",
                        value = 12.34,
                        variant = "variantA",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putBoolean("additionalProp1", true)
                                .putString("additionalProp2", "value")
                                .putInt("additionalProp3", 123)
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return a valid evaluation for String`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getStringDetails("string-flag", "default")
                val want =
                    FlagEvaluationDetails<String>(
                        flagKey = "string-flag",
                        value = "1234value",
                        variant = "variantA",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putBoolean("additionalProp1", true)
                                .putString("additionalProp2", "value")
                                .putInt("additionalProp3", 123)
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return a valid evaluation for List`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got =
                    client.getObjectDetails(
                        "array-flag",
                        Value.List(MutableList(1) { Value.Integer(1234567890) }),
                    )

                val want =
                    FlagEvaluationDetails<Value>(
                        flagKey = "array-flag",
                        value = Value.List(listOf(Value.Integer(1234), Value.Integer(5678))),
                        variant = "variantA",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putBoolean("additionalProp1", true)
                                .putString("additionalProp2", "value")
                                .putInt("additionalProp3", 123)
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return a valid evaluation for Map`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got =
                    client.getObjectDetails(
                        "object-flag",
                        Value.Structure(
                            mapOf(
                                "default" to Value.Boolean(true),
                            ),
                        ),
                    )

                val want =
                    FlagEvaluationDetails<Value>(
                        flagKey = "object-flag",
                        value =
                            Value.Structure(
                                mapOf(
                                    "testValue" to
                                        Value.Structure(
                                            mapOf(
                                                "toto" to Value.Integer(1234),
                                            ),
                                        ),
                                ),
                            ),
                        variant = "variantA",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                        metadata =
                            EvaluationMetadata
                                .builder()
                                .putBoolean("additionalProp1", true)
                                .putString("additionalProp2", "value")
                                .putInt("additionalProp3", 123)
                                .build(),
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return TypeMismatch Bool`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getBooleanDetails("object-flag", false)
                val want =
                    FlagEvaluationDetails<Boolean>(
                        flagKey = "object-flag",
                        value = false,
                        variant = null,
                        reason = "ERROR",
                        errorCode = ErrorCode.TYPE_MISMATCH,
                        errorMessage =
                            "Type mismatch: expect Boolean - Unsupported type for: " +
                                "Structure(structure={testValue=Structure(structure={toto=Integer(integer=1234)})})",
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return TypeMismatch String`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getStringDetails("object-flag", "default")
                val want =
                    FlagEvaluationDetails<String>(
                        flagKey = "object-flag",
                        value = "default",
                        variant = null,
                        reason = "ERROR",
                        errorCode = ErrorCode.TYPE_MISMATCH,
                        errorMessage =
                            "Type mismatch: expect String - Unsupported type for: " +
                                "Structure(structure={testValue=Structure(structure={toto=Integer(integer=1234)})})",
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should return TypeMismatch Double`() =
        runTest {
            val mockEngine =
                mockEngineWithOneResponse(VALID_API_RESPONSE_PAYLOAD)
            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                val got = client.getDoubleDetails("object-flag", 1.233)
                val want =
                    FlagEvaluationDetails<Double>(
                        flagKey = "object-flag",
                        value = 1.233,
                        variant = null,
                        reason = "ERROR",
                        errorCode = ErrorCode.TYPE_MISMATCH,
                        errorMessage =
                            "Type mismatch: expect Double - Unsupported type for: " +
                                "Structure(structure={testValue=Structure(structure={toto=Integer(integer=1234)})})",
                    )
                assertEquals(want, got)
            }
        }

    @Test
    fun `should have different result if waiting for next polling interval`() =
        runTest(
            // TODO: remove
            timeout = 10.minutes,
        ) {
            val mockEngine =
                mockEngineWithTwoResponses(
                    firstContent = VALID_API_SHORT_RESPONSE_PAYLOAD,
                    secondContent = VALID_API_RESPONSE2_PAYLOAD,
                )

            val provider = createOfrepProvider(mockEngine)
            withClient(provider, defaultEvalCtx) { client ->
                runCurrent()
                val got = client.getStringDetails("badge-class2", "default")
                val want =
                    FlagEvaluationDetails<String>(
                        flagKey = "badge-class2",
                        value = "green",
                        variant = "nocolor",
                        reason = "DEFAULT",
                        errorCode = null,
                        errorMessage = null,
                    )
                assertEquals(want, got)
                advanceTimeBy(POLLING_INTERVAL + 1.milliseconds)
                val got2 = client.getStringDetails("badge-class2", "default")
                val want2 =
                    FlagEvaluationDetails<String>(
                        flagKey = "badge-class2",
                        value = "blue",
                        variant = "xxxx",
                        reason = "TARGETING_MATCH",
                        errorCode = null,
                        errorMessage = null,
                    )
                assertEquals(want2, got2)
            }
        }
}
