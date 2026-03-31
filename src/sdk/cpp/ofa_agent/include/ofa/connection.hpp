/**
 * @file connection.hpp
 * @brief Connection Module
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_CONNECTION_HPP
#define OFA_CONNECTION_HPP

#include <string>
#include <memory>
#include <future>
#include "config.hpp"
#include "message.hpp"

namespace ofa {

/**
 * @brief 连接接口
 */
class Connection {
public:
    virtual ~Connection() = default;

    virtual std::future<void> connect(const AgentConfig& config) = 0;
    virtual std::future<void> disconnect() = 0;
    virtual std::future<void> send(const Message& msg) = 0;
    virtual std::future<std::optional<Message>> receive() = 0;
    virtual bool isConnected() const = 0;
};

/**
 * @brief 创建连接
 */
std::unique_ptr<Connection> createConnection(const AgentConfig& config);

/**
 * @brief HTTP连接
 */
class HttpConnection : public Connection {
public:
    HttpConnection() = default;
    ~HttpConnection() override = default;

    std::future<void> connect(const AgentConfig& config) override;
    std::future<void> disconnect() override;
    std::future<void> send(const Message& msg) override;
    std::future<std::optional<Message>> receive() override;
    bool isConnected() const override { return connected_; }

private:
    std::string baseUrl_;
    bool connected_ = false;
};

/**
 * @brief WebSocket连接
 */
class WebSocketConnection : public Connection {
public:
    WebSocketConnection() = default;
    ~WebSocketConnection() override = default;

    std::future<void> connect(const AgentConfig& config) override;
    std::future<void> disconnect() override;
    std::future<void> send(const Message& msg) override;
    std::future<std::optional<Message>> receive() override;
    bool isConnected() const override { return connected_; }

private:
    bool connected_ = false;
};

/**
 * @brief gRPC连接
 */
class GrpcConnection : public Connection {
public:
    GrpcConnection() = default;
    ~GrpcConnection() override = default;

    std::future<void> connect(const AgentConfig& config) override;
    std::future<void> disconnect() override;
    std::future<void> send(const Message& msg) override;
    std::future<std::optional<Message>> receive() override;
    bool isConnected() const override { return connected_; }

private:
    bool connected_ = false;
};

} // namespace ofa

#endif // OFA_CONNECTION_HPP