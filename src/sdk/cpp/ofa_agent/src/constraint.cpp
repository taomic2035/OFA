/**
 * @file constraint.cpp
 * @brief Constraint Checking Implementation
 * Sprint 30: Constraint Engine
 */

#include "ofa/constraint.hpp"
#include <chrono>
#include <algorithm>

namespace ofa {

static int64_t nowMs() {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
        std::chrono::system_clock::now().time_since_epoch()
    ).count();
}

// ==================== ConstraintChecker ====================

ConstraintChecker::ConstraintChecker() {
    loadDefaultRules();
}

ConstraintChecker::~ConstraintChecker() = default;

void ConstraintChecker::addRule(const ConstraintRule& rule) {
    std::lock_guard<std::mutex> lock(mutex_);
    rules_[rule.id] = rule;
}

void ConstraintChecker::removeRule(const std::string& ruleId) {
    std::lock_guard<std::mutex> lock(mutex_);
    rules_.erase(ruleId);
    customCheckers_.erase(ruleId);
}

void ConstraintChecker::setRuleEnabled(const std::string& ruleId, bool enabled) {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = rules_.find(ruleId);
    if (it != rules_.end()) {
        it->second.enabled = enabled;
    }
}

std::vector<ConstraintRule> ConstraintChecker::getRules() const {
    std::lock_guard<std::mutex> lock(mutex_);
    std::vector<ConstraintRule> result;
    for (const auto& pair : rules_) {
        result.push_back(pair.second);
    }
    return result;
}

std::vector<ConstraintRule> ConstraintChecker::getRulesByType(ConstraintType type) const {
    std::lock_guard<std::mutex> lock(mutex_);
    std::vector<ConstraintRule> result;
    for (const auto& pair : rules_) {
        if (pair.second.type == type) {
            result.push_back(pair.second);
        }
    }
    return result;
}

ConstraintReport ConstraintChecker::check(const std::string& taskId,
                                           const std::string& skillId,
                                           const std::string& operation,
                                           const json& data) {
    ConstraintReport report;
    report.taskId = taskId;
    report.skillId = skillId;
    report.operation = operation;
    report.timestamp = nowMs();
    report.allPassed = true;

    std::lock_guard<std::mutex> lock(mutex_);

    // Get applicable rules
    std::vector<ConstraintRule> applicableRules;
    for (const auto& pair : rules_) {
        if (pair.second.enabled) {
            applicableRules.push_back(pair.second);
        }
    }

    // Check recursively
    checkJsonRecursive(data, "", report.results, applicableRules);

    // Determine overall result
    for (const auto& result : report.results) {
        if (!result.passed) {
            report.allPassed = false;
        }
    }

    return report;
}

ConstraintResult ConstraintChecker::checkField(const std::string& field,
                                                const json& value,
                                                const std::vector<ConstraintRule>& rules) {
    ConstraintResult result;
    result.field = field;
    result.passed = true;

    // Check each applicable rule
    for (const auto& rule : rules) {
        if (rule.fields.empty() || rule.fields.find(field) != rule.fields.end()) {
            // Check pattern match
            if (!rule.pattern.empty() && value.is_string()) {
                std::string strValue = value.get<std::string>();
                if (matchPattern(rule.pattern, strValue)) {
                    result.ruleId = rule.id;
                    result.ruleName = rule.name;
                    result.type = rule.type;
                    result.severity = rule.severity;
                    result.passed = false;
                    result.message = rule.description;
                    result.matchedValue = strValue;
                    break;
                }
            }

            // Check custom checker
            auto it = customCheckers_.find(rule.id);
            if (it != customCheckers_.end()) {
                if (!it->second(value)) {
                    result.ruleId = rule.id;
                    result.ruleName = rule.name;
                    result.type = rule.type;
                    result.severity = rule.severity;
                    result.passed = false;
                    result.message = rule.description;
                    break;
                }
            }
        }
    }

    return result;
}

void ConstraintChecker::setCustomChecker(const std::string& ruleId,
                                          std::function<bool(const json&)> checker) {
    std::lock_guard<std::mutex> lock(mutex_);
    customCheckers_[ruleId] = checker;
}

void ConstraintChecker::loadDefaultRules() {
    auto defaults = DefaultConstraints::getAllDefaults();
    for (const auto& rule : defaults) {
        addRule(rule);
    }
}

void ConstraintChecker::loadRules(const json& rulesJson) {
    std::lock_guard<std::mutex> lock(mutex_);

    if (rulesJson.is_array()) {
        for (const auto& ruleJson : rulesJson) {
            ConstraintRule rule = ConstraintRule::fromJson(ruleJson);
            rules_[rule.id] = rule;
        }
    }
}

json ConstraintChecker::exportRules() const {
    std::lock_guard<std::mutex> lock(mutex_);

    json j = json::array();
    for (const auto& pair : rules_) {
        j.push_back(pair.second.toJson());
    }
    return j;
}

void ConstraintChecker::clearRules() {
    std::lock_guard<std::mutex> lock(mutex_);
    rules_.clear();
    customCheckers_.clear();
}

void ConstraintChecker::checkJsonRecursive(const json& data,
                                            const std::string& prefix,
                                            std::vector<ConstraintResult>& results,
                                            const std::vector<ConstraintRule>& applicableRules) {
    if (data.is_object()) {
        for (auto it = data.begin(); it != data.end(); ++it) {
            std::string fieldPath = prefix.empty() ? it.key() : prefix + "." + it.key();
            checkJsonRecursive(it.value(), fieldPath, results, applicableRules);
        }
    } else if (data.is_array()) {
        for (size_t i = 0; i < data.size(); ++i) {
            std::string fieldPath = prefix + "[" + std::to_string(i) + "]";
            checkJsonRecursive(data[i], fieldPath, results, applicableRules);
        }
    } else {
        // Leaf value - check it
        ConstraintResult result = checkField(prefix, data, applicableRules);
        if (!result.passed || result.severity == ConstraintSeverity::Audit) {
            results.push_back(result);
        }
    }
}

bool ConstraintChecker::matchPattern(const std::string& pattern, const std::string& value) {
    try {
        std::regex re(pattern);
        return std::regex_search(value, re);
    } catch (...) {
        return false;
    }
}

// ==================== DefaultConstraints ====================

ConstraintRule DefaultConstraints::createPrivacyRule() {
    ConstraintRule rule;
    rule.id = "privacy-sensitive-fields";
    rule.name = "Sensitive Field Detection";
    rule.type = ConstraintType::Privacy;
    rule.severity = ConstraintSeverity::Block;
    rule.description = "Sensitive data field detected";
    rule.pattern = "";  // Pattern matching on field names instead
    rule.fields = {
        "password", "passwd", "pwd", "secret", "token", "api_key",
        "credit_card", "ssn", "social_security", "phone", "email",
        "address", "birth_date", "name"
    };
    rule.enabled = true;
    rule.createdAt = nowMs();
    return rule;
}

ConstraintRule DefaultConstraints::createFinancialRule() {
    ConstraintRule rule;
    rule.id = "financial-validation";
    rule.name = "Financial Data Validation";
    rule.type = ConstraintType::Financial;
    rule.severity = ConstraintSeverity::Audit;
    rule.description = "Financial transaction detected";
    rule.pattern = "\\d{4}-\\d{4}-\\d{4}-\\d{4}";  // Credit card pattern
    rule.fields = {"credit_card", "card_number", "account"};
    rule.enabled = true;
    rule.createdAt = nowMs();
    return rule;
}

ConstraintRule DefaultConstraints::createSecurityRule() {
    ConstraintRule rule;
    rule.id = "security-command-check";
    rule.name = "Security Command Check";
    rule.type = ConstraintType::Security;
    rule.severity = ConstraintSeverity::Block;
    rule.description = "Potential command injection detected";
    rule.pattern = "(exec|eval|system|shell|cmd|command).*";
    rule.fields = {"command", "cmd", "script", "code"};
    rule.enabled = true;
    rule.createdAt = nowMs();
    return rule;
}

std::vector<ConstraintRule> DefaultConstraints::getAllDefaults() {
    std::vector<ConstraintRule> rules;
    rules.push_back(createPrivacyRule());
    rules.push_back(createFinancialRule());
    rules.push_back(createSecurityRule());

    // Auth rule
    ConstraintRule authRule;
    authRule.id = "auth-required";
    authRule.name = "Authentication Required";
    authRule.type = ConstraintType::Auth;
    authRule.severity = ConstraintSeverity::Block;
    authRule.description = "Operation requires authentication";
    authRule.enabled = true;
    authRule.createdAt = nowMs();
    rules.push_back(authRule);

    // Location rule
    ConstraintRule locRule;
    locRule.id = "location-restriction";
    locRule.name = "Location Restriction";
    locRule.type = ConstraintType::Location;
    locRule.severity = ConstraintSeverity::Warning;
    locRule.description = "Location-based operation restricted";
    locRule.enabled = true;
    locRule.createdAt = nowMs();
    rules.push_back(locRule);

    // Time rule
    ConstraintRule timeRule;
    timeRule.id = "time-restriction";
    timeRule.name = "Time Restriction";
    timeRule.type = ConstraintType::Time;
    timeRule.severity = ConstraintSeverity::Warning;
    timeRule.description = "Operation restricted by time";
    timeRule.enabled = true;
    timeRule.createdAt = nowMs();
    rules.push_back(timeRule);

    return rules;
}

} // namespace ofa