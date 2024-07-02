import XCTest
import Foundation
import OpenFeature
@testable import go_feature_flag_provider

class OfrepApiTests: XCTestCase {
    var defaultEvaluationContext: MutableContext!
    var options = GoFeatureFlagProviderOptions(endpoint: "http://localhost:1031/")
    override func setUp() {
        super.setUp()
        defaultEvaluationContext = MutableContext()
        defaultEvaluationContext.setTargetingKey(targetingKey: "ede04e44-463d-40d1-8fc0-b1d6855578d0")
        defaultEvaluationContext.add(key: "email", value: Value.string("john.doe@gofeatureflag.org"))
        defaultEvaluationContext.add(key: "name", value: Value.string("John Doe"))
        defaultEvaluationContext.add(key: "age", value: Value.integer(2))
        defaultEvaluationContext.add(key: "category", value: Value.double(2.2))
        defaultEvaluationContext.add(key: "struct", value: Value.structure(["test" : Value.string("test")]))
        defaultEvaluationContext.add(key: "list", value: Value.list([Value.string("test1"), Value.string("test2")]))
    }
    override func tearDown() {
        defaultEvaluationContext = nil
        super.tearDown()
    }

    func testShouldReturnAValidEvaluationResponse() async throws{
        let mockResponse = """
            {
              "flags": [
                {
                  "key": "badge-class",
                  "value": "green",
                  "reason": "DEFAULT",
                  "variant": "nocolor"
                },
                {
                  "key": "hide-logo",
                  "value": false,
                  "reason": "STATIC",
                  "variant": "var_false"
                },
                {
                  "key": "title-flag",
                  "value": "GO Feature Flag",
                  "reason": "DEFAULT",
                  "variant": "default_title",
                  "metadata": {
                    "description": "This flag controls the title of the feature flag",
                    "title": "Feature Flag Title"
                  }
                }
              ]
            }
        """
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8))
        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)

        do {
            let (evalResp, response) = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTAssertFalse(evalResp.isError())
            XCTAssertEqual(response.statusCode, 200, "wrong http status")
            XCTAssertEqual(evalResp.flags.count, 3)
            XCTAssertEqual(evalResp.errorCode, nil)
            XCTAssertEqual(evalResp.errorDetails, nil)

            XCTAssertEqual(evalResp.flags[0].key, "badge-class")
            XCTAssertEqual(evalResp.flags[0].value, JSONValue.string("green"))
            XCTAssertEqual(evalResp.flags[0].reason, "DEFAULT")
            XCTAssertEqual(evalResp.flags[0].variant, "nocolor")
            XCTAssertEqual(evalResp.flags[0].errorCode, nil)
            XCTAssertEqual(evalResp.flags[0].errorDetails, nil)

            XCTAssertEqual(evalResp.flags[1].key, "hide-logo")
            XCTAssertEqual(evalResp.flags[1].value, JSONValue.bool(false))
            XCTAssertEqual(evalResp.flags[1].reason, "STATIC")
            XCTAssertEqual(evalResp.flags[1].variant, "var_false")
            XCTAssertEqual(evalResp.flags[1].errorCode, nil)
            XCTAssertEqual(evalResp.flags[1].errorDetails, nil)

            XCTAssertEqual(evalResp.flags[2].key, "title-flag")
            XCTAssertEqual(evalResp.flags[2].value, JSONValue.string("GO Feature Flag"))
            XCTAssertEqual(evalResp.flags[2].reason, "DEFAULT")
            XCTAssertEqual(evalResp.flags[2].variant, "default_title")
            XCTAssertEqual(evalResp.flags[2].errorCode, nil)
            XCTAssertEqual(evalResp.flags[2].errorDetails, nil)
//            XCTAssertEqual(evalResp.flags[2].metadata?["description"], Value.string("This flag controls the title of the feature flag"))
//            XCTAssertEqual(evalResp.flags[2].metadata?["title"], Value.string("Feature Flag Title"))
            XCTAssertEqual(response.value(forHTTPHeaderField: "ETag"), "33a64df551425fcc55e4d42a148795d9f25f89d4")
        } catch {
            XCTFail("exception thrown when doing the evaluation: \(error)")
        }
    }

    func testShouldThrowAnUnauthorizedError() async throws{
        let mockResponse = "{}"
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 401)
        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTFail("Should throw an exception")
        } catch let error as OfrepError {
            switch error {
            case .apiUnauthorizedError(let response):
                XCTAssertNotNil(response)
                break
            default:
                XCTFail("Caught an unexpected OFREP error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowAForbiddenError() async throws{
        let mockResponse = "{}"
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 403)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTFail("Should throw an exception")
        } catch let error as OfrepError {
            switch error {
            case .forbiddenError(let response):
                XCTAssertNotNil(response)
                break
            default:
                XCTFail("Caught an unexpected OFREP error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowTooManyRequest() async throws{
        let mockResponse = "{}"
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 429)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTFail("Should throw an exception")
        } catch let error as OfrepError {
            switch error {
            case .apiTooManyRequestsError(let response):
                XCTAssertNotNil(response)
                XCTAssertEqual(response.allHeaderFields["Retry-Later"] as! String, "120")
                break
            default:
                XCTFail("Caught an unexpected OFREP error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowUnexpectedError() async throws{
        let mockResponse = "{}"
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 500)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTFail("Should throw an exception")
        } catch let error as OfrepError {
            switch error {
            case .unexpectedResponseError(let response):
                XCTAssertNotNil(response)
                break
            default:
                XCTFail("Caught an unexpected OFREP error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldReturnaEvaluationResponseInError() async throws{
        let mockResponse = """
{"errorCode": "INVALID_CONTEXT", "errorDetails":"explanation of the error"}
"""
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 400)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            let (evalResp, httpResp) = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTAssertTrue(evalResp.isError())
            XCTAssertEqual(evalResp.errorCode, ErrorCode.invalidContext)
            XCTAssertEqual(evalResp.errorDetails, "explanation of the error")
            XCTAssertEqual(httpResp.statusCode, 400)
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldReturnaEvaluationResponseIfWeReceiveA304() async throws{
        let mockResponse = ""
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 304)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            let (evalResp, httpResp) = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTAssertFalse(evalResp.isError())
            XCTAssertEqual(httpResp.statusCode, 304)
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowInvalidContextWithNilContext() async throws{
        let mockResponse = "{}"
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 500)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: nil)
            XCTFail("Should throw an exception")
        } catch let error as OpenFeatureError {
            switch error {
            case .invalidContextError:
                break
            default:
                XCTFail("Caught an unexpected OpenFeatureError error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowTargetingKeyMissingErrorWithNoTargetingKey() async throws{
        let mockResponse = "{}"
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 200)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            let ctx = MutableContext()
            ctx.add(key: "email", value: Value.string("john.doe@gofeatureflag.org"))
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: ctx)
            XCTFail("Should throw an exception")
        } catch let error as OpenFeatureError {
            switch error {
            case .targetingKeyMissingError:
                break
            default:
                XCTFail("Caught an unexpected OpenFeatureError error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowUnmarshallErrorWithInvalidJson() async throws{
        let mockResponse = """
            {
              "flags": [
                {
                  "key": "badge-class",
                  "value": "",
                  "reason": "DEFAULT",
                  "variant": "nocolor"
                }
                }
            }
        """
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 200)

        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTFail("Should throw an exception")
        } catch let error as OfrepError {
            switch error {
            case .unmarshallError:
                break
            default:
                XCTFail("Caught an unexpected OpenFeatureError error type: \(error)")
            }
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldThrowWithInvalidOptions() async throws{
        let mockResponse = """
        {
            "flags": [
                {
                    "key": "badge-class",
                    "value": "",
                    "reason": "DEFAULT",
                    "variant": "nocolor"
                }
            ]
        }
        """
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 200)
        let testOptions = GoFeatureFlagProviderOptions(endpoint: "")
        let ofrepAPI = OfrepAPI(networkingService: mockService, options:testOptions)
        do {
            _ = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTFail("Should throw an exception")
        } catch let error as InvalidOptions {
            return
        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

    func testShouldETagShouldNotMatch() async throws{
        let mockResponse = """
        {
            "flags": [
                {
                    "key": "badge-class",
                    "value": "green",
                    "reason": "DEFAULT",
                    "variant": "nocolor"
                }
            ]
        }
        """
        let mockService = MockNetworkingService(mockData:  mockResponse.data(using: .utf8), mockStatus: 200)
        let ofrepAPI = OfrepAPI(networkingService: mockService, options:options)
        do {
            let (_, httpResp) = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTAssertNotNil(httpResp.value(forHTTPHeaderField: "ETag"))
            
            let (_, httpResp2) = try await ofrepAPI.postBulkEvaluateFlags(context: defaultEvaluationContext)
            XCTAssertEqual(httpResp2.statusCode, 304)

        } catch {
            XCTFail("Caught an unexpected error type: \(error)")
        }
    }

}
