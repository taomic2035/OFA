package com.ofa.agent.behavior;

import android.content.Context;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.identity.BehaviorObservation;
import com.ofa.agent.identity.IdentityManager;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * BehaviorCollector - 行为收集器 (v2.4.0)
 *
 * 自动收集用户行为并上报到 IdentityManager 进行性格推断。
 *
 * 行为类型：
 * - decision: 决策类行为（购买、投资等）
 * - interaction: 交互类行为（社交、聊天等）
 * - preference: 偏好类行为（设置、选择等）
 * - activity: 活动类行为（运动、学习等）
 */
public class BehaviorCollector {

    private static final String TAG = "BehaviorCollector";

    private final Context context;
    private final IdentityManager identityManager;
    private final ExecutorService executor;
    private final Handler handler;

    // 行为缓存（批量上报）
    private final List<BehaviorObservation> behaviorBuffer = new ArrayList<>();
    private static final int BUFFER_SIZE = 10;
    private static final long FLUSH_INTERVAL_MS = 60000; // 1分钟

    private Runnable flushRunnable;
    private boolean enabled = false;

    // 行为监听器
    private final List<BehaviorListener> listeners = new ArrayList<>();

    /**
     * 行为监听器
     */
    public interface BehaviorListener {
        void onBehaviorObserved(BehaviorObservation observation);
    }

    /**
     * 创建行为收集器
     */
    public BehaviorCollector(@NonNull Context context, @NonNull IdentityManager identityManager) {
        this.context = context.getApplicationContext();
        this.identityManager = identityManager;
        this.executor = Executors.newSingleThreadExecutor();
        this.handler = new Handler(Looper.getMainLooper());

        // 定时刷新
        flushRunnable = () -> {
            if (enabled) {
                flushBuffer();
                handler.postDelayed(flushRunnable, FLUSH_INTERVAL_MS);
            }
        };
    }

    // === 行为收集方法 ===

    /**
     * 观察决策行为
     */
    public void observeDecision(@NonNull String decisionType, @NonNull Map<String, Object> details) {
        Map<String, Object> context = new HashMap<>();
        context.put("decision_type", decisionType);
        context.putAll(details);

        observe(BehaviorObservation.TYPE_DECISION, context);
    }

    /**
     * 观察交互行为
     */
    public void observeInteraction(@NonNull String interactionType, @NonNull Map<String, Object> details) {
        Map<String, Object> context = new HashMap<>();
        context.put("interaction_type", interactionType);
        context.putAll(details);

        observe(BehaviorObservation.TYPE_INTERACTION, context);
    }

    /**
     * 观察偏好行为
     */
    public void observePreference(@NonNull String preferenceType, @NonNull Map<String, Object> details) {
        Map<String, Object> context = new HashMap<>();
        context.put("preference_type", preferenceType);
        context.putAll(details);

        observe(BehaviorObservation.TYPE_PREFERENCE, context);
    }

    /**
     * 观察活动行为
     */
    public void observeActivity(@NonNull String activityType, @NonNull Map<String, Object> details) {
        Map<String, Object> context = new HashMap<>();
        context.put("activity_type", activityType);
        context.putAll(details);

        observe(BehaviorObservation.TYPE_ACTIVITY, context);
    }

    /**
     * 通用观察方法
     */
    public void observe(@NonNull String type, @NonNull Map<String, Object> context) {
        executor.execute(() -> {
            BehaviorObservation observation = new BehaviorObservation(type, context);

            // 自动推断性格变化
            observation.autoInfer();

            // 添加到缓存
            synchronized (behaviorBuffer) {
                behaviorBuffer.add(observation);

                // 达到阈值时刷新
                if (behaviorBuffer.size() >= BUFFER_SIZE) {
                    flushBuffer();
                }
            }

            // 通知监听器
            notifyListeners(observation);

            Log.d(TAG, "Observed: " + type + ", inferences: " + observation.getInferences().size());
        });
    }

    // === 便捷方法 ===

    /**
     * 记录购买决策
     */
    public void recordPurchase(@NonNull String item, double price, boolean isImpulse) {
        Map<String, Object> details = new HashMap<>();
        details.put("item", item);
        details.put("price", price);
        details.put("is_impulse", isImpulse);

        observeDecision(isImpulse ? "impulse_purchase" : "planned_purchase", details);
    }

    /**
     * 记录社交互动
     */
    public void recordSocialInteraction(@NonNull String type, int participantCount, boolean usedEmoji) {
        Map<String, Object> details = new HashMap<>();
        details.put("participant_count", participantCount);
        details.put("used_emoji", usedEmoji);
        details.put("is_group", participantCount > 1);

        String interactionType = participantCount > 1 ? "group_chats" : "private_chats";
        if (usedEmoji) {
            interactionType = "emoji_heavy";
        }

        observeInteraction(interactionType, details);
    }

    /**
     * 记录偏好设置
     */
    public void recordPreferenceChange(@NonNull String category, @NonNull String key, @NonNull Object value) {
        Map<String, Object> details = new HashMap<>();
        details.put("category", category);
        details.put("key", key);
        details.put("value", value);

        // 判断偏好类型
        String prefType = "routine_following";
        if (key.contains("new") || key.contains("experimental")) {
            prefType = "novel_trying";
        }
        if (key.contains("privacy")) {
            prefType = "privacy_focused";
        }
        if (key.contains("health")) {
            prefType = "healthy_eating";
        }

        observePreference(prefType, details);
    }

    /**
     * 记录活动
     */
    public void recordActivity(@NonNull String activity, long durationMs, boolean isNew) {
        Map<String, Object> details = new HashMap<>();
        details.put("activity", activity);
        details.put("duration_ms", durationMs);
        details.put("is_new", isNew);

        String activityType = isNew ? "exploring_new" : "regular_schedule";
        if (activity.contains("learn") || activity.contains("study")) {
            activityType = "learning_skills";
        }
        if (activity.contains("exercise") || activity.contains("run")) {
            activityType = "regular_schedule";
        }

        observeActivity(activityType, details);
    }

    // === 缓存管理 ===

    /**
     * 刷新缓存到 IdentityManager
     */
    private void flushBuffer() {
        List<BehaviorObservation> toFlush;
        synchronized (behaviorBuffer) {
            if (behaviorBuffer.isEmpty()) return;

            toFlush = new ArrayList<>(behaviorBuffer);
            behaviorBuffer.clear();
        }

        // 上报到 IdentityManager
        for (BehaviorObservation obs : toFlush) {
            identityManager.observeBehavior(obs.getType(), obs.getContext());
        }

        Log.i(TAG, "Flushed " + toFlush.size() + " behaviors to IdentityManager");
    }

    // === 启停控制 ===

    /**
     * 启用收集
     */
    public void enable() {
        if (enabled) return;
        enabled = true;
        handler.post(flushRunnable);
        Log.i(TAG, "BehaviorCollector enabled");
    }

    /**
     * 禁用收集
     */
    public void disable() {
        enabled = false;
        handler.removeCallbacks(flushRunnable);
        flushBuffer(); // 禁用前刷新缓存
        Log.i(TAG, "BehaviorCollector disabled");
    }

    /**
     * 检查是否启用
     */
    public boolean isEnabled() {
        return enabled;
    }

    // === 监听器 ===

    public void addListener(@NonNull BehaviorListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull BehaviorListener listener) {
        listeners.remove(listener);
    }

    private void notifyListeners(BehaviorObservation observation) {
        handler.post(() -> {
            for (BehaviorListener listener : listeners) {
                listener.onBehaviorObserved(observation);
            }
        });
    }

    // === 关闭 ===

    public void shutdown() {
        disable();
        flushBuffer();
        listeners.clear();
        executor.shutdown();
        Log.i(TAG, "BehaviorCollector shutdown");
    }
}