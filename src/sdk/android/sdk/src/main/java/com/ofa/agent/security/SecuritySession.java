package com.ofa.agent.security;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * 安全会话模型 (v3.7.0)
 */
public class SecuritySession {

    private String sessionId;
    private String identityId;
    private String agentId;
    private String sessionKeyId;
    private String securityLevel;
    private long createdAt;
    private long expiresAt;
    private long lastActiveAt;
    private boolean isActive;
    private Map<String, String> deviceInfo;

    public SecuritySession() {
        this.isActive = true;
        this.deviceInfo = new HashMap<>();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static SecuritySession fromJson(@NonNull JSONObject json) throws JSONException {
        SecuritySession session = new SecuritySession();
        session.sessionId = json.optString("session_id");
        session.identityId = json.optString("identity_id");
        session.agentId = json.optString("agent_id");
        session.sessionKeyId = json.optString("session_key_id");
        session.securityLevel = json.optString("security_level");
        session.createdAt = json.optLong("created_at", 0);
        session.expiresAt = json.optLong("expires_at", 0);
        session.lastActiveAt = json.optLong("last_active_at", 0);
        session.isActive = json.optBoolean("is_active", true);

        // 解析设备信息
        JSONObject deviceObj = json.optJSONObject("device_info");
        if (deviceObj != null) {
            for (java.util.Iterator<String> it = deviceObj.keys(); it.hasNext(); ) {
                String k = it.next();
                session.deviceInfo.put(k, deviceObj.optString(k));
            }
        }

        return session;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("session_id", sessionId);
            json.put("identity_id", identityId);
            json.put("agent_id", agentId);
            json.put("session_key_id", sessionKeyId);
            json.put("security_level", securityLevel);
            json.put("created_at", createdAt);
            json.put("expires_at", expiresAt);
            json.put("last_active_at", lastActiveAt);
            json.put("is_active", isActive);

            if (!deviceInfo.isEmpty()) {
                JSONObject deviceObj = new JSONObject();
                for (Map.Entry<String, String> entry : deviceInfo.entrySet()) {
                    deviceObj.put(entry.getKey(), entry.getValue());
                }
                json.put("device_info", deviceObj);
            }
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 检查是否过期
     */
    public boolean isExpired() {
        return System.currentTimeMillis() > expiresAt;
    }

    /**
     * 检查是否有效
     */
    public boolean isValid() {
        return isActive && !isExpired();
    }

    // === Getter/Setter ===

    public String getSessionId() { return sessionId; }
    public void setSessionId(String sessionId) { this.sessionId = sessionId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getAgentId() { return agentId; }
    public void setAgentId(String agentId) { this.agentId = agentId; }

    public String getSessionKeyId() { return sessionKeyId; }
    public void setSessionKeyId(String sessionKeyId) { this.sessionKeyId = sessionKeyId; }

    public String getSecurityLevel() { return securityLevel; }
    public void setSecurityLevel(String securityLevel) { this.securityLevel = securityLevel; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public long getExpiresAt() { return expiresAt; }
    public void setExpiresAt(long expiresAt) { this.expiresAt = expiresAt; }

    public long getLastActiveAt() { return lastActiveAt; }
    public void setLastActiveAt(long lastActiveAt) { this.lastActiveAt = lastActiveAt; }

    public boolean isActive() { return isActive; }
    public void setActive(boolean active) { isActive = active; }

    public Map<String, String> getDeviceInfo() { return deviceInfo; }
    public void setDeviceInfo(Map<String, String> deviceInfo) { this.deviceInfo = deviceInfo; }

    @NonNull
    @Override
    public String toString() {
        return "SecuritySession{" +
                "sessionId='" + sessionId + '\'' +
                ", agentId='" + agentId + '\'' +
                ", securityLevel='" + securityLevel + '\'' +
                ", isActive=" + isActive +
                '}';
    }
}