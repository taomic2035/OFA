package com.ofa.agent.sync;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * 数据版本模型 (v3.6.0)
 */
public class DataVersion {

    private String identityId;
    private String dataType;
    private String dataKey;
    private long version;
    private long updatedAt;
    private String updatedBy;
    private String checksum;
    private boolean isDeleted;

    public DataVersion() {
        this.version = 0;
        this.isDeleted = false;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static DataVersion fromJson(@NonNull JSONObject json) throws JSONException {
        DataVersion dv = new DataVersion();
        dv.identityId = json.optString("identity_id");
        dv.dataType = json.optString("data_type");
        dv.dataKey = json.optString("data_key");
        dv.version = json.optLong("version", 0);
        dv.updatedAt = json.optLong("updated_at", 0);
        dv.updatedBy = json.optString("updated_by");
        dv.checksum = json.optString("checksum");
        dv.isDeleted = json.optBoolean("is_deleted", false);
        return dv;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("identity_id", identityId);
            json.put("data_type", dataType);
            json.put("data_key", dataKey);
            json.put("version", version);
            json.put("updated_at", updatedAt);
            json.put("updated_by", updatedBy);
            json.put("checksum", checksum);
            json.put("is_deleted", isDeleted);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 获取版本键
     */
    @NonNull
    public String getVersionKey() {
        return dataType + "_" + dataKey;
    }

    /**
     * 检查是否过期
     */
    public boolean isStale(long serverVersion) {
        return version < serverVersion;
    }

    /**
     * 检查校验和是否匹配
     */
    public boolean checksumMatches(@Nullable String otherChecksum) {
        if (checksum == null || otherChecksum == null) {
            return false;
        }
        return checksum.equals(otherChecksum);
    }

    // === Getter/Setter ===

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getDataType() { return dataType; }
    public void setDataType(String dataType) { this.dataType = dataType; }

    public String getDataKey() { return dataKey; }
    public void setDataKey(String dataKey) { this.dataKey = dataKey; }

    public long getVersion() { return version; }
    public void setVersion(long version) { this.version = version; }

    public long getUpdatedAt() { return updatedAt; }
    public void setUpdatedAt(long updatedAt) { this.updatedAt = updatedAt; }

    public String getUpdatedBy() { return updatedBy; }
    public void setUpdatedBy(String updatedBy) { this.updatedBy = updatedBy; }

    public String getChecksum() { return checksum; }
    public void setChecksum(String checksum) { this.checksum = checksum; }

    public boolean isDeleted() { return isDeleted; }
    public void setDeleted(boolean deleted) { isDeleted = deleted; }

    @NonNull
    @Override
    public String toString() {
        return "DataVersion{" +
                "dataType='" + dataType + '\'' +
                ", dataKey='" + dataKey + '\'' +
                ", version=" + version +
                ", checksum='" + checksum + '\'' +
                '}';
    }
}