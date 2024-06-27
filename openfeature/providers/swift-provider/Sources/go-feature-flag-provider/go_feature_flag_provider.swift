import OpenFeature
import Combine

struct Metadata: ProviderMetadata {
    var name: String? = "GO Feature Flag provider"
}

final class GoFeatureFlagProvider: FeatureProvider {
    private let eventHandler = EventHandler(ProviderEvent.notReady)

    func initialize(initialContext: (any OpenFeature.EvaluationContext)?) {
        print("initialize")
    }

    func onContextSet(oldContext: (any OpenFeature.EvaluationContext)?,
                      newContext: any OpenFeature.EvaluationContext) {
        print("onContextSet")
    }

    func observe() -> AnyPublisher<OpenFeature.ProviderEvent, Never> {
        return eventHandler.observe()
    }

    var hooks: [any Hook] = []
    var metadata: ProviderMetadata = Metadata()

    func initialize(initialContext: EvaluationContext?) async {
        // add context-aware provider initialisation
    }

    func onContextSet(oldContext: EvaluationContext?, newContext: EvaluationContext) async {
        // add necessary changes on context change
    }

    func getBooleanEvaluation(key: String, defaultValue: Bool,
                              context: EvaluationContext?) throws -> ProviderEvaluation<Bool> {
        return ProviderEvaluation<Bool>(
            value: defaultValue,
            variant: nil,
            reason: nil,
            errorCode: nil,
            errorMessage: nil )
    }

    func getStringEvaluation(key: String, defaultValue: String,
                             context: EvaluationContext?) throws -> ProviderEvaluation<String> {
        return ProviderEvaluation<String>(
            value: defaultValue,
            variant: nil,
            reason: nil,
            errorCode: nil,
            errorMessage: nil )
    }

    func getIntegerEvaluation(key: String, defaultValue: Int64,
                              context: EvaluationContext?) throws -> ProviderEvaluation<Int64> {
        return ProviderEvaluation<Int64>(
            value: defaultValue,
            variant: nil,
            reason: nil,
            errorCode: nil,
            errorMessage: nil )

    }

    func getDoubleEvaluation(key: String, defaultValue: Double,
                             context: EvaluationContext?) throws -> ProviderEvaluation<Double> {
        return ProviderEvaluation<Double>(
            value: defaultValue,
            variant: nil,
            reason: nil,
            errorCode: nil,
            errorMessage: nil )

    }

    func getObjectEvaluation(key: String, defaultValue: Value,
                             context: EvaluationContext?) throws -> ProviderEvaluation<Value> {
        return ProviderEvaluation<Value>(
            value: defaultValue,
            variant: nil,
            reason: nil,
            errorCode: nil,
            errorMessage: nil )
    }

}
