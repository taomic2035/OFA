package com.ofa.agent.offline;

import android.content.Context;
import android.content.SharedPreferences;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.skill.SkillExecutor;

import org.json.JSONObject;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 离线管理器 - 综合管理离线模式
 */
public class OfflineManager {
    private static final String TAG = "OfflineManager";
    private static final String PREFS_NAME = "ofa_offline_prefs";
    private static final String KEY_OFFLINE_MODE = "offline_mode";

    private final Context context;
    private final OfflineLevel level;
    private final LocalScheduler scheduler;
    private final OfflineCache cache;
    private final SharedPreferences prefs;

    private volatile boolean offlineMode;
    private final Handler handler = new Handler(Looper.getMainLooper());
    private SyncCallback syncCallback;
    private final List<OfflineModeListener> modeListeners = new ArrayList<>();

    public OfflineManager(@NonNull Context context, @NonNull OfflineLevel level) {
        this.context = context.getApplicationContext();
        this.level = level;
        this.scheduler = new LocalScheduler(4, level);
        this.cache = new OfflineCache(context);
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
        this.offlineMode = prefs.getBoolean(KEY_OFFLINE_MODE, level == OfflineLevel.L1);
    }

    /**
     * 启动离线管理器
     */
    public void start() {
        scheduler.start();
        Log.i(TAG, "Offline manager started at level " + level.getValue());
    }

    /**
     * 停止离线管理器
     */
    public void stop() {
        scheduler.stop();
        cache.clear();
        Log.i(TAG, "Offline manager stopped");
    }

    /**
     * 设置离线模式
     */
    public void setOfflineMode(boolean offline) {
        this.offlineMode = offline;
        prefs.edit().putBoolean(KEY_OFFLINE_MODE, offline).apply();

        for (OfflineModeListener listener : modeListeners) {
            listener.onModeChanged(offline);
        }

        Log.i(TAG, "Offline mode: " + offline);
    }

    /**
     * 是否处于离线模式
     */
    public boolean isOfflineMode() {
        return offlineMode;
    }

    /**
     * 获取离线等级
     */
    public OfflineLevel getLevel() {
        return level;
    }

    /**
     * 注册技能
     */
    public void registerSkill(@NonNull String skillId, @NonNull SkillExecutor executor, boolean offlineCapable) {
        scheduler.registerSkill(skillId, executor, offlineCapable);
    }

    /**
     * 本地执行任务
     */
    @NonNull
    public String executeLocal(@NonNull String skillId, @Nullable byte[] input) {
        return scheduler.submitTask(skillId, input);
    }

    /**
     * 缓存数据
     */
    public void cacheData(@NonNull String key, @NonNull byte[] data, long expiryMs) {
        cache.put(key, data, expiryMs);
    }

    public void cacheData(@NonNull String key, @NonNull byte[] data) {
        cacheData(key, data, 0);
    }

    /**
     * 获取缓存数据
     */
    @Nullable
    public byte[] getCachedData(@NonNull String key) {
        return cache.get(key);
    }

    /**
     * 获取待同步数据键列表
     */
    @NonNull
    public List<String> getPendingSyncKeys() {
        return cache.getPendingKeys();
    }

    /**
     * 标记已同步
     */
    public void markSynced(@NonNull String key) {
        cache.markSynced(key);
    }

    /**
     * 立即同步
     */
    public boolean syncNow() {
        List<String> pending = getPendingSyncKeys();
        if (pending.isEmpty()) {
            return true;
        }

        if (syncCallback == null) {
            Log.w(TAG, "No sync callback configured");
            return false;
        }

        for (String key : pending) {
            byte[] data = cache.get(key);
            if (data != null) {
                try {
                    syncCallback.onSync(key, data);
                    cache.markSynced(key);
                } catch (Exception e) {
                    Log.e(TAG, "Sync failed for " + key, e);
                    return false;
                }
            }
        }

        return true;
    }

    /**
     * 设置同步回调
     */
    public void setSyncCallback(@Nullable SyncCallback callback) {
        this.syncCallback = callback;
    }

    /**
     * 添加离线模式监听器
     */
    public void addOfflineModeListener(@NonNull OfflineModeListener listener) {
        modeListeners.add(listener);
    }

    /**
     * 移除离线模式监听器
     */
    public void removeOfflineModeListener(@NonNull OfflineModeListener listener) {
        modeListeners.remove(listener);
    }

    /**
     * 获取任务
     */
    @Nullable
    public LocalTask getTask(@NonNull String taskId) {
        return scheduler.getTask(taskId);
    }

    /**
     * 获取统计信息
     */
    @NonNull
    public Stats getStats() {
        return new Stats(
            offlineMode,
            level.getValue(),
            scheduler.getPendingCount(),
            scheduler.getCompletedCount(),
            cache.getPendingCount(),
            cache.hitRate()
        );
    }

    /**
     * 统计信息
     */
    public static class Stats {
        public final boolean offlineMode;
        public final int level;
        public final int pendingTasks;
        public final int completedTasks;
        public final int pendingSync;
        public final double cacheHitRate;

        public Stats(boolean offlineMode, int level, int pendingTasks, int completedTasks, int pendingSync, double cacheHitRate) {
            this.offlineMode = offlineMode;
            this.level = level;
            this.pendingTasks = pendingTasks;
            this.completedTasks = completedTasks;
            this.pendingSync = pendingSync;
            this.cacheHitRate = cacheHitRate;
        }
    }

    public interface SyncCallback {
        void onSync(String key, byte[] data);
    }

    public interface OfflineModeListener {
        void onModeChanged(boolean offline);
    }
}