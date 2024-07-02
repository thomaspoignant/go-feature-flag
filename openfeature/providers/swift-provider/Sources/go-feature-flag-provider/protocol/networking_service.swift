import Foundation

protocol NetworkingService {
    func doRequest(for request: URLRequest) async throws -> (Data, URLResponse)
}

extension URLSession: NetworkingService {
    func doRequest(for request: URLRequest) async throws -> (Data, URLResponse) {
        return try await data(for: request)
    }
}
