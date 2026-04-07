package com.ofa.agent.sync;

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
 * 冲突记录模型 (v3.6.0)
 */
public class ConflictRecord {

    private String conflictId;
    private String identityId;
    private String dataType;
    private String dataKey;
    private DataVersion serverVersion;
    private List<SyncRecord> clientRecords;
    private String strategy;
    private String status;
    private long createdAt;
    private long resolvedAt;
    private Map<String, Object> resolvedValue;

    public ConflictRecord() {
        this.clientRecords = new ArrayList<>();
        this.resolvedValue = new HashMap<>();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static ConflictRecord fromJson(@NonNull JSONObject json) throws JSONException {
        ConflictRecord conflict = new ConflictRecord();
        conflict.conflictId = json.optString("conflict_id");
        conflict.identityId = json.optString("identity_id");
        conflict.dataType = json.optString("data_type");
        conflict.dataKey = json.optString("data_key");
        conflict.strategy = json.optString("strategy");
        conflict.status = json.optString("status");
        conflict.createdAt = json.optLong("created_at", 0);
        conflict.resolvedAt = json.optLong("resolved_at", 0);

        // 解析服务端版本
        JSONObject serverObj = json.optJSONObject("server_version");
        if (serverObj != null) {
            conflict.serverVersion = DataVersion.fromJson(serverObj);
        }

        // 解析客户端记录
        JSONArray clientArray = json.optJSONArray("client_records");
        if (clientArray != null) {
            for (int i = 0; i < clientArray.length(); i++) {
                JSONObject recordObj = clientArray.getJSONObject(i);
                conflict.clientRecords.add(SyncRecord.fromJson(recordObj));
            }
        }

        // 解析解决值
        JSONObject resolvedObj = json.optJSONObject("resolved_value");
        if (resolvedObj != null) {
            for (java.util.Iterator<String> it = resolvedObj.keys(); it.hasNext(); ) {
                String key = it.next();
                Object value = resolvedObj.opt(key);
                if (value != JSONObject.NULL) {
                    conflict.resolvedValue.put(key, value);
                }
            }
        }

        return conflict;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("conflict_id", conflictId);
            json.put("identity_id", identityId);
            json.put("data_type", dataType);
            json.put("data_key", dataKey);
            json.put("strategy", strategy);
            json.put("status", status);
            json.put("created_at", createdAt);
            json.put("resolved_at", resolvedAt);

            if (serverVersion != null) {
                json.put("server_version", serverVersion.toJson());
            }

            if (!clientRecords.isEmpty()) {
                JSONArray clientArray = new JSONArray();
                for (SyncRecord record : clientRecords) {
                    clientArray.put(record.toJson());
                }
                json.put("client_records", clientArray);
            }

            if (!resolvedValue.isEmpty()) {
                JSONObject resolvedObj = new JSONObject();
                for (Map.Entry<String, Object> entry : resolvedValue.entrySet()) {
                    resolvedObj.put(entry.getKey(), entry.getValue());
                }
                json.put("resolved_value", resolvedObj);
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
        map.put("conflict_id", conflictId);
        map.put("data_type", dataType);
        map.put("data_key", dataKey);
        map.put("strategy", strategy);
        map.put("status", status);
        map.put("client_count", clientRecords.size());
        return map;
    }

    /**
     * 是否已解决
     */
    public boolean isResolved() {
        return resolvedAt > 0;
    }

    /**
     * 获取最新的客户端记录
     */
    @Nullable
    public SyncRecord getLatestClientRecord() {
        if (clientRecords.isEmpty()) {
            return null;
        }

        SyncRecord latest = clientRecords.get(0);
        for (SyncRecord record : clientRecords) {
            if (record.getTimestamp() > latest.getTimestamp()) {
                latest = record;
            }
        }
        return latest;
    }

    /**
     * 获取客户端数量
     */
    public int getClientCount() {
        return clientRecords.size();
    }

    /**
     * 添加客户端记录
     */
    public void addClientRecord(@NonNull SyncRecord record) {
        clientRecords.add(record);
    }

    // === Getter/Setter ===

    public String getConflictId() { return conflictId; }
    public void setConflictId(String conflictId) { this.conflictId = conflictId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getDataType() { return dataType; }
    public void setDataType(String dataType) { this.dataType = dataType; }

    public String getDataKey() { return dataKey; }
    public void setDataKey(String dataKey) { this.dataKey = dataKey; }

    public DataVersion getServerVersion() { return serverVersion; }
    public void setServerVersion(DataVersion serverVersion) { this.serverVersion = serverVersion; }

    public List<SyncRecord> getClientRecords() { return clientRecords; }
    public void setClientRecords(List<SyncRecord> clientRecords) { this.clientRecords = clientRecords; }

    public String getStrategy() { return strategy; }
    public void setStrategy(String strategy) { this.strategy = strategy; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public long getResolvedAt() { return resolvedAt; }
    public void setResolvedAt(long resolvedAt) { this.resolvedAt = resolvedAt; }

    public Map<String, Object> getResolvedValue() { return resolvedValue; }
    public void setResolvedValue(Map<String, Object> resolvedValue) { this.resolvedValue = resolvedValue; }

    @NonNull
    @Override
    public String toString() {
        return "ConflictRecord{" +
                "conflictId='" + conflictId + '\'' +
                ", dataType='" + dataType + '\'' +
                ", dataKey='" + dataKey + '\'' +
                ", strategy='" + strategy + '\'' +
                ", status='" + status + '\'' +
                ", clientCount=" + clientRecords.size() +
                '}';
    }
}