package com.ofa.agent.emotion;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 情绪状态客户端 (v4.0.0)
 *
 * 端侧接收 Center 推送的情绪状态，用于调整本地行为和表现。
 * 深层情绪管理和计算在 Center 端 EmotionEngine 完成。
 */
public class EmotionClient {

    private static volatile EmotionClient instance;

    // 当前状态
    private EmotionState currentEmotion;
    private DesireState currentDesire;

    // 监听器
    private final CopyOnWriteArrayList<EmotionStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private EmotionClient() {
        this.currentEmotion = new EmotionState();
        this.currentDesire = new DesireState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static EmotionClient getInstance() {
        if (instance == null) {
            synchronized (EmotionClient.class) {
                if (instance == null) {
                    instance = new EmotionClient();
                }
            }
        }
        return instance;
    }

    /**
     * 初始化
     */
    public void initialize(@Nullable String centerAddress) {
        this.centerAddress = centerAddress;
        this.syncEnabled = centerAddress != null && !centerAddress.isEmpty();
    }

    // === 状态接收 ===

    /**
     * 接收 Center 推送的情绪状态
     */
    public void receiveEmotionState(@NonNull JSONObject emotionJson) {
        try {
            EmotionState newState = EmotionState.fromJson(emotionJson);
            updateEmotionState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收 Center 推送的欲望状态
     */
    public void receiveDesireState(@NonNull JSONObject desireJson) {
        try {
            DesireState newState = DesireState.fromJson(desireJson);
            updateDesireState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收完整的情绪决策上下文
     */
    public void receiveEmotionContext(@NonNull JSONObject contextJson) {
        try {
            // 解析情绪状态
            JSONObject emotionJson = contextJson.optJSONObject("current_emotion");
            if (emotionJson != null) {
                EmotionState emotion = EmotionState.fromJson(emotionJson);
                // 补充影响因子
                JSONObject influence = contextJson.optJSONObject("influence_factors");
                if (influence != null) {
                    emotion.setRiskTolerance(influence.optDouble("risk_tolerance", 0.5));
                    emotion.setSocialTendency(influence.optDouble("social_tendency", 0.5));
                    emotion.setDecisionSpeed(influence.optDouble("decision_speed", 0.5));
                    emotion.setTrustLevel(influence.optDouble("trust_level", 0.5));
                    emotion.setCreativity(influence.optDouble("creativity", 0.5));
                }
                updateEmotionState(emotion);
            }

            // 解析欲望状态
            JSONObject desireJson = contextJson.optJSONObject("current_desire");
            if (desireJson != null) {
                DesireState desire = DesireState.fromJson(desireJson);
                desire.setSatisfactionLevel(contextJson.optDouble("satisfaction_level", 0.5));
                desire.setFrustrationLevel(contextJson.optDouble("frustration_level", 0.3));
                updateDesireState(desire);
            }

        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新情绪状态
     */
    private void updateEmotionState(@NonNull EmotionState newState) {
        EmotionState oldState = this.currentEmotion;
        this.currentEmotion = newState;

        // 通知监听器
        String oldMood = oldState.getCurrentMood();
        String newMood = newState.getCurrentMood();

        for (EmotionStateListener listener : listeners) {
            listener.onEmotionStateChanged(newState);
            if (!oldMood.equals(newMood)) {
                listener.onMoodChanged(oldMood, newMood);
            }
            if (newState.needsAttention()) {
                listener.onEmotionNeedsAttention(newState);
            }
        }
    }

    /**
     * 更新欲望状态
     */
    private void updateDesireState(@NonNull DesireState newState) {
        this.currentDesire = newState;

        // 通知监听器
        for (EmotionStateListener listener : listeners) {
            listener.onDesireStateChanged(newState);
            if (newState.needsAction()) {
                listener.onDesireNeedsAction(newState);
            }
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前情绪状态
     */
    @NonNull
    public EmotionState getCurrentEmotion() {
        return currentEmotion;
    }

    /**
     * 获取当前欲望状态
     */
    @NonNull
    public DesireState getCurrentDesire() {
        return currentDesire;
    }

    /**
     * 获取当前心境
     */
    @NonNull
    public String getCurrentMood() {
        return currentEmotion.getCurrentMood();
    }

    /**
     * 获取情绪描述
     */
    @NonNull
    public String getEmotionDescription() {
        return currentEmotion.getEmotionDescription();
    }

    /**
     * 获取驱动力描述
     */
    @NonNull
    public String getDriveDescription() {
        return currentDesire.getDriveDescription();
    }

    /**
     * 是否正面情绪
     */
    public boolean isPositiveEmotion() {
        return currentEmotion.isPositive();
    }

    /**
     * 是否负面情绪
     */
    public boolean isNegativeEmotion() {
        return currentEmotion.isNegative();
    }

    /**
     * 是否需要关注
     */
    public boolean needsAttention() {
        return currentEmotion.needsAttention();
    }

    /**
     * 是否适合社交
     */
    public boolean isSuitableForSocial() {
        return currentEmotion.isSuitableForSocial();
    }

    /**
     * 是否适合做决策
     */
    public boolean isSuitableForDecision() {
        return currentEmotion.isSuitableForDecision();
    }

    // === 行为调整建议 ===

    /**
     * 获取回复风格建议
     */
    @NonNull
    public String getRecommendedResponseStyle() {
        return currentEmotion.getRecommendedResponseStyle();
    }

    /**
     * 获取行为类型建议
     */
    @NonNull
    public String getRecommendedBehaviorType() {
        return currentDesire.getRecommendedBehaviorType();
    }

    /**
     * 获取风险容忍度
     */
    public double getRiskTolerance() {
        return currentEmotion.getRiskTolerance();
    }

    /**
     * 获取社交倾向
     */
    public double getSocialTendency() {
        return currentEmotion.getSocialTendency();
    }

    /**
     * 获取决策速度
     */
    public double getDecisionSpeed() {
        return currentEmotion.getDecisionSpeed();
    }

    /**
     * 获取信任度
     */
    public double getTrustLevel() {
        return currentEmotion.getTrustLevel();
    }

    /**
     * 获取创造性
     */
    public double getCreativity() {
        return currentEmotion.getCreativity();
    }

    // === 行为上报 ===

    /**
     * 上报事件触发情绪变化（发送到 Center）
     *
     * @param eventType 事件类型
     * @param eventDesc 事件描述
     * @return 上报结果
     */
    @NonNull
    public JSONObject reportEvent(@NonNull String eventType, @NonNull String eventDesc) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "event");
            report.put("event_type", eventType);
            report.put("event_desc", eventDesc);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报欲望满足
     *
     * @param desireType 欲望类型
     * @param amount 满足程度
     * @return 上报结果
     */
    @NonNull
    public JSONObject reportDesireSatisfied(@NonNull String desireType, double amount) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "desire_satisfied");
            report.put("desire_type", desireType);
            report.put("amount", amount);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报行为完成
     *
     * @param actionType 行为类型
     * @param result 结果描述
     * @return 上报结果
     */
    @NonNull
    public JSONObject reportActionCompleted(@NonNull String actionType, @NonNull String result) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "action_completed");
            report.put("action_type", actionType);
            report.put("result", result);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    /**
     * 添加状态监听器
     */
    public void addListener(@NonNull EmotionStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull EmotionStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 情绪状态监听器
     */
    public interface EmotionStateListener {
        /**
         * 情绪状态变化
         */
        void onEmotionStateChanged(@NonNull EmotionState state);

        /**
         * 心境变化
         */
        void onMoodChanged(@NonNull String oldMood, @NonNull String newMood);

        /**
         * 情绪需要关注（高强度负面情绪）
         */
        void onEmotionNeedsAttention(@NonNull EmotionState state);

        /**
         * 欲望状态变化
         */
        void onDesireStateChanged(@NonNull DesireState state);

        /**
         * 欲望需要行动（高挫折度）
         */
        void onDesireNeedsAction(@NonNull DesireState state);
    }

    // === 状态快照 ===

    /**
     * 获取状态快照
     */
    @NonNull
    public JSONObject getStateSnapshot() {
        JSONObject snapshot = new JSONObject();
        try {
            snapshot.put("emotion", currentEmotion.toJson());
            snapshot.put("desire", currentDesire.toJson());
            snapshot.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return snapshot;
    }

    /**
     * 恢复状态快照
     */
    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            JSONObject emotionJson = snapshot.optJSONObject("emotion");
            if (emotionJson != null) {
                currentEmotion = EmotionState.fromJson(emotionJson);
            }
            JSONObject desireJson = snapshot.optJSONObject("desire");
            if (desireJson != null) {
                currentDesire = DesireState.fromJson(desireJson);
            }
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentEmotion = new EmotionState();
        currentDesire = new DesireState();

        for (EmotionStateListener listener : listeners) {
            listener.onEmotionStateChanged(currentEmotion);
            listener.onDesireStateChanged(currentDesire);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "EmotionClient{" +
                "emotion=" + currentEmotion +
                ", desire=" + currentDesire +
                ", listeners=" + listeners.size() +
                '}';
    }
}