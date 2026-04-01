/**
 * @file discovery.cpp
 * @brief OFA P2P 设备发现实现 - 基于 OpenHarmony 分布式软总线
 */

#include "ofa/agent.h"
#include <string>
#include <vector>
#include <unordered_map>
#include <mutex>
#include <atomic>
#include <thread>
#include <chrono>
#include <functional>

namespace ofa {

// 设备信息
struct PeerDevice {
    std::string id;
    std::string name;
    std::string type;      // phone, tablet, watch, tv, etc.
    std::string address;
    bool online;
    int64_t last_seen;
    int latency_ms;
};

// 设备发现实现
struct P2PDiscoveryImpl {
    std::atomic<bool> running{false};
    std::atomic<bool> discovering{false};

    std::unordered_map<std::string, PeerDevice> devices;
    std::mutex devices_mutex;

    // 回调
    OFA_PeerCallback peer_callback = nullptr;
    void* peer_user_data = nullptr;

    // 发现线程
    std::thread discovery_thread;
    std::thread heartbeat_thread;

    int discovery_interval_ms = 5000;
    int heartbeat_interval_ms = 10000;

    P2PDiscoveryImpl() = default;

    ~P2PDiscoveryImpl() {
        stop();
    }

    void startDiscovery() {
        if (discovering) return;
        discovering = true;
        running = true;

        discovery_thread = std::thread([this]() {
            while (running && discovering) {
                performDiscovery();
                std::this_thread::sleep_for(
                    std::chrono::milliseconds(discovery_interval_ms)
                );
            }
        });

        heartbeat_thread = std::thread([this]() {
            while (running) {
                checkDeviceStatus();
                std::this_thread::sleep_for(
                    std::chrono::milliseconds(heartbeat_interval_ms)
                );
            }
        });
    }

    void stopDiscovery() {
        discovering = false;
        if (discovery_thread.joinable()) {
            discovery_thread.join();
        }
    }

    void stop() {
        running = false;
        discovering = false;
        if (discovery_thread.joinable()) {
            discovery_thread.join();
        }
        if (heartbeat_thread.joinable()) {
            heartbeat_thread.join();
        }
    }

    void performDiscovery() {
        // TODO: 实际设备发现
        // 在 OpenHarmony 中使用 DistributedSoftBus API:
        // - SoftBusDetectDevice()
        // - SoftBusSubscribeDeviceFound()

        // 模拟发现设备 (用于开发测试)
        std::vector<PeerDevice> found_devices;

        {
            std::lock_guard<std::mutex> lock(devices_mutex);

            // 模拟发现新设备
            if (devices.size() < 5) {
                PeerDevice new_device;
                new_device.id = "device-" + std::to_string(devices.size());
                new_device.name = "OpenHarmony Device " + std::to_string(devices.size());
                new_device.type = "phone";
                new_device.address = "192.168.1." + std::to_string(100 + devices.size());
                new_device.online = true;
                new_device.last_seen = std::chrono::duration_cast<std::chrono::milliseconds>(
                    std::chrono::system_clock::now().time_since_epoch()
                ).count();
                new_device.latency_ms = 10 + devices.size() * 5;

                devices[new_device.id] = new_device;
                found_devices.push_back(new_device);
            }
        }

        // 通知回调
        for (const auto& device : found_devices) {
            if (peer_callback) {
                peer_callback(device.id.c_str(), device.name.c_str(), true, peer_user_data);
            }
        }
    }

    void checkDeviceStatus() {
        int64_t now = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();

        std::vector<std::string> offline_devices;

        {
            std::lock_guard<std::mutex> lock(devices_mutex);
            for (auto& pair : devices) {
                auto& device = pair.second;
                if (device.online && now - device.last_seen > heartbeat_interval_ms * 3) {
                    device.online = false;
                    offline_devices.push_back(device.id);
                }
            }
        }

        // 通知设备离线
        for (const auto& id : offline_devices) {
            if (peer_callback) {
                std::lock_guard<std::mutex> lock(devices_mutex);
                auto it = devices.find(id);
                if (it != devices.end()) {
                    peer_callback(id.c_str(), it->second.name.c_str(), false, peer_user_data);
                }
            }
        }
    }

    std::vector<PeerDevice> listDevices() {
        std::vector<PeerDevice> result;
        std::lock_guard<std::mutex> lock(devices_mutex);
        for (const auto& pair : devices) {
            if (pair.second.online) {
                result.push_back(pair.second);
            }
        }
        return result;
    }

    PeerDevice* getDevice(const std::string& id) {
        std::lock_guard<std::mutex> lock(devices_mutex);
        auto it = devices.find(id);
        if (it != devices.end()) {
            return &it->second;
        }
        return nullptr;
    }

    void setPeerCallback(OFA_PeerCallback callback, void* user_data) {
        peer_callback = callback;
        peer_user_data = user_data;
    }
};

} // namespace ofa

// C 接口实现

struct OFA_P2PDiscovery {
    ofa::P2PDiscoveryImpl impl;
};

OFA_P2PDiscovery* OFA_P2PDiscovery_Create() {
    return new OFA_P2PDiscovery();
}

void OFA_P2PDiscovery_Destroy(OFA_P2PDiscovery* discovery) {
    if (discovery) {
        discovery->impl.stop();
        delete discovery;
    }
}

OFA_Result OFA_P2PDiscovery_Start(
    OFA_P2PDiscovery* discovery,
    OFA_PeerCallback callback,
    void* user_data
) {
    if (!discovery) return OFA_ERROR;

    discovery->impl.setPeerCallback(callback, user_data);
    discovery->impl.startDiscovery();

    return OFA_OK;
}

void OFA_P2PDiscovery_Stop(OFA_P2PDiscovery* discovery) {
    if (discovery) {
        discovery->impl.stopDiscovery();
    }
}

size_t OFA_P2PDiscovery_GetDeviceCount(const OFA_P2PDiscovery* discovery) {
    if (!discovery) return 0;
    return discovery->impl.listDevices().size();
}

OFA_Result OFA_P2PDiscovery_GetDevices(
    OFA_P2PDiscovery* discovery,
    OFA_PeerInfo** devices,
    size_t* count
) {
    if (!discovery || !devices || !count) return OFA_ERROR;

    auto device_list = discovery->impl.listDevices();
    *count = device_list.size();

    if (device_list.empty()) {
        *devices = nullptr;
        return OFA_OK;
    }

    *devices = static_cast<OFA_PeerInfo*>(malloc(sizeof(OFA_PeerInfo) * device_list.size()));
    for (size_t i = 0; i < device_list.size(); i++) {
        (*devices)[i].id = strdup(device_list[i].id.c_str());
        (*devices)[i].name = strdup(device_list[i].name.c_str());
        (*devices)[i].type = strdup(device_list[i].type.c_str());
        (*devices)[i].online = device_list[i].online;
        (*devices)[i].latency_ms = device_list[i].latency_ms;
    }

    return OFA_OK;
}

OFA_Result OFA_P2PDiscovery_SendToDevice(
    OFA_P2PDiscovery* discovery,
    const char* device_id,
    const uint8_t* data,
    size_t data_len
) {
    if (!discovery || !device_id || !data) return OFA_ERROR;

    // TODO: 实际发送逻辑
    // 在 OpenHarmony 中使用 SoftBusSendData()

    return OFA_OK;
}

void OFA_P2PDiscovery_FreeDevices(OFA_PeerInfo* devices, size_t count) {
    if (devices) {
        for (size_t i = 0; i < count; i++) {
            if (devices[i].id) free(const_cast<char*>(devices[i].id));
            if (devices[i].name) free(const_cast<char*>(devices[i].name));
            if (devices[i].type) free(const_cast<char*>(devices[i].type));
        }
        free(devices);
    }
}