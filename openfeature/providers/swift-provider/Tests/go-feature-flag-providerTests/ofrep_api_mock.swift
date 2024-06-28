import Foundation
@testable import go_feature_flag_provider

class MockNetworkingService: NetworkingService {
    var mockData: Data?
    var mockStatus: Int
    var mockURLResponse: URLResponse?
    var mockError: Error?

    init(mockData: Data? = nil, mockStatus: Int = 200, mockURLResponse: URLResponse? = nil) {
        self.mockData = mockData
        self.mockURLResponse = mockURLResponse
        self.mockStatus = mockStatus
    }

    func doRequest(for request: URLRequest) async throws -> (Data, URLResponse) {
        let data = mockData ?? Data()
        var headers: [String: String]? = nil
        if mockStatus == 429 {
            headers = ["Retry-Later": "120"]
        }

        if mockStatus == 200 {
            headers = ["ETag": "33a64df551425fcc55e4d42a148795d9f25f89d4"]
        }

        if request.value(forHTTPHeaderField: "If-None-Match") == "33a64df551425fcc55e4d42a148795d9f25f89d4" {
            mockStatus = 304
        }

        let response = mockURLResponse ?? HTTPURLResponse(url: request.url!, statusCode: mockStatus, httpVersion: nil, headerFields: headers)!
        return (data, response)
    }
}
