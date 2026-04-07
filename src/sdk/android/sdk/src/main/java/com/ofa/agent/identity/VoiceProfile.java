package com.ofa.agent.identity;

import androidx.annotation.NonNull;

/**
 * VoiceProfile - 语音音色配置
 *
 * 描述用户的语音特征，用于语音合成个性化。
 */
public class VoiceProfile {

    private String id;
    private String voiceType;             // clone/synthetic/preset
    private String presetVoiceId;         // 预设音色ID
    private String cloneReferenceId;      // 克隆参考ID

    // 音色参数
    private double pitch;      // 音高 (0-2, 1.0 为正常)
    private double speed;      // 语速 (0-2, 1.0 为正常)
    private double volume;     // 音量 (0-1)

    // 风格参数
    private String tone;                   // warm/neutral/energetic
    private String accent;                 // 口音
    private double emotionLevel;           // 情感表达程度 (0-1)

    // 语调模式
    private String pausePattern;           // 停顿模式
    private String emphasisStyle;          // 重音风格

    private long createdAt;
    private long updatedAt;

    // 音色类型常量
    public static final String VOICE_TYPE_CLONE = "clone";
    public static final String VOICE_TYPE_SYNTHETIC = "synthetic";
    public static final String VOICE_TYPE_PRESET = "preset";

    // 语调常量
    public static final String TONE_WARM = "warm";
    public static final String TONE_NEUTRAL = "neutral";
    public static final String TONE_ENERGETIC = "energetic";

    /**
     * 创建默认语音配置
     */
    public VoiceProfile() {
        this.id = generateId();
        this.voiceType = VOICE_TYPE_PRESET;
        this.pitch = 1.0;
        this.speed = 1.0;
        this.volume = 0.8;
        this.tone = TONE_WARM;
        this.emotionLevel = 0.6;
        this.pausePattern = "natural";
        this.emphasisStyle = "moderate";
        this.createdAt = System.currentTimeMillis();
        this.updatedAt = System.currentTimeMillis();
    }

    // === 更新方法 ===

    /**
     * 设置音色参数
     */
    public void setVoiceParams(double pitch, double speed, double volume) {
        this.pitch = clamp(pitch, 0, 2);
        this.speed = clamp(speed, 0, 2);
        this.volume = clamp01(volume);
        this.updatedAt = System.currentTimeMillis();
    }

    /**
     * 设置风格参数
     */
    public void setStyleParams(@NonNull String tone, double emotionLevel) {
        this.tone = tone;
        this.emotionLevel = clamp01(emotionLevel);
        this.updatedAt = System.currentTimeMillis();
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

    private String generateId() {
        return "voice_" + System.currentTimeMillis();
    }

    // === Getters ===

    public String getId() { return id; }
    public String getVoiceType() { return voiceType; }
    public String getPresetVoiceId() { return presetVoiceId; }
    public String getCloneReferenceId() { return cloneReferenceId; }

    public double getPitch() { return pitch; }
    public double getSpeed() { return speed; }
    public double getVolume() { return volume; }

    public String getTone() { return tone; }
    public String getAccent() { return accent; }
    public double getEmotionLevel() { return emotionLevel; }

    public String getPausePattern() { return pausePattern; }
    public String getEmphasisStyle() { return emphasisStyle; }

    public long getCreatedAt() { return createdAt; }
    public long getUpdatedAt() { return updatedAt; }

    // === Setters ===

    public void setVoiceType(String voiceType) { this.voiceType = voiceType; }
    public void setPresetVoiceId(String presetVoiceId) { this.presetVoiceId = presetVoiceId; }
    public void setCloneReferenceId(String cloneReferenceId) { this.cloneReferenceId = cloneReferenceId; }
    public void setAccent(String accent) { this.accent = accent; }
    public void setPausePattern(String pausePattern) { this.pausePattern = pausePattern; }
    public void setEmphasisStyle(String emphasisStyle) { this.emphasisStyle = emphasisStyle; }

    /**
     * 转换为 JSON 字符串
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"id\":\"").append(id).append("\",");
        sb.append("\"voice_type\":\"").append(voiceType).append("\",");
        sb.append("\"pitch\":").append(pitch).append(",");
        sb.append("\"speed\":").append(speed).append(",");
        sb.append("\"volume\":").append(volume).append(",");
        sb.append("\"tone\":\"").append(tone).append("\",");
        sb.append("\"emotion_level\":").append(emotionLevel);
        sb.append("}");
        return sb.toString();
    }
}