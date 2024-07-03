import XCTest
import Foundation
import OpenFeature
@testable import go_feature_flag_provider

class ProviderTests: XCTestCase {

    func testShouldReturnAValidEvaluationResponse() async {
        let options = GoFeatureFlagProviderOptions(endpoint: "http://localhost:1031/")
        let provider = GoFeatureFlagProvider(options: options)

        let ctx = MutableContext(
            targetingKey: "ede04e44-463d-40d1-8fc0-b1d6855578d0",
            structure: MutableStructure(attributes: ["product": Value.string("123")]))
        OpenFeatureAPI.shared.setEvaluationContext(evaluationContext: ctx)



        await OpenFeatureAPI.shared.setProviderAndWait(provider: provider)
        let client = OpenFeatureAPI.shared.getClient()


        let expectation = XCTestExpectation(description: "waiting to be ready")
        _ = OpenFeatureAPI.shared.observe().sink { event in
            switch event {
            case ProviderEvent.ready:
                print("ready")
                expectation.fulfill()
                break
            default:
                break
            }
        }

        await fulfillment(of: [expectation], timeout: 3.0)

        let t = Value.structure(["yoyo":Value.string("yoyo")])
        let flagValue = client.getObjectDetails(key: "complex", defaultValue: t)
        print(flagValue)

        


    }


    /*
     should be in FATAL status if 401 error during initialise
     should be in FATAL status if 403 error during initialise
     should be in ERROR status if 429 error during initialise
     should be in ERROR status if targetingKey is missing
     should be in ERROR status if invalid context
     should be in ERROR status if parse error
     should return a FLAG_NOT_FOUND error if the flag does not exist
     should return EvaluationDetails if the flag exists
     should return ParseError if the API return the error
     should send a configuration changed event, when new flag is send
     should call reconciling handler, when context changed
     should call stale handler, when api is not responding
     should not try to call the API before retry-after header
     
     */
}
