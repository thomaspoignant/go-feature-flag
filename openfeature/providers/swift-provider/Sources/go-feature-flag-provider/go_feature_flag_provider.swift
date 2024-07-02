import OpenFeature
import Foundation
import Combine

struct Metadata: ProviderMetadata {
    var name: String? = "GO Feature Flag provider"
}

final class GoFeatureFlagProvider: FeatureProvider {
    private let eventHandler = EventHandler(ProviderEvent.notReady)
    private var evaluationContext: OpenFeature.EvaluationContext?

    private var options: GoFeatureFlagProviderOptions
    private let ofrepAPI: OfrepAPI

    private var inMemoryCache: [String: OfrepEvaluationResponseFlag] = [:]


    init(options: GoFeatureFlagProviderOptions) {
        self.options = options

        // Define network service to use
        var networkService: NetworkingService = URLSession.shared
        if let netSer = self.options.networkService {
            networkService = netSer
        }
        self.ofrepAPI = OfrepAPI(networkingService: networkService, options: self.options)
    }



    func observe() -> AnyPublisher<OpenFeature.ProviderEvent, Never> {
        return eventHandler.observe()
    }

    var hooks: [any Hook] = []
    var metadata: ProviderMetadata = Metadata()

    func initialize(initialContext: (any OpenFeature.EvaluationContext)?) {
        self.evaluationContext = initialContext
        Task {
            try await self.evaluateFlags(context: self.evaluationContext, successEvent: .ready)
        }
    }

    func onContextSet(oldContext: (any OpenFeature.EvaluationContext)?,
                      newContext: any OpenFeature.EvaluationContext) {
        print("onContextSet")
    }


    func getBooleanEvaluation(key: String, defaultValue: Bool,
                              context: EvaluationContext?) throws -> ProviderEvaluation<Bool> {
        guard let flagCached = self.inMemoryCache[key] else {
            throw OpenFeatureError.flagNotFoundError(key: key)
        }
        guard let value = flagCached.value?.asBoolean() else {
            throw OpenFeatureError.typeMismatchError
        }
        return ProviderEvaluation<Bool>(
            value: value,
            variant: flagCached.variant,
            reason: flagCached.reason)
    }


    private func genericEvaluation(key: String) throws -> OfrepEvaluationResponseFlag {
        guard let flagCached = self.inMemoryCache[key] else {
            throw OpenFeatureError.flagNotFoundError(key: key)
        }

        if flagCached.isError() {
            switch flagCached.errorCode{
            case .flagNotFound:
                throw OpenFeatureError.flagNotFoundError(key: key)
            case .invalidContext:
                throw OpenFeatureError.invalidContextError
            case .parseError:
                throw OpenFeatureError.parseError(message: flagCached.errorDetails ?? "parse error")
            case .providerNotReady:
                throw OpenFeatureError.providerNotReadyError
            case .targetingKeyMissing:
                throw OpenFeatureError.targetingKeyMissingError
            case .typeMismatch:
                throw OpenFeatureError.typeMismatchError
            default:
                throw OpenFeatureError.generalError(message: flagCached.errorDetails ?? "general error")
            }
        }

        return flagCached
    }


    func getStringEvaluation(key: String, defaultValue: String,
                             context: EvaluationContext?) throws -> ProviderEvaluation<String> {
        let flagCached = try genericEvaluation(key: key)
        guard let value = flagCached.value?.asString() else {
            throw OpenFeatureError.typeMismatchError
        }
        return ProviderEvaluation<String>(
            value: value,
            variant: flagCached.variant,
            reason: flagCached.reason)
    }

    func getIntegerEvaluation(key: String, defaultValue: Int64,
                              context: EvaluationContext?) throws -> ProviderEvaluation<Int64> {
        let flagCached = try genericEvaluation(key: key)
        guard let value = flagCached.value?.asInteger() else {
            throw OpenFeatureError.typeMismatchError
        }
        return ProviderEvaluation<Int64>(
            value: Int64(value),
            variant: flagCached.variant,
            reason: flagCached.reason)

    }

    func getDoubleEvaluation(key: String, defaultValue: Double,
                             context: EvaluationContext?) throws -> ProviderEvaluation<Double> {
        let flagCached = try genericEvaluation(key: key)
        guard let value = flagCached.value?.asDouble() else {
            throw OpenFeatureError.typeMismatchError
        }
        return ProviderEvaluation<Double>(
            value: value,
            variant: flagCached.variant,
            reason: flagCached.reason)

    }

    func getObjectEvaluation(key: String, defaultValue: Value,
                             context: EvaluationContext?) throws -> ProviderEvaluation<Value> {
        let flagCached = try genericEvaluation(key: key)
        guard let value = flagCached.value?.asObject() else {
            throw OpenFeatureError.typeMismatchError
        }

        // Convert JSONValue to Value
        let encoded = try JSONEncoder().encode(value)
        print(String(data: encoded, encoding: .utf8))
        let decoded = try JSONDecoder().decode(Value.self, from: encoded)

        print(decoded)


//        var flagCached = try genericEvaluation(key: key)
//        guard let value = flagCached.value?.object else {
//            throw OpenFeatureError.typeMismatchError
//        }
        return ProviderEvaluation<Value>(
            value: decoded,
            variant: flagCached.variant,
            reason: flagCached.reason)
    }

    private func evaluateFlags(context: EvaluationContext?, successEvent: ProviderEvent) async throws {
        do {
            let (ofrepEvalResponse, httpResp) = try await self.ofrepAPI.postBulkEvaluateFlags(context: context)

            if httpResp.statusCode == 304 {
                return // Do nothing the value is the same
            }

            if ofrepEvalResponse.isError(){
                switch ofrepEvalResponse.errorCode {
                case .providerNotReady:
                    throw OpenFeatureError.providerNotReadyError
                case .parseError:
                    throw OpenFeatureError.parseError(message: ofrepEvalResponse.errorDetails ?? "impossible to parse")
                case .targetingKeyMissing:
                    throw OpenFeatureError.targetingKeyMissingError
                case .invalidContext:
                    throw OpenFeatureError.invalidContextError
                default:
                    throw OpenFeatureError.generalError(message: ofrepEvalResponse.errorDetails ?? "")
                }
            }

            var inMemoryCacheNew: [String:OfrepEvaluationResponseFlag] = [:]
            for flag in ofrepEvalResponse.flags {
                if let key = flag.key {
                    inMemoryCacheNew[key] = flag
                }
            }
            self.inMemoryCache = inMemoryCacheNew
        } catch let error as OfrepError{
            print(error)
        } catch {
            print(error)
            self.eventHandler.send(.error)
            throw error
        }

        self.eventHandler.send(successEvent)
    }
}
