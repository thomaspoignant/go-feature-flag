package dev.openfeature.kotlin.contrib.providers.ofrep.integration

import dev.openfeature.kotlin.contrib.providers.ofrep.OfrepProvider
import dev.openfeature.kotlin.contrib.providers.ofrep.bean.OfrepOptions
import dev.openfeature.kotlin.sdk.EvaluationMetadata
import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.Value
import dev.openfeature.kotlin.sdk.exceptions.OpenFeatureError
import kotlinx.coroutines.test.runTest
import kotlin.test.AfterTest
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNull

/**
 * Uses an actual OFREP backend launched as a Docker container, see the `dockerCompose` configuration
 * in `build.gradle.kts`.
 */
class IntegrationTest {
    private lateinit var ofrepProvider: OfrepProvider

    @BeforeTest
    fun setUp() {
        ofrepProvider =
            OfrepProvider(
                OfrepOptions(
                    endpoint = "http://localhost:1031",
                ),
            )
    }

    @AfterTest
    fun tearDown() {
        ofrepProvider.shutdown()
    }

    @Test
    fun `should fetch String flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            val evaluation = ofrepProvider.getStringEvaluation(KEY_STATIC_STRING, defaultValue = DEFAULT_STRING, context = null)

            assertEquals(SIMPLE_VALUE_STRING, evaluation.value)
            assertEquals(DEFAULT_VARIATION, evaluation.variant)
            assertEquals(STATIC_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch Boolean flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            val evaluation = ofrepProvider.getBooleanEvaluation(KEY_STATIC_BOOLEAN, defaultValue = DEFAULT_BOOLEAN, context = null)

            assertEquals(SIMPLE_VALUE_BOOLEAN, evaluation.value)
            assertEquals(DEFAULT_VARIATION, evaluation.variant)
            assertEquals(STATIC_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch Integer flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            val evaluation = ofrepProvider.getIntegerEvaluation(KEY_STATIC_INTEGER, defaultValue = DEFAULT_INTEGER, context = null)

            assertEquals(SIMPLE_VALUE_INTEGER, evaluation.value)
            assertEquals(DEFAULT_VARIATION, evaluation.variant)
            assertEquals(STATIC_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch Double flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            val evaluation = ofrepProvider.getDoubleEvaluation(KEY_STATIC_DOUBLE, defaultValue = DEFAULT_DOUBLE, context = null)

            assertEquals(SIMPLE_VALUE_DOUBLE, evaluation.value)
            assertEquals(DEFAULT_VARIATION, evaluation.variant)
            assertEquals(STATIC_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch Object flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            val evaluation = ofrepProvider.getObjectEvaluation(KEY_STATIC_OBJECT, defaultValue = DEFAULT_OBJECT, context = null)

            assertEquals(SIMPLE_VALUE_OBJECT, evaluation.value)
            assertEquals(DEFAULT_VARIATION, evaluation.variant)
            assertEquals(STATIC_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should return not found for non-existent String flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            assertFailsWith<OpenFeatureError.FlagNotFoundError> {
                ofrepProvider.getStringEvaluation(KEY_NON_EXISTENT, defaultValue = DEFAULT_STRING, context = null)
            }
        }

    @Test
    fun `should return not found for non-existent Boolean flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            assertFailsWith<OpenFeatureError.FlagNotFoundError> {
                ofrepProvider.getBooleanEvaluation(KEY_NON_EXISTENT, defaultValue = DEFAULT_BOOLEAN, context = null)
            }
        }

    @Test
    fun `should return not found for non-existent Integer flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            assertFailsWith<OpenFeatureError.FlagNotFoundError> {
                ofrepProvider.getIntegerEvaluation(KEY_NON_EXISTENT, defaultValue = DEFAULT_INTEGER, context = null)
            }
        }

    @Test
    fun `should return not found for non-existent Object flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            assertFailsWith<OpenFeatureError.FlagNotFoundError> {
                ofrepProvider.getObjectEvaluation(KEY_NON_EXISTENT, defaultValue = DEFAULT_OBJECT, context = null)
            }
        }

    @Test
    fun `should return not found for non-existent Double flag`() =
        runTest {
            ofrepProvider.initialize(SIMPLE_CONTEXT)

            assertFailsWith<OpenFeatureError.FlagNotFoundError> {
                ofrepProvider.getDoubleEvaluation(KEY_NON_EXISTENT, defaultValue = DEFAULT_DOUBLE, context = null)
            }
        }

    @Test
    fun `should fetch targeted String flag`() =
        runTest {
            ofrepProvider.initialize(TARGETED_CONTEXT)

            val evaluation = ofrepProvider.getStringEvaluation(KEY_TARGETED_STRING, defaultValue = DEFAULT_STRING, context = null)

            assertEquals(TARGETED_VALUE_STRING, evaluation.value)
            assertEquals(TARGETED_VARIATION, evaluation.variant)
            assertEquals(TARGETING_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch targeted Boolean flag`() =
        runTest {
            ofrepProvider.initialize(TARGETED_CONTEXT)

            val evaluation = ofrepProvider.getBooleanEvaluation(KEY_TARGETED_BOOLEAN, defaultValue = DEFAULT_BOOLEAN, context = null)

            assertEquals(TARGETED_VALUE_BOOLEAN, evaluation.value)
            assertEquals(TARGETED_VARIATION, evaluation.variant)
            assertEquals(TARGETING_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch targeted Integer flag`() =
        runTest {
            ofrepProvider.initialize(TARGETED_CONTEXT)

            val evaluation = ofrepProvider.getIntegerEvaluation(KEY_TARGETED_INTEGER, defaultValue = DEFAULT_INTEGER, context = null)

            assertEquals(TARGETED_VALUE_INTEGER, evaluation.value)
            assertEquals(TARGETED_VARIATION, evaluation.variant)
            assertEquals(TARGETING_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch targeted Double flag`() =
        runTest {
            ofrepProvider.initialize(TARGETED_CONTEXT)

            val evaluation = ofrepProvider.getDoubleEvaluation(KEY_TARGETED_DOUBLE, defaultValue = DEFAULT_DOUBLE, context = null)

            assertEquals(TARGETED_VALUE_DOUBLE, evaluation.value)
            assertEquals(TARGETED_VARIATION, evaluation.variant)
            assertEquals(TARGETING_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    @Test
    fun `should fetch targeted Object flag`() =
        runTest {
            ofrepProvider.initialize(TARGETED_CONTEXT)

            val evaluation = ofrepProvider.getObjectEvaluation(KEY_TARGETED_OBJECT, defaultValue = DEFAULT_OBJECT, context = null)

            assertEquals(TARGETED_VALUE_OBJECT, evaluation.value)
            assertEquals(TARGETED_VARIATION, evaluation.variant)
            assertEquals(TARGETING_REASON, evaluation.reason)
            assertNull(evaluation.errorCode)
            assertNull(evaluation.errorMessage)
            assertEquals(EvaluationMetadata.EMPTY, evaluation.metadata)
        }

    companion object {
        private const val TARGETING_KEY = "123"
        private val SIMPLE_CONTEXT = ImmutableContext(TARGETING_KEY)
        private val TARGETED_CONTEXT =
            ImmutableContext(TARGETING_KEY, mapOf("targeted" to Value.Boolean(true)))

        private const val KEY_STATIC_STRING = "static-string-flag"
        private const val KEY_STATIC_BOOLEAN = "static-boolean-flag"
        private const val KEY_STATIC_INTEGER = "static-integer-flag"
        private const val KEY_STATIC_DOUBLE = "static-double-flag"
        private const val KEY_STATIC_OBJECT = "static-object-flag"

        private const val KEY_NON_EXISTENT = "non-existent-flag"

        private const val KEY_TARGETED_STRING = "targeted-string-flag"
        private const val KEY_TARGETED_BOOLEAN = "targeted-boolean-flag"
        private const val KEY_TARGETED_INTEGER = "targeted-integer-flag"
        private const val KEY_TARGETED_DOUBLE = "targeted-double-flag"
        private const val KEY_TARGETED_OBJECT = "targeted-object-flag"

        private const val SIMPLE_VALUE_STRING = "foo"
        private const val SIMPLE_VALUE_BOOLEAN = true
        private const val SIMPLE_VALUE_INTEGER = 123
        private const val SIMPLE_VALUE_DOUBLE = 3.14
        private val SIMPLE_VALUE_OBJECT =
            Value.Structure(
                mapOf(
                    "string" to Value.String("foo"),
                    "boolean" to Value.Boolean(true),
                    "number" to Value.Double(2.78),
                    "list" to
                        Value.List(
                            listOf(
                                Value.String("one"),
                                Value.String("two"),
                                Value.String("three"),
                            ),
                        ),
                ),
            )

        private const val TARGETED_VALUE_STRING = "bar"
        private const val TARGETED_VALUE_BOOLEAN = false
        private const val TARGETED_VALUE_INTEGER = 999
        private const val TARGETED_VALUE_DOUBLE = 999.9
        private val TARGETED_VALUE_OBJECT =
            Value.Structure(
                mapOf(
                    "special" to Value.Boolean(true),
                ),
            )

        private const val DEFAULT_STRING = "default"
        private const val DEFAULT_BOOLEAN = false
        private const val DEFAULT_INTEGER = -1
        private const val DEFAULT_DOUBLE = -1.0
        private val DEFAULT_OBJECT = Value.Null

        private const val DEFAULT_VARIATION = "default_variation"
        private const val TARGETED_VARIATION = "targeted_variation"

        private const val STATIC_REASON = "STATIC"
        private const val TARGETING_REASON = "TARGETING_MATCH"
    }
}
