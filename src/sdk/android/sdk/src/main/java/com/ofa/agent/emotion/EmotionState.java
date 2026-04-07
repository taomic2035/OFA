package com.ofa.agent.emotion;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * 情绪状态模型 (v4.0.0)
 *
 * 端侧只接收 Center 推送的情绪状态，用于表现和行为调整。
 * 深层情绪管理和计算在 Center 端 EmotionEngine 完成。
 */
public class EmotionState {

    // === 七情 (0-1) ===
    private double joy;      // 喜
    private double anger;    // 怒
    private double sadness;  // 哀
    private double fear;     // 惧
    private double love;     // 爱
    private double disgust;  // 恶
    private double desire;   // 欲

    // === 情绪元数据 ===
    private double intensity;     // 情绪强度
    private double valence;       // 情感效价 (-1到1)
    private double arousal;       // 唤醒度

    // === 当前心境 ===
    private String currentMood;   // 当前心境
    private String moodTrend;     // 情绪趋势
    private String lastTrigger;   // 最后触发原因

    // === 时间属性 ===
    private long timestamp;
    private int duration;         // 持续时间(分钟)

    // === 影响因子 (用于调整行为) ===
    private double riskTolerance;       // 风险容忍度
    private double socialTendency;      // 社交倾向
    private double decisionSpeed;       // 决策速度
    private double trustLevel;          // 信任度
    private double creativity;          // 创造性

    public EmotionState() {
        // 默认中性状态
        this.joy = 0.5;
        this.anger = 0.1;
        this.sadness = 0.1;
        this.fear = 0.1;
        this.love = 0.3;
        this.disgust = 0.1;
        this.desire = 0.3;
        this.intensity = 0.3;
        this.valence = 0.0;
        this.arousal = 0.3;
        this.currentMood = "neutral";
        this.moodTrend = "stable";
        this.timestamp = System.currentTimeMillis();
        this.duration = 0;

        // 默认影响因子
        this.riskTolerance = 0.5;
        this.socialTendency = 0.5;
        this.decisionSpeed = 0.5;
        this.trustLevel = 0.5;
        this.creativity = 0.5;
    }

    /**
     * 从 JSON 解析 (Center 推送的状态)
     */
    @NonNull
    public static EmotionState fromJson(@NonNull JSONObject json) throws JSONException {
        EmotionState state = new EmotionState();

        // 七情
        state.joy = json.optDouble("joy", 0.5);
        state.anger = json.optDouble("anger", 0.1);
        state.sadness = json.optDouble("sadness", 0.1);
        state.fear = json.optDouble("fear", 0.1);
        state.love = json.optDouble("love", 0.3);
        state.disgust = json.optDouble("disgust", 0.1);
        state.desire = json.optDouble("desire", 0.3);

        // 元数据
        state.intensity = json.optDouble("intensity", 0.3);
        state.valence = json.optDouble("valence", 0.0);
        state.arousal = json.optDouble("arousal", 0.3);

        // 心境
        state.currentMood = json.optString("current_mood", "neutral");
        state.moodTrend = json.optString("mood_trend", "stable");
        state.lastTrigger = json.optString("last_trigger", "");

        // 时间
        state.timestamp = json.optLong("timestamp", System.currentTimeMillis());
        state.duration = json.optInt("duration", 0);

        // 影响因子
        JSONObject influence = json.optJSONObject("influence_factors");
        if (influence != null) {
            state.riskTolerance = influence.optDouble("risk_tolerance", 0.5);
            state.socialTendency = influence.optDouble("social_tendency", 0.5);
            state.decisionSpeed = influence.optDouble("decision_speed", 0.5);
            state.trustLevel = influence.optDouble("trust_level", 0.5);
            state.creativity = influence.optDouble("creativity", 0.5);
        }

        return state;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            // 七情
            json.put("joy", joy);
            json.put("anger", anger);
            json.put("sadness", sadness);
            json.put("fear", fear);
            json.put("love", love);
            json.put("disgust", disgust);
            json.put("desire", desire);

            // 元数据
            json.put("intensity", intensity);
            json.put("valence", valence);
            json.put("arousal", arousal);

            // 心境
            json.put("current_mood", currentMood);
            json.put("mood_trend", moodTrend);
            json.put("last_trigger", lastTrigger);

            // 时间
            json.put("timestamp", timestamp);
            json.put("duration", duration);

            // 影响因子
            JSONObject influence = new JSONObject();
            influence.put("risk_tolerance", riskTolerance);
            influence.put("social_tendency", socialTendency);
            influence.put("decision_speed", decisionSpeed);
            influence.put("trust_level", trustLevel);
            influence.put("creativity", creativity);
            json.put("influence_factors", influence);

        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 获取主导情绪
     */
    @NonNull
    public String getDominantEmotion() {
        double max = joy;
        String dominant = "joy";

        if (anger > max) { max = anger; dominant = "anger"; }
        if (sadness > max) { max = sadness; dominant = "sadness"; }
        if (fear > max) { max = fear; dominant = "fear"; }
        if (love > max) { max = love; dominant = "love"; }
        if (disgust > max) { max = disgust; dominant = "disgust"; }
        if (desire > max) { dominant = "desire"; }

        return dominant;
    }

    /**
     * 获取情绪描述（用于表现）
     */
    @NonNull
    public String getEmotionDescription() {
        String dominant = getDominantEmotion();
        String level;

        if (intensity > 0.7) {
            level = "强烈";
        } else if (intensity > 0.4) {
            level = "中等";
        } else {
            level = "轻微";
        }

        switch (dominant) {
            case "joy":
                return level + "的愉悦感";
            case "anger":
                return level + "的愤怒";
            case "sadness":
                return level + "的悲伤";
            case "fear":
                return level + "的担忧";
            case "love":
                return level + "的喜爱";
            case "disgust":
                return level + "的不满";
            case "desire":
                return level + "的欲望";
            default:
                return level + "的情绪";
        }
    }

    /**
     * 是否正面情绪为主
     */
    public boolean isPositive() {
        return valence > 0.2;
    }

    /**
     * 是否负面情绪为主
     */
    public boolean isNegative() {
        return valence < -0.2;
    }

    /**
     * 是否高唤醒状态（激动、紧张）
     */
    public boolean isHighArousal() {
        return arousal > 0.6;
    }

    /**
     * 是否低唤醒状态（平静、低落）
     */
    public boolean isLowArousal() {
        return arousal < 0.3;
    }

    /**
     * 是否需要关注（高强度负面情绪）
     */
    public boolean needsAttention() {
        return intensity > 0.6 && isNegative();
    }

    /**
     * 是否适合社交互动
     */
    public boolean isSuitableForSocial() {
        return socialTendency > 0.4 && !isNegative();
    }

    /**
     * 是否适合做决策
     */
    public boolean isSuitableForDecision() {
        // 高强度情绪不适合做重要决策
        return intensity < 0.5;
    }

    /**
     * 获取推荐回复风格
     */
    @NonNull
    public String getRecommendedResponseStyle() {
        if (isPositive() && isHighArousal()) {
            return "energetic";  // 热情活跃
        } else if (isPositive() && isLowArousal()) {
            return "calm";       // 平静温和
        } else if (isNegative() && isHighArousal()) {
            return "supportive"; // 支持安慰
        } else if (isNegative() && isLowArousal()) {
            return "gentle";     // 温柔体贴
        } else {
            return "neutral";    // 中性
        }
    }

    // === Getter/Setter ===

    public double getJoy() { return joy; }
    public void setJoy(double joy) { this.joy = clamp01(joy); }

    public double getAnger() { return anger; }
    public void setAnger(double anger) { this.anger = clamp01(anger); }

    public double getSadness() { return sadness; }
    public void setSadness(double sadness) { this.sadness = clamp01(sadness); }

    public double getFear() { return fear; }
    public void setFear(double fear) { this.fear = clamp01(fear); }

    public double getLove() { return love; }
    public void setLove(double love) { this.love = clamp01(love); }

    public double getDisgust() { return disgust; }
    public void setDisgust(double disgust) { this.disgust = clamp01(disgust); }

    public double getDesire() { return desire; }
    public void setDesire(double desire) { this.desire = clamp01(desire); }

    public double getIntensity() { return intensity; }
    public void setIntensity(double intensity) { this.intensity = clamp01(intensity); }

    public double getValence() { return valence; }
    public void setValence(double valence) { this.valence = clamp(valence, -1, 1); }

    public double getArousal() { return arousal; }
    public void setArousal(double arousal) { this.arousal = clamp01(arousal); }

    public String getCurrentMood() { return currentMood; }
    public void setCurrentMood(String currentMood) { this.currentMood = currentMood; }

    public String getMoodTrend() { return moodTrend; }
    public void setMoodTrend(String moodTrend) { this.moodTrend = moodTrend; }

    public String getLastTrigger() { return lastTrigger; }
    public void setLastTrigger(String lastTrigger) { this.lastTrigger = lastTrigger; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    public int getDuration() { return duration; }
    public void setDuration(int duration) { this.duration = duration; }

    public double getRiskTolerance() { return riskTolerance; }
    public void setRiskTolerance(double riskTolerance) { this.riskTolerance = clamp01(riskTolerance); }

    public double getSocialTendency() { return socialTendency; }
    public void setSocialTendency(double socialTendency) { this.socialTendency = clamp01(socialTendency); }

    public double getDecisionSpeed() { return decisionSpeed; }
    public void setDecisionSpeed(double decisionSpeed) { this.decisionSpeed = clamp01(decisionSpeed); }

    public double getTrustLevel() { return trustLevel; }
    public void setTrustLevel(double trustLevel) { this.trustLevel = clamp01(trustLevel); }

    public double getCreativity() { return creativity; }
    public void setCreativity(double creativity) { this.creativity = clamp01(creativity); }

    @NonNull
    @Override
    public String toString() {
        return "EmotionState{" +
                "mood='" + currentMood + '\'' +
                ", dominant='" + getDominantEmotion() + '\'' +
                ", intensity=" + intensity +
                ", valence=" + valence +
                '}';
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        return clamp(value, 0, 1);
    }

    private double clamp(double value, double min, double max) {
        if (value < min) return min;
        if (value > max) return max;
        return value;
    }
}