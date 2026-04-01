import Foundation

/// 内置离线技能
public struct BuiltinSkills {

    /// 注册所有内置技能到离线管理器
    public static func register(to manager: OfflineManager) async {
        await manager.registerSkill(EchoSkill(), offlineCapable: true)
        await manager.registerSkill(TextProcessSkill(), offlineCapable: true)
        await manager.registerSkill(CalculatorSkill(), offlineCapable: true)
        await manager.registerSkill(TimestampSkill(), offlineCapable: true)
        await manager.registerSkill(JSONFormatSkill(), offlineCapable: true)
        await manager.registerSkill(HashSkill(), offlineCapable: true)
    }
}

/// 回显技能
public struct EchoSkill: SkillExecutor {
    public let skillId = "echo"
    public let skillName = "Echo"
    public let category = "test"

    public init() {}

    public func execute(_ input: Data?) async throws -> Data {
        input ?? Data()
    }
}

/// 文本处理技能
public struct TextProcessSkill: SkillExecutor {
    public let skillId = "text.process"
    public let skillName = "Text Processing"
    public let category = "text"

    public init() {}

    public func execute(_ input: Data?) async throws -> Data {
        guard let input = input,
              let text = String(data: input, encoding: .utf8) else {
            throw OFAError.executionFailed("Invalid input")
        }

        // 解析操作: "operation:text"
        let parts = text.split(separator: ":", maxSplits: 1)
        let result: String

        if parts.count == 2 {
            let op = String(parts[0])
            let content = String(parts[1])

            switch op.lowercased() {
            case "uppercase":
                result = content.uppercased()
            case "lowercase":
                result = content.lowercased()
            case "reverse":
                result = String(content.reversed())
            case "length":
                result = String(content.count)
            case "trim":
                result = content.trimmingCharacters(in: .whitespacesAndNewlines)
            default:
                result = content
            }
        } else {
            result = text
        }

        return result.data(using: .utf8) ?? Data()
    }
}

/// 计算器技能
public struct CalculatorSkill: SkillExecutor {
    public let skillId = "calculator"
    public let skillName = "Calculator"
    public let category = "math"

    public init() {}

    public func execute(_ input: Data?) async throws -> Data {
        guard let input = input,
              let text = String(data: input, encoding: .utf8) else {
            throw OFAError.executionFailed("Invalid input")
        }

        let parts = text.split(separator: " ")
        var result = ""

        if parts.count == 2 {
            // 单操作数
            let op = String(parts[0])
            guard let a = Double(String(parts[1])) else {
                throw OFAError.executionFailed("Invalid number")
            }

            switch op.lowercased() {
            case "sqrt":
                result = String(sqrt(a))
            case "sin":
                result = String(sin(a))
            case "cos":
                result = String(cos(a))
            case "abs":
                result = String(abs(a))
            default:
                throw OFAError.executionFailed("Unknown operation")
            }
        } else if parts.count >= 3 {
            // 双操作数
            guard let a = Double(String(parts[0])),
                  let b = Double(String(parts[2])) else {
                throw OFAError.executionFailed("Invalid numbers")
            }

            let op = String(parts[1])

            switch op {
            case "add", "+":
                result = String(a + b)
            case "sub", "-":
                result = String(a - b)
            case "mul", "*":
                result = String(a * b)
            case "div", "/":
                guard b != 0 else {
                    throw OFAError.executionFailed("Division by zero")
                }
                result = String(a / b)
            case "pow":
                result = String(pow(a, b))
            case "mod", "%":
                result = String(Int(a) % Int(b))
            default:
                throw OFAError.executionFailed("Unknown operation")
            }
        } else {
            throw OFAError.executionFailed("Invalid expression")
        }

        return result.data(using: .utf8) ?? Data()
    }
}

/// 时间戳技能
public struct TimestampSkill: SkillExecutor {
    public let skillId = "timestamp"
    public let skillName = "Timestamp"
    public let category = "time"

    public init() {}

    public func execute(_ input: Data?) async throws -> Data {
        let now = Date()
        let formatter = DateFormatter()

        if let input = input,
           let op = String(data: input, encoding: .utf8),
           op.lowercased() == "format" {
            formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        } else {
            formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        }

        let result = formatter.string(from: now)
        return result.data(using: .utf8) ?? Data()
    }
}

/// JSON 格式化技能
public struct JSONFormatSkill: SkillExecutor {
    public let skillId = "json.format"
    public let skillName = "JSON Formatter"
    public let category = "data"

    public init() {}

    public func execute(_ input: Data?) async throws -> Data {
        guard let input = input else {
            throw OFAError.executionFailed("Invalid input")
        }

        guard let json = try? JSONSerialization.jsonObject(with: input),
              let formatted = try? JSONSerialization.data(withJSONObject: json, options: .prettyPrinted) else {
            throw OFAError.executionFailed("Invalid JSON")
        }

        return formatted
    }
}

/// 哈希技能
public struct HashSkill: SkillExecutor {
    public let skillId = "hash.simple"
    public let skillName = "Simple Hash"
    public let category = "crypto"

    public init() {}

    public func execute(_ input: Data?) async throws -> Data {
        guard let input = input else {
            throw OFAError.executionFailed("Invalid input")
        }

        // 简单哈希 (生产环境应使用 CryptoKit)
        var hash: UInt32 = 0
        for byte in input {
            hash = hash &* 31 &+ UInt32(byte)
        }

        let result = String(hash)
        return result.data(using: .utf8) ?? Data()
    }
}