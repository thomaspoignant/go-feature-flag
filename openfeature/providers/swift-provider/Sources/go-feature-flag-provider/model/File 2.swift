import Foundation

public struct NetworkStruct: Equatable {
    public init(fields: [String: NetworkValue]) {
        self.fields = fields
    }
    public var fields: [String: NetworkValue]
}

public enum NetworkValue: Equatable {
    case null
    case string(String)
    case number(Double)
    case boolean(Bool)
    case structure(NetworkStruct)
    case list([NetworkValue])
}

extension NetworkValue: Codable {
    public func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()

        switch self {
        case .null:
            try container.encodeNil()
        case .number(let double):
            try container.encode(double)
        case .string(let string):
            try container.encode(string)
        case .boolean(let boolean):
            try container.encode(boolean)
        case .structure(let structure):
            try container.encode(structure)
        case .list(let list):
            try container.encode(list)
        }
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if container.decodeNil() {
            self = .null
        } else if let double = try? container.decode(Double.self) {
            self = .number(double)
        } else if let string = try? container.decode(String.self) {
            self = .string(string)
        } else if let bool = try? container.decode(Bool.self) {
            self = .boolean(bool)
        } else if let object = try? container.decode(NetworkStruct.self) {
            self = .structure(object)
        } else if let list = try? container.decode([NetworkValue].self) {
            self = .list(list)
        } else {
            throw DecodingError.dataCorrupted(
                .init(codingPath: decoder.codingPath, debugDescription: "Invalid data"))
        }
    }
}

extension NetworkStruct: Codable {
    public func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        try container.encode(fields)
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        self.fields = try container.decode([String: NetworkValue].self)
    }
}
