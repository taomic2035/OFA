/**
 * @file constraint.hpp
 * @brief Constraint Checking for OFA C++ SDK
 * @version 8.1.0
 * Sprint 30: Constraint Engine
 */

#ifndef OFA_CONSTRAINT_HPP
#define OFA_CONSTRAINT_HPP

#include <string>
#include <vector>
#include <map>
#include <set>
#include <memory>
#include <functional>
#include <regex>

#include "types.hpp"
#include "error.hpp"

namespace ofa {

/**
 * @brief 约束类型
 */
enum class ConstraintType {
    Privacy,        // 隐私约束
    Financial,      // 财务约束
    Security,       // 安全约束
    Auth,           // 认证约束
    Location,       // 位置约束
    Time,           // 时间约束
    Custom          // 自定义约束
};

/**
 * @brief 约束严重级别
 */
enum class ConstraintSeverity {
    Warning,        // 警告
    Block,          // 阻止执行
    Audit           // 审计记录
};

/**
 * @brief 约束规则
 */
struct ConstraintRule {
    std::string id;
    std::string name;
    ConstraintType type;
    ConstraintSeverity severity;
    std::string description;
    std::string pattern;           // 正则匹配模式
    std::set<std::string> fields;  // 适用的字段名
    bool enabled = true;
    int64_t createdAt;

    json toJson() const {
        json j;
        j["id"] = id;
        j["name"] = name;
        j["type"] = static_cast<int>(type);
        j["severity"] = static_cast<int>(severity);
        j["description"] = description;
        j["pattern"] = pattern;
        j["fields"] = std::vector<std::string>(fields.begin(), fields.end());
        j["enabled"] = enabled;
        j["createdAt"] = createdAt;
        return j;
    }

    static ConstraintRule fromJson(const json& j) {
        ConstraintRule rule;
        rule.id = j["id"].get<std::string>();
        rule.name = j["name"].get<std::string>();
        rule.type = static_cast<ConstraintType>(j["type"].get<int>());
        rule.severity = static_cast<ConstraintSeverity>(j["severity"].get<int>());
        rule.description = j.value("description", "");
        rule.pattern = j.value("pattern", "");
        auto fieldsVec = j.value("fields", std::vector<std::string>{});
        rule.fields = std::set<std::string>(fieldsVec.begin(), fieldsVec.end());
        rule.enabled = j.value("enabled", true);
        rule.createdAt = j.value("createdAt", 0);
        return rule;
    }
};

/**
 * @brief 约束检查结果
 */
struct ConstraintResult {
    std::string ruleId;
    std::string ruleName;
    ConstraintType type;
    ConstraintSeverity severity;
    bool passed;
    std::string message;
    std::string field;
    std::string matchedValue;

    json toJson() const {
        json j;
        j["ruleId"] = ruleId;
        j["ruleName"] = ruleName;
        j["type"] = static_cast<int>(type);
        j["severity"] = static_cast<int>(severity);
        j["passed"] = passed;
        j["message"] = message;
        j["field"] = field;
        j["matchedValue"] = matchedValue;
        return j;
    }
};

/**
 * @brief 约束检查报告
 */
struct ConstraintReport {
    std::string taskId;
    std::string skillId;
    std::string operation;
    bool allPassed;
    std::vector<ConstraintResult> results;
    int64_t timestamp;

    json toJson() const {
        json j;
        j["taskId"] = taskId;
        j["skillId"] = skillId;
        j["operation"] = operation;
        j["allPassed"] = allPassed;
        j["results"] = json::array();
        for (const auto& r : results) {
            j["results"].push_back(r.toJson());
        }
        j["timestamp"] = timestamp;
        return j;
    }

    bool hasBlockers() const {
        for (const auto& r : results) {
            if (!r.passed && r.severity == ConstraintSeverity::Block) {
                return true;
            }
        }
        return false;
    }

    std::vector<std::string> getBlockerMessages() const {
        std::vector<std::string> messages;
        for (const auto& r : results) {
            if (!r.passed && r.severity == ConstraintSeverity::Block) {
                messages.push_back(r.message);
            }
        }
        return messages;
    }
};

/**
 * @brief 约束检查器
 */
class ConstraintChecker {
public:
    explicit ConstraintChecker();
    ~ConstraintChecker();

    /**
     * @brief 添加规则
     */
    void addRule(const ConstraintRule& rule);

    /**
     * @brief 移除规则
     */
    void removeRule(const std::string& ruleId);

    /**
     * @brief 启用/禁用规则
     */
    void setRuleEnabled(const std::string& ruleId, bool enabled);

    /**
     * @brief 获取所有规则
     */
    std::vector<ConstraintRule> getRules() const;

    /**
     * @brief 获取特定类型的规则
     */
    std::vector<ConstraintRule> getRulesByType(ConstraintType type) const;

    /**
     * @brief 检查数据
     */
    ConstraintReport check(const std::string& taskId,
                           const std::string& skillId,
                           const std::string& operation,
                           const json& data);

    /**
     * @brief 检查单个字段
     */
    ConstraintResult checkField(const std::string& field,
                                const json& value,
                                const std::vector<ConstraintRule>& rules);

    /**
     * @brief 设置自定义检查器
     */
    void setCustomChecker(const std::string& ruleId,
                          std::function<bool(const json&)> checker);

    /**
     * @brief 加载默认规则
     */
    void loadDefaultRules();

    /**
     * @brief 从JSON加载规则
     */
    void loadRules(const json& rulesJson);

    /**
     * @brief 导出规则到JSON
     */
    json exportRules() const;

    /**
     * @brief 清除所有规则
     */
    void clearRules();

private:
    std::map<std::string, ConstraintRule> rules_;
    std::map<std::string, std::function<bool(const json&)>> customCheckers_;
    mutable std::mutex mutex_;

    void checkJsonRecursive(const json& data,
                            const std::string& prefix,
                            std::vector<ConstraintResult>& results,
                            const std::vector<ConstraintRule>& applicableRules);

    bool matchPattern(const std::string& pattern, const std::string& value);
};

/**
 * @brief 内置约束规则
 */
namespace DefaultConstraints {

/**
 * @brief 创建隐私规则
 */
ConstraintRule createPrivacyRule();

/**
 * @brief 创建财务规则
 */
ConstraintRule createFinancialRule();

/**
 * @brief 创建安全规则
 */
ConstraintRule createSecurityRule();

/**
 * @brief 获取所有默认规则
 */
std::vector<ConstraintRule> getAllDefaults();

} // namespace DefaultConstraints

} // namespace ofa

#endif // OFA_CONSTRAINT_HPP