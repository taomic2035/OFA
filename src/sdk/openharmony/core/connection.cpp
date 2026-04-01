/**
 * @file connection.cpp
 * @brief OFA Center 连接管理实现
 */

#include "ofa/agent.h"
#include <string>
#include <mutex>
#include <atomic>
#include <thread>
#include <chrono>
#include <functional>

namespace ofa {

// 连接状态
enum class ConnectionState {
    DISCONNECTED,
    CONNECTING,
    CONNECTED,
    RECONNECTING
};

// 连接管理器实现
struct ConnectionManager {
    std::string center_addr;
    std::atomic<ConnectionState> state{ConnectionState::DISCONNECTED};
    std::atomic<bool> auto_reconnect{true};
    int reconnect_interval_ms{5000};
    int heartbeat_interval_ms{30000};

    // 回调
    std::function<void()> on_connected;
    std::function<void()> on_disconnected;
    std::function<void(const std::string&)> on_error;

    // 线程控制
    std::thread heartbeat_thread;
    std::thread reconnect_thread;
    std::atomic<bool> running{false};
    std::mutex mutex;

    // 认证信息
    std::string token;
    std::string agent_id;

    ConnectionManager() = default;

    ~ConnectionManager() {
        stop();
    }

    void start() {
        if (running) return;
        running = true;

        // 启动连接线程
        connect();

        // 启动心跳线程
        heartbeat_thread = std::thread([this]() {
            while (running) {
                if (state == ConnectionState::CONNECTED) {
                    sendHeartbeat();
                }
                std::this_thread::sleep_for(
                    std::chrono::milliseconds(heartbeat_interval_ms)
                );
            }
        });
    }

    void stop() {
        running = false;
        if (heartbeat_thread.joinable()) {
            heartbeat_thread.join();
        }
        if (reconnect_thread.joinable()) {
            reconnect_thread.join();
        }
        state = ConnectionState::DISCONNECTED;
    }

    void connect() {
        state = ConnectionState::CONNECTING;

        // TODO: 实际 gRPC/HTTP 连接
        // 在 OpenHarmony 中使用分布式软总线

        // 模拟连接成功
        state = ConnectionState::CONNECTED;
        if (on_connected) {
            on_connected();
        }
    }

    void disconnect() {
        state = ConnectionState::DISCONNECTED;
        if (on_disconnected) {
            on_disconnected();
        }

        // 自动重连
        if (auto_reconnect && running) {
            startReconnect();
        }
    }

    void startReconnect() {
        if (reconnect_thread.joinable()) return;

        reconnect_thread = std::thread([this]() {
            while (running && state != ConnectionState::CONNECTED) {
                std::this_thread::sleep_for(
                    std::chrono::milliseconds(reconnect_interval_ms)
                );
                if (running) {
                    state = ConnectionState::RECONNECTING;
                    connect();
                }
            }
        });
    }

    void sendHeartbeat() {
        // TODO: 实际心跳发送
        // 在 OpenHarmony 中通过分布式软总线发送
    }

    bool isConnected() const {
        return state == ConnectionState::CONNECTED;
    }
};

} // namespace ofa

// C 接口实现

struct OFA_Connection {
    ofa::ConnectionManager manager;
};

OFA_Connection* OFA_Connection_Create(const char* center_addr) {
    if (!center_addr) return nullptr;

    auto* conn = new OFA_Connection();
    conn->manager.center_addr = center_addr;
    return conn;
}

OFA_Result OFA_Connection_Start(OFA_Connection* conn) {
    if (!conn) return OFA_ERROR;

    conn->manager.start();
    return OFA_OK;
}

OFA_Result OFA_Connection_Stop(OFA_Connection* conn) {
    if (!conn) return OFA_ERROR;

    conn->manager.stop();
    return OFA_OK;
}

void OFA_Connection_Destroy(OFA_Connection* conn) {
    if (conn) {
        conn->manager.stop();
        delete conn;
    }
}

bool OFA_Connection_IsConnected(const OFA_Connection* conn) {
    if (!conn) return false;
    return conn->manager.isConnected();
}

OFA_Result OFA_Connection_SetToken(OFA_Connection* conn, const char* token) {
    if (!conn || !token) return OFA_ERROR;

    conn->manager.token = token;
    return OFA_OK;
}

void OFA_Connection_SetCallbacks(
    OFA_Connection* conn,
    OFA_ConnectionCallback on_connected,
    OFA_ConnectionCallback on_disconnected,
    OFA_ConnectionErrorCallback on_error,
    void* user_data
) {
    if (!conn) return;

    if (on_connected) {
        conn->manager.on_connected = [on_connected, user_data]() {
            on_connected(user_data);
        };
    }
    if (on_disconnected) {
        conn->manager.on_disconnected = [on_disconnected, user_data]() {
            on_disconnected(user_data);
        };
    }
    if (on_error) {
        conn->manager.on_error = [on_error, user_data](const std::string& err) {
            on_error(err.c_str(), user_data);
        };
    }
}