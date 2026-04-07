package com.ofa.agent.identity;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.core.CenterConnection;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * IdentityManager - 身份管理器（核心）
 *
 * 管理用户个人身份，包括：
 * - 本地身份存储
 * - Center 同步
 * - 行为观察收集
 * - 性格推断
 *
 * 实现"万物都是我"的核心理念：所有设备共享同一人格。
 */
public class IdentityManager {

    private static final String TAG = "IdentityManager";

    // 推断阈值
    private static final int INFERENCE_THRESHOLD = 10;

    private final Context context;
    private final LocalIdentityStore localStore;
    private IdentitySyncService syncService;

    private PersonalIdentity currentIdentity;
    private DecisionContext decisionContext;

    private final List<BehaviorObservation> pendingObservations = new ArrayList<>();
    private final List<IdentityChangeListener> listeners = new ArrayList<>();

    private boolean initialized = false;

    /**
     * 身份变更监听器
     */
    public interface IdentityChangeListener {
        void onIdentityChanged(PersonalIdentity identity);
        void onPersonalityUpdated(Personality personality);
        void onSyncCompleted(IdentitySyncService.SyncResult result);
    }

    /**
     * 同步结果
     */
    public static class SyncResult {
        public final boolean success;
        public final String error;

        public SyncResult(boolean success, String error) {
            this.success = success;
            this.error = error;
        }

        public static SyncResult success() {
            return new SyncResult(true, null);
        }

        public static SyncResult failure(String error) {
            return new SyncResult(false, error);
        }
    }

    /**
     * 创建身份管理器
     */
    public IdentityManager(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.localStore = new LocalIdentityStore(context);

        // 加载本地身份
        loadLocalIdentity();
    }

    /**
     * 初始化
     */
    public void initialize() {
        if (initialized) return;

        Log.i(TAG, "Initializing IdentityManager...");

        // 加载本地身份
        if (currentIdentity == null) {
            currentIdentity = localStore.loadIdentity();

            if (currentIdentity == null) {
                // 创建默认身份
                currentIdentity = new PersonalIdentity();
                localStore.saveIdentity(currentIdentity);
                Log.i(TAG, "Created default identity: " + currentIdentity.getId());
            } else {
                Log.i(TAG, "Loaded identity: " + currentIdentity.getId() + " v" + currentIdentity.getVersion());
            }
        }

        // 创建决策上下文
        decisionContext = new DecisionContext(currentIdentity);

        initialized = true;
        Log.i(TAG, "IdentityManager initialized");
    }

    /**
     * 设置同步服务
     */
    public void enableSync(@Nullable String centerAddress, int centerPort) {
        if (syncService != null) {
            syncService.disableSync();
        }

        syncService = new IdentitySyncService(context, localStore, centerAddress, centerPort);
        syncService.setStatusListener(new IdentitySyncService.SyncStatusListener() {
            @Override
            public void onSyncStarted() {
                Log.d(TAG, "Sync started");
            }

            @Override
            public void onSyncCompleted(IdentitySyncService.SyncResult result) {
                Log.i(TAG, "Sync completed: success=" + result.success + ", conflict=" + result.conflict);
                if (result.success && result.identity != null) {
                    currentIdentity = result.identity;
                    decisionContext = new DecisionContext(currentIdentity);
                    notifyIdentityChanged();
                }
                notifySyncCompleted(IdentitySyncService.SyncResult.success(result.identity));
            }

            @Override
            public void onSyncFailed(String error) {
                Log.e(TAG, "Sync failed: " + error);
                notifySyncCompleted(IdentitySyncService.SyncResult.failure(error));
            }
        });

        syncService.enableSync();
        Log.i(TAG, "Sync enabled with Center: " + centerAddress + ":" + centerPort);
    }

    /**
     * 禁用同步
     */
    public void disableSync() {
        if (syncService != null) {
            syncService.disableSync();
            syncService = null;
        }
        Log.i(TAG, "Sync disabled");
    }

    // === 身份操作 ===

    /**
     * 获取当前身份
     */
    @Nullable
    public PersonalIdentity getIdentity() {
        return currentIdentity;
    }

    /**
     * 更新身份
     */
    public void updateIdentity(@NonNull PersonalIdentity identity) {
        this.currentIdentity = identity;
        localStore.saveIdentity(identity);
        decisionContext = new DecisionContext(identity);
        notifyIdentityChanged();
    }

    /**
     * 更新基本信息
     */
    public void updateBasicInfo(@NonNull String name, @Nullable String nickname,
                                @Nullable String avatar, @Nullable String location) {
        if (currentIdentity == null) return;

        currentIdentity.setName(name);
        if (nickname != null) currentIdentity.setNickname(nickname);
        if (avatar != null) currentIdentity.setAvatar(avatar);
        if (location != null) currentIdentity.setLocation(location);

        localStore.saveIdentity(currentIdentity);
        notifyIdentityChanged();
    }

    /**
     * 更新性格特质
     */
    public void updatePersonality(@NonNull Map<String, Double> updates) {
        if (currentIdentity == null) return;

        currentIdentity.updatePersonality(updates);
        localStore.saveIdentity(currentIdentity);
        notifyPersonalityUpdated();
    }

    /**
     * 设置 MBTI 类型
     */
    public void setMBTIType(@NonNull String mbtiType) {
        if (currentIdentity == null || currentIdentity.getPersonality() == null) return;

        currentIdentity.getPersonality().setMBTIType(mbtiType);
        localStore.saveIdentity(currentIdentity);
        notifyPersonalityUpdated();
    }

    /**
     * 更新价值观
     */
    public void updateValueSystem(@NonNull Map<String, Double> updates) {
        if (currentIdentity == null) return;

        currentIdentity.updateValueSystem(updates);
        localStore.saveIdentity(currentIdentity);
        notifyIdentityChanged();
    }

    /**
     * 添加兴趣
     */
    public void addInterest(@NonNull Interest interest) {
        if (currentIdentity == null) return;

        currentIdentity.addInterest(interest);
        localStore.saveIdentity(currentIdentity);
        notifyIdentityChanged();
    }

    /**
     * 移除兴趣
     */
    public boolean removeInterest(@NonNull String interestId) {
        if (currentIdentity == null) return false;

        boolean removed = currentIdentity.removeInterest(interestId);
        if (removed) {
            localStore.saveIdentity(currentIdentity);
            notifyIdentityChanged();
        }
        return removed;
    }

    // === 同步操作 ===

    /**
     * 同步到 Center
     */
    @NonNull
    public CompletableFuture<SyncResult> syncToCenter() {
        if (syncService == null) {
            return CompletableFuture.completedFuture(SyncResult.failure("Sync not enabled"));
        }

        return syncService.syncToCenter().thenApply(result ->
            result.success ? SyncResult.success() : SyncResult.failure(result.error));
    }

    /**
     * 从 Center 恢复
     */
    @NonNull
    public CompletableFuture<PersonalIdentity> restoreFromCenter(@NonNull String identityId) {
        if (syncService == null) {
            return CompletableFuture.completedFuture(null);
        }

        return syncService.restoreFromCenter(identityId).thenApply(identity -> {
            if (identity != null) {
                currentIdentity = identity;
                decisionContext = new DecisionContext(identity);
                notifyIdentityChanged();
            }
            return identity;
        });
    }

    // === 行为观察 ===

    /**
     * 观察行为
     */
    public void observeBehavior(@NonNull String type, @NonNull Map<String, Object> context) {
        BehaviorObservation observation = new BehaviorObservation(type, context);

        // 自动推断性格变化
        observation.autoInfer();

        // 添加到待处理列表
        pendingObservations.add(observation);

        // 如果达到阈值，触发推断
        if (pendingObservations.size() >= INFERENCE_THRESHOLD) {
            triggerInference();
        }

        // 上报到 Center
        if (syncService != null && syncService.isSyncEnabled()) {
            syncService.reportBehavior(observation);
        }

        Log.d(TAG, "Behavior observed: " + type);
    }

    /**
     * 触发性格推断
     */
    private void triggerInference() {
        if (currentIdentity == null || currentIdentity.getPersonality() == null) return;

        Personality personality = currentIdentity.getPersonality();

        // 累积推断结果
        for (BehaviorObservation obs : pendingObservations) {
            Map<String, Double> inferences = obs.getInferences();

            // 应用推断（带收敛机制）
            applyInferencesWithConvergence(personality, inferences);
        }

        // 更新观察计数
        personality.incrementObservedCount();
        personality.setLastInferredAt(System.currentTimeMillis());

        // 计算稳定度
        int observedCount = personality.getObservedCount();
        double stability = Math.min(1.0, observedCount / 100.0);
        personality.setStabilityScore(stability);

        // 清空待处理列表
        pendingObservations.clear();

        // 保存
        localStore.saveIdentity(currentIdentity);

        Log.i(TAG, "Personality inferred: observations=" + observedCount + ", stability=" + stability);
        notifyPersonalityUpdated();
    }

    /**
     * 应用推断结果（带收敛机制）
     *
     * 收敛公式：new_value = current + (inference * (1 - stability) * weight)
     */
    private void applyInferencesWithConvergence(@NonNull Personality personality,
                                                  @NonNull Map<String, Double> inferences) {
        double stability = personality.getStabilityScore();
        double weight = 0.5; // 单次观察权重

        for (Map.Entry<String, Double> entry : inferences.entrySet()) {
            String key = entry.getKey();
            double inference = entry.getValue();

            // 收敛因子：稳定度越高，变化越小
            double convergenceFactor = 1 - stability;
            double delta = inference * convergenceFactor * weight;

            personality.updateTrait(key, getCurrentTraitValue(personality, key) + delta);
        }
    }

    /**
     * 获取当前特质值
     */
    private double getCurrentTraitValue(@NonNull Personality personality, @NonNull String key) {
        switch (key) {
            case "openness":
                return personality.getOpenness();
            case "conscientiousness":
                return personality.getConscientiousness();
            case "extraversion":
                return personality.getExtraversion();
            case "agreeableness":
                return personality.getAgreeableness();
            case "neuroticism":
                return personality.getNeuroticism();
            default:
                return personality.getCustomTraits().getOrDefault(key, 0.5);
        }
    }

    // === 决策上下文 ===

    /**
     * 获取决策上下文
     */
    @Nullable
    public DecisionContext getDecisionContext() {
        return decisionContext;
    }

    /**
     * 生成 AI Prompt 上下文
     */
    @NonNull
    public String generatePromptContext() {
        if (decisionContext == null) {
            return "";
        }
        return decisionContext.generatePromptContext();
    }

    // === 监听器 ===

    /**
     * 添加监听器
     */
    public void addListener(@NonNull IdentityChangeListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除监听器
     */
    public void removeListener(@NonNull IdentityChangeListener listener) {
        listeners.remove(listener);
    }

    private void notifyIdentityChanged() {
        for (IdentityChangeListener listener : listeners) {
            listener.onIdentityChanged(currentIdentity);
        }
    }

    private void notifyPersonalityUpdated() {
        for (IdentityChangeListener listener : listeners) {
            listener.onPersonalityUpdated(currentIdentity.getPersonality());
        }
    }

    private void notifySyncCompleted(IdentitySyncService.SyncResult result) {
        for (IdentityChangeListener listener : listeners) {
            listener.onSyncCompleted(result);
        }
    }

    // === 辅助方法 ===

    private void loadLocalIdentity() {
        currentIdentity = localStore.loadIdentity();
    }

    /**
     * 检查是否已初始化
     */
    public boolean isInitialized() {
        return initialized;
    }

    /**
     * 检查是否有身份
     */
    public boolean hasIdentity() {
        return currentIdentity != null;
    }

    /**
     * 获取身份 ID
     */
    @Nullable
    public String getIdentityId() {
        return currentIdentity != null ? currentIdentity.getId() : null;
    }

    /**
     * 关闭
     */
    public void shutdown() {
        disableSync();
        listeners.clear();
        pendingObservations.clear();

        // 保存当前状态
        if (currentIdentity != null) {
            localStore.saveIdentity(currentIdentity);
        }

        initialized = false;
        Log.i(TAG, "IdentityManager shutdown");
    }
}