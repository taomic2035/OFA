package com.ofa.agent.speech;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * SpeechContent状态模型 (v5.2.0)
 *
 * 端侧接收 Center 推送的表达内容状态，用于内容生成和表达。
 * 深层内容管理在 Center 端 SpeechContentEngine 完成。
 */
public class SpeechContentState {

    // === 内容风格 ===
    private ContentStyle contentStyle;

    // === 表达深度 ===
    private ExpressionDepth expressionDepth;

    // === 文化表达 ===
    private CulturalExpression culturalExpression;

    // === 社交表达 ===
    private SocialExpression socialExpression;

    // === 决策上下文 ===
    private ContentDecisionContext decisionContext;

    // === 版本信息 ===
    private long version;
    private long timestamp;

    public SpeechContentState() {
        this.contentStyle = new ContentStyle();
        this.expressionDepth = new ExpressionDepth();
        this.culturalExpression = new CulturalExpression();
        this.socialExpression = new SocialExpression();
        this.decisionContext = new ContentDecisionContext();
    }

    // === Getters ===

    @NonNull
    public ContentStyle getContentStyle() {
        return contentStyle;
    }

    @NonNull
    public ExpressionDepth getExpressionDepth() {
        return expressionDepth;
    }

    @NonNull
    public CulturalExpression getCulturalExpression() {
        return culturalExpression;
    }

    @NonNull
    public SocialExpression getSocialExpression() {
        return socialExpression;
    }

    @NonNull
    public ContentDecisionContext getDecisionContext() {
        return decisionContext;
    }

    public long getVersion() {
        return version;
    }

    public long getTimestamp() {
        return timestamp;
    }

    // === Setters ===

    public void setContentStyle(@NonNull ContentStyle contentStyle) {
        this.contentStyle = contentStyle;
    }

    public void setExpressionDepth(@NonNull ExpressionDepth expressionDepth) {
        this.expressionDepth = expressionDepth;
    }

    public void setCulturalExpression(@NonNull CulturalExpression culturalExpression) {
        this.culturalExpression = culturalExpression;
    }

    public void setSocialExpression(@NonNull SocialExpression socialExpression) {
        this.socialExpression = socialExpression;
    }

    public void setDecisionContext(@NonNull ContentDecisionContext decisionContext) {
        this.decisionContext = decisionContext;
    }

    public void setVersion(long version) {
        this.version = version;
    }

    public void setTimestamp(long timestamp) {
        this.timestamp = timestamp;
    }

    // === JSON 序列化 ===

    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("content_style", contentStyle.toJson());
            json.put("expression_depth", expressionDepth.toJson());
            json.put("cultural_expression", culturalExpression.toJson());
            json.put("social_expression", socialExpression.toJson());
            json.put("decision_context", decisionContext.toJson());
            json.put("version", version);
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    @NonNull
    public static SpeechContentState fromJson(@NonNull JSONObject json) throws JSONException {
        SpeechContentState state = new SpeechContentState();

        if (json.has("content_style")) {
            state.contentStyle = ContentStyle.fromJson(json.getJSONObject("content_style"));
        }
        if (json.has("expression_depth")) {
            state.expressionDepth = ExpressionDepth.fromJson(json.getJSONObject("expression_depth"));
        }
        if (json.has("cultural_expression")) {
            state.culturalExpression = CulturalExpression.fromJson(json.getJSONObject("cultural_expression"));
        }
        if (json.has("social_expression")) {
            state.socialExpression = SocialExpression.fromJson(json.getJSONObject("social_expression"));
        }
        if (json.has("decision_context")) {
            state.decisionContext = ContentDecisionContext.fromJson(json.getJSONObject("decision_context"));
        }
        if (json.has("version")) {
            state.version = json.getLong("version");
        }
        if (json.has("timestamp")) {
            state.timestamp = json.getLong("timestamp");
        }

        return state;
    }

    // === 内部模型类 ===

    /**
     * 内容风格
     */
    public static class ContentStyle {
        public String toneStyle;           // formal, casual, professional, friendly
        public double toneConsistency;     // 0-1

        public String languageLevel;       // simple, moderate, sophisticated
        public String technicalLevel;      // layman, intermediate, expert

        public double directness;          // 0-1, 直接程度
        public double euphemismUsage;      // 0-1, 委婉语使用
        public double metaphorUsage;       // 0-1, 隐喻使用
        public double humorTendency;       // 0-1, 幽默倾向

        public String emotionalColoring;   // neutral, warm, cool, passionate
        public double enthusiasmLevel;     // 0-1

        public String persuasionStyle;     // logical, emotional, balanced
        public String evidenceType;        // data, anecdote, expert, mixed

        public ContentStyle() {
            this.toneStyle = "neutral";
            this.toneConsistency = 0.7;
            this.languageLevel = "moderate";
            this.technicalLevel = "intermediate";
            this.directness = 0.5;
            this.euphemismUsage = 0.3;
            this.metaphorUsage = 0.3;
            this.humorTendency = 0.3;
            this.emotionalColoring = "neutral";
            this.enthusiasmLevel = 0.5;
            this.persuasionStyle = "balanced";
            this.evidenceType = "mixed";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("tone_style", toneStyle);
                json.put("tone_consistency", toneConsistency);
                json.put("language_level", languageLevel);
                json.put("technical_level", technicalLevel);
                json.put("directness", directness);
                json.put("euphemism_usage", euphemismUsage);
                json.put("metaphor_usage", metaphorUsage);
                json.put("humor_tendency", humorTendency);
                json.put("emotional_coloring", emotionalColoring);
                json.put("enthusiasm_level", enthusiasmLevel);
                json.put("persuasion_style", persuasionStyle);
                json.put("evidence_type", evidenceType);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ContentStyle fromJson(@NonNull JSONObject json) throws JSONException {
            ContentStyle style = new ContentStyle();
            if (json.has("tone_style")) style.toneStyle = json.getString("tone_style");
            if (json.has("tone_consistency")) style.toneConsistency = json.getDouble("tone_consistency");
            if (json.has("language_level")) style.languageLevel = json.getString("language_level");
            if (json.has("technical_level")) style.technicalLevel = json.getString("technical_level");
            if (json.has("directness")) style.directness = json.getDouble("directness");
            if (json.has("euphemism_usage")) style.euphemismUsage = json.getDouble("euphemism_usage");
            if (json.has("metaphor_usage")) style.metaphorUsage = json.getDouble("metaphor_usage");
            if (json.has("humor_tendency")) style.humorTendency = json.getDouble("humor_tendency");
            if (json.has("emotional_coloring")) style.emotionalColoring = json.getString("emotional_coloring");
            if (json.has("enthusiasm_level")) style.enthusiasmLevel = json.getDouble("enthusiasm_level");
            if (json.has("persuasion_style")) style.persuasionStyle = json.getString("persuasion_style");
            if (json.has("evidence_type")) style.evidenceType = json.getString("evidence_type");
            return style;
        }

        public boolean isFormalTone() {
            return "formal".equals(toneStyle) || "professional".equals(toneStyle);
        }

        public boolean isHighDirectness() {
            return directness > 0.6;
        }

        public boolean isHumorous() {
            return humorTendency > 0.5;
        }
    }

    /**
     * 表达深度
     */
    public static class ExpressionDepth {
        public String thinkingDepth;       // surface, moderate, deep, philosophical
        public String abstractionLevel;    // concrete, mixed, abstract
        public String complexityLevel;     // simple, moderate, complex

        public double selfDisclosureLevel; // 0-1, 自我暴露程度
        public double intimacyThreshold;   // 0-1, 亲密阈值
        public double vulnerabilityLevel;  // 0-1, 脆弱性展示

        public double reflectionTendency;  // 0-1, 反思倾向
        public double selfAwarenessLevel;  // 0-1, 自我意识水平

        public String professionalDepth;   // task_focused, analytical, strategic
        public String personalDepth;       // surface, moderate, deep

        public ExpressionDepth() {
            this.thinkingDepth = "moderate";
            this.abstractionLevel = "mixed";
            this.complexityLevel = "moderate";
            this.selfDisclosureLevel = 0.5;
            this.intimacyThreshold = 0.5;
            this.vulnerabilityLevel = 0.3;
            this.reflectionTendency = 0.5;
            this.selfAwarenessLevel = 0.5;
            this.professionalDepth = "analytical";
            this.personalDepth = "moderate";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("thinking_depth", thinkingDepth);
                json.put("abstraction_level", abstractionLevel);
                json.put("complexity_level", complexityLevel);
                json.put("self_disclosure_level", selfDisclosureLevel);
                json.put("intimacy_threshold", intimacyThreshold);
                json.put("vulnerability_level", vulnerabilityLevel);
                json.put("reflection_tendency", reflectionTendency);
                json.put("self_awareness_level", selfAwarenessLevel);
                json.put("professional_depth", professionalDepth);
                json.put("personal_depth", personalDepth);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ExpressionDepth fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionDepth depth = new ExpressionDepth();
            if (json.has("thinking_depth")) depth.thinkingDepth = json.getString("thinking_depth");
            if (json.has("abstraction_level")) depth.abstractionLevel = json.getString("abstraction_level");
            if (json.has("complexity_level")) depth.complexityLevel = json.getString("complexity_level");
            if (json.has("self_disclosure_level")) depth.selfDisclosureLevel = json.getDouble("self_disclosure_level");
            if (json.has("intimacy_threshold")) depth.intimacyThreshold = json.getDouble("intimacy_threshold");
            if (json.has("vulnerability_level")) depth.vulnerabilityLevel = json.getDouble("vulnerability_level");
            if (json.has("reflection_tendency")) depth.reflectionTendency = json.getDouble("reflection_tendency");
            if (json.has("self_awareness_level")) depth.selfAwarenessLevel = json.getDouble("self_awareness_level");
            if (json.has("professional_depth")) depth.professionalDepth = json.getString("professional_depth");
            if (json.has("personal_depth")) depth.personalDepth = json.getString("personal_depth");
            return depth;
        }

        public boolean isDeepThinking() {
            return "deep".equals(thinkingDepth) || "philosophical".equals(thinkingDepth);
        }

        public boolean isHighSelfDisclosure() {
            return selfDisclosureLevel > 0.6;
        }
    }

    /**
     * 文化表达
     */
    public static class CulturalExpression {
        public boolean highContextCommunication;
        public double indirectExpression;    // 0-1
        public double faceSaving;            // 0-1

        public double respectLevel;          // 0-1
        public double hierarchyAwareness;    // 0-1
        public String honorificUsage;        // none, light, moderate, heavy

        public double tabooAwareness;        // 0-1
        public List<String> sensitiveTopics;

        public double collectivistExpression; // 0-1
        public double groupReferenceUsage;    // 0-1

        public CulturalExpression() {
            this.highContextCommunication = false;
            this.indirectExpression = 0.3;
            this.faceSaving = 0.4;
            this.respectLevel = 0.5;
            this.hierarchyAwareness = 0.4;
            this.honorificUsage = "light";
            this.tabooAwareness = 0.5;
            this.sensitiveTopics = new ArrayList<>();
            this.collectivistExpression = 0.4;
            this.groupReferenceUsage = 0.4;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("high_context_communication", highContextCommunication);
                json.put("indirect_expression", indirectExpression);
                json.put("face_saving", faceSaving);
                json.put("respect_level", respectLevel);
                json.put("hierarchy_awareness", hierarchyAwareness);
                json.put("honorific_usage", honorificUsage);
                json.put("taboo_awareness", tabooAwareness);
                json.put("collectivist_expression", collectivistExpression);
                json.put("group_reference_usage", groupReferenceUsage);

                JSONArray topics = new JSONArray();
                for (String t : sensitiveTopics) topics.put(t);
                json.put("sensitive_topics", topics);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static CulturalExpression fromJson(@NonNull JSONObject json) throws JSONException {
            CulturalExpression expr = new CulturalExpression();
            if (json.has("high_context_communication")) expr.highContextCommunication = json.getBoolean("high_context_communication");
            if (json.has("indirect_expression")) expr.indirectExpression = json.getDouble("indirect_expression");
            if (json.has("face_saving")) expr.faceSaving = json.getDouble("face_saving");
            if (json.has("respect_level")) expr.respectLevel = json.getDouble("respect_level");
            if (json.has("hierarchy_awareness")) expr.hierarchyAwareness = json.getDouble("hierarchy_awareness");
            if (json.has("honorific_usage")) expr.honorificUsage = json.getString("honorific_usage");
            if (json.has("taboo_awareness")) expr.tabooAwareness = json.getDouble("taboo_awareness");
            if (json.has("collectivist_expression")) expr.collectivistExpression = json.getDouble("collectivist_expression");
            if (json.has("group_reference_usage")) expr.groupReferenceUsage = json.getDouble("group_reference_usage");

            if (json.has("sensitive_topics")) {
                JSONArray arr = json.getJSONArray("sensitive_topics");
                for (int i = 0; i < arr.length(); i++) expr.sensitiveTopics.add(arr.getString(i));
            }
            return expr;
        }

        public boolean isHighContext() {
            return highContextCommunication || indirectExpression > 0.6;
        }

        public boolean isHighRespect() {
            return respectLevel > 0.6;
        }
    }

    /**
     * 社交表达
     */
    public static class SocialExpression {
        public String professionalTone;     // authoritative, collaborative, supportive
        public double expertiseDisplay;     // 0-1
        public double humilityExpression;   // 0-1

        public String classExpression;      // understated, moderate, aspirational
        public String networkingStyle;      // reserved, balanced, proactive

        public double roleConsistency;      // 0-1
        public String authorityExpression;  // formal, earned, collaborative

        public double identityConfidence;   // 0-1
        public double authenticExpression;  // 0-1

        public SocialExpression() {
            this.professionalTone = "collaborative";
            this.expertiseDisplay = 0.5;
            this.humilityExpression = 0.5;
            this.classExpression = "moderate";
            this.networkingStyle = "balanced";
            this.roleConsistency = 0.7;
            this.authorityExpression = "earned";
            this.identityConfidence = 0.6;
            this.authenticExpression = 0.7;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("professional_tone", professionalTone);
                json.put("expertise_display", expertiseDisplay);
                json.put("humility_expression", humilityExpression);
                json.put("class_expression", classExpression);
                json.put("networking_style", networkingStyle);
                json.put("role_consistency", roleConsistency);
                json.put("authority_expression", authorityExpression);
                json.put("identity_confidence", identityConfidence);
                json.put("authentic_expression", authenticExpression);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static SocialExpression fromJson(@NonNull JSONObject json) throws JSONException {
            SocialExpression expr = new SocialExpression();
            if (json.has("professional_tone")) expr.professionalTone = json.getString("professional_tone");
            if (json.has("expertise_display")) expr.expertiseDisplay = json.getDouble("expertise_display");
            if (json.has("humility_expression")) expr.humilityExpression = json.getDouble("humility_expression");
            if (json.has("class_expression")) expr.classExpression = json.getString("class_expression");
            if (json.has("networking_style")) expr.networkingStyle = json.getString("networking_style");
            if (json.has("role_consistency")) expr.roleConsistency = json.getDouble("role_consistency");
            if (json.has("authority_expression")) expr.authorityExpression = json.getString("authority_expression");
            if (json.has("identity_confidence")) expr.identityConfidence = json.getDouble("identity_confidence");
            if (json.has("authentic_expression")) expr.authenticExpression = json.getDouble("authentic_expression");
            return expr;
        }

        public boolean isAuthoritative() {
            return "authoritative".equals(professionalTone);
        }

        public boolean isHighConfidence() {
            return identityConfidence > 0.6;
        }
    }

    /**
     * 内容决策上下文
     */
    public static class ContentDecisionContext {
        public String recommendedTone;
        public String recommendedFormality;
        public String recommendedDepth;
        public String recommendedLength;
        public double recommendedDirectness;

        public ContentSceneAdaptation sceneAdaptation;
        public ContentEmotionAdaptation emotionAdaptation;
        public ContentSocialAdaptation socialAdaptation;
        public ContentCulturalAdaptation culturalAdaptation;

        public String openingSuggestion;
        public String closingSuggestion;
        public List<String> keyTopicsToAvoid;
        public List<String> keyTopicsToInclude;

        public ContentDecisionContext() {
            this.recommendedTone = "neutral";
            this.recommendedFormality = "neutral";
            this.recommendedDepth = "moderate";
            this.recommendedLength = "medium";
            this.recommendedDirectness = 0.5;
            this.sceneAdaptation = new ContentSceneAdaptation();
            this.emotionAdaptation = new ContentEmotionAdaptation();
            this.socialAdaptation = new ContentSocialAdaptation();
            this.culturalAdaptation = new ContentCulturalAdaptation();
            this.keyTopicsToAvoid = new ArrayList<>();
            this.keyTopicsToInclude = new ArrayList<>();
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("recommended_tone", recommendedTone);
                json.put("recommended_formality", recommendedFormality);
                json.put("recommended_depth", recommendedDepth);
                json.put("recommended_length", recommendedLength);
                json.put("recommended_directness", recommendedDirectness);
                json.put("scene_adaptation", sceneAdaptation.toJson());
                json.put("emotion_adaptation", emotionAdaptation.toJson());
                json.put("social_adaptation", socialAdaptation.toJson());
                json.put("cultural_adaptation", culturalAdaptation.toJson());
                json.put("opening_suggestion", openingSuggestion);
                json.put("closing_suggestion", closingSuggestion);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ContentDecisionContext fromJson(@NonNull JSONObject json) throws JSONException {
            ContentDecisionContext ctx = new ContentDecisionContext();
            if (json.has("recommended_tone")) ctx.recommendedTone = json.getString("recommended_tone");
            if (json.has("recommended_formality")) ctx.recommendedFormality = json.getString("recommended_formality");
            if (json.has("recommended_depth")) ctx.recommendedDepth = json.getString("recommended_depth");
            if (json.has("recommended_length")) ctx.recommendedLength = json.getString("recommended_length");
            if (json.has("recommended_directness")) ctx.recommendedDirectness = json.getDouble("recommended_directness");
            if (json.has("scene_adaptation")) ctx.sceneAdaptation = ContentSceneAdaptation.fromJson(json.getJSONObject("scene_adaptation"));
            if (json.has("emotion_adaptation")) ctx.emotionAdaptation = ContentEmotionAdaptation.fromJson(json.getJSONObject("emotion_adaptation"));
            if (json.has("social_adaptation")) ctx.socialAdaptation = ContentSocialAdaptation.fromJson(json.getJSONObject("social_adaptation"));
            if (json.has("cultural_adaptation")) ctx.culturalAdaptation = ContentCulturalAdaptation.fromJson(json.getJSONObject("cultural_adaptation"));
            if (json.has("opening_suggestion")) ctx.openingSuggestion = json.getString("opening_suggestion");
            if (json.has("closing_suggestion")) ctx.closingSuggestion = json.getString("closing_suggestion");
            return ctx;
        }
    }

    /**
     * 场景适应
     */
    public static class ContentSceneAdaptation {
        public String scene;
        public String toneAdjust;
        public String formalityAdjust;
        public String depthAdjust;
        public String lengthPreference;

        public ContentSceneAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("scene", scene);
                json.put("tone_adjust", toneAdjust);
                json.put("formality_adjust", formalityAdjust);
                json.put("depth_adjust", depthAdjust);
                json.put("length_preference", lengthPreference);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ContentSceneAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ContentSceneAdaptation ca = new ContentSceneAdaptation();
            if (json.has("scene")) ca.scene = json.getString("scene");
            if (json.has("tone_adjust")) ca.toneAdjust = json.getString("tone_adjust");
            if (json.has("formality_adjust")) ca.formalityAdjust = json.getString("formality_adjust");
            if (json.has("depth_adjust")) ca.depthAdjust = json.getString("depth_adjust");
            if (json.has("length_preference")) ca.lengthPreference = json.getString("length_preference");
            return ca;
        }
    }

    /**
     * 情绪适应
     */
    public static class ContentEmotionAdaptation {
        public String currentEmotion;
        public String emotionalColoring;
        public double expressionIntensity;
        public String wordChoice;
        public String sentenceStyle;

        public ContentEmotionAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("current_emotion", currentEmotion);
                json.put("emotional_coloring", emotionalColoring);
                json.put("expression_intensity", expressionIntensity);
                json.put("word_choice", wordChoice);
                json.put("sentence_style", sentenceStyle);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ContentEmotionAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ContentEmotionAdaptation ca = new ContentEmotionAdaptation();
            if (json.has("current_emotion")) ca.currentEmotion = json.getString("current_emotion");
            if (json.has("emotional_coloring")) ca.emotionalColoring = json.getString("emotional_coloring");
            if (json.has("expression_intensity")) ca.expressionIntensity = json.getDouble("expression_intensity");
            if (json.has("word_choice")) ca.wordChoice = json.getString("word_choice");
            if (json.has("sentence_style")) ca.sentenceStyle = json.getString("sentence_style");
            return ca;
        }
    }

    /**
     * 社交适应
     */
    public static class ContentSocialAdaptation {
        public String socialContext;
        public double respectLevel;
        public String honorificUsage;
        public String selfReferenceStyle;
        public String otherReferenceStyle;

        public ContentSocialAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("social_context", socialContext);
                json.put("respect_level", respectLevel);
                json.put("honorific_usage", honorificUsage);
                json.put("self_reference_style", selfReferenceStyle);
                json.put("other_reference_style", otherReferenceStyle);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ContentSocialAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ContentSocialAdaptation ca = new ContentSocialAdaptation();
            if (json.has("social_context")) ca.socialContext = json.getString("social_context");
            if (json.has("respect_level")) ca.respectLevel = json.getDouble("respect_level");
            if (json.has("honorific_usage")) ca.honorificUsage = json.getString("honorific_usage");
            if (json.has("self_reference_style")) ca.selfReferenceStyle = json.getString("self_reference_style");
            if (json.has("other_reference_style")) ca.otherReferenceStyle = json.getString("other_reference_style");
            return ca;
        }
    }

    /**
     * 文化适应
     */
    public static class ContentCulturalAdaptation {
        public String culturalContext;
        public double indirectnessLevel;
        public double faceSavingLevel;
        public double collectivistEmphasis;

        public ContentCulturalAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("cultural_context", culturalContext);
                json.put("indirectness_level", indirectnessLevel);
                json.put("face_saving_level", faceSavingLevel);
                json.put("collectivist_emphasis", collectivistEmphasis);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static ContentCulturalAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ContentCulturalAdaptation ca = new ContentCulturalAdaptation();
            if (json.has("cultural_context")) ca.culturalContext = json.getString("cultural_context");
            if (json.has("indirectness_level")) ca.indirectnessLevel = json.getDouble("indirectness_level");
            if (json.has("face_saving_level")) ca.faceSavingLevel = json.getDouble("face_saving_level");
            if (json.has("collectivist_emphasis")) ca.collectivistEmphasis = json.getDouble("collectivist_emphasis");
            return ca;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "SpeechContentState{" +
                "tone='" + contentStyle.toneStyle + '\'' +
                ", directness=" + contentStyle.directness +
                ", depth='" + expressionDepth.thinkingDepth + '\'' +
                '}';
    }
}