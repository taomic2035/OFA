/**
 * @file builtin.cpp
 * @brief Built-in Skills Implementation
 * Sprint 29: C++ Agent SDK
 */

#include "ofa/skills.hpp"
#include <algorithm>
#include <cmath>
#include <cctype>

namespace ofa {

/**
 * @brief 注册内置技能
 */
void registerBuiltinSkills(SkillExecutor& executor) {
    // Echo技能
    executor.registerSkill("echo", [](const std::string& op, const json& input) -> json {
        return input;
    });

    // 文本处理技能
    executor.registerSkill("text.process", [](const std::string& op, const json& input) -> json {
        std::string text = input.value("text", "");

        if (op == "uppercase") {
            std::transform(text.begin(), text.end(), text.begin(), ::toupper);
            return text;
        } else if (op == "lowercase") {
            std::transform(text.begin(), text.end(), text.begin(), ::tolower);
            return text;
        } else if (op == "reverse") {
            std::reverse(text.begin(), text.end());
            return text;
        } else if (op == "length") {
            return text.length();
        }

        return text;
    });

    // 计算器技能
    executor.registerSkill("calculator", [](const std::string& op, const json& input) -> json {
        double a = input.value("a", 0.0);
        double b = input.value("b", 0.0);

        if (op == "add") return a + b;
        if (op == "sub") return a - b;
        if (op == "mul") return a * b;
        if (op == "div") {
            if (b == 0) throw std::runtime_error("Division by zero");
            return a / b;
        }
        if (op == "pow") return std::pow(a, b);
        if (op == "sqrt") return std::sqrt(a);
        if (op == "mod") return std::fmod(a, b);
        if (op == "abs") return std::abs(a);

        return 0.0;
    });

    // JSON处理技能
    executor.registerSkill("json.process", [](const std::string& op, const json& input) -> json {
        json data = input.value("data", json::object());

        if (op == "parse") {
            if (data.is_string()) {
                return json::parse(data.get<std::string>());
            }
            return data;
        } else if (op == "stringify") {
            return data.dump(input.value("indent", 2));
        } else if (op == "get_keys") {
            if (data.is_object()) {
                std::vector<std::string> keys;
                for (auto& [key, _] : data.items()) {
                    keys.push_back(key);
                }
                return keys;
            }
            return json::array();
        } else if (op == "get_values") {
            if (data.is_object()) {
                std::vector<json> values;
                for (auto& [_, val] : data.items()) {
                    values.push_back(val);
                }
                return values;
            }
            return json::array();
        }

        return data;
    });
}

} // namespace ofa