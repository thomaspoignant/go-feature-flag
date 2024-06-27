import XCTest
import Foundation
import OpenFeature
@testable import go_feature_flag_provider

class OfrepApiTests: XCTestCase {
    func testValid() async throws{
        let mockResponse = EvaluationResponse(flags: [
            EvaluationResponseFlag(value: nil, key: nil, reason: "ERROR", variant: nil, errorCode: "FLAG_NOT_FOUND", errorDetails: nil)
        ])
        
        let expectedData = toJsonString(mockResponse: mockResponse).data(using: .utf8)!
        let mockService = MockNetworkingService(mockData: expectedData)
        let options = GoFeatureFlagProviderOptions(endpoint: "http://localhost:1031/")
        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        
        let evalCtx = MutableContext()
        evalCtx.setTargetingKey(targetingKey: "1")
        evalCtx.add(key: "email", value: Value.string("john.doe@gofeatureflag.org"))
        
        do {
            let (evalResp, response) = try await ofrepAPI.bulkEvaluation(context: evalCtx)
            
            
        } catch {
            print(error)
        }
    }
    
    
    
    func toJsonString(mockResponse: EvaluationResponse)->String {
        do{
            let jsonData = try JSONEncoder().encode(mockResponse)
            guard let jsonString = String(data: jsonData, encoding: .utf8) else {
                XCTFail("impossible to convert EvaluationResponse")
                return ""
            }
            return jsonString
        } catch {
            XCTFail("error while encoding EvaluationResponse")
            return ""
        }
    }
    
}
    
    
