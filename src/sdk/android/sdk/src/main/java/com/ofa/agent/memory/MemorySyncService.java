package com.ofa.agent.memory;

import android.content.Context;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * MemorySyncService - 记忆同步服务 (v2.2.0)
 *
 * 负责跨设备记忆同步：
 * - 上传本地记忆变更到 Center
 * - 下载远程记忆变更到本地
 * - 冲突检测与解决
 */
public class MemorySyncService {

    private static final String TAG = "MemorySyncService";

    private final Context context;
    private final UserMemoryManager memoryManager;
    private final String centerAddress;
    private final int centerPort;
    private final String identityId;

    private final Handler handler = new Handler(Looper.getMainLooper());
    private boolean syncEnabled = false;
    private Runnable syncRunnable;
    private static final long SYNC_INTERVAL_MS = 300000; // 5分钟同步一次

    private SyncStatusListener statusListener;
    private long lastSyncTimestamp = 0;
    private long localVersion = 0;

    /**
     * 同步状态监听器
     */
    public interface SyncStatusListener {
        void onSyncStarted();
        void onSyncCompleted(SyncResult result);
        void onSyncFailed(String error);
    }

    /**
     * 同步结果
     */
    public static class SyncResult {
        public final boolean success;
        public final int uploadedCount;
        public final int downloadedCount;
        public final int conflictCount;
        public final String error;

        public SyncResult(boolean success, int uploaded, int downloaded, int conflicts, String error) {
            this.success = success;
            this.uploadedCount = uploaded;
            this.downloadedCount = downloaded;
            this.conflictCount = conflicts;
            this.error = error;
        }

        public static SyncResult success(int uploaded, int downloaded, int conflicts) {
            return new SyncResult(true, uploaded, downloaded, conflicts, null);
        }

        public static SyncResult failure(String error) {
            return new SyncResult(false, 0, 0, 0, error);
        }
    }

    /**
     * 创建记忆同步服务
     */
    public MemorySyncService(@NonNull Context context,
                              @NonNull UserMemoryManager memoryManager,
                              @NonNull String identityId,
                              @Nullable String centerAddress,
                              int centerPort) {
        this.context = context.getApplicationContext();
        this.memoryManager = memoryManager;
        this.identityId = identityId;
        this.centerAddress = centerAddress != null ? centerAddress : "localhost";
        this.centerPort = centerPort;

        // 定时同步任务
        syncRunnable = () -> {
            if (syncEnabled) {
                sync();
                handler.postDelayed(syncRunnable, SYNC_INTERVAL_MS);
            }
        };
    }

    // === 同步操作 ===

    /**
     * 执行同步
     */
    @NonNull
    public CompletableFuture<SyncResult> sync() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                notifySyncStarted();

                // 1. 收集本地变更
                List<MemoryEntry> localChanges = collectLocalChanges();

                // 2. 构建同步请求
                String requestJson = buildSyncRequest(localChanges);

                // 3. 发送同步请求
                String responseJson = sendSyncRequest(requestJson);

                // 4. 解析响应并应用变更
                SyncResult result = processSyncResponse(responseJson);

                // 更新同步时间戳
                lastSyncTimestamp = System.currentTimeMillis();

                notifySyncCompleted(result);
                return result;

            } catch (Exception e) {
                Log.e(TAG, "Sync failed", e);
                notifySyncFailed(e.getMessage());
                return SyncResult.failure(e.getMessage());
            }
        });
    }

    /**
     * 全量同步（从 Center 恢复）
     */
    @NonNull
    public CompletableFuture<SyncResult> fullSync() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                notifySyncStarted();

                // 发送全量同步请求
                String responseJson = sendFullSyncRequest();

                // 解析并应用
                SyncResult result = processSyncResponse(responseJson);

                lastSyncTimestamp = System.currentTimeMillis();
                localVersion = System.currentTimeMillis();

                notifySyncCompleted(result);
                return result;

            } catch (Exception e) {
                Log.e(TAG, "Full sync failed", e);
                notifySyncFailed(e.getMessage());
                return SyncResult.failure(e.getMessage());
            }
        });
    }

    // === 收集变更 ===

    private List<MemoryEntry> collectLocalChanges() {
        List<MemoryEntry> changes = new ArrayList<>();

        // 获取自上次同步以来的变更
        // 简化实现：获取所有记忆（实际应该追踪变更）
        List<MemoryEntry> allEntries = new ArrayList<>();

        // 这里简化处理，获取常用记忆
        for (String key : memoryManager.getStats().hotKeyCount > 0 ?
                getHotKeys() : new ArrayList<>()) {
            MemoryEntry entry = memoryManager.getLastAction(key);
            if (entry != null) {
                changes.add(entry);
            }
        }

        return changes;
    }

    private List<String> getHotKeys() {
        // 简化实现：返回空列表
        return new ArrayList<>();
    }

    // === 网络通信 ===

    private String buildSyncRequest(@NonNull List<MemoryEntry> changes) {
        JSONObject request = new JSONObject();
        try {
            request.put("identity_id", identityId);
            request.put("version", localVersion);

            JSONArray memories = new JSONArray();
            for (MemoryEntry entry : changes) {
                memories.put(entryToJson(entry));
            }
            request.put("memories", memories);

        } catch (Exception e) {
            Log.e(TAG, "Failed to build sync request", e);
        }
        return request.toString();
    }

    private String sendSyncRequest(@NonNull String requestJson) throws Exception {
        String url = "http://" + centerAddress + ":" + centerPort + "/api/v1/memory/sync";
        return sendHttpPost(url, requestJson);
    }

    private String sendFullSyncRequest() throws Exception {
        String url = "http://" + centerAddress + ":" + centerPort + "/api/v1/memory/" + identityId;
        return sendHttpGet(url);
    }

    private String sendHttpPost(@NonNull String url, @NonNull String body) throws Exception {
        HttpURLConnection conn = (HttpURLConnection) new URL(url).openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json");
        conn.setDoOutput(true);
        conn.setConnectTimeout(10000);
        conn.setReadTimeout(30000);

        OutputStream os = conn.getOutputStream();
        os.write(body.getBytes("UTF-8"));
        os.close();

        int responseCode = conn.getResponseCode();
        if (responseCode != 200) {
            throw new Exception("HTTP error: " + responseCode);
        }

        java.io.InputStream is = conn.getInputStream();
        java.io.BufferedReader reader = new java.io.BufferedReader(
            new java.io.InputStreamReader(is, "UTF-8"));
        StringBuilder response = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            response.append(line);
        }
        reader.close();
        conn.disconnect();

        return response.toString();
    }

    private String sendHttpGet(@NonNull String url) throws Exception {
        HttpURLConnection conn = (HttpURLConnection) new URL(url).openConnection();
        conn.setRequestMethod("GET");
        conn.setRequestProperty("Accept", "application/json");
        conn.setConnectTimeout(10000);
        conn.setReadTimeout(30000);

        int responseCode = conn.getResponseCode();
        if (responseCode != 200) {
            throw new Exception("HTTP error: " + responseCode);
        }

        java.io.InputStream is = conn.getInputStream();
        java.io.BufferedReader reader = new java.io.BufferedReader(
            new java.io.InputStreamReader(is, "UTF-8"));
        StringBuilder response = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            response.append(line);
        }
        reader.close();
        conn.disconnect();

        return response.toString();
    }

    // === 响应处理 ===

    private SyncResult processSyncResponse(@NonNull String responseJson) {
        try {
            JSONObject response = new JSONObject(responseJson);

            if (!response.optBoolean("success", false)) {
                return SyncResult.failure(response.optString("error", "Sync failed"));
            }

            int uploaded = response.optInt("uploaded_count", 0);
            int downloaded = 0;
            int conflicts = 0;

            // 解析下载的记忆
            if (response.has("memories")) {
                JSONArray memories = response.getJSONArray("memories");
                for (int i = 0; i < memories.length(); i++) {
                    MemoryEntry entry = jsonToMemoryEntry(memories.getJSONObject(i));
                    if (entry != null) {
                        // 检查冲突
                        MemoryEntry local = memoryManager.getLastAction(entry.getKey());
                        if (local != null && local.getTimestamp() > entry.getTimestamp()) {
                            // 本地更新，保留本地
                            conflicts++;
                        } else {
                            // 应用远程变更
                            memoryManager.remember(entry);
                            downloaded++;
                        }
                    }
                }
            }

            // 更新版本
            localVersion = response.optLong("version", System.currentTimeMillis());

            return SyncResult.success(uploaded, downloaded, conflicts);

        } catch (Exception e) {
            Log.e(TAG, "Failed to process sync response", e);
            return SyncResult.failure("Parse error: " + e.getMessage());
        }
    }

    // === JSON 转换 ===

    private JSONObject entryToJson(@NonNull MemoryEntry entry) {
        JSONObject json = new JSONObject();
        try {
            json.put("key", entry.getKey());
            json.put("value", entry.getValue());
            json.put("category", entry.getCategory());
            json.put("timestamp", entry.getTimestamp());
            json.put("count", entry.getCount());
            json.put("score", entry.getScore());
        } catch (Exception e) {
            Log.e(TAG, "Failed to convert entry to JSON", e);
        }
        return json;
    }

    @Nullable
    private MemoryEntry jsonToMemoryEntry(@NonNull JSONObject json) {
        try {
            return new MemoryEntry.Builder()
                .key(json.getString("key"))
                .value(json.getString("value"))
                .category(json.optString("category", "general"))
                .timestamp(json.optLong("timestamp", System.currentTimeMillis()))
                .count(json.optInt("count", 1))
                .score((float) json.optDouble("score", 1.0))
                .build();
        } catch (Exception e) {
            Log.e(TAG, "Failed to parse memory entry", e);
            return null;
        }
    }

    // === 启停控制 ===

    /**
     * 启用同步
     */
    public void enableSync() {
        syncEnabled = true;
        handler.post(syncRunnable);
        Log.i(TAG, "Memory sync enabled");
    }

    /**
     * 禁用同步
     */
    public void disableSync() {
        syncEnabled = false;
        handler.removeCallbacks(syncRunnable);
        Log.i(TAG, "Memory sync disabled");
    }

    /**
     * 检查是否启用同步
     */
    public boolean isSyncEnabled() {
        return syncEnabled;
    }

    // === 监听器 ===

    private void notifySyncStarted() {
        if (statusListener != null) {
            handler.post(() -> statusListener.onSyncStarted());
        }
    }

    private void notifySyncCompleted(SyncResult result) {
        if (statusListener != null) {
            handler.post(() -> statusListener.onSyncCompleted(result));
        }
    }

    private void notifySyncFailed(String error) {
        if (statusListener != null) {
            handler.post(() -> statusListener.onSyncFailed(error));
        }
    }

    public void setStatusListener(@Nullable SyncStatusListener listener) {
        this.statusListener = listener;
    }
}