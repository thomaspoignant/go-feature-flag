import Foundation
import OpenFeature

struct EvaluationRequest {
    var context: [String: AnyHashable?] = [:]

    mutating func setTargetingKey(targetingKey: String) {
        context["targetingKey"] = targetingKey
    }

    mutating func addContext(toAdd: [String: AnyHashable?]) {
        for (key, value) in toAdd {
            context[key] = value
        }
    }

    func asObjectMap() -> [String: AnyHashable?] {
        var container: [String: AnyHashable?] = [:]
        container["context"] = context
        return container
    }

    func asJSONData() throws -> Data {
        do {
            let filteredDictionary = asObjectMap().compactMapValues { $0 }
            let jsonData = try JSONSerialization.data(withJSONObject: filteredDictionary, options: [])
            return jsonData
        } catch {
            // TODO: catch the errors and remap them
            throw error
        }
    }

    static func convertEvaluationContext(context: EvaluationContext) -> EvaluationRequest {
        var requestBody = EvaluationRequest()
        requestBody.setTargetingKey(targetingKey:context.getTargetingKey())
        requestBody.addContext(toAdd:context.asObjectMap())
        return requestBody
    }
}
