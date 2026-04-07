package com.ofa.agent.voice;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Voice状态客户端 (v5.1.0)
 *
 * 端侧接收 Center 推送的 Voice 状态，用于语音合成和输出。
 * 深层 Voice 管理在 Center 端 VoiceEngine 完成。
 */
public class VoiceClient {

    private static volatile VoiceClient instance;

    // 当前状态
    private VoiceState currentState;

    // 监听器
    private final CopyOnWriteArrayList<VoiceStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private VoiceClient() {
        this.currentState = new VoiceState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static VoiceClient getInstance() {
        if (instance == null) {
            synchronized (VoiceClient.class) {
                if (instance == null) {
                    instance = new VoiceClient();
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
     * 接收 Center 推送的 Voice 状态
     */
    public void receiveVoiceState(@NonNull JSONObject stateJson) {
        try {
            VoiceState newState = VoiceState.fromJson(stateJson);
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
            VoiceState newState = VoiceState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull VoiceState newState) {
        this.currentState = newState;

        for (VoiceStateListener listener : listeners) {
            listener.onVoiceStateChanged(newState);
        }
    }

    // === 状态获取 ===

    @NonNull
    public VoiceState getCurrentState() {
        return currentState;
    }

    // === 语音特征获取 ===

    @NonNull
    public VoiceState.VoiceCharacteristics getVoiceCharacteristics() {
        return currentState.getVoiceCharacteristics();
    }

    public String getVoiceType() {
        return currentState.getVoiceCharacteristics().voiceType;
    }

    public String getVoiceAge() {
        return currentState.getVoiceCharacteristics().voiceAge;
    }

    public double getBasePitch() {
        return currentState.getVoiceCharacteristics().basePitch;
    }

    public double getSpeakingRate() {
        return currentState.getVoiceCharacteristics().speakingRate;
    }

    public double getBaseVolume() {
        return currentState.getVoiceCharacteristics().baseVolume;
    }

    public String getTimbreType() {
        return currentState.getVoiceCharacteristics().timbreType;
    }

    public boolean isMaleVoice() {
        return currentState.getVoiceCharacteristics().isMaleVoice();
    }

    public boolean isFemaleVoice() {
        return currentState.getVoiceCharacteristics().isFemaleVoice();
    }

    public boolean isYoungVoice() {
        return currentState.getVoiceCharacteristics().isYoungVoice();
    }

    public boolean isFastRate() {
        return currentState.getVoiceCharacteristics().isFastRate();
    }

    // === 语音风格获取 ===

    @NonNull
    public VoiceState.VoiceStyle getVoiceStyle() {
        return currentState.getVoiceStyle();
    }

    public String getRegionalAccent() {
        return currentState.getVoiceStyle().regionalAccent;
    }

    public String getFormalityLevel() {
        return currentState.getVoiceStyle().formalityLevel;
    }

    public String getCommunicationStyle() {
        return currentState.getVoiceStyle().communicationStyle;
    }

    public double getVoiceAuthority() {
        return currentState.getVoiceStyle().voiceAuthority;
    }

    public double getVoiceWarmth() {
        return currentState.getVoiceStyle().voiceWarmth;
    }

    public boolean isFormalStyle() {
        return currentState.getVoiceStyle().isFormalStyle();
    }

    public boolean isDirectCommunication() {
        return currentState.getVoiceStyle().isDirectCommunication();
    }

    // === 情绪语音获取 ===

    @NonNull
    public VoiceState.EmotionalVoice getEmotionalVoice() {
        return currentState.getEmotionalVoice();
    }

    public double getEmotionalExpressiveness() {
        return currentState.getEmotionalVoice().emotionalExpressiveness;
    }

    public String getEmotionRegulation() {
        return currentState.getEmotionalVoice().emotionRegulation;
    }

    public boolean isHighExpressiveness() {
        return currentState.getEmotionalVoice().isHighExpressiveness();
    }

    public boolean isAmplifyMode() {
        return currentState.getEmotionalVoice().isAmplifyMode();
    }

    // === 语言模式获取 ===

    @NonNull
    public VoiceState.SpeechPatterns getSpeechPatterns() {
        return currentState.getSpeechPatterns();
    }

    public String getVocabularyLevel() {
        return currentState.getSpeechPatterns().vocabularyLevel;
    }

    public String getSentenceLength() {
        return currentState.getSpeechPatterns().sentenceLength;
    }

    public String getThoughtOrganization() {
        return currentState.getSpeechPatterns().thoughtOrganization;
    }

    public boolean isSimpleVocabulary() {
        return currentState.getSpeechPatterns().isSimpleVocabulary();
    }

    public boolean isStructuredThinking() {
        return currentState.getSpeechPatterns().isStructuredThinking();
    }

    // === TTS配置获取 ===

    @NonNull
    public VoiceState.TTSConfiguration getTtsConfig() {
        return currentState.getTtsConfig();
    }

    public String getEngineType() {
        return currentState.getTtsConfig().engineType;
    }

    public String getEngineName() {
        return currentState.getTtsConfig().engineName;
    }

    public int getSampleRate() {
        return currentState.getTtsConfig().sampleRate;
    }

    public String getAudioFormat() {
        return currentState.getTtsConfig().audioFormat;
    }

    public String getQualityLevel() {
        return currentState.getTtsConfig().qualityLevel;
    }

    public boolean isHighQuality() {
        return currentState.getTtsConfig().isHighQuality();
    }

    public boolean isRealTimeMode() {
        return currentState.getTtsConfig().isRealTimeMode();
    }

    // === 决策上下文获取 ===

    @NonNull
    public VoiceState.VoiceDecisionContext getDecisionContext() {
        return currentState.getDecisionContext();
    }

    public double getRecommendedPitch() {
        return currentState.getDecisionContext().recommendedPitch;
    }

    public double getRecommendedRate() {
        return currentState.getDecisionContext().recommendedRate;
    }

    public double getRecommendedVolume() {
        return currentState.getDecisionContext().recommendedVolume;
    }

    public String getRecommendedTone() {
        return currentState.getDecisionContext().recommendedTone;
    }

    public String getRecommendedFormality() {
        return currentState.getDecisionContext().recommendedFormality;
    }

    // === 情绪适应获取 ===

    @NonNull
    public VoiceState.VoiceEmotionAdaptation getEmotionAdaptation() {
        return currentState.getDecisionContext().emotionAdaptation;
    }

    public String getCurrentEmotion() {
        return currentState.getDecisionContext().emotionAdaptation.currentEmotion;
    }

    public double getPitchShift() {
        return currentState.getDecisionContext().emotionAdaptation.pitchShift;
    }

    public double getRateMultiplier() {
        return currentState.getDecisionContext().emotionAdaptation.rateMultiplier;
    }

    // === 场景适应获取 ===

    @NonNull
    public VoiceState.VoiceSceneAdaptation getSceneAdaptation() {
        return currentState.getDecisionContext().sceneAdaptation;
    }

    public String getCurrentScene() {
        return currentState.getDecisionContext().sceneAdaptation.scene;
    }

    public String getArticulationMode() {
        return currentState.getDecisionContext().sceneAdaptation.articulationMode;
    }

    // === 社交适应获取 ===

    @NonNull
    public VoiceState.VoiceSocialAdaptation getSocialAdaptation() {
        return currentState.getDecisionContext().socialAdaptation;
    }

    public String getSocialContext() {
        return currentState.getDecisionContext().socialAdaptation.socialContext;
    }

    public double getAuthorityLevel() {
        return currentState.getDecisionContext().socialAdaptation.authorityLevel;
    }

    public double getWarmthLevel() {
        return currentState.getDecisionContext().socialAdaptation.warmthLevel;
    }

    // === 决策辅助 ===

    /**
     * 获取适合当前情绪的音高建议
     */
    @NonNull
    public String getPitchRecommendation() {
        double pitch = getRecommendedPitch();
        String emotion = getCurrentEmotion();

        if (emotion == null || emotion.isEmpty()) {
            return "使用标准音高";
        }

        double shift = getPitchShift();
        if (shift > 2.0) {
            return "提高音调，表现兴奋或紧张";
        } else if (shift < -2.0) {
            return "降低音调，表现沉稳或低落";
        } else {
            return "保持自然音高";
        }
    }

    /**
     * 获取适合当前场景的语速建议
     */
    @NonNull
    public String getRateRecommendation() {
        double rate = getRecommendedRate();
        String scene = getCurrentScene();

        if ("meeting".equals(scene) || "presentation".equals(scene)) {
            return "稍慢语速，清晰表达";
        } else if ("casual".equals(scene)) {
            return "自然语速，轻松交流";
        } else {
            return "标准语速";
        }
    }

    /**
     * 获取适合当前社交语境的风格建议
     */
    @NonNull
    public String getStyleRecommendation() {
        String socialContext = getSocialContext();
        double authority = getAuthorityLevel();
        double warmth = getWarmthLevel();

        if ("professional".equals(socialContext) || "formal".equals(socialContext)) {
            return "专业正式的语音风格";
        } else if ("intimate".equals(socialContext)) {
            return "温暖亲切的语音风格";
        } else if (authority > 0.6) {
            return "自信权威的语音风格";
        } else if (warmth > 0.6) {
            return "友好温暖的语音风格";
        } else {
            return "自然平衡的语音风格";
        }
    }

    /**
     * 获取 TTS 配置建议
     */
    @NonNull
    public String getTTSRecommendation() {
        String quality = getQualityLevel();
        String mode = getEngineType();

        if (isHighQuality()) {
            return "高质量合成模式，适合正式场合";
        } else if (isRealTimeMode()) {
            return "实时合成模式，适合对话场景";
        } else {
            return "平衡模式";
        }
    }

    /**
     * 获取语音综合建议
     */
    @NonNull
    public String getVoiceAdvice() {
        StringBuilder advice = new StringBuilder();

        // 音高建议
        advice.append(getPitchRecommendation()).append("。");

        // 语速建议
        advice.append(getRateRecommendation()).append("。");

        // 风格建议
        advice.append(getStyleRecommendation()).append("。");

        // 表达性建议
        if (isHighExpressiveness()) {
            advice.append("可以展现丰富的语音情感。");
        }

        return advice.toString();
    }

    // === 行为上报 ===

    /**
     * 上报语音变化
     */
    @NonNull
    public JSONObject reportVoiceChange(@NonNull String changeType, @NonNull String newValue) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "voice_change");
            report.put("change_type", changeType);
            report.put("new_value", newValue);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报情绪影响
     */
    @NonNull
    public JSONObject reportEmotionInfluence(@NonNull String emotion, double intensity) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "emotion_influence");
            report.put("emotion", emotion);
            report.put("intensity", intensity);
            report.put("pitch_shift", getPitchShift());
            report.put("rate_multiplier", getRateMultiplier());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报场景变化
     */
    @NonNull
    public JSONObject reportSceneChange(@NonNull String newScene, @NonNull String previousScene) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "scene_change");
            report.put("new_scene", newScene);
            report.put("previous_scene", previousScene);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    public void addListener(@NonNull VoiceStateListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull VoiceStateListener listener) {
        listeners.remove(listener);
    }

    public void clearListeners() {
        listeners.clear();
    }

    /**
     * Voice状态监听器
     */
    public interface VoiceStateListener {
        void onVoiceStateChanged(@NonNull VoiceState state);
    }

    // === 状态快照 ===

    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = VoiceState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    public void reset() {
        currentState = new VoiceState();

        for (VoiceStateListener listener : listeners) {
            listener.onVoiceStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "VoiceClient{" +
                "voiceType='" + getVoiceType() + '\'' +
                ", voiceAge='" + getVoiceAge() + '\'' +
                ", pitch=" + getBasePitch() +
                ", rate=" + getSpeakingRate() +
                ", scene='" + getCurrentScene() + '\'' +
                '}';
    }
}