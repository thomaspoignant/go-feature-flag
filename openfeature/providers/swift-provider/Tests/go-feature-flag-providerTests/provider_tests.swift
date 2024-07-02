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

}
