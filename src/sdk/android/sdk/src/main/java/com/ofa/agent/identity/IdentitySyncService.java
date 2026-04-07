package com.ofa.agent.identity;

import android.content.Context;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.core.CenterConnection;

import org.json.JSONObject;

import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.concurrent.CompletableFuture;

/**
 * IdentitySyncService - 身份同步服务
 *
 * 负责 Agent 与 Center 之间的身份同步：
 * 1. 上传本地变更到 Center
 * 2. 下载远程变更到本地
 * 3. 处理同步冲突
 */
public class IdentitySyncService {

    private static final String TAG = "IdentitySyncService";

    private final Context context;
    private final LocalIdentityStore localStore;
    private final String centerAddress;
    private final int centerPort;

    private CenterConnection centerConnection;
    private final Handler handler = new Handler(Looper.getMainLooper());

    private boolean syncEnabled = false;
    private Runnable syncRunnable;
    private static final long SYNC_INTERVAL_MS = 60000; // 1 分钟同步一次

    private SyncStatusListener statusListener;

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
        public final boolean conflict;
        public final PersonalIdentity identity;
        public final String error;

        public SyncResult(boolean success, boolean conflict, PersonalIdentity identity, String error) {
            this.success = success;
            this.conflict = conflict;
            this.identity = identity;
            this.error = error;
        }

        public static SyncResult success(PersonalIdentity identity) {
            return new SyncResult(true, false, identity, null);
        }

        public static SyncResult conflict(PersonalIdentity identity) {
            return new SyncResult(true, true, identity, null);
        }

        public static SyncResult failure(String error) {
            return new SyncResult(false, false, null, error);
        }
    }

    /**
     * 创建同步服务
     */
    public IdentitySyncService(@NonNull Context context,
                               @NonNull LocalIdentityStore localStore,
                               @Nullable String centerAddress,
                               int centerPort) {
        this.context = context.getApplicationContext();
        this.localStore = localStore;
        this.centerAddress = centerAddress != null ? centerAddress : "localhost";
        this.centerPort = centerPort;

        // 定时同步任务
        syncRunnable = () -> {
            if (syncEnabled) {
                syncToCenter();
                handler.postDelayed(syncRunnable, SYNC_INTERVAL_MS);
            }
        };
    }

    /**
     * 设置 Center 连接
     */
    public void setCenterConnection(@Nullable CenterConnection connection) {
        this.centerConnection = connection;
    }

    // === 同步操作 ===

    /**
     * 同步到 Center（上传本地变更）
     */
    @NonNull
    public CompletableFuture<SyncResult> syncToCenter() {
        return CompletableFuture.supplyAsync(() -> {
            PersonalIdentity localIdentity = localStore.loadIdentity();
            if (localIdentity == null) {
                return SyncResult.failure("No local identity to sync");
            }

            try {
                notifySyncStarted();

                // 构建同步请求
                String requestJson = buildSyncRequest(localIdentity);

                // 发送同步请求
                String responseJson = sendSyncRequest(requestJson);

                // 解析响应
                SyncResult result = parseSyncResponse(responseJson, localIdentity);

                if (result.success && result.identity != null) {
                    // 更新本地存储
                    localStore.saveIdentity(result.identity);
                }

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
     * 从 Center 恢复（下载远程身份）
     */
    @NonNull
    public CompletableFuture<PersonalIdentity> restoreFromCenter(@NonNull String identityId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                // 发送获取请求
                String responseJson = sendGetRequest(identityId);

                // 解析身份
                PersonalIdentity identity = PersonalIdentity.fromJson(responseJson);
                if (identity != null) {
                    // 保存到本地
                    localStore.saveIdentity(identity);
                    Log.i(TAG, "Identity restored from Center: " + identity.getId());
                }

                return identity;

            } catch (Exception e) {
                Log.e(TAG, "Restore failed", e);
                return null;
            }
        });
    }

    /**
     * 上报行为观察
     */
    public void reportBehavior(@NonNull BehaviorObservation observation) {
        if (!syncEnabled || centerConnection == null || !centerConnection.isConnected()) {
            Log.w(TAG, "Cannot report behavior: not connected");
            return;
        }

        try {
            // TODO: 实现行为上报
            Log.d(TAG, "Behavior reported: " + observation.getType());

        } catch (Exception e) {
            Log.w(TAG, "Failed to report behavior", e);
        }
    }

    // === 启停控制 ===

    /**
     * 启用同步
     */
    public void enableSync() {
        syncEnabled = true;
        handler.post(syncRunnable);
        Log.i(TAG, "Sync enabled");
    }

    /**
     * 禁用同步
     */
    public void disableSync() {
        syncEnabled = false;
        handler.removeCallbacks(syncRunnable);
        Log.i(TAG, "Sync disabled");
    }

    /**
     * 检查是否启用同步
     */
    public boolean isSyncEnabled() {
        return syncEnabled;
    }

    // === HTTP 通信 ===

    private String buildSyncRequest(@NonNull PersonalIdentity identity) {
        JSONObject request = new JSONObject();
        try {
            request.put("agent_id", identity.getId());
            request.put("identity", new JSONObject(identity.toJson()));
            request.put("version", identity.getVersion());
            request.put("sync_type", "full");
        } catch (Exception e) {
            Log.e(TAG, "Failed to build sync request", e);
        }
        return request.toString();
    }

    private String sendSyncRequest(@NonNull String requestJson) throws Exception {
        String url = "http://" + centerAddress + ":" + centerPort + "/api/v1/identity/sync";
        return sendHttpPost(url, requestJson);
    }

    private String sendGetRequest(@NonNull String identityId) throws Exception {
        String url = "http://" + centerAddress + ":" + centerPort + "/api/v1/identity/" + identityId;
        return sendHttpGet(url);
    }

    private String sendHttpPost(@NonNull String url, @NonNull String body) throws Exception {
        HttpURLConnection conn = (HttpURLConnection) new URL(url).openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json");
        conn.setDoOutput(true);

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

    // === 响应解析 ===

    private SyncResult parseSyncResponse(@NonNull String responseJson,
                                          @NonNull PersonalIdentity localIdentity) {
        try {
            JSONObject response = new JSONObject(responseJson);

            boolean success = response.optBoolean("success", false);
            boolean conflict = response.optBoolean("conflict", false);

            if (!success) {
                return SyncResult.failure(response.optString("error", "Sync failed"));
            }

            // 解析返回的身份
            if (response.has("identity")) {
                JSONObject identityJson = response.getJSONObject("identity");
                PersonalIdentity remoteIdentity = PersonalIdentity.fromJson(identityJson.toString());

                if (conflict) {
                    // 冲突处理：取最新版本
                    remoteIdentity = resolveConflict(localIdentity, remoteIdentity);
                    return SyncResult.conflict(remoteIdentity);
                }

                return SyncResult.success(remoteIdentity);
            }

            return SyncResult.success(localIdentity);

        } catch (Exception e) {
            Log.e(TAG, "Failed to parse sync response", e);
            return SyncResult.failure("Parse error: " + e.getMessage());
        }
    }

    /**
     * 冲突解决策略
     *
     * 规则：
     * - 同一 key: 取最新时间戳的值
     * - 计数字段: 累加各设备计数
     * - 分数字段: 取平均值
     */
    private PersonalIdentity resolveConflict(@NonNull PersonalIdentity local,
                                              @NonNull PersonalIdentity remote) {
        // 简化版：取 updatedAt 更新的
        if (local.getUpdatedAt() >= remote.getUpdatedAt()) {
            return local;
        } else {
            return remote;
        }
    }

    // === 监听器通知 ===

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