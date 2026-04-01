import Foundation

/// 约束类型
public struct ConstraintType: OptionSet, Sendable {
    public let rawValue: Int

    public init(rawValue: Int) {
        self.rawValue = rawValue
    }

    public static let none = ConstraintType(rawValue: 0)
    public static let privacy = ConstraintType(rawValue: 1)
    public static let financial = ConstraintType(rawValue: 2)
    public static let security = ConstraintType(rawValue: 4)
    public static let authRequired = ConstraintType(rawValue: 8)
    public static let location = ConstraintType(rawValue: 16)
    public static let personal = ConstraintType(rawValue: 32)
    public static let device = ConstraintType(rawValue: 64)
}

/// 约束检查结果
public struct ConstraintResult: Sendable {
    public let allowed: Bool
    public let violated: ConstraintType
    public let reason: String?
    public let requiresAuth: Bool
    public var suggestions: [String]

    public init(allowed: Bool, violated: ConstraintType = .none, reason: String? = nil, requiresAuth: Bool = false) {
        self.allowed = allowed
        self.violated = violated
        self.reason = reason
        self.requiresAuth = requiresAuth
        self.suggestions = []
    }

    public mutating func addSuggestion(_ suggestion: String) {
        suggestions.append(suggestion)
    }
}

/// 约束规则
public struct ConstraintRule: Sendable {
    public let name: String
    public let type: ConstraintType
    public let actionPattern: String?
    public let dataPattern: String?
    public let offlineRestricted: Bool
    public let requiresAuth: Bool
    public let message: String

    public init(
        name: String,
        type: ConstraintType,
        actionPattern: String? = nil,
        dataPattern: String? = nil,
        offlineRestricted: Bool = false,
        requiresAuth: Bool = false,
        message: String = ""
    ) {
        self.name = name
        self.type = type
        self.actionPattern = actionPattern
        self.dataPattern = dataPattern
        self.offlineRestricted = offlineRestricted
        self.requiresAuth = requiresAuth
        self.message = message
    }
}

/// 约束检查器
public actor ConstraintChecker {
    private var rules: [ConstraintRule] = []
    private var offlineRestrictedActions: Set<String> = []
    private var sensitiveFields: Set<String> = []
    private var offlineMode = false

    public init() {
        loadDefaultRules()
        loadSensitiveFields()
    }

    private func loadDefaultRules() {
        // 财务操作
        addRule(ConstraintRule(
            name: "financial_operations",
            type: .financial,
            actionPattern: "(payment|transfer|withdraw|pay)",
            offlineRestricted: true,
            requiresAuth: true,
            message: "Financial operations require online mode and authorization"
        ))

        // 隐私数据
        addRule(ConstraintRule(
            name: "privacy_data",
            type: .privacy,
            dataPattern: "(idcard|id_card|身份证|passport|护照)",
            message: "Data contains sensitive personal information"
        ))

        // 位置信息
        addRule(ConstraintRule(
            name: "location_data",
            type: .location,
            dataPattern: "(location|gps|latitude|longitude|经纬度)",
            requiresAuth: true,
            message: "Location data sharing requires authorization"
        ))

        // 安全操作
        addRule(ConstraintRule(
            name: "security_operations",
            type: .security,
            actionPattern: "(delete|password|auth|login|logout)",
            offlineRestricted: true,
            requiresAuth: true,
            message: "Security operations require online mode and authorization"
        ))
    }

    private func loadSensitiveFields() {
        let fields = [
            "idcard", "id_card", "身份证", "passport", "护照",
            "phone", "mobile", "电话", "手机",
            "email", "邮箱",
            "address", "地址",
            "bank_account", "银行卡",
            "password", "密码",
            "token", "令牌",
            "secret", "密钥",
            "location", "gps", "位置"
        ]
        sensitiveFields = Set(fields.map { $0.lowercased() })
    }

    /// 添加规则
    public func addRule(_ rule: ConstraintRule) {
        rules.append(rule)

        if rule.offlineRestricted, let pattern = rule.actionPattern {
            let cleanPattern = pattern
                .replacingOccurrences(of: "(", with: "")
                .replacingOccurrences(of: ")", with: "")
            for action in cleanPattern.split(separator: "|") {
                offlineRestrictedActions.insert(action.trimmingCharacters(in: .whitespaces).lowercased())
            }
        }
    }

    /// 移除规则
    public func removeRule(_ name: String) {
        rules.removeAll { $0.name == name }
    }

    /// 设置离线模式
    public func setOfflineMode(_ offline: Bool) {
        offlineMode = offline
    }

    /// 检查约束
    public func check(action: String, data: String? = nil) -> ConstraintResult {
        var result = ConstraintResult(allowed: true)

        // 1. 检查离线受限操作
        if offlineMode {
            let actionLower = action.lowercased()
            for restricted in offlineRestrictedActions {
                if actionLower.contains(restricted) {
                    result = ConstraintResult(
                        allowed: false,
                        violated: [.financial, .security],
                        reason: "Action '\(action)' requires online mode",
                        requiresAuth: false
                    )
                    result.suggestions.append("Connect to network or use alternative offline action")
                    return result
                }
            }
        }

        // 2. 应用规则
        for rule in rules {
            let ruleResult = applyRule(rule, action: action, data: data)
            if !ruleResult.allowed {
                return ruleResult
            }
        }

        // 3. 检查敏感数据
        if let data = data {
            let dataResult = checkSensitiveData(data)
            if !dataResult.allowed {
                return dataResult
            }
        }

        return result
    }

    private func applyRule(_ rule: ConstraintRule, action: String, data: String?) -> ConstraintResult {
        var result = ConstraintResult(allowed: true)

        // 检查操作模式
        if let actionPattern = rule.actionPattern {
            if !action.range(of: actionPattern, options: .regularExpression)?.isEmpty ?? true {
                return result
            }
        }

        // 检查数据模式
        if let dataPattern = rule.dataPattern, let data = data {
            if !data.range(of: dataPattern, options: .regularExpression)?.isEmpty ?? true {
                return result
            }
        }

        // 离线限制
        if rule.offlineRestricted && offlineMode {
            return ConstraintResult(
                allowed: false,
                violated: rule.type,
                reason: rule.message,
                requiresAuth: rule.requiresAuth
            )
        }

        // 授权要求
        if rule.requiresAuth {
            // TODO: 检查用户授权状态
            return ConstraintResult(
                allowed: false,
                violated: .authRequired,
                reason: rule.message,
                requiresAuth: true
            )
        }

        return result
    }

    private func checkSensitiveData(_ data: String) -> ConstraintResult {
        let dataLower = data.lowercased()

        for field in sensitiveFields {
            if dataLower.contains(field) {
                var type: ConstraintType = .privacy
                var reason = "Data contains sensitive information"

                if field.contains("bank") || field.contains("card") {
                    type = .financial
                    reason = "Data contains financial information"
                } else if field.contains("location") || field.contains("gps") {
                    type = .location
                    reason = "Data contains location information"
                } else if field.contains("password") || field.contains("token") || field.contains("secret") {
                    type = .security
                    reason = "Data contains security credentials"
                }

                return ConstraintResult(
                    allowed: false,
                    violated: type,
                    reason: reason
                )
            }
        }

        return ConstraintResult(allowed: true)
    }

    /// 添加敏感字段
    public func addSensitiveField(_ field: String) {
        sensitiveFields.insert(field.lowercased())
    }

    /// 移除敏感字段
    public func removeSensitiveField(_ field: String) {
        sensitiveFields.remove(field.lowercased())
    }

    /// 获取离线受限操作
    public func getOfflineRestrictedActions() -> Set<String> {
        offlineRestrictedActions
    }

    /// 获取规则列表
    public func getRules() -> [ConstraintRule] {
        rules
    }

    /// 是否允许
    public func isAllowed(action: String, data: String? = nil) -> Bool {
        check(action: action, data: data).allowed
    }
}