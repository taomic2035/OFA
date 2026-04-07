package com.ofa.agent.state;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.Map;

/**
 * 状态变更事件模型 (v3.1.0)
 */
public class StateChange {

    private String agentId;
    private String identityId;
    private String changeType;
    private DeviceState oldState;
    private DeviceState newState;
    private long timestamp;
    private long version;

    public StateChange() {}

    /**
     * 创建状态变更事件
     */
    @NonNull
    public static StateChange create(@NonNull String agentId, @NonNull String identityId,
                                     @NonNull String changeType,
                                     @Nullable DeviceState oldState,
                                     @NonNull DeviceState newState) {
        StateChange change = new StateChange();
        change.agentId = agentId;
        change.identityId = identityId;
        change.changeType = changeType;
        change.oldState = oldState;
        change.newState = newState;
        change.timestamp = System.currentTimeMillis();
        return change;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static StateChange fromJson(@NonNull JSONObject json) throws JSONException {
        StateChange change = new StateChange();

        change.agentId = json.optString("agent_id");
        change.identityId = json.optString("identity_id");
        change.changeType = json.optString("change_type");
        change.timestamp = json.optLong("timestamp", System.currentTimeMillis());
        change.version = json.optLong("version", 0);

        JSONObject oldStateJson = json.optJSONObject("old_state");
        if (oldStateJson != null) {
            change.oldState = DeviceState.fromJson(oldStateJson);
        }

        JSONObject newStateJson = json.optJSONObject("new_state");
        if (newStateJson != null) {
            change.newState = DeviceState.fromJson(newStateJson);
        }

        return change;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();

        json.put("agent_id", agentId);
        json.put("identity_id", identityId);
        json.put("change_type", changeType);
        json.put("timestamp", timestamp);
        json.put("version", version);

        if (oldState != null) {
            json.put("old_state", oldState.toJson());
        }

        if (newState != null) {
            json.put("new_state", newState.toJson());
        }

        return json;
    }

    /**
     * 转换为 Map（用于消息传递）
     */
    @NonNull
    public Map<String, Object> toMap() {
        java.util.HashMap<String, Object> map = new java.util.HashMap<>();

        map.put("agent_id", agentId);
        map.put("identity_id", identityId);
        map.put("change_type", changeType);
        map.put("timestamp", timestamp);
        map.put("version", version);

        if (newState != null) {
            try {
                map.put("new_state", newState.toJson());
            } catch (JSONException e) {
                // ignore
            }
        }

        if (oldState != null) {
            try {
                map.put("old_state", oldState.toJson());
            } catch (JSONException e) {
                // ignore
            }
        }

        return map;
    }

    // Getter/Setter

    public String getAgentId() { return agentId; }
    public void setAgentId(String agentId) { this.agentId = agentId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getChangeType() { return changeType; }
    public void setChangeType(String changeType) { this.changeType = changeType; }

    public DeviceState getOldState() { return oldState; }
    public void setOldState(DeviceState oldState) { this.oldState = oldState; }

    public DeviceState getNewState() { return newState; }
    public void setNewState(DeviceState newState) { this.newState = newState; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    public long getVersion() { return version; }
    public void setVersion(long version) { this.version = version; }

    // 辅助方法

    /**
     * 检查是否是上线变更
     */
    public boolean isOnlineChange() {
        return DeviceState.CHANGE_ONLINE.equals(changeType);
    }

    /**
     * 检查是否是离线变更
     */
    public boolean isOfflineChange() {
        return DeviceState.CHANGE_OFFLINE.equals(changeType);
    }

    /**
     * 检查是否是电池变更
     */
    public boolean isBatteryChange() {
        return DeviceState.CHANGE_BATTERY.equals(changeType);
    }

    /**
     * 检查是否是网络变更
     */
    public boolean isNetworkChange() {
        return DeviceState.CHANGE_NETWORK.equals(changeType);
    }

    /**
     * 检查是否是场景变更
     */
    public boolean isSceneChange() {
        return DeviceState.CHANGE_SCENE.equals(changeType);
    }

    @NonNull
    @Override
    public String toString() {
        return "StateChange{" +
                "agentId='" + agentId + '\'' +
                ", changeType='" + changeType + '\'' +
                ", timestamp=" + timestamp +
                '}';
    }
}