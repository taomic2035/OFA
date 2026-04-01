/**
 * @file builtin.cpp
 * @brief OFA OpenHarmony 内置离线技能
 */

#include "ofa/agent.h"
#include <cstring>
#include <string>
#include <vector>
#include <algorithm>
#include <cmath>
#include <cstdlib>

namespace ofa {

// 文本处理技能
static OFA_Result TextProcessHandler(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
) {
    if (!input || !output || !output_len) return OFA_ERROR;

    std::string text(reinterpret_cast<const char*>(input), input_len);
    std::string result;

    // 解析操作: 输入格式 "operation:text"
    size_t colon_pos = text.find(':');
    if (colon_pos == std::string::npos) {
        result = text; // 默认原样返回
    } else {
        std::string op = text.substr(0, colon_pos);
        std::string content = text.substr(colon_pos + 1);

        if (op == "uppercase") {
            std::transform(content.begin(), content.end(), std::back_inserter(result), ::toupper);
        } else if (op == "lowercase") {
            std::transform(content.begin(), content.end(), std::back_inserter(result), ::tolower);
        } else if (op == "reverse") {
            result = std::string(content.rbegin(), content.rend());
        } else if (op == "length") {
            result = std::to_string(content.length());
        } else if (op == "trim") {
            size_t start = content.find_first_not_of(" \t\n\r");
            size_t end = content.find_last_not_of(" \t\n\r");
            if (start != std::string::npos && end != std::string::npos) {
                result = content.substr(start, end - start + 1);
            } else {
                result = "";
            }
        } else {
            result = content;
        }
    }

    *output_len = result.length();
    *output = static_cast<uint8_t*>(malloc(*output_len));
    memcpy(*output, result.c_str(), *output_len);

    return OFA_OK;
}

// JSON 格式化技能 (简化版)
static OFA_Result JsonFormatHandler(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
) {
    if (!input || !output || !output_len) return OFA_ERROR;

    std::string json(reinterpret_cast<const char*>(input), input_len);

    // 简化的 JSON 格式化
    std::string result;
    int indent = 0;
    bool in_string = false;

    for (size_t i = 0; i < json.length(); i++) {
        char c = json[i];

        if (c == '"' && (i == 0 || json[i-1] != '\\')) {
            in_string = !in_string;
            result += c;
        } else if (in_string) {
            result += c;
        } else if (c == '{' || c == '[') {
            result += c;
            result += '\n';
            indent++;
            for (int j = 0; j < indent; j++) result += '  ';
        } else if (c == '}' || c == ']') {
            result += '\n';
            indent--;
            for (int j = 0; j < indent; j++) result += '  ';
            result += c;
        } else if (c == ',') {
            result += c;
            result += '\n';
            for (int j = 0; j < indent; j++) result += '  ';
        } else if (c == ':') {
            result += c;
            result += ' ';
        } else if (c != ' ' && c != '\t' && c != '\n' && c != '\r') {
            result += c;
        }
    }

    *output_len = result.length();
    *output = static_cast<uint8_t*>(malloc(*output_len));
    memcpy(*output, result.c_str(), *output_len);

    return OFA_OK;
}

// 计算器技能
static OFA_Result CalculatorHandler(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
) {
    if (!input || !output || !output_len) return OFA_ERROR;

    std::string expr(reinterpret_cast<const char*>(input), input_len);
    std::string result;

    // 解析表达式: 格式 "a op b" 或 "op a"
    std::vector<std::string> parts;
    std::string token;
    for (char c : expr) {
        if (c == ' ') {
            if (!token.empty()) {
                parts.push_back(token);
                token.clear();
            }
        } else {
            token += c;
        }
    }
    if (!token.empty()) parts.push_back(token);

    if (parts.size() < 2) {
        result = "Invalid expression";
    } else if (parts.size() == 2) {
        // 单操作数: sqrt, sin, cos
        std::string op = parts[0];
        double a = std::atof(parts[1].c_str());

        if (op == "sqrt") {
            result = std::to_string(std::sqrt(a));
        } else if (op == "sin") {
            result = std::to_string(std::sin(a));
        } else if (op == "cos") {
            result = std::to_string(std::cos(a));
        } else if (op == "abs") {
            result = std::to_string(std::abs(a));
        } else {
            result = "Unknown operation";
        }
    } else if (parts.size() >= 3) {
        // 双操作数: add, sub, mul, div, pow
        double a = std::atof(parts[0].c_str());
        std::string op = parts[1];
        double b = std::atof(parts[2].c_str());

        if (op == "add" || op == "+") {
            result = std::to_string(a + b);
        } else if (op == "sub" || op == "-") {
            result = std::to_string(a - b);
        } else if (op == "mul" || op == "*") {
            result = std::to_string(a * b);
        } else if (op == "div" || op == "/") {
            if (b == 0) {
                result = "Division by zero";
            } else {
                result = std::to_string(a / b);
            }
        } else if (op == "pow") {
            result = std::to_string(std::pow(a, b));
        } else if (op == "mod" || op == "%") {
            result = std::to_string(static_cast<int>(a) % static_cast<int>(b));
        } else {
            result = "Unknown operation";
        }
    }

    *output_len = result.length();
    *output = static_cast<uint8_t*>(malloc(*output_len));
    memcpy(*output, result.c_str(), *output_len);

    return OFA_OK;
}

// 回显技能
static OFA_Result EchoHandler(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
) {
    if (!input || !output || !output_len) return OFA_ERROR;

    *output_len = input_len;
    *output = static_cast<uint8_t*>(malloc(input_len));
    memcpy(*output, input, input_len);

    return OFA_OK;
}

// 时间戳技能
static OFA_Result TimestampHandler(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
) {
    std::string result;

    // 获取当前时间戳
    auto now = std::chrono::system_clock::now();
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        now.time_since_epoch()
    ).count();

    if (input && input_len > 0) {
        std::string op(reinterpret_cast<const char*>(input), input_len);
        if (op == "format") {
            // 格式化时间戳
            time_t t = std::chrono::system_clock::to_time_t(now);
            char buf[64];
            strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M:%S", localtime(&t));
            result = buf;
        } else {
            result = std::to_string(ms);
        }
    } else {
        result = std::to_string(ms);
    }

    *output_len = result.length();
    *output = static_cast<uint8_t*>(malloc(*output_len));
    memcpy(*output, result.c_str(), *output_len);

    return OFA_OK;
}

// 哈希技能 (简化版)
static OFA_Result HashHandler(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
) {
    if (!input || !output || !output_len) return OFA_ERROR;

    // 简化哈希 (实际应使用 MD5/SHA)
    uint32_t hash = 0;
    for (size_t i = 0; i < input_len; i++) {
        hash = hash * 31 + input[i];
    }

    std::string result = std::to_string(hash);

    *output_len = result.length();
    *output = static_cast<uint8_t*>(malloc(*output_len));
    memcpy(*output, result.c_str(), *output_len);

    return OFA_OK;
}

} // namespace ofa

// 内置技能注册
void OFA_RegisterBuiltinSkills(OFA_Agent* agent) {
    if (!agent) return;

    // 文本处理
    OFA_Skill text_skill = {};
    text_skill.id = "text.process";
    text_skill.name = "Text Processing";
    text_skill.category = "text";
    text_skill.offline_capable = true;
    text_skill.handler = ofa::TextProcessHandler;
    OFA_Agent_RegisterSkill(agent, &text_skill);

    // JSON 格式化
    OFA_Skill json_skill = {};
    json_skill.id = "json.format";
    json_skill.name = "JSON Formatter";
    json_skill.category = "data";
    json_skill.offline_capable = true;
    json_skill.handler = ofa::JsonFormatHandler;
    OFA_Agent_RegisterSkill(agent, &json_skill);

    // 计算器
    OFA_Skill calc_skill = {};
    calc_skill.id = "calculator";
    calc_skill.name = "Calculator";
    calc_skill.category = "math";
    calc_skill.offline_capable = true;
    calc_skill.handler = ofa::CalculatorHandler;
    OFA_Agent_RegisterSkill(agent, &calc_skill);

    // 回显
    OFA_Skill echo_skill = {};
    echo_skill.id = "echo";
    echo_skill.name = "Echo";
    echo_skill.category = "test";
    echo_skill.offline_capable = true;
    echo_skill.handler = ofa::EchoHandler;
    OFA_Agent_RegisterSkill(agent, &echo_skill);

    // 时间戳
    OFA_Skill ts_skill = {};
    ts_skill.id = "timestamp";
    ts_skill.name = "Timestamp";
    ts_skill.category = "time";
    ts_skill.offline_capable = true;
    ts_skill.handler = ofa::TimestampHandler;
    OFA_Agent_RegisterSkill(agent, &ts_skill);

    // 哈希
    OFA_Skill hash_skill = {};
    hash_skill.id = "hash.simple";
    hash_skill.name = "Simple Hash";
    hash_skill.category = "crypto";
    hash_skill.offline_capable = true;
    hash_skill.handler = ofa::HashHandler;
    OFA_Agent_RegisterSkill(agent, &hash_skill);
}