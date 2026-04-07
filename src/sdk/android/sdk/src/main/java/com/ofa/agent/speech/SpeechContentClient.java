package com.ofa.agent.speech;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * SpeechContent状态客户端 (v5.2.0)
 *
 * 端侧接收 Center 推送的表达内容状态，用于内容生成和表达。
 * 深层内容管理在 Center 端 SpeechContentEngine 完成。
 */
public class SpeechContentClient {

    private static volatile SpeechContentClient instance;

    // 当前状态
    private SpeechContentState currentState;

    // 监听器
    private final CopyOnWriteArrayList<SpeechContentStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private SpeechContentClient() {
        this.currentState = new SpeechContentState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static SpeechContentClient getInstance() {
        if (instance == null) {
            synchronized (SpeechContentClient.class) {
                if (instance == null) {
                    instance = new SpeechContentClient();
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
     * 接收 Center 推送的 SpeechContent 状态
     */
    public void receiveSpeechContentState(@NonNull JSONObject stateJson) {
        try {
            SpeechContentState newState = SpeechContentState.fromJson(stateJson);
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
            SpeechContentState newState = SpeechContentState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull SpeechContentState newState) {
        this.currentState = newState;

        for (SpeechContentStateListener listener : listeners) {
            listener.onSpeechContentStateChanged(newState);
        }
    }

    // === 状态获取 ===

    @NonNull
    public SpeechContentState getCurrentState() {
        return currentState;
    }

    // === 内容风格获取 ===

    @NonNull
    public SpeechContentState.ContentStyle getContentStyle() {
        return currentState.getContentStyle();
    }

    public String getToneStyle() {
        return currentState.getContentStyle().toneStyle;
    }

    public String getLanguageLevel() {
        return currentState.getContentStyle().languageLevel;
    }

    public double getDirectness() {
        return currentState.getContentStyle().directness;
    }

    public double getHumorTendency() {
        return currentState.getContentStyle().humorTendency;
    }

    public String getEmotionalColoring() {
        return currentState.getContentStyle().emotionalColoring;
    }

    public double getEnthusiasmLevel() {
        return currentState.getContentStyle().enthusiasmLevel;
    }

    public boolean isFormalTone() {
        return currentState.getContentStyle().isFormalTone();
    }

    public boolean isHighDirectness() {
        return currentState.getContentStyle().isHighDirectness();
    }

    public boolean isHumorous() {
        return currentState.getContentStyle().isHumorous();
    }

    // === 表达深度获取 ===

    @NonNull
    public SpeechContentState.ExpressionDepth getExpressionDepth() {
        return currentState.getExpressionDepth();
    }

    public String getThinkingDepth() {
        return currentState.getExpressionDepth().thinkingDepth;
    }

    public double getSelfDisclosureLevel() {
        return currentState.getExpressionDepth().selfDisclosureLevel;
    }

    public boolean isDeepThinking() {
        return currentState.getExpressionDepth().isDeepThinking();
    }

    // === 文化表达获取 ===

    @NonNull
    public SpeechContentState.CulturalExpression getCulturalExpression() {
        return currentState.getCulturalExpression();
    }

    public double getIndirectExpression() {
        return currentState.getCulturalExpression().indirectExpression;
    }

    public double getRespectLevel() {
        return currentState.getCulturalExpression().respectLevel;
    }

    public String getHonorificUsage() {
        return currentState.getCulturalExpression().honorificUsage;
    }

    public boolean isHighContext() {
        return currentState.getCulturalExpression().isHighContext();
    }

    public List<String> getSensitiveTopics() {
        return currentState.getCulturalExpression().sensitiveTopics;
    }

    // === 社交表达获取 ===

    @NonNull
    public SpeechContentState.SocialExpression getSocialExpression() {
        return currentState.getSocialExpression();
    }

    public String getProfessionalTone() {
        return currentState.getSocialExpression().professionalTone;
    }

    public double getIdentityConfidence() {
        return currentState.getSocialExpression().identityConfidence;
    }

    public boolean isHighConfidence() {
        return currentState.getSocialExpression().isHighConfidence();
    }

    // === 决策上下文获取 ===

    @NonNull
    public SpeechContentState.ContentDecisionContext getDecisionContext() {
        return currentState.getDecisionContext();
    }

    public String getRecommendedTone() {
        return currentState.getDecisionContext().recommendedTone;
    }

    public String getRecommendedFormality() {
        return currentState.getDecisionContext().recommendedFormality;
    }

    public String getRecommendedDepth() {
        return currentState.getDecisionContext().recommendedDepth;
    }

    public String getRecommendedLength() {
        return currentState.getDecisionContext().recommendedLength;
    }

    public double getRecommendedDirectness() {
        return currentState.getDecisionContext().recommendedDirectness;
    }

    public String getOpeningSuggestion() {
        return currentState.getDecisionContext().openingSuggestion;
    }

    public String getClosingSuggestion() {
        return currentState.getDecisionContext().closingSuggestion;
    }

    // === 场景适应 ===

    @NonNull
    public SpeechContentState.ContentSceneAdaptation getSceneAdaptation() {
        return currentState.getDecisionContext().sceneAdaptation;
    }

    public String getCurrentScene() {
        return currentState.getDecisionContext().sceneAdaptation.scene;
    }

    // === 情绪适应 ===

    @NonNull
    public SpeechContentState.ContentEmotionAdaptation getEmotionAdaptation() {
        return currentState.getDecisionContext().emotionAdaptation;
    }

    public String getCurrentEmotion() {
        return currentState.getDecisionContext().emotionAdaptation.currentEmotion;
    }

    public double getExpressionIntensity() {
        return currentState.getDecisionContext().emotionAdaptation.expressionIntensity;
    }

    // === 社交适应 ===

    @NonNull
    public SpeechContentState.ContentSocialAdaptation getSocialAdaptation() {
        return currentState.getDecisionContext().socialAdaptation;
    }

    public String getSocialContext() {
        return currentState.getDecisionContext().socialAdaptation.socialContext;
    }

    // === 文化适应 ===

    @NonNull
    public SpeechContentState.ContentCulturalAdaptation getCulturalAdaptation() {
        return currentState.getDecisionContext().culturalAdaptation;
    }

    // === 决策辅助 ===

    /**
     * 获取适合当前场景的语调建议
     */
    @NonNull
    public String getToneRecommendation() {
        String tone = getRecommendedTone();
        String scene = getCurrentScene();

        if ("meeting".equals(scene) || "presentation".equals(scene)) {
            return "建议使用专业、正式的语调";
        } else if ("casual".equals(scene)) {
            return "建议使用轻松、友好的语调";
        } else if ("enthusiastic".equals(tone)) {
            return "建议使用热情、积极的语调";
        } else {
            return "建议使用" + tone + "语调";
        }
    }

    /**
     * 获取适合当前场景的正式度建议
     */
    @NonNull
    public String getFormalityRecommendation() {
        String formality = getRecommendedFormality();
        String socialContext = getSocialContext();

        if ("professional".equals(socialContext) || "formal".equals(socialContext)) {
            return "建议使用正式语言，保持专业态度";
        } else if ("intimate".equals(socialContext)) {
            return "可以使用轻松、随意的表达方式";
        } else {
            return "建议使用" + formality + "的表达方式";
        }
    }

    /**
     * 获取适合当前场景的深度建议
     */
    @NonNull
    public String getDepthRecommendation() {
        String depth = getRecommendedDepth();
        String socialContext = getSocialContext();

        if ("intimate".equals(socialContext)) {
            return "可以深入表达个人想法和感受";
        } else if ("public".equals(socialContext)) {
            return "建议保持适度深度，避免过于私密的话题";
        } else {
            return "建议使用" + depth + "的表达深度";
        }
    }

    /**
     * 获取表达建议
     */
    @NonNull
    public String getExpressionRecommendation() {
        StringBuilder recommendation = new StringBuilder();

        // 语调建议
        recommendation.append(getToneRecommendation()).append("。");

        // 直接度建议
        double directness = getRecommendedDirectness();
        if (directness > 0.6) {
            recommendation.append("可以直接、明确地表达观点。");
        } else if (directness < 0.4) {
            recommendation.append("建议使用委婉、含蓄的表达方式。");
        }

        // 情感色彩建议
        String coloring = getEmotionalColoring();
        if ("warm".equals(coloring)) {
            recommendation.append("表达可以温暖、亲切。");
        } else if ("passionate".equals(coloring)) {
            recommendation.append("表达可以充满热情。");
        }

        return recommendation.toString();
    }

    /**
     * 获取开场白建议
     */
    @NonNull
    public String getOpeningRecommendation() {
        String opening = getOpeningSuggestion();
        if (opening != null && !opening.isEmpty()) {
            return opening;
        }

        String socialContext = getSocialContext();
        if ("professional".equals(socialContext)) {
            return "您好，有什么我可以帮助您的吗？";
        } else if ("intimate".equals(socialContext)) {
            return "嗨，好久不见！";
        } else {
            return "你好！";
        }
    }

    /**
     * 获取结束语建议
     */
    @NonNull
    public String getClosingRecommendation() {
        String closing = getClosingSuggestion();
        if (closing != null && !closing.isEmpty()) {
            return closing;
        }

        String socialContext = getSocialContext();
        if ("professional".equals(socialContext)) {
            return "感谢您的时间，祝您愉快。";
        } else if ("intimate".equals(socialContext)) {
            return "下次再聊！";
        } else {
            return "再见！";
        }
    }

    /**
     * 获取需要避免的话题
     */
    @NonNull
    public List<String> getTopicsToAvoid() {
        return currentState.getDecisionContext().keyTopicsToAvoid;
    }

    /**
     * 检查话题是否敏感
     */
    public boolean isSensitiveTopic(@NonNull String topic) {
        List<String> sensitive = getSensitiveTopics();
        for (String s : sensitive) {
            if (topic.contains(s) || s.contains(topic)) {
                return true;
            }
        }
        return false;
    }

    /**
     * 获取综合表达建议
     */
    @NonNull
    public String getComprehensiveAdvice() {
        StringBuilder advice = new StringBuilder();

        // 开场建议
        advice.append("开场：").append(getOpeningRecommendation()).append("\n");

        // 表达风格
        advice.append("风格：").append(getExpressionRecommendation()).append("\n");

        // 深度建议
        advice.append("深度：").append(getDepthRecommendation()).append("\n");

        // 结束语建议
        advice.append("结束：").append(getClosingRecommendation());

        return advice.toString();
    }

    // === 行为上报 ===

    /**
     * 上报内容生成
     */
    @NonNull
    public JSONObject reportContentGeneration(@NonNull String contentType, @NonNull String topic) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "content_generation");
            report.put("content_type", contentType);
            report.put("topic", topic);
            report.put("tone_used", getToneStyle());
            report.put("formality", getRecommendedFormality());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报风格调整
     */
    @NonNull
    public JSONObject reportStyleAdjustment(@NonNull String newTone, @NonNull String reason) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "style_adjustment");
            report.put("new_tone", newTone);
            report.put("reason", reason);
            report.put("previous_tone", getToneStyle());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报深度调整
     */
    @NonNull
    public JSONObject reportDepthAdjustment(@NonNull String newDepth, @NonNull String context) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "depth_adjustment");
            report.put("new_depth", newDepth);
            report.put("context", context);
            report.put("previous_depth", getThinkingDepth());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    public void addListener(@NonNull SpeechContentStateListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull SpeechContentStateListener listener) {
        listeners.remove(listener);
    }

    public void clearListeners() {
        listeners.clear();
    }

    /**
     * SpeechContent状态监听器
     */
    public interface SpeechContentStateListener {
        void onSpeechContentStateChanged(@NonNull SpeechContentState state);
    }

    // === 状态快照 ===

    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = SpeechContentState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    public void reset() {
        currentState = new SpeechContentState();

        for (SpeechContentStateListener listener : listeners) {
            listener.onSpeechContentStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "SpeechContentClient{" +
                "tone='" + getToneStyle() + '\'' +
                ", directness=" + getDirectness() +
                ", depth='" + getThinkingDepth() + '\'' +
                ", scene='" + getCurrentScene() + '\'' +
                '}';
    }
}