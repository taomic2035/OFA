package com.ofa.agent.sync;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 数据同步客户端 (v3.6.0)
 *
 * 负责：
 * - 增量数据同步
 * - 冲突检测与解决
 * - 版本管理
 * - 离线队列管理
 */
public class DataSyncClient {

    private static final String TAG = "DataSyncClient";

    // 消息类型
    private static final String MSG_SYNC_RECORD = "sync_record";
    private static final String MSG_SYNC_BATCH = "sync_batch";
    private static final String MSG_SYNC_STATUS = "sync_status";
    private static final String MSG_VERSION_QUERY = "version_query";
    private static final String MSG_CONFLICT_RESOLVE = "conflict_resolve";

    // 操作类型
    public static final String OP_CREATE = "create";
    public static final String OP_UPDATE = "update";
    public static final String OP_DELETE = "delete";
    public static final String OP_MERGE = "merge";

    // 状态
    public static final String STATUS_PENDING = "pending";
    public static final String STATUS_SYNCING = "syncing";
    public static final String STATUS_COMPLETED = "completed";
    public static final String STATUS_FAILED = "failed";
    public static final String STATUS_CONFLICT = "conflict";

    // 冲突策略
    public static final String CONFLICT_LAST_WRITE = "last_write";
    public static final String CONFLICT_FIRST_WRITE = "first_write";
    public static final String CONFLICT_MERGE = "merge";
    public static final String CONFLICT_SERVER = "server";
    public static final String CONFLICT_CLIENT = "client";

    // 本地缓存
    private final Map<String, SyncRecord> recordCache;
    private final Map<String, DataVersion> versionCache;
    private final List<SyncRecord> pendingQueue;
    private final Map<String, ConflictRecord> conflicts;
    private final List<SyncListener> listeners;

    private MessageBus messageBus;
    private String agentId;
    private String identityId;

    // 配置
    private DataSyncConfig config;

    // 同步状态
    private boolean isSyncing = false;
    private String currentBatchId = null;

    /**
     * 同步监听器
     */
    public interface SyncListener {
        void onSyncStarted(@NonNull String batchId);
        void onSyncCompleted(@NonNull String batchId, int processed);
        void onSyncFailed(@NonNull String batchId, @NonNull String error);
        void onConflictDetected(@NonNull ConflictRecord conflict);
        void onConflictResolved(@NonNull String conflictId);
        void onRecordSynced(@NonNull SyncRecord record);
        void onVersionChanged(@NonNull String dataType, @NonNull String dataKey, long version);
    }

    /**
     * 配置
     */
    public static class DataSyncConfig {
        public int maxBatchSize = 100;
        public int maxRetryCount = 3;
        public long syncInterval = 30000;
        public String conflictStrategy = CONFLICT_LAST_WRITE;
        public boolean enableAutoMerge = true;
    }

    public DataSyncClient() {
        this.recordCache = new ConcurrentHashMap<>();
        this.versionCache = new ConcurrentHashMap<>();
        this.pendingQueue = new CopyOnWriteArrayList<>();
        this.conflicts = new ConcurrentHashMap<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.config = new DataSyncConfig();
    }

    /**
     * 初始化
     */
    public void initialize(@NonNull String agentId, @NonNull String identityId,
                           @Nullable MessageBus messageBus) {
        this.agentId = agentId;
        this.identityId = identityId;
        this.messageBus = messageBus;

        if (messageBus != null) {
            messageBus.addListener(this::handleMessage);
        }

        Log.i(TAG, "DataSyncClient initialized for " + agentId);
    }

    // === 增量同步 ===

    /**
     * 创建同步记录
     */
    @NonNull
    public SyncRecord createRecord(@NonNull String dataType, @NonNull String dataKey,
                                   @NonNull String operation, @NonNull Map<String, Object> newValue) {
        String recordId = generateRecordId(dataType, dataKey);

        // 获取当前版本
        String versionKey = getVersionKey(dataType, dataKey);
        long currentVersion = 0;
        DataVersion version = versionCache.get(versionKey);
        if (version != null) {
            currentVersion = version.getVersion();
        }

        SyncRecord record = new SyncRecord();
        record.setRecordId(recordId);
        record.setIdentityId(identityId);
        record.setAgentId(agentId);
        record.setDataType(dataType);
        record.setDataKey(dataKey);
        record.setOperation(operation);
        record.setVersion(currentVersion + 1);
        record.setTimestamp(System.currentTimeMillis());
        record.setNewValue(newValue);
        record.setChecksum(calculateChecksum(newValue));
        record.setStatus(STATUS_PENDING);
        record.setRetryCount(0);

        // 缓存记录
        recordCache.put(recordId, record);
        pendingQueue.add(record);

        // 发送到 Center
        sendRecordToCenter(record);

        Log.d(TAG, "Created sync record: " + recordId);

        return record;
    }

    /**
     * 批量同步
     */
    public void syncBatch(@NonNull List<SyncRecord> records, @Nullable SyncCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        String batchId = generateBatchId();

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_SYNC_BATCH);
        payload.put("batch_id", batchId);
        payload.put("identity_id", identityId);
        payload.put("agent_id", agentId);

        // 序列化记录
        JSONArray recordsArray = new JSONArray();
        for (SyncRecord record : records) {
            recordsArray.put(record.toJson());
        }
        payload.put("records", recordsArray);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        isSyncing = true;
        currentBatchId = batchId;

        notifySyncStarted(batchId);

        if (callback != null) {
            callback.onSuccess(batchId);
        }
    }

    /**
     * 获取待同步记录
     */
    @NonNull
    public List<SyncRecord> getPendingRecords() {
        List<SyncRecord> pending = new ArrayList<>();
        for (SyncRecord record : pendingQueue) {
            if (STATUS_PENDING.equals(record.getStatus())) {
                pending.add(record);
            }
        }
        return pending;
    }

    /**
     * 获取同步队列大小
     */
    public int getQueueSize() {
        return pendingQueue.size();
    }

    // === 版本管理 ===

    /**
     * 获取版本
     */
    @Nullable
    public DataVersion getVersion(@NonNull String dataType, @NonNull String dataKey) {
        String versionKey = getVersionKey(dataType, dataKey);
        return versionCache.get(versionKey);
    }

    /**
     * 查询服务器版本
     */
    public void queryServerVersion(@NonNull String dataType, @NonNull String dataKey,
                                    @Nullable VersionCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_VERSION_QUERY);
        payload.put("identity_id", identityId);
        payload.put("data_type", dataType);
        payload.put("data_key", dataKey);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        if (callback != null) {
            callback.onSuccess(null); // 异步返回
        }
    }

    /**
     * 更新本地版本
     */
    public void updateLocalVersion(@NonNull String dataType, @NonNull String dataKey,
                                   long version, @NonNull String checksum) {
        String versionKey = getVersionKey(dataType, dataKey);

        DataVersion dv = new DataVersion();
        dv.setIdentityId(identityId);
        dv.setDataType(dataType);
        dv.setDataKey(dataKey);
        dv.setVersion(version);
        dv.setChecksum(checksum);
        dv.setUpdatedAt(System.currentTimeMillis());

        versionCache.put(versionKey, dv);

        notifyVersionChanged(dataType, dataKey, version);
    }

    // === 冲突处理 ===

    /**
     * 获取冲突列表
     */
    @NonNull
    public List<ConflictRecord> getConflicts() {
        List<ConflictRecord> conflictList = new ArrayList<>();
        for (ConflictRecord conflict : conflicts.values()) {
            if (STATUS_CONFLICT.equals(conflict.getStatus())) {
                conflictList.add(conflict);
            }
        }
        return conflictList;
    }

    /**
     * 解决冲突
     */
    public void resolveConflict(@NonNull String conflictId, @NonNull String strategy,
                                @Nullable Map<String, Object> resolvedValue,
                                @Nullable ConflictCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        ConflictRecord conflict = conflicts.get(conflictId);
        if (conflict == null) {
            if (callback != null) {
                callback.onError("Conflict not found: " + conflictId);
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_CONFLICT_RESOLVE);
        payload.put("conflict_id", conflictId);
        payload.put("identity_id", identityId);
        payload.put("strategy", strategy);
        if (resolvedValue != null) {
            payload.put("resolved_value", resolvedValue);
        }

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        if (callback != null) {
            callback.onSuccess();
        }
    }

    /**
     * 自动合并冲突
     */
    @Nullable
    public Map<String, Object> autoMerge(@NonNull ConflictRecord conflict) {
        String strategy = conflict.getStrategy();
        if (strategy == null || strategy.isEmpty()) {
            strategy = config.conflictStrategy;
        }

        switch (strategy) {
            case CONFLICT_CLIENT:
                // 客户端优先
                if (conflict.getClientRecords() != null && !conflict.getClientRecords().isEmpty()) {
                    return conflict.getClientRecords().get(0).getNewValue();
                }
                break;

            case CONFLICT_SERVER:
                // 服务端优先
                if (conflict.getServerVersion() != null) {
                    Map<String, Object> result = new HashMap<>();
                    result.put("version", conflict.getServerVersion().getVersion());
                    result.put("source", "server");
                    return result;
                }
                break;

            case CONFLICT_LAST_WRITE:
                // 最后写入胜
                if (conflict.getClientRecords() != null && !conflict.getClientRecords().isEmpty()) {
                    SyncRecord clientRecord = conflict.getClientRecords().get(0);
                    if (conflict.getServerVersion() != null) {
                        if (clientRecord.getTimestamp() > conflict.getServerVersion().getUpdatedAt()) {
                            return clientRecord.getNewValue();
                        }
                    }
                }
                break;

            case CONFLICT_MERGE:
                // 深度合并
                return deepMerge(conflict);
        }

        return null;
    }

    private Map<String, Object> deepMerge(ConflictRecord conflict) {
        Map<String, Object> result = new HashMap<>();

        // 从服务端版本开始
        if (conflict.getServerVersion() != null) {
            result.put("server_version", conflict.getServerVersion().getVersion());
        }

        // 合并客户端数据
        if (conflict.getClientRecords() != null) {
            for (SyncRecord record : conflict.getClientRecords()) {
                if (record.getNewValue() != null) {
                    for (Map.Entry<String, Object> entry : record.getNewValue().entrySet()) {
                        if (!"server_version".equals(entry.getKey())) {
                            result.put(entry.getKey(), entry.getValue());
                        }
                    }
                }
            }
        }

        result.put("merged", true);
        result.put("merged_at", System.currentTimeMillis());

        return result;
    }

    // === 重试机制 ===

    /**
     * 重试失败的记录
     */
    public int retryFailedRecords() {
        int retried = 0;

        for (SyncRecord record : recordCache.values()) {
            if (STATUS_FAILED.equals(record.getStatus()) &&
                record.getRetryCount() < config.maxRetryCount) {

                record.setStatus(STATUS_PENDING);
                record.setRetryCount(record.getRetryCount() + 1);
                record.setErrorMessage(null);

                pendingQueue.add(record);
                sendRecordToCenter(record);
                retried++;
            }
        }

        Log.d(TAG, "Retried " + retried + " failed records");
        return retried;
    }

    // === 消息处理 ===

    private void handleMessage(@NonNull Message message) {
        if (message.payload == null) {
            return;
        }

        Object typeObj = message.payload.get("type");
        if (!"sync_event".equals(typeObj)) {
            return;
        }

        try {
            String event = (String) message.payload.get("event");
            if (event == null) {
                return;
            }

            switch (event) {
                case "batch_completed":
                    handleBatchCompleted(message.payload);
                    break;

                case "batch_failed":
                    handleBatchFailed(message.payload);
                    break;

                case "record_synced":
                    handleRecordSynced(message.payload);
                    break;

                case "conflict_detected":
                    handleConflictDetected(message.payload);
                    break;

                case "conflict_resolved":
                    handleConflictResolved(message.payload);
                    break;

                case "version_update":
                    handleVersionUpdate(message.payload);
                    break;
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to handle sync event", e);
        }
    }

    private void handleBatchCompleted(Map<String, Object> payload) {
        String batchId = (String) payload.get("batch_id");
        int processed = (Integer) payload.getOrDefault("processed", 0);

        isSyncing = false;
        currentBatchId = null;

        notifySyncCompleted(batchId, processed);
    }

    private void handleBatchFailed(Map<String, Object> payload) {
        String batchId = (String) payload.get("batch_id");
        String error = (String) payload.get("error");

        isSyncing = false;
        currentBatchId = null;

        notifySyncFailed(batchId, error != null ? error : "Unknown error");
    }

    private void handleRecordSynced(Map<String, Object> payload) {
        Object recordObj = payload.get("record");
        if (recordObj instanceof JSONObject) {
            try {
                SyncRecord record = SyncRecord.fromJson((JSONObject) recordObj);
                recordCache.put(record.getRecordId(), record);
                pendingQueue.removeIf(r -> r.getRecordId().equals(record.getRecordId()));
                notifyRecordSynced(record);
            } catch (JSONException e) {
                Log.e(TAG, "Failed to parse synced record", e);
            }
        }
    }

    private void handleConflictDetected(Map<String, Object> payload) {
        Object conflictObj = payload.get("conflict");
        if (conflictObj instanceof JSONObject) {
            try {
                ConflictRecord conflict = ConflictRecord.fromJson((JSONObject) conflictObj);
                conflicts.put(conflict.getConflictId(), conflict);
                notifyConflictDetected(conflict);
            } catch (JSONException e) {
                Log.e(TAG, "Failed to parse conflict", e);
            }
        }
    }

    private void handleConflictResolved(Map<String, Object> payload) {
        String conflictId = (String) payload.get("conflict_id");

        ConflictRecord conflict = conflicts.get(conflictId);
        if (conflict != null) {
            conflict.setStatus(STATUS_COMPLETED);
            notifyConflictResolved(conflictId);
        }
    }

    private void handleVersionUpdate(Map<String, Object> payload) {
        String dataType = (String) payload.get("data_type");
        String dataKey = (String) payload.get("data_key");
        Long version = ((Number) payload.get("version")).longValue();
        String checksum = (String) payload.get("checksum");

        updateLocalVersion(dataType, dataKey, version, checksum);
    }

    // === 发送到 Center ===

    private void sendRecordToCenter(SyncRecord record) {
        if (messageBus == null) {
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_SYNC_RECORD);
        payload.put("record", record.toJson());

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);
    }

    // === 监听器管理 ===

    public void addListener(@NonNull SyncListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull SyncListener listener) {
        listeners.remove(listener);
    }

    private void notifySyncStarted(String batchId) {
        for (SyncListener l : listeners) {
            l.onSyncStarted(batchId);
        }
    }

    private void notifySyncCompleted(String batchId, int processed) {
        for (SyncListener l : listeners) {
            l.onSyncCompleted(batchId, processed);
        }
    }

    private void notifySyncFailed(String batchId, String error) {
        for (SyncListener l : listeners) {
            l.onSyncFailed(batchId, error);
        }
    }

    private void notifyConflictDetected(ConflictRecord conflict) {
        for (SyncListener l : listeners) {
            l.onConflictDetected(conflict);
        }
    }

    private void notifyConflictResolved(String conflictId) {
        for (SyncListener l : listeners) {
            l.onConflictResolved(conflictId);
        }
    }

    private void notifyRecordSynced(SyncRecord record) {
        for (SyncListener l : listeners) {
            l.onRecordSynced(record);
        }
    }

    private void notifyVersionChanged(String dataType, String dataKey, long version) {
        for (SyncListener l : listeners) {
            l.onVersionChanged(dataType, dataKey, version);
        }
    }

    // === 统计信息 ===

    @NonNull
    public SyncStats getStats() {
        SyncStats stats = new SyncStats();
        stats.totalRecords = recordCache.size();
        stats.pendingRecords = getPendingRecords().size();
        stats.versionCount = versionCache.size();
        stats.conflictCount = getConflicts().size();
        stats.isSyncing = isSyncing;

        return stats;
    }

    // === 辅助方法 ===

    private String generateRecordId(String dataType, String dataKey) {
        return "record-" + identityId + "-" + dataType + "-" + dataKey + "-" + System.currentTimeMillis();
    }

    private String generateBatchId() {
        return "batch-" + identityId + "-" + System.currentTimeMillis();
    }

    private String generateMessageId() {
        return "sync_msg_" + System.currentTimeMillis() + "_" + agentId;
    }

    private String getVersionKey(String dataType, String dataKey) {
        return dataType + "_" + dataKey;
    }

    private String calculateChecksum(Map<String, Object> data) {
        if (data == null) {
            return "";
        }
        return Integer.toHexString(data.hashCode());
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        recordCache.clear();
        versionCache.clear();
        pendingQueue.clear();
        conflicts.clear();
        listeners.clear();

        isSyncing = false;
        currentBatchId = null;

        Log.i(TAG, "DataSyncClient cleaned up");
    }

    // === 回调接口 ===

    public interface SyncCallback {
        void onSuccess(@Nullable String batchId);
        void onError(@NonNull String error);
    }

    public interface VersionCallback {
        void onSuccess(@Nullable DataVersion version);
        void onError(@NonNull String error);
    }

    public interface ConflictCallback {
        void onSuccess();
        void onError(@NonNull String error);
    }

    /**
     * 统计信息
     */
    public static class SyncStats {
        public int totalRecords;
        public int pendingRecords;
        public int versionCount;
        public int conflictCount;
        public boolean isSyncing;

        @NonNull
        @Override
        public String toString() {
            return "SyncStats{" +
                    "total=" + totalRecords +
                    ", pending=" + pendingRecords +
                    ", versions=" + versionCount +
                    ", conflicts=" + conflictCount +
                    ", syncing=" + isSyncing +
                    '}';
        }
    }
}