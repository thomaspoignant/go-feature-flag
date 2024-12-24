import dev.openfeature.sdk.EvaluationMetadata
import dev.openfeature.sdk.ProviderEvaluation
import dev.openfeature.sdk.Value
import dev.openfeature.sdk.exceptions.ErrorCode
import dev.openfeature.sdk.exceptions.OpenFeatureError
import kotlin.reflect.KClass

data class OfrepApiResponse(
    val flags: List<FlagDto>? = null,
    val errorCode: ErrorCode?,
    val errorDetails: String?
)

data class FlagDto(
    val value: Any,
    val key: String,
    val reason: String,
    val variant: String,
    val errorCode: ErrorCode?,
    val errorDetails: String?,
    val metadata: Map<String, Any>? = emptyMap()
) {
    fun isError(): Boolean {
        return errorCode != null
    }

    fun <T : Any> toProviderEvaluation(expectedType: KClass<*>): ProviderEvaluation<T> {
        if (!expectedType.isInstance(value)) {
            val isSpecialCase =
                expectedType == Int::class && value is Long && value.toInt().toLong() == value
            if (!isSpecialCase) {
                throw OpenFeatureError.TypeMismatchError("Type mismatch: expect ${expectedType.simpleName} - Unsupported type for: $value")
            }
        }

        if (expectedType == Int::class) {
            val typedValue = (value as Long).toInt()
            return ProviderEvaluation(
                value = typedValue as T,
                reason = reason,
                variant = variant,
                errorCode = errorCode,
                errorMessage = errorDetails,
                metadata = convertMetadata(metadata)
            )
        }

        if (expectedType == Object::class) {
            if (value is List<*>) {
                val typedValue = Value.List(convertList(value as List<Any>))
                return ProviderEvaluation(
                    value = typedValue as T,
                    reason = reason,
                    variant = variant,
                    errorCode = errorCode,
                    errorMessage = errorDetails,
                    metadata = convertMetadata(metadata)
                )
            } else if (value is Map<*, *>) {
                val typedValue = convertObjectToStructure(value)
                return ProviderEvaluation(
                    value = typedValue as T,
                    reason = reason,
                    variant = variant,
                    errorCode = errorCode,
                    errorMessage = errorDetails,
                    metadata = convertMetadata(metadata)
                )
            } else {
                throw IllegalArgumentException("Unsupported type for: $value")
            }
        }



        @Suppress("unchecked_cast")
        return ProviderEvaluation(
            value = value as T,
            reason = reason,
            variant = variant,
            errorCode = errorCode,
            errorMessage = errorDetails,
            metadata = convertMetadata(metadata)
        )
    }

    private fun convertMetadata(inputMap: Map<String, Any>?): EvaluationMetadata {
        //check that inputMap is null or empty
        if (inputMap.isNullOrEmpty()) {
            return EvaluationMetadata.EMPTY
        }

        val metadataBuilder = EvaluationMetadata.builder()
        inputMap.forEach { entry ->
            // switch case on entry.value types
            when (entry.value) {
                is String -> {
                    metadataBuilder.putString(entry.key, entry.value as String)
                }

                is Boolean -> {
                    metadataBuilder.putBoolean(entry.key, entry.value as Boolean)
                }

                is Int -> {
                    metadataBuilder.putInt(entry.key, entry.value as Int)
                }

                is Long -> {
                    metadataBuilder.putInt(entry.key, (entry.value as Long).toInt())
                }

                is Double -> {
                    metadataBuilder.putDouble(entry.key, entry.value as Double)
                }
            }
        }

        return metadataBuilder.build()
    }

    private fun convertList(inputList: List<*>): List<Value> {
        return inputList.map { item ->
            when (item) {
                is String -> Value.String(item)
                is Boolean -> Value.Boolean(item)
                is Long -> Value.Integer(item.toInt())
                is Double -> Value.Double(item)
                is java.util.Date -> Value.Date(item)
                is Map<*, *> -> {
                    @Suppress("unchecked_cast")
                    Value.Structure(item as Map<String, Value>)
                }

                is List<*> -> {
                    @Suppress("unchecked_cast")
                    Value.List(convertList(item as List<Any>))
                }

                else -> throw IllegalArgumentException(
                    "Unsupported type for: $item"
                )
            }
        }
    }

    private fun convertObjectToStructure(obj: Any): Value.Structure {
        if (obj !is Map<*, *>) {
            throw IllegalArgumentException("Object must be a Map")
        }
        val convertedMap = obj.entries.associate { (key, value) ->
            if (key !is String) {
                throw IllegalArgumentException("Map key must be a String")
            }
            key to when (value) {
                is String -> Value.String(value)
                is Boolean -> Value.Boolean(value)
                is Long -> Value.Integer(value.toInt())
                is Double -> Value.Double(value)
                is java.util.Date -> Value.Date(value)
                is Map<*, *> -> convertObjectToStructure(value)
                is List<*> -> Value.List(convertList(value as List<Any>))
                else -> throw IllegalArgumentException("Unsupported type for: $value")
            }
        }
        return Value.Structure(convertedMap)
    }

}
