package com.ofa.agent.behavior;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 情绪行为状态客户端 (v4.5.0)
 *
 * 端侧接收 Center 推送的情绪行为状态，用于调整决策和表达。
 * 深层情绪行为管理在 Center 端 EmotionBehaviorEngine 完成。
 */
public class EmotionBehaviorClient {

    private static volatile EmotionBehaviorClient instance;

    // 当前状态
    private EmotionBehaviorState currentState;

    // 监听器
    private final CopyOnWriteArrayList<EmotionBehaviorStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private EmotionBehaviorClient() {
        this.currentState = new EmotionBehaviorState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static EmotionBehaviorClient getInstance() {
        if (instance == null) {
            synchronized (EmotionBehaviorClient.class) {
                if (instance == null) {
                    instance = new EmotionBehaviorClient();
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
     * 接收 Center 推送的情绪行为状态
     */
    public void receiveEmotionBehaviorState(@NonNull JSONObject stateJson) {
        try {
            EmotionBehaviorState newState = EmotionBehaviorState.fromJson(stateJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收决策上下文
     */
    public void receiveDecisionContext(@NonNull JSONObject contextJson) {
        try {
            EmotionBehaviorState newState = EmotionBehaviorState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull EmotionBehaviorState newState) {
        this.currentState = newState;

        for (EmotionBehaviorStateListener listener : listeners) {
            listener.onEmotionBehaviorStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前情绪行为状态
     */
    @NonNull
    public EmotionBehaviorState getCurrentState() {
        return currentState;
    }

    // === 决策影响获取 ===

    /**
     * 获取决策影响
     */
    @NonNull
    public EmotionBehaviorState.DecisionInfluence getDecisionInfluence() {
        return currentState.getDecisionInfluence();
    }

    /**
     * 获取风险承受度
     */
    public double getRiskTolerance() {
        return currentState.getDecisionInfluence().riskTolerance;
    }

    /**
     * 获取冲动控制
     */
    public double getImpulseControl() {
        return currentState.getDecisionInfluence().impulseControl;
    }

    /**
     * 获取社交趋近度
     */
    public double getSocialApproach() {
        return currentState.getDecisionInfluence().socialApproach;
    }

    /**
     * 获取信任水平
     */
    public double getTrustLevel() {
        return currentState.getDecisionInfluence().trustLevel;
    }

    /**
     * 获取合作倾向
     */
    public double getCooperationTendency() {
        return currentState.getDecisionInfluence().cooperationTendency;
    }

    /**
     * 获取决策风格
     */
    @NonNull
    public String getDecisionStyle() {
        return currentState.getDecisionInfluence().getDecisionStyle();
    }

    // === 表达影响获取 ===

    /**
     * 获取表达影响
     */
    @NonNull
    public EmotionBehaviorState.ExpressionInfluence getExpressionInfluence() {
        return currentState.getExpressionInfluence();
    }

    /**
     * 获取温暖程度
     */
    public double getWarmthLevel() {
        return currentState.getExpressionInfluence().warmthLevel;
    }

    /**
     * 获取热情程度
     */
    public double getEnthusiasmLevel() {
        return currentState.getExpressionInfluence().enthusiasmLevel;
    }

    /**
     * 获取表达倾向
     */
    @NonNull
    public String getExpressionTendency() {
        return currentState.getExpressionInfluence().expressionTendency;
    }

    /**
     * 获取沟通风格
     */
    @NonNull
    public String getCommunicationStyle() {
        return currentState.getExpressionInfluence().getCommunicationStyle();
    }

    /**
     * 获取回应速度
     */
    @NonNull
    public String getResponseSpeed() {
        return currentState.getExpressionInfluence().responseSpeed;
    }

    // === 行为指导获取 ===

    /**
     * 获取行为指导
     */
    @NonNull
    public EmotionBehaviorState.BehaviorGuidance getGuidance() {
        return currentState.getGuidance();
    }

    /**
     * 是否应延迟决策
     */
    public boolean shouldDelayDecision() {
        return currentState.shouldDelayDecision();
    }

    /**
     * 获取延迟原因
     */
    @Nullable
    public String getDelayReason() {
        return currentState.getGuidance().delayReason;
    }

    /**
     * 是否应表达情绪
     */
    public boolean shouldExpressEmotion() {
        return currentState.shouldExpressEmotion();
    }

    /**
     * 是否应社交互动
     */
    public boolean shouldInteract() {
        return currentState.getGuidance().shouldInteract;
    }

    /**
     * 获取风险警告
     */
    @Nullable
    public String getRiskWarning() {
        return currentState.getRiskWarning();
    }

    /**
     * 获取冲动风险
     */
    public double getImpulseRisk() {
        return currentState.getGuidance().impulseRisk;
    }

    /**
     * 获取调节建议
     */
    @Nullable
    public String getRegulationSuggestion() {
        return currentState.getRegulationSuggestion();
    }

    // === 情绪状态获取 ===

    /**
     * 获取情绪状态摘要
     */
    @NonNull
    public EmotionBehaviorState.EmotionStateSummary getEmotionState() {
        return currentState.getEmotionState();
    }

    /**
     * 获取主导情绪
     */
    @NonNull
    public String getDominantEmotion() {
        return currentState.getEmotionState().dominantEmotion;
    }

    /**
     * 获取情绪强度
     */
    public double getEmotionIntensity() {
        return currentState.getEmotionState().intensity;
    }

    /**
     * 是否积极情绪
     */
    public boolean isPositiveEmotion() {
        return currentState.getEmotionState().isPositive();
    }

    /**
     * 是否消极情绪
     */
    public boolean isNegativeEmotion() {
        return currentState.getEmotionState().isNegative();
    }

    /**
     * 是否高唤醒
     */
    public boolean isHighArousal() {
        return currentState.getEmotionState().isHighArousal();
    }

    // === 推荐获取 ===

    /**
     * 获取推荐行为
     */
    @NonNull
    public List<EmotionBehaviorState.BehaviorRecommendation> getRecommendedBehaviors() {
        return currentState.getRecommendedBehaviors();
    }

    /**
     * 获取推荐应对策略
     */
    @NonNull
    public List<EmotionBehaviorState.CopingRecommendation> getRecommendedCoping() {
        return currentState.getRecommendedCoping();
    }

    // === 决策辅助 ===

    /**
     * 是否建议做重大决策
     */
    public boolean shouldMakeMajorDecision() {
        // 情绪强度过高或冲动风险高时不建议
        if (getEmotionIntensity() > 0.7) return false;
        if (getImpulseRisk() > 0.6) return false;
        if (shouldDelayDecision()) return false;
        return true;
    }

    /**
     * 是否建议社交互动
     */
    public boolean shouldSocialize() {
        // 社交趋近且不回避时建议
        return isSociallyApproach() && shouldInteract();
    }

    /**
     * 是否建议尝试新事物
     */
    public boolean shouldTryNewThings() {
        EmotionBehaviorState.DecisionInfluence d = currentState.getDecisionInfluence();
        return d.noveltySeeking > 0.6 && d.riskTolerance > 0.5;
    }

    /**
     * 获取推荐决策方式
     */
    @NonNull
    public String getRecommendedDecisionApproach() {
        String style = getDecisionStyle();
        switch (style) {
            case "intuitive":
                return "跟随直觉，快速决策";
            case "analytical":
                return "深思熟虑，分析利弊";
            case "decisive":
                return "果断决策，行动优先";
            default:
                return "平衡分析，适时决策";
        }
    }

    /**
     * 获取推荐表达方式
     */
    @NonNull
    public String getRecommendedExpressionApproach() {
        String tendency = getExpressionTendency();
        switch (tendency) {
            case "express":
                return "直接表达情感";
            case "suppress":
                return "内敛含蓄";
            case "mask":
                return "适度隐藏真实情绪";
            default:
                return "根据情境灵活表达";
        }
    }

    /**
     * 获取推荐社交方式
     */
    @NonNull
    public String getRecommendedSocialApproach() {
        if (isSociallyApproach()) {
            return "主动社交，积极互动";
        } else {
            return "保持距离，适度独处";
        }
    }

    /**
     * 获取风险提示
     */
    @Nullable
    public String getRiskAlert() {
        StringBuilder alert = new StringBuilder();

        if (getImpulseRisk() > 0.6) {
            alert.append("冲动风险较高，建议冷静后再决策。");
        }

        if (getEmotionIntensity() > 0.7) {
            String emotion = getDominantEmotion();
            if ("anger".equals(emotion) || "fear".equals(emotion)) {
                alert.append("情绪强度较高，可能影响判断。");
            }
        }

        String riskWarning = getRiskWarning();
        if (riskWarning != null && !riskWarning.isEmpty()) {
            alert.append(riskWarning);
        }

        return alert.length() > 0 ? alert.toString() : null;
    }

    // === 行为上报 ===

    /**
     * 上报决策行为
     */
    @NonNull
    public JSONObject reportDecision(@NonNull String decisionType, boolean wasImpulsive, double satisfaction) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "decision_report");
            report.put("decision_type", decisionType);
            report.put("was_impulsive", wasImpulsive);
            report.put("satisfaction", satisfaction);
            report.put("emotion_state", getDominantEmotion());
            report.put("emotion_intensity", getEmotionIntensity());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报表达行为
     */
    @NonNull
    public JSONObject reportExpression(@NonNull String expressionType, boolean wasAuthentic, double effectiveness) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "expression_report");
            report.put("expression_type", expressionType);
            report.put("was_authentic", wasAuthentic);
            report.put("effectiveness", effectiveness);
            report.put("emotion_state", getDominantEmotion());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报应对策略使用
     */
    @NonNull
    public JSONObject reportCopingUsed(@NonNull String strategyId, boolean wasEffective) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "coping_used");
            report.put("strategy_id", strategyId);
            report.put("was_effective", wasEffective);
            report.put("emotion_state", getDominantEmotion());
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
    public void addListener(@NonNull EmotionBehaviorStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull EmotionBehaviorStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 情绪行为状态监听器
     */
    public interface EmotionBehaviorStateListener {
        /**
         * 情绪行为状态变化
         */
        void onEmotionBehaviorStateChanged(@NonNull EmotionBehaviorState state);
    }

    // === 状态快照 ===

    /**
     * 获取状态快照
     */
    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    /**
     * 恢复状态快照
     */
    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = EmotionBehaviorState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new EmotionBehaviorState();

        for (EmotionBehaviorStateListener listener : listeners) {
            listener.onEmotionBehaviorStateChanged(currentState);
        }
    }

    // === 私有方法 ===

    private boolean isSociallyApproach() {
        return currentState.getDecisionInfluence().isSociallyApproach();
    }

    @NonNull
    @Override
    public String toString() {
        return "EmotionBehaviorClient{" +
                "emotion=" + getDominantEmotion() +
                ", intensity=" + getEmotionIntensity() +
                ", riskTolerance=" + getRiskTolerance() +
                '}';
    }
}