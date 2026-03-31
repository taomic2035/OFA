import Foundation
import SwiftProtobuf
import Darwin.math

/// Protocol for skill executors
public protocol SkillExecutor: Sendable {
    /// Unique skill identifier
    var skillId: String { get }

    /// Human-readable skill name
    var skillName: String { get }

    /// Skill category (text, data, math, utility, etc.)
    var category: String { get }

    /// Execute the skill
    /// - Parameter input: Input data (JSON)
    /// - Returns: Output data (JSON)
    func execute(_ input: Data) async throws -> Data
}

// MARK: - Built-in Skills

/// Echo skill for testing
public final class EchoSkill: SkillExecutor, @unchecked Sendable {
    public let skillId = "echo"
    public let skillName = "Echo"
    public let category = "utility"

    public init() {}

    public func execute(_ input: Data) async throws -> Data {
        let inputString = String(data: input, encoding: .utf8) ?? ""

        let result: [String: Any] = [
            "echo": inputString,
            "length": input.count
        ]

        return try JSONSerialization.data(withJSONObject: result)
    }
}

/// Text processing skill
public final class TextProcessSkill: SkillExecutor, @unchecked Sendable {
    public let skillId = "text.process"
    public let skillName = "Text Process"
    public let category = "text"

    public init() {}

    public func execute(_ input: Data) async throws -> Data {
        guard let json = try JSONSerialization.jsonObject(with: input) as? [String: Any],
              let text = json["text"] as? String else {
            throw OFAError.executionFailed("Missing 'text' field")
        }

        let operation = (json["operation"] as? String ?? "uppercase").lowercased()

        let result: String

        switch operation {
        case "uppercase":
            result = text.uppercased()
        case "lowercase":
            result = text.lowercased()
        case "reverse":
            result = String(text.reversed())
        case "length":
            let output: [String: Any] = ["result": text.count]
            return try JSONSerialization.data(withJSONObject: output)
        default:
            throw OFAError.executionFailed("Unknown operation: \(operation)")
        }

        let output: [String: Any] = ["result": result]
        return try JSONSerialization.data(withJSONObject: output)
    }
}

/// Calculator skill
public final class CalculatorSkill: SkillExecutor, @unchecked Sendable {
    public let skillId = "calculator"
    public let skillName = "Calculator"
    public let category = "math"

    public init() {}

    public func execute(_ input: Data) async throws -> Data {
        guard let json = try JSONSerialization.jsonObject(with: input) as? [String: Any],
              let operation = json["operation"] as? String else {
            throw OFAError.executionFailed("Missing 'operation' field")
        }

        let a = (json["a"] as? Double) ?? 0
        let b = (json["b"] as? Double) ?? 0

        let result: Double

        switch operation.lowercased() {
        case "add":
            result = a + b
        case "sub":
            result = a - b
        case "mul":
            result = a * b
        case "div":
            guard b != 0 else {
                throw OFAError.executionFailed("Division by zero")
            }
            result = a / b
        case "pow":
            result = pow(a, b)
        case "sqrt":
            result = sqrt(a)
        default:
            throw OFAError.executionFailed("Unknown operation: \(operation)")
        }

        let output: [String: Any] = [
            "result": result,
            "operation": operation
        ]
        return try JSONSerialization.data(withJSONObject: output)
    }
}