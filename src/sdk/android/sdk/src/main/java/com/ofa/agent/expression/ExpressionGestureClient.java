package com.ofa.agent.expression;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * ExpressionGesture状态客户端 (v5.3.0)
 *
 * 端侧接收 Center 推送的表情动作状态，用于动画渲染和表达。
 * 深层表情动作管理在 Center 端 ExpressionGestureEngine 完成。
 */
public class ExpressionGestureClient {

    private static volatile ExpressionGestureClient instance;

    // 当前状态
    private ExpressionGestureState currentState;

    // 监听器
    private final CopyOnWriteArrayList<ExpressionGestureStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private ExpressionGestureClient() {
        this.currentState = new ExpressionGestureState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static ExpressionGestureClient getInstance() {
        if (instance == null) {
            synchronized (ExpressionGestureClient.class) {
                if (instance == null) {
                    instance = new ExpressionGestureClient();
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
     * 接收 Center 推送的 ExpressionGesture 状态
     */
    public void receiveExpressionGestureState(@NonNull JSONObject stateJson) {
        try {
            ExpressionGestureState newState = ExpressionGestureState.fromJson(stateJson);
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
            ExpressionGestureState newState = ExpressionGestureState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull ExpressionGestureState newState) {
        this.currentState = newState;

        for (ExpressionGestureStateListener listener : listeners) {
            listener.onExpressionGestureStateChanged(newState);
        }
    }

    // === 状态获取 ===

    @NonNull
    public ExpressionGestureState getCurrentState() {
        return currentState;
    }

    // === Facial Expression Settings ===

    @NonNull
    public ExpressionGestureState.FacialExpressionSettings getFacialExpressionSettings() {
        return currentState.getFacialExpressionSettings();
    }

    public String getDefaultExpression() {
        return currentState.getFacialExpressionSettings().defaultExpression;
    }

    public double getExpressionIntensity() {
        return currentState.getFacialExpressionSettings().expressionIntensity;
    }

    public double getExpressionRange() {
        return currentState.getFacialExpressionSettings().expressionRange;
    }

    public double getBlinkRate() {
        return currentState.getFacialExpressionSettings().blinkRate;
    }

    public double getSmileTendency() {
        return currentState.getFacialExpressionSettings().smileTendency;
    }

    public boolean isHighExpressiveness() {
        return currentState.getFacialExpressionSettings().isHighExpressiveness();
    }

    public boolean isSubtleExpression() {
        return currentState.getFacialExpressionSettings().isSubtleExpression();
    }

    // === Body Gesture Settings ===

    @NonNull
    public ExpressionGestureState.BodyGestureSettings getBodyGestureSettings() {
        return currentState.getBodyGestureSettings();
    }

    public String getDefaultPosture() {
        return currentState.getBodyGestureSettings().defaultPosture;
    }

    public double getGestureIntensity() {
        return currentState.getBodyGestureSettings().gestureIntensity;
    }

    public double getGestureRange() {
        return currentState.getBodyGestureSettings().gestureRange;
    }

    public boolean isExpressiveGestures() {
        return currentState.getBodyGestureSettings().isExpressiveGestures();
    }

    public boolean isMinimalGestures() {
        return currentState.getBodyGestureSettings().isMinimalGestures();
    }

    // === Emotion Expression Mapping ===

    @NonNull
    public ExpressionGestureState.EmotionExpressionMapping getEmotionExpressionMapping() {
        return currentState.getEmotionExpressionMapping();
    }

    @Nullable
    public ExpressionGestureState.ExpressionMapping getMappingForEmotion(@NonNull String emotion) {
        return currentState.getEmotionExpressionMapping().getMappingForEmotion(emotion);
    }

    public boolean hasEmotionMapping(@NonNull String emotion) {
        return currentState.getEmotionExpressionMapping().hasMapping(emotion);
    }

    // === Social Gesture Settings ===

    @NonNull
    public ExpressionGestureState.SocialGestureSettings getSocialGestureSettings() {
        return currentState.getSocialGestureSettings();
    }

    public String getGreetingGesture() {
        return currentState.getSocialGestureSettings().greetingGesture;
    }

    public String getPartingGesture() {
        return currentState.getSocialGestureSettings().partingGesture;
    }

    public String getAgreementGesture() {
        return currentState.getSocialGestureSettings().agreementGesture;
    }

    public String getDisagreementGesture() {
        return currentState.getSocialGestureSettings().disagreementGesture;
    }

    public double getTouchComfortLevel() {
        return currentState.getSocialGestureSettings().touchComfortLevel;
    }

    public String getPreferredDistance() {
        return currentState.getSocialGestureSettings().preferredDistance;
    }

    // === Animation Settings ===

    @NonNull
    public ExpressionGestureState.AnimationSettings getAnimationSettings() {
        return currentState.getAnimationSettings();
    }

    public String getAnimationStyle() {
        return currentState.getAnimationSettings().animationStyle;
    }

    public boolean isLipSyncEnabled() {
        return currentState.getAnimationSettings().lipSyncEnabled;
    }

    public String getLipSyncQuality() {
        return currentState.getAnimationSettings().lipSyncQuality;
    }

    public boolean isBlinkAnimationEnabled() {
        return currentState.getAnimationSettings().blinkAnimationEnabled;
    }

    public boolean isBreathingAnimationEnabled() {
        return currentState.getAnimationSettings().breathingAnimationEnabled;
    }

    public double getBreathingRate() {
        return currentState.getAnimationSettings().breathingRate;
    }

    // === Context ===

    @NonNull
    public ExpressionGestureState.ExpressionGestureContext getContext() {
        return currentState.getContext();
    }

    // === Current Expression ===

    @NonNull
    public ExpressionGestureState.ExpressionState getCurrentExpression() {
        return currentState.getContext().currentExpression;
    }

    public String getCurrentExpressionName() {
        return currentState.getContext().currentExpression.expressionName;
    }

    public double getCurrentExpressionIntensity() {
        return currentState.getContext().currentExpression.intensity;
    }

    // === Current Gesture ===

    @NonNull
    public ExpressionGestureState.GestureState getCurrentGesture() {
        return currentState.getContext().currentGesture;
    }

    public String getCurrentGestureName() {
        return currentState.getContext().currentGesture.gestureName;
    }

    // === Recommended Expression ===

    @NonNull
    public ExpressionGestureState.ExpressionState getRecommendedExpression() {
        return currentState.getContext().recommendedExpression;
    }

    @NonNull
    public ExpressionGestureState.GestureState getRecommendedGesture() {
        return currentState.getContext().recommendedGesture;
    }

    // === Scene Adaptation ===

    @NonNull
    public ExpressionGestureState.ExpressionSceneAdaptation getSceneAdaptation() {
        return currentState.getContext().sceneAdaptation;
    }

    public String getCurrentScene() {
        return currentState.getContext().sceneAdaptation.scene;
    }

    public double getSceneFormalityLevel() {
        return currentState.getContext().sceneAdaptation.formalityLevel;
    }

    // === Emotion Adaptation ===

    @NonNull
    public ExpressionGestureState.ExpressionEmotionAdaptation getEmotionAdaptation() {
        return currentState.getContext().emotionAdaptation;
    }

    public String getEmotionForExpression() {
        return currentState.getContext().emotionAdaptation.currentEmotion;
    }

    public String getTargetExpression() {
        return currentState.getContext().emotionAdaptation.targetExpression;
    }

    public String getTargetGesture() {
        return currentState.getContext().emotionAdaptation.targetGesture;
    }

    // === Social Adaptation ===

    @NonNull
    public ExpressionGestureState.ExpressionSocialAdaptation getSocialAdaptation() {
        return currentState.getContext().socialAdaptation;
    }

    public String getSocialContext() {
        return currentState.getContext().socialAdaptation.socialContext;
    }

    // === Animation State ===

    @NonNull
    public ExpressionGestureState.AnimationState getAnimationState() {
        return currentState.getContext().animationState;
    }

    public String getCurrentAnimation() {
        return currentState.getContext().animationState.currentAnimation;
    }

    public double getAnimationProgress() {
        return currentState.getContext().animationState.animationProgress;
    }

    // === 决策辅助 ===

    /**
     * 获取表情建议
     */
    @NonNull
    public String getExpressionRecommendation() {
        ExpressionGestureState.ExpressionState recommended = getRecommendedExpression();
        String scene = getCurrentScene();
        String emotion = getEmotionForExpression();

        StringBuilder recommendation = new StringBuilder();

        // 基于场景
        if ("meeting".equals(scene) || "presentation".equals(scene)) {
            recommendation.append("建议保持专业、自然的表情。");
        } else if ("casual".equals(scene)) {
            recommendation.append("可以使用轻松、友好的表情。");
        }

        // 基于情绪
        if (!"neutral".equals(emotion)) {
            ExpressionGestureState.ExpressionMapping mapping = getMappingForEmotion(emotion);
            if (mapping != null) {
                recommendation.append(" 当前情绪(").append(emotion).append(")建议表情: ")
                    .append(mapping.expressionType);
            }
        }

        return recommendation.toString();
    }

    /**
     * 获取动作建议
     */
    @NonNull
    public String getGestureRecommendation() {
        ExpressionGestureState.GestureState recommended = getRecommendedGesture();
        String scene = getCurrentScene();
        String socialContext = getSocialContext();

        StringBuilder recommendation = new StringBuilder();

        // 基于社交场景
        if ("professional".equals(socialContext)) {
            recommendation.append("建议使用正式姿态，适度手势。");
        } else if ("intimate".equals(socialContext)) {
            recommendation.append("可以使用放松姿态，自然手势。");
        }

        // 基于场景
        double formality = getSceneFormalityLevel();
        if (formality > 0.7) {
            recommendation.append(" 保持端庄、正式的姿态。");
        } else if (formality < 0.3) {
            recommendation.append(" 可以放松姿态，增加手势表达。");
        }

        return recommendation.toString();
    }

    /**
     * 获取问候动作建议
     */
    @NonNull
    public String getGreetingRecommendation() {
        String socialContext = getSocialContext();
        String greetingGesture = getGreetingGesture();

        if ("professional".equals(socialContext)) {
            return "建议使用握手或点头作为问候";
        } else if ("intimate".equals(socialContext)) {
            return "可以使用挥手或拥抱作为问候";
        } else {
            return "建议使用" + greetingGesture + "作为问候";
        }
    }

    /**
     * 获取告别动作建议
     */
    @NonNull
    public String getPartingRecommendation() {
        String socialContext = getSocialContext();
        String partingGesture = getPartingGesture();

        if ("professional".equals(socialContext)) {
            return "建议使用点头或握手作为告别";
        } else if ("intimate".equals(socialContext)) {
            return "可以使用挥手或拥抱作为告别";
        } else {
            return "建议使用" + partingGesture + "作为告别";
        }
    }

    /**
     * 获取眼部表情建议
     */
    @NonNull
    public String getEyeExpressionRecommendation() {
        ExpressionGestureState.FacialExpressionSettings settings = getFacialExpressionSettings();
        double eyeContact = settings.eyeContactTendency;
        double blinkRate = settings.blinkRate;

        StringBuilder recommendation = new StringBuilder();

        if (eyeContact > 0.7) {
            recommendation.append("保持自然的眼神接触，传递真诚感。");
        } else if (eyeContact < 0.3) {
            recommendation.append("可以适度减少眼神接触，保持舒适距离。");
        }

        if (blinkRate > 20) {
            recommendation.append(" 注意眨眼频率，避免显得紧张。");
        } else if (blinkRate < 12) {
            recommendation.append(" 可以增加眨眼频率，显得更自然。");
        }

        return recommendation.toString();
    }

    /**
     * 获取微笑建议
     */
    @NonNull
    public String getSmileRecommendation() {
        ExpressionGestureState.FacialExpressionSettings settings = getFacialExpressionSettings();
        double smileTendency = settings.smileTendency;
        String smileType = settings.smileType;

        if (smileTendency > 0.7) {
            return "建议保持" + smileType + "微笑，展现亲和力";
        } else if (smileTendency > 0.4) {
            return "适度微笑，保持友好但不过分";
        } else {
            return "保持中性表情，根据情况适度微笑";
        }
    }

    /**
     * 获取姿态建议
     */
    @NonNull
    public String getPostureRecommendation() {
        ExpressionGestureState.BodyGestureSettings settings = getBodyGestureSettings();
        String posture = settings.defaultPosture;
        String scene = getCurrentScene();

        if ("meeting".equals(scene) || "presentation".equals(scene)) {
            return "建议保持自信、端正的姿态";
        } else if ("casual".equals(scene)) {
            return "可以使用放松、自然的姿态";
        } else {
            return "建议保持" + posture + "姿态";
        }
    }

    /**
     * 获取手势使用建议
     */
    @NonNull
    public String getHandGestureRecommendation() {
        ExpressionGestureState.BodyGestureSettings settings = getBodyGestureSettings();
        String style = settings.handGestureStyle;
        boolean enabled = settings.handGestureEnabled;

        if (!enabled) {
            return "建议减少手势使用，保持简洁";
        }

        if ("minimal".equals(style)) {
            return "建议使用简洁、有限的手势";
        } else if ("expressive".equals(style)) {
            return "可以使用丰富的手势增强表达";
        } else {
            return "适度使用手势辅助表达";
        }
    }

    /**
     * 获取社交姿态综合建议
     */
    @NonNull
    public String getComprehensiveSocialAdvice() {
        StringBuilder advice = new StringBuilder();

        // 问候建议
        advice.append("问候：").append(getGreetingRecommendation()).append("\n");

        // 表情建议
        advice.append("表情：").append(getExpressionRecommendation()).append("\n");

        // 姿态建议
        advice.append("姿态：").append(getPostureRecommendation()).append("\n");

        // 手势建议
        advice.append("手势：").append(getHandGestureRecommendation()).append("\n");

        // 告别建议
        advice.append("告别：").append(getPartingRecommendation());

        return advice.toString();
    }

    /**
     * 获取动画渲染建议
     */
    @NonNull
    public String getAnimationRecommendation() {
        ExpressionGestureState.AnimationSettings settings = getAnimationSettings();
        StringBuilder recommendation = new StringBuilder();

        // Lip sync
        if (settings.lipSyncEnabled) {
            recommendation.append("启用唇形同步，质量: ").append(settings.lipSyncQuality);
        }

        // Blink
        if (settings.blinkAnimationEnabled) {
            recommendation.append(", 自然眨眼动画");
        }

        // Breathing
        if (settings.breathingAnimationEnabled) {
            recommendation.append(", 呼吸动画频率: ").append(settings.breathingRate).append("/min");
        }

        // Eye movement
        if (settings.eyeMovementEnabled) {
            recommendation.append(", 眼球运动");
        }

        return recommendation.toString();
    }

    // === 行为上报 ===

    /**
     * 上报表情变化
     */
    @NonNull
    public JSONObject reportExpressionChange(@NonNull String newExpression, @NonNull String trigger) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "expression_change");
            report.put("new_expression", newExpression);
            report.put("trigger", trigger);
            report.put("previous_expression", getCurrentExpressionName());
            report.put("intensity", getCurrentExpressionIntensity());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报动作变化
     */
    @NonNull
    public JSONObject reportGestureChange(@NonNull String newGesture, @NonNull String trigger) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "gesture_change");
            report.put("new_gesture", newGesture);
            report.put("trigger", trigger);
            report.put("previous_gesture", getCurrentGestureName());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报动画完成
     */
    @NonNull
    public JSONObject reportAnimationComplete(@NonNull String animationName) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "animation_complete");
            report.put("animation_name", animationName);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    public void addListener(@NonNull ExpressionGestureStateListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull ExpressionGestureStateListener listener) {
        listeners.remove(listener);
    }

    public void clearListeners() {
        listeners.clear();
    }

    /**
     * ExpressionGesture状态监听器
     */
    public interface ExpressionGestureStateListener {
        void onExpressionGestureStateChanged(@NonNull ExpressionGestureState state);
    }

    // === 状态快照 ===

    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = ExpressionGestureState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    public void reset() {
        currentState = new ExpressionGestureState();

        for (ExpressionGestureStateListener listener : listeners) {
            listener.onExpressionGestureStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "ExpressionGestureClient{" +
                "expression='" + getCurrentExpressionName() + '\'' +
                ", intensity=" + getCurrentExpressionIntensity() +
                ", gesture='" + getCurrentGestureName() + '\'' +
                ", scene='" + getCurrentScene() + '\'' +
                ", emotion='" + getEmotionForExpression() + '\'' +
                '}';
    }
}