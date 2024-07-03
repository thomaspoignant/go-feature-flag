import Foundation
import OpenFeature

// Define a Codable enum that can represent any type of JSON value
enum JSONValue: Codable, Equatable {
    case string(String)
    case integer(Int64)
    case double(Double)
    case object([String: JSONValue])
    case array([JSONValue])
    case bool(Bool)
    case null

    // Decode the JSON based on its type
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if let stringValue = try? container.decode(String.self) {
            self = .string(stringValue)
        } else if let intValue = try? container.decode(Int64.self) {
            self = .integer(intValue)
        } else if let doubleValue = try? container.decode(Double.self) {
            self = .double(doubleValue)
        } else if let boolValue = try? container.decode(Bool.self) {
            self = .bool(boolValue)
        } else if let objectValue = try? container.decode([String: JSONValue].self) {
            self = .object(objectValue)
        } else if let arrayValue = try? container.decode([JSONValue].self) {
            self = .array(arrayValue)
        } else if container.decodeNil() {
            self = .null
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode JSONValue")
        }
    }

    // Encode the JSON based on its type
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch self {
        case .string(let value):
            try container.encode(value)
        case .integer(let value):
            try container.encode(value)
        case .double(let value):
            try container.encode(value)
        case .object(let value):
            try container.encode(value)
        case .array(let value):
            try container.encode(value)
        case .bool(let value):
            try container.encode(value)
        case .null:
            try container.encodeNil()
        }
    }

    func asString() -> String? {
        if case .string(let value) = self {
            return value
        }
        return nil
    }

    func asBoolean() -> Bool? {
        if case .bool(let value) = self {
            return value
        }
        return nil
    }

    func asInteger() -> Int64? {
        if case .integer(let value) = self {
            return value
        }
        return nil
    }

    func asDouble() -> Double? {
        if case .double(let value) = self {
            return value
        }
        return nil
    }

    func asObject() -> [String:JSONValue]? {
        if case .object(let value) = self {
            return value
        }
        return nil
    }

    func asArray() -> [JSONValue]? {
        if case .array(let value) = self {
            return value
        }
        return nil
    }

    func toValue() -> Value {
            switch self {
            case .string(let string):
                return .string(string)
            case .integer(let integer):
                return .integer(integer)
            case .double(let double):
                return .double(double)
            case .bool(let bool):
                return .boolean(bool)
            case .object(let object):
                let transformedObject = object.mapValues { $0.toValue() }
                return .structure(transformedObject)
            case .array(let array):
                let transformedArray = array.map { $0.toValue() }
                return .list(transformedArray)
            case .null:
                return .null
            }
        }
}
