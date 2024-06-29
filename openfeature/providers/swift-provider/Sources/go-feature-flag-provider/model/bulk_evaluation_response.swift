import OpenFeature

struct EvaluationResponseDTO: Codable {
    var flags: [EvaluationResponseFlagDTO]?
    let errorCode: String?
    let errorDetails: String?
}

struct EvaluationResponseFlagDTO: Codable {
    let value: JSONValue?
    let key: String?
    let reason: String?
    let variant: String?
    let errorCode: String?
    let errorDetails: String?
//    let metadata: [String:Value]?
}

struct OfrepEvaluationResponse{
    let flags: [OfrepEvaluationResponseFlag]
    let errorCode: ErrorCode?
    let errorDetails: String?


    func isError() -> Bool {
        return errorCode != nil
    }

    static func fromEvaluationResponseDTO(dto: EvaluationResponseDTO) -> OfrepEvaluationResponse {
        var flagsConverted: [OfrepEvaluationResponseFlag] = []
        var errCode: ErrorCode? = nil
        let errDetails = dto.errorDetails

        if let flagsDTO = dto.flags {
            for flag in flagsDTO {
                var errorCode: ErrorCode? = nil
                if let erroCodeValue = flag.errorCode {
                    errorCode = convertErrorCode(code: erroCodeValue)
                }

                flagsConverted.append(OfrepEvaluationResponseFlag(
                    value: flag.value,
                    key: flag.key,
                    reason: flag.reason,
                    variant: flag.variant,
                    errorCode: errorCode,
                    errorDetails: flag.errorDetails
//                    metadata: flag.metadata
                ))
            }
        }

        if let errorCode = dto.errorCode {
            errCode = convertErrorCode(code: errorCode)
        }

        return OfrepEvaluationResponse(flags: flagsConverted, errorCode: errCode, errorDetails: errDetails)
    }


    static func convertErrorCode(code: String) -> ErrorCode {
        switch code {
        case "PROVIDER_NOT_READY":
            return ErrorCode.providerNotReady
        case "FLAG_NOT_FOUND":
            return ErrorCode.flagNotFound
        case "PARSE_ERROR":
            return ErrorCode.parseError
        case "TYPE_MISMATCH":
            return ErrorCode.typeMismatch
        case "TARGETING_KEY_MISSING":
            return ErrorCode.targetingKeyMissing
        case "INVALID_CONTEXT":
            return ErrorCode.invalidContext
        default:
            return ErrorCode.general
        }
    }
}

struct OfrepEvaluationResponseFlag {
    let value: JSONValue?
    let key: String?
    let reason: String?
    let variant: String?
    let errorCode: ErrorCode?
    let errorDetails: String?
//    let metadata: [String:Value]?

    func isError() -> Bool {
        return errorCode != nil
    }
}
