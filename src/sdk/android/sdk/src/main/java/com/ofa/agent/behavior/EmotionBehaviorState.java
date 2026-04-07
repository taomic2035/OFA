package com.ofa.agent.behavior;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * 情绪行为状态模型 (v4.5.0)
 *
 * 端侧只接收 Center 推送的情绪行为状态，用于调整决策和表达。
 * 深层情绪行为管理在 Center 端 EmotionBehaviorEngine 完成。
 */
public class EmotionBehaviorState {

    // === 决策影响 ===
    private DecisionInfluence decisionInfluence;

    // === 表达影响 ===
    private ExpressionInfluence expressionInfluence;

    // === 行为指导 ===
    private BehaviorGuidance guidance;

    // === 推荐行为 ===
    private List<BehaviorRecommendation> recommendedBehaviors;

    // === 推荐应对策略 ===
    private List<CopingRecommendation> recommendedCoping;

    // === 当前情绪状态 ===
    private EmotionStateSummary emotionState;

    // === 时间属性 ===
    private long timestamp;

    public EmotionBehaviorState() {
        this.decisionInfluence = new DecisionInfluence();
        this.expressionInfluence = new ExpressionInfluence();
        this.guidance = new BehaviorGuidance();
        this.recommendedBehaviors = new ArrayList<>();
        this.recommendedCoping = new ArrayList<>();
        this.emotionState = new EmotionStateSummary();
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static EmotionBehaviorState fromJson(@NonNull JSONObject json) throws JSONException {
        EmotionBehaviorState state = new EmotionBehaviorState();

        // 决策影响
        JSONObject decisionJson = json.optJSONObject("decision_influence");
        if (decisionJson != null) {
            state.decisionInfluence = DecisionInfluence.fromJson(decisionJson);
        }

        // 表达影响
        JSONObject expressionJson = json.optJSONObject("expression_influence");
        if (expressionJson != null) {
            state.expressionInfluence = ExpressionInfluence.fromJson(expressionJson);
        }

        // 行为指导
        JSONObject guidanceJson = json.optJSONObject("behavior_guidance");
        if (guidanceJson != null) {
            state.guidance = BehaviorGuidance.fromJson(guidanceJson);
        }

        // 推荐行为
        JSONArray behaviorsJson = json.optJSONArray("recommended_behaviors");
        if (behaviorsJson != null) {
            state.recommendedBehaviors = new ArrayList<>();
            for (int i = 0; i < behaviorsJson.length(); i++) {
                state.recommendedBehaviors.add(BehaviorRecommendation.fromJson(behaviorsJson.getJSONObject(i)));
            }
        }

        // 推荐应对策略
        JSONArray copingJson = json.optJSONArray("recommended_coping_strategies");
        if (copingJson != null) {
            state.recommendedCoping = new ArrayList<>();
            for (int i = 0; i < copingJson.length(); i++) {
                state.recommendedCoping.add(CopingRecommendation.fromJson(copingJson.getJSONObject(i)));
            }
        }

        // 情绪状态
        JSONObject emotionJson = json.optJSONObject("current_emotion_state");
        if (emotionJson != null) {
            state.emotionState = EmotionStateSummary.fromJson(emotionJson);
        }

        state.timestamp = json.optLong("timestamp", System.currentTimeMillis());

        return state;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            if (decisionInfluence != null) {
                json.put("decision_influence", decisionInfluence.toJson());
            }
            if (expressionInfluence != null) {
                json.put("expression_influence", expressionInfluence.toJson());
            }
            if (guidance != null) {
                json.put("behavior_guidance", guidance.toJson());
            }
            if (recommendedBehaviors != null) {
                JSONArray behaviorsArray = new JSONArray();
                for (BehaviorRecommendation b : recommendedBehaviors) {
                    behaviorsArray.put(b.toJson());
                }
                json.put("recommended_behaviors", behaviorsArray);
            }
            if (recommendedCoping != null) {
                JSONArray copingArray = new JSONArray();
                for (CopingRecommendation c : recommendedCoping) {
                    copingArray.put(c.toJson());
                }
                json.put("recommended_coping_strategies", copingArray);
            }
            if (emotionState != null) {
                json.put("current_emotion_state", emotionState.toJson());
            }
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    // === 便捷方法 ===

    /**
     * 是否高风险偏好
     */
    public boolean isHighRiskTolerance() {
        return decisionInfluence != null && decisionInfluence.isHighRiskTolerance();
    }

    /**
     * 是否冲动
     */
    public boolean isImpulsive() {
        return decisionInfluence != null && decisionInfluence.isImpulsive();
    }

    /**
     * 是否社交趋近
     */
    public boolean isSociallyApproach() {
        return decisionInfluence != null && decisionInfluence.isSociallyApproach();
    }

    /**
     * 是否表达型
     */
    public boolean isExpressive() {
        return expressionInfluence != null && expressionInfluence.isExpressive();
    }

    /**
     * 是否应延迟决策
     */
    public boolean shouldDelayDecision() {
        return guidance != null && guidance.shouldDelay;
    }

    /**
     * 是否应表达情绪
     */
    public boolean shouldExpressEmotion() {
        return guidance != null && guidance.shouldExpress;
    }

    /**
     * 获取风险警告
     */
    @Nullable
    public String getRiskWarning() {
        return guidance != null ? guidance.riskWarning : null;
    }

    /**
     * 获取调节建议
     */
    @Nullable
    public String getRegulationSuggestion() {
        return guidance != null ? guidance.regulationSuggestion : null;
    }

    // === Getter/Setter ===

    public DecisionInfluence getDecisionInfluence() { return decisionInfluence; }
    public void setDecisionInfluence(DecisionInfluence decisionInfluence) { this.decisionInfluence = decisionInfluence; }

    public ExpressionInfluence getExpressionInfluence() { return expressionInfluence; }
    public void setExpressionInfluence(ExpressionInfluence expressionInfluence) { this.expressionInfluence = expressionInfluence; }

    public BehaviorGuidance getGuidance() { return guidance; }
    public void setGuidance(BehaviorGuidance guidance) { this.guidance = guidance; }

    public List<BehaviorRecommendation> getRecommendedBehaviors() { return recommendedBehaviors; }
    public void setRecommendedBehaviors(List<BehaviorRecommendation> recommendedBehaviors) { this.recommendedBehaviors = recommendedBehaviors; }

    public List<CopingRecommendation> getRecommendedCoping() { return recommendedCoping; }
    public void setRecommendedCoping(List<CopingRecommendation> recommendedCoping) { this.recommendedCoping = recommendedCoping; }

    public EmotionStateSummary getEmotionState() { return emotionState; }
    public void setEmotionState(EmotionStateSummary emotionState) { this.emotionState = emotionState; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "EmotionBehaviorState{" +
                "emotion=" + (emotionState != null ? emotionState.dominantEmotion : "null") +
                ", riskTolerance=" + (decisionInfluence != null ? decisionInfluence.riskTolerance : 0) +
                '}';
    }

    /**
     * 决策影响
     */
    public static class DecisionInfluence {
        public double riskTolerance;
        public double riskAversion;
        public double impulseControl;
        public double delayedGratification;
        public double socialApproach;
        public double socialAvoidance;
        public double trustLevel;
        public double cooperationTendency;
        public double decisionSpeed;
        public double deliberationLevel;
        public double decisiveness;
        public double noveltySeeking;
        public double familiarityPreference;
        public double qualityFocus;
        public double priceSensitivity;
        public String dominantEmotion;
        public double emotionIntensity;

        public DecisionInfluence() {
            this.riskTolerance = 0.5;
            this.riskAversion = 0.5;
            this.impulseControl = 0.6;
            this.delayedGratification = 0.5;
            this.socialApproach = 0.6;
            this.socialAvoidance = 0.4;
            this.trustLevel = 0.5;
            this.cooperationTendency = 0.6;
            this.decisionSpeed = 0.5;
            this.deliberationLevel = 0.5;
            this.decisiveness = 0.5;
            this.noveltySeeking = 0.5;
            this.familiarityPreference = 0.5;
            this.qualityFocus = 0.5;
            this.priceSensitivity = 0.5;
            this.dominantEmotion = "neutral";
            this.emotionIntensity = 0.3;
        }

        public static DecisionInfluence fromJson(JSONObject json) {
            DecisionInfluence d = new DecisionInfluence();
            d.riskTolerance = json.optDouble("risk_tolerance", 0.5);
            d.riskAversion = json.optDouble("risk_aversion", 0.5);
            d.impulseControl = json.optDouble("impulse_control", 0.6);
            d.delayedGratification = json.optDouble("delayed_gratification", 0.5);
            d.socialApproach = json.optDouble("social_approach", 0.6);
            d.socialAvoidance = json.optDouble("social_avoidance", 0.4);
            d.trustLevel = json.optDouble("trust_level", 0.5);
            d.cooperationTendency = json.optDouble("cooperation_tendency", 0.6);
            d.decisionSpeed = json.optDouble("decision_speed", 0.5);
            d.deliberationLevel = json.optDouble("deliberation_level", 0.5);
            d.decisiveness = json.optDouble("decisiveness", 0.5);
            d.noveltySeeking = json.optDouble("novelty_seeking", 0.5);
            d.familiarityPreference = json.optDouble("familiarity_preference", 0.5);
            d.qualityFocus = json.optDouble("quality_focus", 0.5);
            d.priceSensitivity = json.optDouble("price_sensitivity", 0.5);
            d.dominantEmotion = json.optString("dominant_emotion", "neutral");
            d.emotionIntensity = json.optDouble("emotion_intensity", 0.3);
            return d;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("risk_tolerance", riskTolerance);
                json.put("risk_aversion", riskAversion);
                json.put("impulse_control", impulseControl);
                json.put("delayed_gratification", delayedGratification);
                json.put("social_approach", socialApproach);
                json.put("social_avoidance", socialAvoidance);
                json.put("trust_level", trustLevel);
                json.put("cooperation_tendency", cooperationTendency);
                json.put("decision_speed", decisionSpeed);
                json.put("deliberation_level", deliberationLevel);
                json.put("decisiveness", decisiveness);
                json.put("novelty_seeking", noveltySeeking);
                json.put("familiarity_preference", familiarityPreference);
                json.put("quality_focus", qualityFocus);
                json.put("price_sensitivity", priceSensitivity);
                json.put("dominant_emotion", dominantEmotion);
                json.put("emotion_intensity", emotionIntensity);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isHighRiskTolerance() { return riskTolerance > 0.6; }
        public boolean isImpulsive() { return impulseControl < 0.4 && decisionSpeed > 0.6; }
        public boolean isSociallyApproach() { return socialApproach > socialAvoidance; }

        public String getDecisionStyle() {
            if (decisionSpeed > 0.7 && deliberationLevel < 0.4) return "intuitive";
            if (deliberationLevel > 0.7 && decisionSpeed < 0.4) return "analytical";
            if (decisiveness > 0.6) return "decisive";
            return "balanced";
        }
    }

    /**
     * 表达影响
     */
    public static class ExpressionInfluence {
        public String toneStyle;
        public double formalityLevel;
        public double warmthLevel;
        public double enthusiasmLevel;
        public String wordChoice;
        public String sentenceLength;
        public double complexityLevel;
        public double metaphorUse;
        public double emojiUsage;
        public String emojiType;
        public double exclamationUse;
        public String responseSpeed;
        public double proactiveness;
        public double detailLevel;
        public double humorLevel;
        public String voiceTone;
        public String speechSpeed;
        public String volumeLevel;
        public double pauseFrequency;
        public String underlyingEmotion;
        public String expressionTendency;

        public ExpressionInfluence() {
            this.toneStyle = "casual";
            this.formalityLevel = 0.4;
            this.warmthLevel = 0.6;
            this.enthusiasmLevel = 0.5;
            this.wordChoice = "neutral";
            this.sentenceLength = "medium";
            this.complexityLevel = 0.5;
            this.metaphorUse = 0.3;
            this.emojiUsage = 0.4;
            this.emojiType = "varied";
            this.exclamationUse = 0.3;
            this.responseSpeed = "thoughtful";
            this.proactiveness = 0.5;
            this.detailLevel = 0.5;
            this.humorLevel = 0.4;
            this.voiceTone = "modulated";
            this.speechSpeed = "normal";
            this.volumeLevel = "normal";
            this.pauseFrequency = 0.3;
            this.underlyingEmotion = "neutral";
            this.expressionTendency = "express";
        }

        public static ExpressionInfluence fromJson(JSONObject json) {
            ExpressionInfluence e = new ExpressionInfluence();
            e.toneStyle = json.optString("tone_style", "casual");
            e.formalityLevel = json.optDouble("formality_level", 0.4);
            e.warmthLevel = json.optDouble("warmth_level", 0.6);
            e.enthusiasmLevel = json.optDouble("enthusiasm_level", 0.5);
            e.wordChoice = json.optString("word_choice", "neutral");
            e.sentenceLength = json.optString("sentence_length", "medium");
            e.complexityLevel = json.optDouble("complexity_level", 0.5);
            e.metaphorUse = json.optDouble("metaphor_use", 0.3);
            e.emojiUsage = json.optDouble("emoji_usage", 0.4);
            e.emojiType = json.optString("emoji_type", "varied");
            e.exclamationUse = json.optDouble("exclamation_use", 0.3);
            e.responseSpeed = json.optString("response_speed", "thoughtful");
            e.proactiveness = json.optDouble("proactiveness", 0.5);
            e.detailLevel = json.optDouble("detail_level", 0.5);
            e.humorLevel = json.optDouble("humor_level", 0.4);
            e.voiceTone = json.optString("voice_tone", "modulated");
            e.speechSpeed = json.optString("speech_speed", "normal");
            e.volumeLevel = json.optString("volume_level", "normal");
            e.pauseFrequency = json.optDouble("pause_frequency", 0.3);
            e.underlyingEmotion = json.optString("underlying_emotion", "neutral");
            e.expressionTendency = json.optString("expression_tendency", "express");
            return e;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("tone_style", toneStyle);
                json.put("formality_level", formalityLevel);
                json.put("warmth_level", warmthLevel);
                json.put("enthusiasm_level", enthusiasmLevel);
                json.put("word_choice", wordChoice);
                json.put("sentence_length", sentenceLength);
                json.put("complexity_level", complexityLevel);
                json.put("metaphor_use", metaphorUse);
                json.put("emoji_usage", emojiUsage);
                json.put("emoji_type", emojiType);
                json.put("exclamation_use", exclamationUse);
                json.put("response_speed", responseSpeed);
                json.put("proactiveness", proactiveness);
                json.put("detail_level", detailLevel);
                json.put("humor_level", humorLevel);
                json.put("voice_tone", voiceTone);
                json.put("speech_speed", speechSpeed);
                json.put("volume_level", volumeLevel);
                json.put("pause_frequency", pauseFrequency);
                json.put("underlying_emotion", underlyingEmotion);
                json.put("expression_tendency", expressionTendency);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isExpressive() { return warmthLevel > 0.6 && enthusiasmLevel > 0.5; }
        public boolean isReserved() { return warmthLevel < 0.4 && enthusiasmLevel < 0.4; }

        public String getCommunicationStyle() {
            if (warmthLevel > 0.6 && humorLevel > 0.5) return "warm_friendly";
            if (formalityLevel > 0.7) return "professional";
            if (detailLevel > 0.7) return "detailed";
            if ("immediate".equals(responseSpeed)) return "responsive";
            return "balanced";
        }
    }

    /**
     * 行为指导
     */
    public static class BehaviorGuidance {
        public String decisionStyle;
        public boolean shouldDelay;
        public String delayReason;
        public String expressionStyle;
        public boolean shouldExpress;
        public String expressionChannel;
        public String socialApproach;
        public boolean shouldInteract;
        public String riskWarning;
        public double impulseRisk;
        public String regulationSuggestion;

        public BehaviorGuidance() {}

        public static BehaviorGuidance fromJson(JSONObject json) {
            BehaviorGuidance g = new BehaviorGuidance();
            g.decisionStyle = json.optString("decision_style", "balanced");
            g.shouldDelay = json.optBoolean("should_delay", false);
            g.delayReason = json.optString("delay_reason", "");
            g.expressionStyle = json.optString("expression_style", "balanced");
            g.shouldExpress = json.optBoolean("should_express", true);
            g.expressionChannel = json.optString("expression_channel", "");
            g.socialApproach = json.optString("social_approach", "moderate");
            g.shouldInteract = json.optBoolean("should_interact", true);
            g.riskWarning = json.optString("risk_warning", "");
            g.impulseRisk = json.optDouble("impulse_risk", 0);
            g.regulationSuggestion = json.optString("regulation_suggestion", "");
            return g;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("decision_style", decisionStyle);
                json.put("should_delay", shouldDelay);
                json.put("delay_reason", delayReason);
                json.put("expression_style", expressionStyle);
                json.put("should_express", shouldExpress);
                json.put("expression_channel", expressionChannel);
                json.put("social_approach", socialApproach);
                json.put("should_interact", shouldInteract);
                json.put("risk_warning", riskWarning);
                json.put("impulse_risk", impulseRisk);
                json.put("regulation_suggestion", regulationSuggestion);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }
    }

    /**
     * 行为推荐
     */
    public static class BehaviorRecommendation {
        public String behaviorType;
        public String behaviorName;
        public String reason;
        public double urgency;
        public double appropriateness;

        public BehaviorRecommendation() {}

        public static BehaviorRecommendation fromJson(JSONObject json) {
            BehaviorRecommendation b = new BehaviorRecommendation();
            b.behaviorType = json.optString("behavior_type", "");
            b.behaviorName = json.optString("behavior_name", "");
            b.reason = json.optString("reason", "");
            b.urgency = json.optDouble("urgency", 0.5);
            b.appropriateness = json.optDouble("appropriateness", 0.5);
            return b;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("behavior_type", behaviorType);
                json.put("behavior_name", behaviorName);
                json.put("reason", reason);
                json.put("urgency", urgency);
                json.put("appropriateness", appropriateness);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }
    }

    /**
     * 应对策略推荐
     */
    public static class CopingRecommendation {
        public String strategyId;
        public String strategyName;
        public String reason;
        public double effectiveness;
        public int priority;

        public CopingRecommendation() {}

        public static CopingRecommendation fromJson(JSONObject json) {
            CopingRecommendation c = new CopingRecommendation();
            c.strategyId = json.optString("strategy_id", "");
            c.strategyName = json.optString("strategy_name", "");
            c.reason = json.optString("reason", "");
            c.effectiveness = json.optDouble("effectiveness", 0.5);
            c.priority = json.optInt("priority", 1);
            return c;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("strategy_id", strategyId);
                json.put("strategy_name", strategyName);
                json.put("reason", reason);
                json.put("effectiveness", effectiveness);
                json.put("priority", priority);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }
    }

    /**
     * 情绪状态摘要
     */
    public static class EmotionStateSummary {
        public String dominantEmotion;
        public double intensity;
        public double valence;
        public double arousal;

        public EmotionStateSummary() {
            this.dominantEmotion = "neutral";
            this.intensity = 0.3;
            this.valence = 0;
            this.arousal = 0.3;
        }

        public static EmotionStateSummary fromJson(JSONObject json) {
            EmotionStateSummary e = new EmotionStateSummary();
            e.dominantEmotion = json.optString("dominant_emotion", "neutral");
            e.intensity = json.optDouble("intensity", 0.3);
            e.valence = json.optDouble("valence", 0);
            e.arousal = json.optDouble("arousal", 0.3);
            return e;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("dominant_emotion", dominantEmotion);
                json.put("intensity", intensity);
                json.put("valence", valence);
                json.put("arousal", arousal);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isPositive() { return valence > 0.3; }
        public boolean isNegative() { return valence < -0.3; }
        public boolean isHighArousal() { return arousal > 0.6; }
        public boolean isLowArousal() { return arousal < 0.4; }
    }
}