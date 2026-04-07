package com.ofa.agent.sync;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * 同步记录模型 (v3.6.0)
 */
public class SyncRecord {

    private String recordId;
    private String identityId;
    private String agentId;
    private String dataType;
    private String dataKey;
    private String operation;
    private long version;
    private long timestamp;
    private Map<String, Object> oldValue;
    private Map<String, Object> newValue;
    private String checksum;
    private String status;
    private String conflictWith;
    private long resolvedAt;
    private String resolvedBy;
    private int retryCount;
    private String errorMessage;

    public SyncRecord() {
        this.retryCount = 0;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static SyncRecord fromJson(@NonNull JSONObject json) throws JSONException {
        SyncRecord record = new SyncRecord();
        record.recordId = json.optString("record_id");
        record.identityId = json.optString("identity_id");
        record.agentId = json.optString("agent_id");
        record.dataType = json.optString("data_type");
        record.dataKey = json.optString("data_key");
        record.operation = json.optString("operation");
        record.version = json.optLong("version", 0);
        record.timestamp = json.optLong("timestamp", 0);
        record.checksum = json.optString("checksum");
        record.status = json.optString("status");
        record.conflictWith = json.optString("conflict_with");
        record.resolvedAt = json.optLong("resolved_at", 0);
        record.resolvedBy = json.optString("resolved_by");
        record.retryCount = json.optInt("retry_count", 0);
        record.errorMessage = json.optString("error_message");

        // 解析值
        JSONObject oldObj = json.optJSONObject("old_value");
        if (oldObj != null) {
            record.oldValue = jsonToMap(oldObj);
        }

        JSONObject newObj = json.optJSONObject("new_value");
        if (newObj != null) {
            record.newValue = jsonToMap(newObj);
        }

        return record;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("record_id", recordId);
            json.put("identity_id", identityId);
            json.put("agent_id", agentId);
            json.put("data_type", dataType);
            json.put("data_key", dataKey);
            json.put("operation", operation);
            json.put("version", version);
            json.put("timestamp", timestamp);
            json.put("checksum", checksum);
            json.put("status", status);
            json.put("conflict_with", conflictWith);
            json.put("resolved_at", resolvedAt);
            json.put("resolved_by", resolvedBy);
            json.put("retry_count", retryCount);
            json.put("error_message", errorMessage);

            if (oldValue != null) {
                json.put("old_value", mapToJson(oldValue));
            }
            if (newValue != null) {
                json.put("new_value", mapToJson(newValue));
            }
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 转换为 Map
     */
    @NonNull
    public Map<String, Object> toMap() {
        Map<String, Object> map = new HashMap<>();
        map.put("record_id", recordId);
        map.put("data_type", dataType);
        map.put("data_key", dataKey);
        map.put("operation", operation);
        map.put("version", version);
        map.put("status", status);
        return map;
    }

    private static Map<String, Object> jsonToMap(JSONObject json) {
        Map<String, Object> map = new HashMap<>();
        for (java.util.Iterator<String> it = json.keys(); it.hasNext(); ) {
            String key = it.next();
            Object value = json.opt(key);
            if (value != JSONObject.NULL) {
                map.put(key, value);
            }
        }
        return map;
    }

    private static JSONObject mapToJson(Map<String, Object> map) {
        JSONObject json = new JSONObject();
        try {
            for (Map.Entry<String, Object> entry : map.entrySet()) {
                json.put(entry.getKey(), entry.getValue());
            }
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    // === Getter/Setter ===

    public String getRecordId() { return recordId; }
    public void setRecordId(String recordId) { this.recordId = recordId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getAgentId() { return agentId; }
    public void setAgentId(String agentId) { this.agentId = agentId; }

    public String getDataType() { return dataType; }
    public void setDataType(String dataType) { this.dataType = dataType; }

    public String getDataKey() { return dataKey; }
    public void setDataKey(String dataKey) { this.dataKey = dataKey; }

    public String getOperation() { return operation; }
    public void setOperation(String operation) { this.operation = operation; }

    public long getVersion() { return version; }
    public void setVersion(long version) { this.version = version; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    public Map<String, Object> getOldValue() { return oldValue; }
    public void setOldValue(Map<String, Object> oldValue) { this.oldValue = oldValue; }

    public Map<String, Object> getNewValue() { return newValue; }
    public void setNewValue(Map<String, Object> newValue) { this.newValue = newValue; }

    public String getChecksum() { return checksum; }
    public void setChecksum(String checksum) { this.checksum = checksum; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public String getConflictWith() { return conflictWith; }
    public void setConflictWith(String conflictWith) { this.conflictWith = conflictWith; }

    public long getResolvedAt() { return resolvedAt; }
    public void setResolvedAt(long resolvedAt) { this.resolvedAt = resolvedAt; }

    public String getResolvedBy() { return resolvedBy; }
    public void setResolvedBy(String resolvedBy) { this.resolvedBy = resolvedBy; }

    public int getRetryCount() { return retryCount; }
    public void setRetryCount(int retryCount) { this.retryCount = retryCount; }

    public String getErrorMessage() { return errorMessage; }
    public void setErrorMessage(String errorMessage) { this.errorMessage = errorMessage; }

    @NonNull
    @Override
    public String toString() {
        return "SyncRecord{" +
                "recordId='" + recordId + '\'' +
                ", dataType='" + dataType + '\'' +
                ", dataKey='" + dataKey + '\'' +
                ", operation='" + operation + '\'' +
                ", version=" + version +
                ", status='" + status + '\'' +
                '}';
    }
}