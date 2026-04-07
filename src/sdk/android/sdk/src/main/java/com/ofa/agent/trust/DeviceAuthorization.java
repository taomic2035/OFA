package com.ofa.agent.trust;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * 设备授权信息 (v2.8.0)
 *
 * 表示设备从 Center 获得的授权，包含信任级别和权限列表。
 */
public class DeviceAuthorization {

    private String agentId;
    private String identityId;
    private TrustLevel trustLevel;
    private List<String> permissions;
    private long grantedAt;
    private String grantedBy;
    private Long expiresAt;
    private long lastVerified;
    private int verificationCount;

    public DeviceAuthorization() {
        this.permissions = new ArrayList<>();
        this.trustLevel = TrustLevel.NONE;
    }

    // Getters
    public String getAgentId() {
        return agentId;
    }

    public String getIdentityId() {
        return identityId;
    }

    public TrustLevel getTrustLevel() {
        return trustLevel;
    }

    public List<String> getPermissions() {
        return permissions;
    }

    public long getGrantedAt() {
        return grantedAt;
    }

    public String getGrantedBy() {
        return grantedBy;
    }

    public Long getExpiresAt() {
        return expiresAt;
    }

    public long getLastVerified() {
        return lastVerified;
    }

    public int getVerificationCount() {
        return verificationCount;
    }

    // Setters
    public void setAgentId(String agentId) {
        this.agentId = agentId;
    }

    public void setIdentityId(String identityId) {
        this.identityId = identityId;
    }

    public void setTrustLevel(TrustLevel trustLevel) {
        this.trustLevel = trustLevel;
    }

    public void setPermissions(List<String> permissions) {
        this.permissions = permissions != null ? permissions : new ArrayList<>();
    }

    public void setGrantedAt(long grantedAt) {
        this.grantedAt = grantedAt;
    }

    public void setGrantedBy(String grantedBy) {
        this.grantedBy = grantedBy;
    }

    public void setExpiresAt(Long expiresAt) {
        this.expiresAt = expiresAt;
    }

    public void setLastVerified(long lastVerified) {
        this.lastVerified = lastVerified;
    }

    public void setVerificationCount(int verificationCount) {
        this.verificationCount = verificationCount;
    }

    /**
     * 检查授权是否过期
     */
    public boolean isExpired() {
        if (expiresAt == null) {
            return false;
        }
        return System.currentTimeMillis() > expiresAt;
    }

    /**
     * 检查是否有指定权限
     */
    public boolean hasPermission(String permission) {
        // 先检查权限列表
        for (String p : permissions) {
            if (p.equals(permission) || p.equals("admin")) {
                return true;
            }
        }
        // 再检查信任级别
        return trustLevel.hasPermission(permission);
    }

    /**
     * 检查是否为主设备
     */
    public boolean isPrimary() {
        return trustLevel == TrustLevel.PRIMARY;
    }

    /**
     * 检查是否可以写入
     */
    public boolean canWrite() {
        return hasPermission("write");
    }

    /**
     * 检查是否可以同步
     */
    public boolean canSync() {
        return hasPermission("sync");
    }

    /**
     * 检查是否可以管理
     */
    public boolean canAdmin() {
        return hasPermission("admin");
    }

    // JSON 序列化
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("agent_id", agentId);
        json.put("identity_id", identityId);
        json.put("trust_level", trustLevel.getName());
        json.put("permissions", new JSONArray(permissions));
        json.put("granted_at", grantedAt);
        json.put("granted_by", grantedBy);
        if (expiresAt != null) {
            json.put("expires_at", expiresAt);
        }
        json.put("last_verified", lastVerified);
        json.put("verification_count", verificationCount);
        return json;
    }

    // JSON 反序列化
    public static DeviceAuthorization fromJson(JSONObject json) throws JSONException {
        DeviceAuthorization auth = new DeviceAuthorization();
        auth.setAgentId(json.optString("agent_id"));
        auth.setIdentityId(json.optString("identity_id"));
        auth.setTrustLevel(TrustLevel.fromName(json.optString("trust_level")));
        auth.setGrantedAt(json.optLong("granted_at"));
        auth.setGrantedBy(json.optString("granted_by"));
        if (json.has("expires_at")) {
            auth.setExpiresAt(json.getLong("expires_at"));
        }
        auth.setLastVerified(json.optLong("last_verified"));
        auth.setVerificationCount(json.optInt("verification_count"));

        // 解析权限列表
        List<String> perms = new ArrayList<>();
        JSONArray permsArray = json.optJSONArray("permissions");
        if (permsArray != null) {
            for (int i = 0; i < permsArray.length(); i++) {
                perms.add(permsArray.getString(i));
            }
        }
        auth.setPermissions(perms);

        return auth;
    }

    @Override
    public String toString() {
        return "DeviceAuthorization{" +
                "agentId='" + agentId + '\'' +
                ", identityId='" + identityId + '\'' +
                ", trustLevel=" + trustLevel +
                ", permissions=" + permissions +
                '}';
    }
}