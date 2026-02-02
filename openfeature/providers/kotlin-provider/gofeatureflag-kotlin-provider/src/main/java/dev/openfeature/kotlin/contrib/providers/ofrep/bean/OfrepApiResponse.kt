@file:OptIn(ExperimentalTime::class)

package dev.openfeature.kotlin.contrib.providers.ofrep.bean

import dev.openfeature.kotlin.contrib.providers.ofrep.serialization.EvaluationMetadataSerializer
import dev.openfeature.kotlin.contrib.providers.ofrep.serialization.ValueSerializer
import dev.openfeature.kotlin.sdk.EvaluationMetadata
import dev.openfeature.kotlin.sdk.ProviderEvaluation
import dev.openfeature.kotlin.sdk.Value
import dev.openfeature.kotlin.sdk.exceptions.ErrorCode
import dev.openfeature.kotlin.sdk.exceptions.OpenFeatureError
import kotlinx.serialization.Serializable
import kotlin.time.ExperimentalTime
import kotlin.time.Instant

@Serializable
internal data class OfrepApiResponse(
    val flags: List<FlagDto>? = null,
    val errorCode: ErrorCode? = null,
    val errorDetails: String? = null,
)

@Serializable
internal data class FlagDto(
    @Serializable(with = ValueSerializer::class)
    val value: Value? = null,
    val key: String,
    val reason: String? = null,
    val variant: String? = null,
    val errorCode: ErrorCode? = null,
    val errorDetails: String? = null,
    @Serializable(with = EvaluationMetadataSerializer::class)
    val metadata: EvaluationMetadata = EvaluationMetadata.EMPTY,
) {
    fun isError(): Boolean = errorCode != null
}

@OptIn(ExperimentalTime::class)
inline fun <reified T> Value.toPrimitive(): T {
    val value: T? =
        when (T::class) {
            Boolean::class -> asBoolean() as T?
            String::class -> asString() as T?
            Int::class -> asInteger() as T?
            Double::class ->
                // doubles might have been serialized as integers
                (asDouble() ?: asInteger()?.toDouble()) as T?

            Instant::class ->
                // Instants might have been serialized as a string
                (asInstant() ?: asString()?.let { Instant.parse(it) }) as T?
            else -> error("toPrimitive not implemented for ${T::class}")
        }
    return value ?: throw OpenFeatureError.TypeMismatchError(
        "Type mismatch: expect ${T::class.simpleName} - Unsupported type for: $this",
    )
}

internal inline fun <reified T> FlagDto.toProviderEvaluation(default: T): ProviderEvaluation<T> {
    val convertedValue: T? = if (T::class == Value::class) value as T else value?.toPrimitive()
    return ProviderEvaluation(
        value = convertedValue ?: default,
        reason = reason,
        variant = variant,
        errorCode = errorCode,
        errorMessage = errorDetails,
        metadata = metadata,
    )
}
