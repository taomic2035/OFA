package com.ofa.agent.security;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * 安全通道模型 (v3.7.0)
 */
public class SecureChannel {

    private String channelId;
    private String identityId;
    private String sourceAgent;
    private String targetAgent;
    private String channelKeyId;
    private String securityLevel;
    private long createdAt;
    private long expiresAt;
    private boolean isActive;
    private int messageCount;

    public SecureChannel() {
        this.isActive = true;
        this.messageCount = 0;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static SecureChannel fromJson(@NonNull JSONObject json) throws JSONException {
        SecureChannel channel = new SecureChannel();
        channel.channelId = json.optString("channel_id");
        channel.identityId = json.optString("identity_id");
        channel.sourceAgent = json.optString("source_agent");
        channel.targetAgent = json.optString("target_agent");
        channel.channelKeyId = json.optString("channel_key_id");
        channel.securityLevel = json.optString("security_level");
        channel.createdAt = json.optLong("created_at", 0);
        channel.expiresAt = json.optLong("expires_at", 0);
        channel.isActive = json.optBoolean("is_active", true);
        channel.messageCount = json.optInt("message_count", 0);

        return channel;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("channel_id", channelId);
            json.put("identity_id", identityId);
            json.put("source_agent", sourceAgent);
            json.put("target_agent", targetAgent);
            json.put("channel_key_id", channelKeyId);
            json.put("security_level", securityLevel);
            json.put("created_at", createdAt);
            json.put("expires_at", expiresAt);
            json.put("is_active", isActive);
            json.put("message_count", messageCount);
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

    /**
     * 增加消息计数
     */
    public void incrementMessageCount() {
        this.messageCount++;
    }

    // === Getter/Setter ===

    public String getChannelId() { return channelId; }
    public void setChannelId(String channelId) { this.channelId = channelId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getSourceAgent() { return sourceAgent; }
    public void setSourceAgent(String sourceAgent) { this.sourceAgent = sourceAgent; }

    public String getTargetAgent() { return targetAgent; }
    public void setTargetAgent(String targetAgent) { this.targetAgent = targetAgent; }

    public String getChannelKeyId() { return channelKeyId; }
    public void setChannelKeyId(String channelKeyId) { this.channelKeyId = channelKeyId; }

    public String getSecurityLevel() { return securityLevel; }
    public void setSecurityLevel(String securityLevel) { this.securityLevel = securityLevel; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public long getExpiresAt() { return expiresAt; }
    public void setExpiresAt(long expiresAt) { this.expiresAt = expiresAt; }

    public boolean isActive() { return isActive; }
    public void setActive(boolean active) { isActive = active; }

    public int getMessageCount() { return messageCount; }
    public void setMessageCount(int messageCount) { this.messageCount = messageCount; }

    @NonNull
    @Override
    public String toString() {
        return "SecureChannel{" +
                "channelId='" + channelId + '\'' +
                ", source='" + sourceAgent + '\'' +
                ", target='" + targetAgent + '\'' +
                ", messages=" + messageCount +
                '}';
    }
}