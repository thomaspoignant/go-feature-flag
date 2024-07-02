import Foundation

struct GoFeatureFlagProviderOptions {
    let endpoint: String
    let networkService: NetworkingService? = URLSession.shared
}
