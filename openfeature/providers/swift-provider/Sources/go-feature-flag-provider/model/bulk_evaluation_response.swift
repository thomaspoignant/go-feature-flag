import OpenFeature

struct EvaluationResponse: Codable {
    var flags: [EvaluationResponseFlag]
}

struct EvaluationResponseFlag: Codable {
    let value: JSONValue?
    let key: String?
    let reason: String?
    let variant: String?
    let errorCode: String?
    let errorDetails: String?
}
