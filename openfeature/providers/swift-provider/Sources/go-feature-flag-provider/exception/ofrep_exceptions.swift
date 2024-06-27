//
//  File.swift
//  
//
//  Created by thomas.poignant on 27/06/2024.
//

import Foundation

enum OfrepError: Error {
    case httpResponseCastError
    case unmarshallError(error: Error)
    case apiUnauthorizedError(response: HTTPURLResponse)
    case forbiddenError(response: HTTPURLResponse)
    case apiTooManyRequestsError(response: HTTPURLResponse)
    case unexpectedResponseError(response: HTTPURLResponse)
}
