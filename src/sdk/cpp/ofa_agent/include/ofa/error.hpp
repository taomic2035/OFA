/**
 * @file error.hpp
 * @brief Error Handling
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_ERROR_HPP
#define OFA_ERROR_HPP

#include <string>
#include <stdexcept>

namespace ofa {

/**
 * @brief OFA异常
 */
class OfAException : public std::runtime_error {
public:
    explicit OfAException(const std::string& message)
        : std::runtime_error(message) {}
};

/**
 * @brief 连接错误
 */
class ConnectionException : public OfAException {
public:
    explicit ConnectionException(const std::string& message)
        : OfAException(message) {}
};

/**
 * @brief 技能未找到
 */
class SkillNotFoundException : public OfAException {
public:
    explicit SkillNotFoundException(const std::string& skillId)
        : OfAException("Skill not found: " + skillId) {}
};

/**
 * @brief 无效操作
 */
class InvalidOperationException : public OfAException {
public:
    explicit InvalidOperationException(const std::string& operation)
        : OfAException("Invalid operation: " + operation) {}
};

/**
 * @brief 协议错误
 */
class ProtocolException : public OfAException {
public:
    explicit ProtocolException(const std::string& message)
        : OfAException(message) {}
};

/**
 * @brief 超时
 */
class TimeoutException : public OfAException {
public:
    TimeoutException() : OfAException("Operation timed out") {}
};

} // namespace ofa

#endif // OFA_ERROR_HPP