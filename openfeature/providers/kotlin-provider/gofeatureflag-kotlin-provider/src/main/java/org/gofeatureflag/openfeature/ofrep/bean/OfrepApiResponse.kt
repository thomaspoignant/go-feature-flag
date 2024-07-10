import dev.openfeature.sdk.exceptions.ErrorCode

data class OfrepApiResponse(
    val flags: List<FlagDto>,
    val errorCode: ErrorCode?,
    val errorDetails: String?
)

data class FlagDto(
    val value: Any,
    val key: String,
    val reason: String,
    val variant: String,
    val errorCode: ErrorCode?,
    val errorDetails: String?
)
