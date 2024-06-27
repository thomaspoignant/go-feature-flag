import Foundation
@testable import go_feature_flag_provider

class MockNetworkingService: NetworkingService {
    var mockData: Data?
    var mockURLResponse: URLResponse?
    var mockError: Error?

    init(mockData: Data? = nil, mockURLResponse: URLResponse? = nil, mockError: Error? = nil) {
      self.mockData = mockData
      self.mockURLResponse = mockURLResponse
      self.mockError = mockError
    }

    func doRequest(for request: URLRequest) async throws -> (Data, URLResponse) {
      if let error = mockError {
          throw error
      }

      let data = mockData ?? Data()
      let response = mockURLResponse ?? HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!

      return (data, response)
    }
}
