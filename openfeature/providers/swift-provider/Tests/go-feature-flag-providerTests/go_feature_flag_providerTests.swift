import XCTest
import OpenFeature
@testable import go_feature_flag_provider

class MovieAPIClientTests: XCTestCase {
//    var apiClient: MovieAPIClient!
//    override func setUp() {
//        super.setUp()
//        // Use the mock networking service during testing
////        let mockNetworkingService = MockNetworkingService()
//        let realNetworkingService = URLSession.shared
//        let options = GoFeatureFlagProviderOptions(endpoint: "http://localhost:1031/")
//        apiClient = MovieAPIClient(
//            networkingService: realNetworkingService,
//            options: options)
//    }
//    override func tearDown() {
//        apiClient = nil
//        super.tearDown()
//    }
//
//    
//    func testFetchMovieData() async {
//        do{
//            let evalCtx = MutableContext()
//            evalCtx.setTargetingKey(targetingKey: "1")
//            evalCtx.add(key: "email", value: Value.string("john.doe@gofeatureflag.org"))
//            
//            let (data, response) = try await apiClient.bulkEvaluation(context: evalCtx)
//            print(data, response)
//        } catch{
//            print(type(of: error))
//            print(error)
//        }
//      
//        
//        
////            do {
//////                let url = URL(string: "https://jsonplaceholder.typicode.com/todos/1")!
//////                let data = try await fetchData(from: url)
//////                // Process the data
//////                print("Data received: \(data)")
////                
////                try await apiClient.bulkEvaluation()
////                print("yoyo")
//////                print("Data received: \(data)")
////
////               
////                
////                
////            } catch {
////                // Handle the error
////                print("Error: \(error)")
////            }
//        
//        
//
////        let expectation = self.expectation(description: "Fetch movie data")
////        apiClient.fetchMovieData { result in
////            switch result {
////            case .success(let movie):
////                // Assert the properties of the movie object you expect to receive
////                print(movie)
////                XCTAssertEqual(movie.variant, "toto")
////            case .failure(let error):
////                XCTFail("Error: \(error.localizedDescription)")
////            }
////            expectation.fulfill()
////        }
////        wait(for: [expectation], timeout: 5.0)
//    }
}
