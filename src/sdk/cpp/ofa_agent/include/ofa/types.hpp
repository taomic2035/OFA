/**
 * @file types.hpp
 * @brief Common Types
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_TYPES_HPP
#define OFA_TYPES_HPP

#include <string>
#include <vector>
#include <map>
#include <functional>
#include <future>

// JSON library
#include <nlohmann/json.hpp>

namespace ofa {

using json = nlohmann::json;

/**
 * @brief 技能处理器
 */
using SkillHandler = std::function<json(const std::string& operation, const json& input)>;

/**
 * @brief 消息处理器
 */
using MessageHandler = std::function<void(const json& msg)>;

/**
 * @brief 版本信息
 */
struct Version {
    static constexpr const char* STRING = "8.1.0";
    static constexpr int MAJOR = 8;
    static constexpr int MINOR = 1;
    static constexpr int PATCH = 0;
};

} // namespace ofa

#endif // OFA_TYPES_HPP