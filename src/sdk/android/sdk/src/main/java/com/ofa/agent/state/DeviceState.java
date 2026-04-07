package com.ofa.agent.state;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 设备状态模型 (v3.1.0)
 *
 * 表示设备的完整状态信息，包括：
 * - 连接状态
 * - 电源状态
 * - 网络状态
 * - 场景状态
 * - 能力状态
 */
public class DeviceState {

    // === 连接状态 ===
    private String agentId;
    private String identityId;
    private String deviceType;     // mobile, tablet, watch, iot
    private String deviceName;
    private boolean online;
    private long lastSeen;

    // === 电源状态 ===
    private int batteryLevel;      // 0-100
    private boolean charging;
    private boolean powerSaver;

    // === 网络状态 ===
    private String networkType;    // wifi, cellular, offline
    private int networkStrength;   // 0-100
    private boolean roaming;

    // === 场景状态 ===
    private String scene;          // running, walking, driving, meeting, idle
    private Map<String, Object> sceneContext;

    // === 能力状态 ===
    private List<String> capabilities;
    private List<String> activeApps;

    // === 位置信息 ===
    private DeviceLocation location;

    // === 优先级与信任 ===
    private String trustLevel;
    private int priority;

    // === 状态时间戳 ===
    private long stateVersion;
    private long updatedAt;

    // === 扩展状态 ===
    private Map<String, Object> metadata;

    // === 状态变更类型 ===
    public static final String CHANGE_ONLINE = "online";
    public static final String CHANGE_OFFLINE = "offline";
    public static final String CHANGE_BATTERY = "battery";
    public static final String CHANGE_NETWORK = "network";
    public static final String CHANGE_SCENE = "scene";
    public static final String CHANGE_LOCATION = "location";
    public static final String CHANGE_CAPABILITIES = "capabilities";
    public static final String CHANGE_FULL = "full";

    public DeviceState() {
        this.capabilities = new ArrayList<>();
        this.activeApps = new ArrayList<>();
        this.sceneContext = new HashMap<>();
        this.metadata = new HashMap<>();
        this.scene = "idle";
        this.networkType = "unknown";
        this.trustLevel = "low";
        this.priority = 50;
    }

    // === 工厂方法 ===

    /**
     * 创建基础设备状态
     */
    @NonNull
    public static DeviceState create(@NonNull String agentId, @NonNull String identityId,
                                     @NonNull String deviceType, @NonNull String deviceName) {
        DeviceState state = new DeviceState();
        state.agentId = agentId;
        state.identityId = identityId;
        state.deviceType = deviceType;
        state.deviceName = deviceName;
        state.online = true;
        state.lastSeen = System.currentTimeMillis();
        state.updatedAt = System.currentTimeMillis();
        return state;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static DeviceState fromJson(@NonNull JSONObject json) throws JSONException {
        DeviceState state = new DeviceState();

        state.agentId = json.optString("agent_id");
        state.identityId = json.optString("identity_id");
        state.deviceType = json.optString("device_type");
        state.deviceName = json.optString("device_name");
        state.online = json.optBoolean("online", false);
        state.lastSeen = json.optLong("last_seen", System.currentTimeMillis());
        state.batteryLevel = json.optInt("battery_level", 100);
        state.charging = json.optBoolean("charging", false);
        state.powerSaver = json.optBoolean("power_saver", false);
        state.networkType = json.optString("network_type", "unknown");
        state.networkStrength = json.optInt("network_strength", 0);
        state.roaming = json.optBoolean("roaming", false);
        state.scene = json.optString("scene", "idle");
        state.trustLevel = json.optString("trust_level", "low");
        state.priority = json.optInt("priority", 50);
        state.stateVersion = json.optLong("state_version", 0);
        state.updatedAt = json.optLong("updated_at", System.currentTimeMillis());

        // 解析 sceneContext
        JSONObject sceneContextJson = json.optJSONObject("scene_context");
        if (sceneContextJson != null) {
            state.sceneContext = new HashMap<>();
            for (java.util.Iterator<String> it = sceneContextJson.keys(); it.hasNext(); ) {
                String key = it.next();
                state.sceneContext.put(key, sceneContextJson.get(key));
            }
        }

        // 解析 capabilities
        JSONArray capabilitiesArray = json.optJSONArray("capabilities");
        if (capabilitiesArray != null) {
            state.capabilities = new ArrayList<>();
            for (int i = 0; i < capabilitiesArray.length(); i++) {
                state.capabilities.add(capabilitiesArray.getString(i));
            }
        }

        // 解析 activeApps
        JSONArray activeAppsArray = json.optJSONArray("active_apps");
        if (activeAppsArray != null) {
            state.activeApps = new ArrayList<>();
            for (int i = 0; i < activeAppsArray.length(); i++) {
                state.activeApps.add(activeAppsArray.getString(i));
            }
        }

        // 解析 location
        JSONObject locationJson = json.optJSONObject("location");
        if (locationJson != null) {
            state.location = DeviceLocation.fromJson(locationJson);
        }

        // 解析 metadata
        JSONObject metadataJson = json.optJSONObject("metadata");
        if (metadataJson != null) {
            state.metadata = new HashMap<>();
            for (java.util.Iterator<String> it = metadataJson.keys(); it.hasNext(); ) {
                String key = it.next();
                state.metadata.put(key, metadataJson.get(key));
            }
        }

        return state;
    }

    // === JSON 序列化 ===

    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();

        json.put("agent_id", agentId);
        json.put("identity_id", identityId);
        json.put("device_type", deviceType);
        json.put("device_name", deviceName);
        json.put("online", online);
        json.put("last_seen", lastSeen);
        json.put("battery_level", batteryLevel);
        json.put("charging", charging);
        json.put("power_saver", powerSaver);
        json.put("network_type", networkType);
        json.put("network_strength", networkStrength);
        json.put("roaming", roaming);
        json.put("scene", scene);
        json.put("trust_level", trustLevel);
        json.put("priority", priority);
        json.put("state_version", stateVersion);
        json.put("updated_at", updatedAt);

        if (sceneContext != null && !sceneContext.isEmpty()) {
            JSONObject sceneContextJson = new JSONObject();
            for (Map.Entry<String, Object> e : sceneContext.entrySet()) {
                sceneContextJson.put(e.getKey(), e.getValue());
            }
            json.put("scene_context", sceneContextJson);
        }

        if (capabilities != null && !capabilities.isEmpty()) {
            json.put("capabilities", new JSONArray(capabilities));
        }

        if (activeApps != null && !activeApps.isEmpty()) {
            json.put("active_apps", new JSONArray(activeApps));
        }

        if (location != null) {
            json.put("location", location.toJson());
        }

        if (metadata != null && !metadata.isEmpty()) {
            JSONObject metadataJson = new JSONObject();
            for (Map.Entry<String, Object> e : metadata.entrySet()) {
                metadataJson.put(e.getKey(), e.getValue());
            }
            json.put("metadata", metadataJson);
        }

        return json;
    }

    // === 状态变更检测 ===

    /**
     * 检测与旧状态的变更类型
     */
    @NonNull
    public String detectChangeType(@Nullable DeviceState oldState) {
        if (oldState == null) {
            return CHANGE_FULL;
        }

        // 上线/离线变更
        if (oldState.online != online) {
            return online ? CHANGE_ONLINE : CHANGE_OFFLINE;
        }

        // 电池变更
        if (oldState.batteryLevel != batteryLevel ||
            oldState.charging != charging ||
            oldState.powerSaver != powerSaver) {
            return CHANGE_BATTERY;
        }

        // 网络变更
        if (!oldState.networkType.equals(networkType) ||
            oldState.networkStrength != networkStrength) {
            return CHANGE_NETWORK;
        }

        // 场景变更
        if (!oldState.scene.equals(scene)) {
            return CHANGE_SCENE;
        }

        // 能力变更
        if (!equalLists(oldState.capabilities, capabilities)) {
            return CHANGE_CAPABILITIES;
        }

        return CHANGE_FULL;
    }

    private boolean equalLists(List<String> a, List<String> b) {
        if (a == null && b == null) return true;
        if (a == null || b == null) return false;
        if (a.size() != b.size()) return false;
        for (int i = 0; i < a.size(); i++) {
            if (!a.get(i).equals(b.get(i))) return false;
        }
        return true;
    }

    // === Getter/Setter ===

    public String getAgentId() { return agentId; }
    public void setAgentId(String agentId) { this.agentId = agentId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getDeviceType() { return deviceType; }
    public void setDeviceType(String deviceType) { this.deviceType = deviceType; }

    public String getDeviceName() { return deviceName; }
    public void setDeviceName(String deviceName) { this.deviceName = deviceName; }

    public boolean isOnline() { return online; }
    public void setOnline(boolean online) { this.online = online; }

    public long getLastSeen() { return lastSeen; }
    public void setLastSeen(long lastSeen) { this.lastSeen = lastSeen; }

    public int getBatteryLevel() { return batteryLevel; }
    public void setBatteryLevel(int batteryLevel) {
        if (batteryLevel < 0) batteryLevel = 0;
        if (batteryLevel > 100) batteryLevel = 100;
        this.batteryLevel = batteryLevel;
    }

    public boolean isCharging() { return charging; }
    public void setCharging(boolean charging) { this.charging = charging; }

    public boolean isPowerSaver() { return powerSaver; }
    public void setPowerSaver(boolean powerSaver) { this.powerSaver = powerSaver; }

    public String getNetworkType() { return networkType; }
    public void setNetworkType(String networkType) { this.networkType = networkType; }

    public int getNetworkStrength() { return networkStrength; }
    public void setNetworkStrength(int networkStrength) {
        if (networkStrength < 0) networkStrength = 0;
        if (networkStrength > 100) networkStrength = 100;
        this.networkStrength = networkStrength;
    }

    public boolean isRoaming() { return roaming; }
    public void setRoaming(boolean roaming) { this.roaming = roaming; }

    public String getScene() { return scene; }
    public void setScene(String scene) { this.scene = scene; }

    public Map<String, Object> getSceneContext() { return sceneContext; }
    public void setSceneContext(Map<String, Object> sceneContext) { this.sceneContext = sceneContext; }

    public List<String> getCapabilities() { return capabilities; }
    public void setCapabilities(List<String> capabilities) { this.capabilities = capabilities; }

    public List<String> getActiveApps() { return activeApps; }
    public void setActiveApps(List<String> activeApps) { this.activeApps = activeApps; }

    public DeviceLocation getLocation() { return location; }
    public void setLocation(DeviceLocation location) { this.location = location; }

    public String getTrustLevel() { return trustLevel; }
    public void setTrustLevel(String trustLevel) { this.trustLevel = trustLevel; }

    public int getPriority() { return priority; }
    public void setPriority(int priority) {
        if (priority < 0) priority = 0;
        if (priority > 100) priority = 100;
        this.priority = priority;
    }

    public long getStateVersion() { return stateVersion; }
    public void setStateVersion(long stateVersion) { this.stateVersion = stateVersion; }

    public long getUpdatedAt() { return updatedAt; }
    public void setUpdatedAt(long updatedAt) { this.updatedAt = updatedAt; }

    public Map<String, Object> getMetadata() { return metadata; }
    public void setMetadata(Map<String, Object> metadata) { this.metadata = metadata; }

    // === 辅助方法 ===

    /**
     * 检查电池是否低电量 (< 20%)
     */
    public boolean isLowBattery() {
        return batteryLevel < 20 && !charging;
    }

    /**
     * 检查网络是否可用
     */
    public boolean hasNetwork() {
        return !networkType.equals("offline") && networkStrength > 0;
    }

    /**
     * 检查是否有 WiFi
     */
    public boolean hasWiFi() {
        return networkType.equals("wifi");
    }

    /**
     * 检查是否有蜂窝网络
     */
    public boolean hasCellular() {
        return networkType.equals("cellular");
    }

    /**
     * 检查是否处于指定场景
     */
    public boolean isInScene(@NonNull String targetScene) {
        return scene.equals(targetScene);
    }

    /**
     * 检查是否有指定能力
     */
    public boolean hasCapability(@NonNull String capability) {
        return capabilities.contains(capability);
    }

    /**
     * 添加能力
     */
    public void addCapability(@NonNull String capability) {
        if (!capabilities.contains(capability)) {
            capabilities.add(capability);
        }
    }

    /**
     * 移除能力
     */
    public void removeCapability(@NonNull String capability) {
        capabilities.remove(capability);
    }

    /**
     * 添加场景上下文
     */
    public void addSceneContext(@NonNull String key, Object value) {
        sceneContext.put(key, value);
    }

    /**
     * 添加元数据
     */
    public void addMetadata(@NonNull String key, Object value) {
        metadata.put(key, value);
    }

    @NonNull
    @Override
    public String toString() {
        return "DeviceState{" +
                "agentId='" + agentId + '\'' +
                ", deviceType='" + deviceType + '\'' +
                ", online=" + online +
                ", battery=" + batteryLevel + "%" +
                (charging ? "(charging)" : "") +
                ", network=" + networkType +
                ", scene='" + scene + '\'' +
                ", priority=" + priority +
                '}';
    }
}