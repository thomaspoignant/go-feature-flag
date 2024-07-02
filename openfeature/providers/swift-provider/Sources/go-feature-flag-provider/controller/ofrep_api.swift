import Foundation
import OpenFeature

class OfrepAPI {
    private let networkingService: NetworkingService
    private var etag: String = ""
    private let options: GoFeatureFlagProviderOptions
    
    init(networkingService: NetworkingService, options: GoFeatureFlagProviderOptions) {
        self.networkingService = networkingService
        self.options = options
    }
    
    func postBulkEvaluateFlags(context: EvaluationContext?) async throws -> (OfrepEvaluationResponse, HTTPURLResponse) {
        guard let context = context else {
            throw OpenFeatureError.invalidContextError
        }
        try validateContext(context: context)
        
        guard let url = URL(string: options.endpoint) else {
            throw InvalidOptions.invalidEndpoint(message: "endpoint [" + options.endpoint + "] is not valid")
        }
        let ofrepURL = url.appendingPathComponent("ofrep/v1/evaluate/flags")
        var request = URLRequest(url: ofrepURL)
        request.httpMethod = "POST"
        request.httpBody = try EvaluationRequest.convertEvaluationContext(context: context).asJSONData()
        request.setValue(
            "application/json",
            forHTTPHeaderField: "Content-Type"
        )
        
        if etag != "" {
            request.setValue(etag, forHTTPHeaderField: "If-None-Match")
        }
        
        let (data, response) = try await networkingService.doRequest(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw OfrepError.httpResponseCastError
        }
        
        if httpResponse.statusCode == 401 {
            throw OfrepError.apiUnauthorizedError(response: httpResponse)
        }
        if httpResponse.statusCode == 403 {
            throw OfrepError.forbiddenError(response: httpResponse)
        }
        if httpResponse.statusCode == 429 {
            throw OfrepError.apiTooManyRequestsError(response: httpResponse)
        }
        if httpResponse.statusCode > 400 {
            throw OfrepError.unexpectedResponseError(response: httpResponse)
        }
        if httpResponse.statusCode == 304 {
            return (OfrepEvaluationResponse(flags: [], errorCode: nil, errorDetails: nil), httpResponse)
        }
        
        // Store ETag to use it in the next request
        if let etagHeaderValue = httpResponse.value(forHTTPHeaderField: "ETag") {
            if etagHeaderValue != "" && httpResponse.statusCode == 200 {
                etag = etagHeaderValue
            }
        }
        
        do {
            let dto = try JSONDecoder().decode(EvaluationResponseDTO.self, from: data)
            let evaluationResponse = OfrepEvaluationResponse.fromEvaluationResponseDTO(dto: dto)
            return (evaluationResponse, httpResponse)
        } catch {
            throw OfrepError.unmarshallError(error: error)
        }
    }
    
    private func validateContext(context: EvaluationContext) throws {
        let targetingKey = context.getTargetingKey()
        if targetingKey.isEmpty {
            throw OpenFeatureError.targetingKeyMissingError
        }
    }
}
