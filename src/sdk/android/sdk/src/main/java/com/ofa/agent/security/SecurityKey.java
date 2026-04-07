package com.ofa.agent.security;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * 安全密钥模型 (v3.7.0)
 */
public class SecurityKey {

    private String keyId;
    private String keyType;
    private String algorithm;
    private byte[] keyData;
    private String identityId;
    private String agentId;
    private long createdAt;
    private long expiresAt;
    private boolean isActive;
    private Map<String, String> metadata;

    public SecurityKey() {
        this.isActive = true;
        this.metadata = new HashMap<>();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static SecurityKey fromJson(@NonNull JSONObject json) throws JSONException {
        SecurityKey key = new SecurityKey();
        key.keyId = json.optString("key_id");
        key.keyType = json.optString("key_type");
        key.algorithm = json.optString("algorithm");
        key.identityId = json.optString("identity_id");
        key.agentId = json.optString("agent_id");
        key.createdAt = json.optLong("created_at", 0);
        key.expiresAt = json.optLong("expires_at", 0);
        key.isActive = json.optBoolean("is_active", true);

        // 解析元数据
        JSONObject metaObj = json.optJSONObject("metadata");
        if (metaObj != null) {
            for (java.util.Iterator<String> it = metaObj.keys(); it.hasNext(); ) {
                String k = it.next();
                key.metadata.put(k, metaObj.optString(k));
            }
        }

        return key;
    }

    /**
     * 转换为 JSON (不含密钥数据)
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("key_id", keyId);
            json.put("key_type", keyType);
            json.put("algorithm", algorithm);
            json.put("identity_id", identityId);
            json.put("agent_id", agentId);
            json.put("created_at", createdAt);
            json.put("expires_at", expiresAt);
            json.put("is_active", isActive);

            if (!metadata.isEmpty()) {
                JSONObject metaObj = new JSONObject();
                for (Map.Entry<String, String> entry : metadata.entrySet()) {
                    metaObj.put(entry.getKey(), entry.getValue());
                }
                json.put("metadata", metaObj);
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
        if (expiresAt == 0) {
            return false;
        }
        return System.currentTimeMillis() > expiresAt;
    }

    /**
     * 检查是否有效
     */
    public boolean isValid() {
        return isActive && !isExpired();
    }

    // === Getter/Setter ===

    public String getKeyId() { return keyId; }
    public void setKeyId(String keyId) { this.keyId = keyId; }

    public String getKeyType() { return keyType; }
    public void setKeyType(String keyType) { this.keyType = keyType; }

    public String getAlgorithm() { return algorithm; }
    public void setAlgorithm(String algorithm) { this.algorithm = algorithm; }

    public byte[] getKeyData() { return keyData; }
    public void setKeyData(byte[] keyData) { this.keyData = keyData; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getAgentId() { return agentId; }
    public void setAgentId(String agentId) { this.agentId = agentId; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public long getExpiresAt() { return expiresAt; }
    public void setExpiresAt(long expiresAt) { this.expiresAt = expiresAt; }

    public boolean isActive() { return isActive; }
    public void setActive(boolean active) { isActive = active; }

    public Map<String, String> getMetadata() { return metadata; }
    public void setMetadata(Map<String, String> metadata) { this.metadata = metadata; }

    @NonNull
    @Override
    public String toString() {
        return "SecurityKey{" +
                "keyId='" + keyId + '\'' +
                ", keyType='" + keyType + '\'' +
                ", algorithm='" + algorithm + '\'' +
                ", isActive=" + isActive +
                '}';
    }
}