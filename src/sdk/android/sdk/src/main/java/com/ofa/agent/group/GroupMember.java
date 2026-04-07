package com.ofa.agent.group;

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
 * 群组成员模型 (v3.5.0)
 */
public class GroupMember {

    // 角色类型
    public static final String ROLE_OWNER = "owner";
    public static final String ROLE_ADMIN = "admin";
    public static final String ROLE_MEMBER = "member";
    public static final String ROLE_GUEST = "guest";

    private String agentId;
    private String deviceName;
    private String deviceType;
    private String role;
    private long joinedAt;
    private long lastActiveAt;
    private boolean online;
    private int priority;
    private List<String> capabilities;

    public GroupMember() {
        this.role = ROLE_MEMBER;
        this.online = true;
        this.priority = 50;
        this.capabilities = new ArrayList<>();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static GroupMember fromJson(@NonNull JSONObject json) throws JSONException {
        GroupMember member = new GroupMember();
        member.agentId = json.optString("agent_id");
        member.deviceName = json.optString("device_name");
        member.deviceType = json.optString("device_type");
        member.role = json.optString("role", ROLE_MEMBER);
        member.joinedAt = json.optLong("joined_at", 0);
        member.lastActiveAt = json.optLong("last_active_at", 0);
        member.online = json.optBoolean("is_online", true);
        member.priority = json.optInt("priority", 50);

        // 解析能力
        JSONArray capsArray = json.optJSONArray("capabilities");
        if (capsArray != null) {
            for (int i = 0; i < capsArray.length(); i++) {
                member.capabilities.add(capsArray.getString(i));
            }
        }

        return member;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("agent_id", agentId);
        json.put("device_name", deviceName);
        json.put("device_type", deviceType);
        json.put("role", role);
        json.put("joined_at", joinedAt);
        json.put("last_active_at", lastActiveAt);
        json.put("is_online", online);
        json.put("priority", priority);

        if (!capabilities.isEmpty()) {
            JSONArray capsArray = new JSONArray();
            for (String cap : capabilities) {
                capsArray.put(cap);
            }
            json.put("capabilities", capsArray);
        }

        return json;
    }

    /**
     * 转换为 Map
     */
    @NonNull
    public Map<String, Object> toMap() {
        Map<String, Object> map = new HashMap<>();
        map.put("agent_id", agentId);
        map.put("device_name", deviceName);
        map.put("device_type", deviceType);
        map.put("role", role);
        map.put("is_online", online);
        map.put("priority", priority);
        return map;
    }

    // === 角色检查 ===

    /**
     * 是否为群主
     */
    public boolean isOwner() {
        return ROLE_OWNER.equals(role);
    }

    /**
     * 是否为管理员
     */
    public boolean isAdmin() {
        return ROLE_ADMIN.equals(role) || isOwner();
    }

    /**
     * 是否为访客
     */
    public boolean isGuest() {
        return ROLE_GUEST.equals(role);
    }

    /**
     * 检查是否有能力
     */
    public boolean hasCapability(@NonNull String capability) {
        return capabilities.contains(capability);
    }

    // === Getter/Setter ===

    public String getAgentId() { return agentId; }
    public void setAgentId(String agentId) { this.agentId = agentId; }

    public String getDeviceName() { return deviceName; }
    public void setDeviceName(String deviceName) { this.deviceName = deviceName; }

    public String getDeviceType() { return deviceType; }
    public void setDeviceType(String deviceType) { this.deviceType = deviceType; }

    public String getRole() { return role; }
    public void setRole(String role) { this.role = role; }

    public long getJoinedAt() { return joinedAt; }
    public void setJoinedAt(long joinedAt) { this.joinedAt = joinedAt; }

    public long getLastActiveAt() { return lastActiveAt; }
    public void setLastActiveAt(long lastActiveAt) { this.lastActiveAt = lastActiveAt; }

    public boolean isOnline() { return online; }
    public void setOnline(boolean online) { this.online = online; }

    public int getPriority() { return priority; }
    public void setPriority(int priority) { this.priority = priority; }

    public List<String> getCapabilities() { return capabilities; }
    public void setCapabilities(List<String> capabilities) { this.capabilities = capabilities; }

    public void addCapability(@NonNull String capability) {
        if (!capabilities.contains(capability)) {
            capabilities.add(capability);
        }
    }

    public void removeCapability(@NonNull String capability) {
        capabilities.remove(capability);
    }

    @NonNull
    @Override
    public String toString() {
        return "GroupMember{" +
                "agentId='" + agentId + '\'' +
                ", deviceName='" + deviceName + '\'' +
                ", role='" + role + '\'' +
                ", online=" + online +
                '}';
    }
}