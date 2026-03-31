/**
 * @file config.hpp
 * @brief Agent Configuration
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_CONFIG_HPP
#define OFA_CONFIG_HPP

#include <string>
#include <vector>
#include <map>

namespace ofa {

/**
 * @brief Agent配置
 */
struct AgentConfig {
    std::string agentId;
    std::string name;
    std::string centerUrl;
    std::string connectionType; // "grpc", "http", "websocket"
    uint32_t heartbeatInterval = 30;
    uint32_t reconnectInterval = 5;
    uint32_t maxReconnectAttempts = 10;
    std::vector<std::string> skills;
    std::map<std::string, std::string> metadata;
    bool tlsEnabled = false;
    uint32_t timeout = 30;

    /**
     * @brief 默认配置
     */
    static AgentConfig Default() {
        return AgentConfig{
            .agentId = "",
            .name = "C++ Agent",
            .centerUrl = "localhost:9090",
            .connectionType = "grpc",
            .heartbeatInterval = 30,
            .reconnectInterval = 5,
            .maxReconnectAttempts = 10,
            .skills = {},
            .metadata = {},
            .tlsEnabled = false,
            .timeout = 30
        };
    }
};

/**
 * @brief 配置构建器
 */
class AgentConfigBuilder {
public:
    AgentConfigBuilder& agentId(const std::string& id) {
        config_.agentId = id;
        return *this;
    }

    AgentConfigBuilder& name(const std::string& name) {
        config_.name = name;
        return *this;
    }

    AgentConfigBuilder& centerUrl(const std::string& url) {
        config_.centerUrl = url;
        return *this;
    }

    AgentConfigBuilder& connectionType(const std::string& type) {
        config_.connectionType = type;
        return *this;
    }

    AgentConfigBuilder& heartbeatInterval(uint32_t interval) {
        config_.heartbeatInterval = interval;
        return *this;
    }

    AgentConfigBuilder& addSkill(const std::string& skill) {
        config_.skills.push_back(skill);
        return *this;
    }

    AgentConfigBuilder& metadata(const std::string& key, const std::string& value) {
        config_.metadata[key] = value;
        return *this;
    }

    AgentConfigBuilder& tls(bool enabled) {
        config_.tlsEnabled = enabled;
        return *this;
    }

    AgentConfig build() {
        return config_;
    }

private:
    AgentConfig config_ = AgentConfig::Default();
};

} // namespace ofa

#endif // OFA_CONFIG_HPP