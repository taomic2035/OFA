package com.ofa.agent.voice;

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
 * Voice状态模型 (v5.1.0)
 *
 * 端侧接收 Center 推送的 Voice 状态，用于语音合成和输出。
 * 深层 Voice 管理在 Center 端 VoiceEngine 完成。
 */
public class VoiceState {

    // === 语音特征 ===
    private VoiceCharacteristics voiceCharacteristics;

    // === 语音风格 ===
    private VoiceStyle voiceStyle;

    // === 情绪语音 ===
    private EmotionalVoice emotionalVoice;

    // === 语言模式 ===
    private SpeechPatterns speechPatterns;

    // === TTS配置 ===
    private TTSConfiguration ttsConfig;

    // === 决策上下文 ===
    private VoiceDecisionContext decisionContext;

    // === 版本信息 ===
    private long version;
    private long timestamp;

    // === 构造函数 ===

    public VoiceState() {
        this.voiceCharacteristics = new VoiceCharacteristics();
        this.voiceStyle = new VoiceStyle();
        this.emotionalVoice = new EmotionalVoice();
        this.speechPatterns = new SpeechPatterns();
        this.ttsConfig = new TTSConfiguration();
        this.decisionContext = new VoiceDecisionContext();
    }

    // === Getters ===

    @NonNull
    public VoiceCharacteristics getVoiceCharacteristics() {
        return voiceCharacteristics;
    }

    @NonNull
    public VoiceStyle getVoiceStyle() {
        return voiceStyle;
    }

    @NonNull
    public EmotionalVoice getEmotionalVoice() {
        return emotionalVoice;
    }

    @NonNull
    public SpeechPatterns getSpeechPatterns() {
        return speechPatterns;
    }

    @NonNull
    public TTSConfiguration getTtsConfig() {
        return ttsConfig;
    }

    @NonNull
    public VoiceDecisionContext getDecisionContext() {
        return decisionContext;
    }

    public long getVersion() {
        return version;
    }

    public long getTimestamp() {
        return timestamp;
    }

    // === Setters ===

    public void setVoiceCharacteristics(@NonNull VoiceCharacteristics voiceCharacteristics) {
        this.voiceCharacteristics = voiceCharacteristics;
    }

    public void setVoiceStyle(@NonNull VoiceStyle voiceStyle) {
        this.voiceStyle = voiceStyle;
    }

    public void setEmotionalVoice(@NonNull EmotionalVoice emotionalVoice) {
        this.emotionalVoice = emotionalVoice;
    }

    public void setSpeechPatterns(@NonNull SpeechPatterns speechPatterns) {
        this.speechPatterns = speechPatterns;
    }

    public void setTtsConfig(@NonNull TTSConfiguration ttsConfig) {
        this.ttsConfig = ttsConfig;
    }

    public void setDecisionContext(@NonNull VoiceDecisionContext decisionContext) {
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
            json.put("voice_characteristics", voiceCharacteristics.toJson());
            json.put("voice_style", voiceStyle.toJson());
            json.put("emotional_voice", emotionalVoice.toJson());
            json.put("speech_patterns", speechPatterns.toJson());
            json.put("tts_config", ttsConfig.toJson());
            json.put("decision_context", decisionContext.toJson());
            json.put("version", version);
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    @NonNull
    public static VoiceState fromJson(@NonNull JSONObject json) throws JSONException {
        VoiceState state = new VoiceState();

        if (json.has("voice_characteristics")) {
            state.voiceCharacteristics = VoiceCharacteristics.fromJson(json.getJSONObject("voice_characteristics"));
        }
        if (json.has("voice_style")) {
            state.voiceStyle = VoiceStyle.fromJson(json.getJSONObject("voice_style"));
        }
        if (json.has("emotional_voice")) {
            state.emotionalVoice = EmotionalVoice.fromJson(json.getJSONObject("emotional_voice"));
        }
        if (json.has("speech_patterns")) {
            state.speechPatterns = SpeechPatterns.fromJson(json.getJSONObject("speech_patterns"));
        }
        if (json.has("tts_config")) {
            state.ttsConfig = TTSConfiguration.fromJson(json.getJSONObject("tts_config"));
        }
        if (json.has("decision_context")) {
            state.decisionContext = VoiceDecisionContext.fromJson(json.getJSONObject("decision_context"));
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
     * 语音特征
     */
    public static class VoiceCharacteristics {
        public String voiceType;        // male, female, neutral
        public String voiceAge;         // child, young, adult, senior
        public String voiceQuality;     // clear, breathy, husky

        public double basePitch;        // Hz
        public double pitchRange;       // semitones
        public String pitchVariation;   // flat, moderate, dynamic

        public double speakingRate;     // words per minute
        public String rateVariation;    // steady, varied
        public double pauseFrequency;   // 0-1

        public double baseVolume;       // 0-1
        public double volumeRange;      // 0-1
        public String volumeVariation;  // steady, dynamic

        public String timbreType;       // warm, bright, dark, neutral
        public String resonance;        // low, medium, high
        public double breathiness;      // 0-1
        public double roughness;        // 0-1

        public String articulationStyle; // precise, relaxed, casual
        public double accentStrength;   // 0-1
        public double distinctiveness;  // 0-1

        public VoiceCharacteristics() {
            this.voiceType = "neutral";
            this.voiceAge = "adult";
            this.voiceQuality = "clear";
            this.basePitch = 150.0;
            this.pitchRange = 12.0;
            this.pitchVariation = "moderate";
            this.speakingRate = 150.0;
            this.rateVariation = "varied";
            this.pauseFrequency = 0.3;
            this.baseVolume = 0.7;
            this.volumeRange = 0.3;
            this.volumeVariation = "dynamic";
            this.timbreType = "neutral";
            this.resonance = "medium";
            this.breathiness = 0.1;
            this.roughness = 0.1;
            this.articulationStyle = "relaxed";
            this.accentStrength = 0.3;
            this.distinctiveness = 0.5;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("voice_type", voiceType);
                json.put("voice_age", voiceAge);
                json.put("voice_quality", voiceQuality);
                json.put("base_pitch", basePitch);
                json.put("pitch_range", pitchRange);
                json.put("pitch_variation", pitchVariation);
                json.put("speaking_rate", speakingRate);
                json.put("rate_variation", rateVariation);
                json.put("pause_frequency", pauseFrequency);
                json.put("base_volume", baseVolume);
                json.put("volume_range", volumeRange);
                json.put("volume_variation", volumeVariation);
                json.put("timbre_type", timbreType);
                json.put("resonance", resonance);
                json.put("breathiness", breathiness);
                json.put("roughness", roughness);
                json.put("articulation_style", articulationStyle);
                json.put("accent_strength", accentStrength);
                json.put("distinctiveness", distinctiveness);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static VoiceCharacteristics fromJson(@NonNull JSONObject json) throws JSONException {
            VoiceCharacteristics chars = new VoiceCharacteristics();
            if (json.has("voice_type")) chars.voiceType = json.getString("voice_type");
            if (json.has("voice_age")) chars.voiceAge = json.getString("voice_age");
            if (json.has("voice_quality")) chars.voiceQuality = json.getString("voice_quality");
            if (json.has("base_pitch")) chars.basePitch = json.getDouble("base_pitch");
            if (json.has("pitch_range")) chars.pitchRange = json.getDouble("pitch_range");
            if (json.has("pitch_variation")) chars.pitchVariation = json.getString("pitch_variation");
            if (json.has("speaking_rate")) chars.speakingRate = json.getDouble("speaking_rate");
            if (json.has("rate_variation")) chars.rateVariation = json.getString("rate_variation");
            if (json.has("pause_frequency")) chars.pauseFrequency = json.getDouble("pause_frequency");
            if (json.has("base_volume")) chars.baseVolume = json.getDouble("base_volume");
            if (json.has("volume_range")) chars.volumeRange = json.getDouble("volume_range");
            if (json.has("volume_variation")) chars.volumeVariation = json.getString("volume_variation");
            if (json.has("timbre_type")) chars.timbreType = json.getString("timbre_type");
            if (json.has("resonance")) chars.resonance = json.getString("resonance");
            if (json.has("breathiness")) chars.breathiness = json.getDouble("breathiness");
            if (json.has("roughness")) chars.roughness = json.getDouble("roughness");
            if (json.has("articulation_style")) chars.articulationStyle = json.getString("articulation_style");
            if (json.has("accent_strength")) chars.accentStrength = json.getDouble("accent_strength");
            if (json.has("distinctiveness")) chars.distinctiveness = json.getDouble("distinctiveness");
            return chars;
        }

        /**
         * 是否男性声音
         */
        public boolean isMaleVoice() {
            return "male".equals(voiceType);
        }

        /**
         * 是否女性声音
         */
        public boolean isFemaleVoice() {
            return "female".equals(voiceType);
        }

        /**
         * 是否年轻声音
         */
        public boolean isYoungVoice() {
            return "child".equals(voiceAge) || "young".equals(voiceAge);
        }

        /**
         * 是否快语速
         */
        public boolean isFastRate() {
            return speakingRate > 180;
        }
    }

    /**
     * 语音风格
     */
    public static class VoiceStyle {
        public String regionalAccent;   // standard, regional, dialect
        public String accentRegion;     // e.g., "beijing"
        public double accentIntensity;  // 0-1

        public String formalityLevel;   // casual, neutral, formal
        public double formalityAdjust;  // -1 to 1

        public String communicationStyle; // direct, indirect
        public double indirectnessLevel;  // 0-1

        public boolean professionalVoice;
        public double voiceAuthority;   // 0-1
        public double voiceWarmth;      // 0-1

        public boolean adaptiveStyle;
        public double styleConsistency; // 0-1

        public VoiceStyle() {
            this.regionalAccent = "standard";
            this.accentIntensity = 0.3;
            this.formalityLevel = "neutral";
            this.formalityAdjust = 0.0;
            this.communicationStyle = "direct";
            this.indirectnessLevel = 0.3;
            this.professionalVoice = false;
            this.voiceAuthority = 0.5;
            this.voiceWarmth = 0.5;
            this.adaptiveStyle = true;
            this.styleConsistency = 0.8;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("regional_accent", regionalAccent);
                json.put("accent_region", accentRegion);
                json.put("accent_intensity", accentIntensity);
                json.put("formality_level", formalityLevel);
                json.put("formality_adjust", formalityAdjust);
                json.put("communication_style", communicationStyle);
                json.put("indirectness_level", indirectnessLevel);
                json.put("professional_voice", professionalVoice);
                json.put("voice_authority", voiceAuthority);
                json.put("voice_warmth", voiceWarmth);
                json.put("adaptive_style", adaptiveStyle);
                json.put("style_consistency", styleConsistency);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static VoiceStyle fromJson(@NonNull JSONObject json) throws JSONException {
            VoiceStyle style = new VoiceStyle();
            if (json.has("regional_accent")) style.regionalAccent = json.getString("regional_accent");
            if (json.has("accent_region")) style.accentRegion = json.getString("accent_region");
            if (json.has("accent_intensity")) style.accentIntensity = json.getDouble("accent_intensity");
            if (json.has("formality_level")) style.formalityLevel = json.getString("formality_level");
            if (json.has("formality_adjust")) style.formalityAdjust = json.getDouble("formality_adjust");
            if (json.has("communication_style")) style.communicationStyle = json.getString("communication_style");
            if (json.has("indirectness_level")) style.indirectnessLevel = json.getDouble("indirectness_level");
            if (json.has("professional_voice")) style.professionalVoice = json.getBoolean("professional_voice");
            if (json.has("voice_authority")) style.voiceAuthority = json.getDouble("voice_authority");
            if (json.has("voice_warmth")) style.voiceWarmth = json.getDouble("voice_warmth");
            if (json.has("adaptive_style")) style.adaptiveStyle = json.getBoolean("adaptive_style");
            if (json.has("style_consistency")) style.styleConsistency = json.getDouble("style_consistency");
            return style;
        }

        /**
         * 是否正式风格
         */
        public boolean isFormalStyle() {
            return "formal".equals(formalityLevel) || formalityAdjust > 0.5;
        }

        /**
         * 是否直接沟通
         */
        public boolean isDirectCommunication() {
            return "direct".equals(communicationStyle) && indirectnessLevel < 0.4;
        }
    }

    /**
     * 情绪语音
     */
    public static class EmotionalVoice {
        public double emotionalExpressiveness; // 0-1
        public double emotionalContagion;      // 0-1
        public double pitchModulationRange;    // semitones
        public double rateModulationRange;     // % change
        public double volumeModulation;        // 0-1
        public String emotionRegulation;       // suppress, express, amplify
        public double baselineStability;       // 0-1

        public EmotionalVoice() {
            this.emotionalExpressiveness = 0.6;
            this.emotionalContagion = 0.5;
            this.pitchModulationRange = 6.0;
            this.rateModulationRange = 0.3;
            this.volumeModulation = 0.2;
            this.emotionRegulation = "express";
            this.baselineStability = 0.7;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("emotional_expressiveness", emotionalExpressiveness);
                json.put("emotional_contagion", emotionalContagion);
                json.put("pitch_modulation_range", pitchModulationRange);
                json.put("rate_modulation_range", rateModulationRange);
                json.put("volume_modulation", volumeModulation);
                json.put("emotion_regulation", emotionRegulation);
                json.put("baseline_stability", baselineStability);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static EmotionalVoice fromJson(@NonNull JSONObject json) throws JSONException {
            EmotionalVoice ev = new EmotionalVoice();
            if (json.has("emotional_expressiveness")) ev.emotionalExpressiveness = json.getDouble("emotional_expressiveness");
            if (json.has("emotional_contagion")) ev.emotionalContagion = json.getDouble("emotional_contagion");
            if (json.has("pitch_modulation_range")) ev.pitchModulationRange = json.getDouble("pitch_modulation_range");
            if (json.has("rate_modulation_range")) ev.rateModulationRange = json.getDouble("rate_modulation_range");
            if (json.has("volume_modulation")) ev.volumeModulation = json.getDouble("volume_modulation");
            if (json.has("emotion_regulation")) ev.emotionRegulation = json.getString("emotion_regulation");
            if (json.has("baseline_stability")) ev.baselineStability = json.getDouble("baseline_stability");
            return ev;
        }

        /**
         * 是否高表达性
         */
        public boolean isHighExpressiveness() {
            return emotionalExpressiveness > 0.6;
        }

        /**
         * 是否放大情绪
         */
        public boolean isAmplifyMode() {
            return "amplify".equals(emotionRegulation);
        }
    }

    /**
     * 语言模式
     */
    public static class SpeechPatterns {
        public String vocabularyLevel;    // simple, moderate, sophisticated
        public double jargonUsage;        // 0-1
        public double idiomUsage;         // 0-1

        public String sentenceLength;     // short, medium, long
        public String sentenceComplexity; // simple, compound, complex

        public List<String> fillerWords;
        public double fillerFrequency;    // 0-1

        public String thoughtOrganization; // linear, associative, structured

        public List<String> listeningResponses;
        public double responseFrequency;   // 0-1

        public double pauseBeforeSpeaking; // 0-1

        public SpeechPatterns() {
            this.vocabularyLevel = "moderate";
            this.jargonUsage = 0.3;
            this.idiomUsage = 0.4;
            this.sentenceLength = "medium";
            this.sentenceComplexity = "compound";
            this.fillerWords = new ArrayList<>();
            this.fillerFrequency = 0.2;
            this.thoughtOrganization = "structured";
            this.listeningResponses = new ArrayList<>();
            this.responseFrequency = 0.4;
            this.pauseBeforeSpeaking = 0.2;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("vocabulary_level", vocabularyLevel);
                json.put("jargon_usage", jargonUsage);
                json.put("idiom_usage", idiomUsage);
                json.put("sentence_length", sentenceLength);
                json.put("sentence_complexity", sentenceComplexity);
                json.put("filler_frequency", fillerFrequency);
                json.put("thought_organization", thoughtOrganization);
                json.put("response_frequency", responseFrequency);
                json.put("pause_before_speaking", pauseBeforeSpeaking);

                JSONArray fillers = new JSONArray();
                for (String f : fillerWords) fillers.put(f);
                json.put("filler_words", fillers);

                JSONArray responses = new JSONArray();
                for (String r : listeningResponses) responses.put(r);
                json.put("listening_responses", responses);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static SpeechPatterns fromJson(@NonNull JSONObject json) throws JSONException {
            SpeechPatterns sp = new SpeechPatterns();
            if (json.has("vocabulary_level")) sp.vocabularyLevel = json.getString("vocabulary_level");
            if (json.has("jargon_usage")) sp.jargonUsage = json.getDouble("jargon_usage");
            if (json.has("idiom_usage")) sp.idiomUsage = json.getDouble("idiom_usage");
            if (json.has("sentence_length")) sp.sentenceLength = json.getString("sentence_length");
            if (json.has("sentence_complexity")) sp.sentenceComplexity = json.getString("sentence_complexity");
            if (json.has("filler_frequency")) sp.fillerFrequency = json.getDouble("filler_frequency");
            if (json.has("thought_organization")) sp.thoughtOrganization = json.getString("thought_organization");
            if (json.has("response_frequency")) sp.responseFrequency = json.getDouble("response_frequency");
            if (json.has("pause_before_speaking")) sp.pauseBeforeSpeaking = json.getDouble("pause_before_speaking");

            if (json.has("filler_words")) {
                JSONArray arr = json.getJSONArray("filler_words");
                for (int i = 0; i < arr.length(); i++) sp.fillerWords.add(arr.getString(i));
            }
            if (json.has("listening_responses")) {
                JSONArray arr = json.getJSONArray("listening_responses");
                for (int i = 0; i < arr.length(); i++) sp.listeningResponses.add(arr.getString(i));
            }
            return sp;
        }

        /**
         * 是否简单词汇
         */
        public boolean isSimpleVocabulary() {
            return "simple".equals(vocabularyLevel);
        }

        /**
         * 是否结构化思维
         */
        public boolean isStructuredThinking() {
            return "structured".equals(thoughtOrganization);
        }
    }

    /**
     * TTS配置
     */
    public static class TTSConfiguration {
        public String engineType;        // built_in, cloud, hybrid
        public String engineName;
        public String engineVersion;

        public String voiceModelId;
        public String voiceModelName;
        public boolean customVoice;

        public int sampleRate;           // Hz
        public int bitDepth;             // 16, 24
        public int channels;             // 1, 2
        public String audioFormat;       // wav, mp3, ogg

        public String qualityLevel;      // low, medium, high
        public String latencyMode;       // real_time, balanced, quality

        public boolean ssmlEnabled;
        public boolean streamingEnabled;
        public boolean cacheEnabled;

        public TTSConfiguration() {
            this.engineType = "built_in";
            this.engineName = "ofa_tts";
            this.engineVersion = "1.0.0";
            this.voiceModelId = "default";
            this.customVoice = false;
            this.sampleRate = 22050;
            this.bitDepth = 16;
            this.channels = 1;
            this.audioFormat = "wav";
            this.qualityLevel = "high";
            this.latencyMode = "balanced";
            this.ssmlEnabled = true;
            this.streamingEnabled = true;
            this.cacheEnabled = true;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("engine_type", engineType);
                json.put("engine_name", engineName);
                json.put("engine_version", engineVersion);
                json.put("voice_model_id", voiceModelId);
                json.put("voice_model_name", voiceModelName);
                json.put("custom_voice", customVoice);
                json.put("sample_rate", sampleRate);
                json.put("bit_depth", bitDepth);
                json.put("channels", channels);
                json.put("audio_format", audioFormat);
                json.put("quality_level", qualityLevel);
                json.put("latency_mode", latencyMode);
                json.put("ssml_enabled", ssmlEnabled);
                json.put("streaming_enabled", streamingEnabled);
                json.put("cache_enabled", cacheEnabled);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static TTSConfiguration fromJson(@NonNull JSONObject json) throws JSONException {
            TTSConfiguration config = new TTSConfiguration();
            if (json.has("engine_type")) config.engineType = json.getString("engine_type");
            if (json.has("engine_name")) config.engineName = json.getString("engine_name");
            if (json.has("engine_version")) config.engineVersion = json.getString("engine_version");
            if (json.has("voice_model_id")) config.voiceModelId = json.getString("voice_model_id");
            if (json.has("voice_model_name")) config.voiceModelName = json.getString("voice_model_name");
            if (json.has("custom_voice")) config.customVoice = json.getBoolean("custom_voice");
            if (json.has("sample_rate")) config.sampleRate = json.getInt("sample_rate");
            if (json.has("bit_depth")) config.bitDepth = json.getInt("bit_depth");
            if (json.has("channels")) config.channels = json.getInt("channels");
            if (json.has("audio_format")) config.audioFormat = json.getString("audio_format");
            if (json.has("quality_level")) config.qualityLevel = json.getString("quality_level");
            if (json.has("latency_mode")) config.latencyMode = json.getString("latency_mode");
            if (json.has("ssml_enabled")) config.ssmlEnabled = json.getBoolean("ssml_enabled");
            if (json.has("streaming_enabled")) config.streamingEnabled = json.getBoolean("streaming_enabled");
            if (json.has("cache_enabled")) config.cacheEnabled = json.getBoolean("cache_enabled");
            return config;
        }

        /**
         * 是否高质量
         */
        public boolean isHighQuality() {
            return "high".equals(qualityLevel) || "ultra".equals(qualityLevel);
        }

        /**
         * 是否实时模式
         */
        public boolean isRealTimeMode() {
            return "real_time".equals(latencyMode);
        }
    }

    /**
     * Voice决策上下文
     */
    public static class VoiceDecisionContext {
        public double recommendedPitch;
        public double recommendedRate;
        public double recommendedVolume;
        public String recommendedTone;
        public String recommendedFormality;

        public VoiceEmotionAdaptation emotionAdaptation;
        public VoiceSceneAdaptation sceneAdaptation;
        public VoiceSocialAdaptation socialAdaptation;

        public VoiceDecisionContext() {
            this.recommendedPitch = 150.0;
            this.recommendedRate = 150.0;
            this.recommendedVolume = 0.7;
            this.recommendedTone = "neutral";
            this.recommendedFormality = "neutral";
            this.emotionAdaptation = new VoiceEmotionAdaptation();
            this.sceneAdaptation = new VoiceSceneAdaptation();
            this.socialAdaptation = new VoiceSocialAdaptation();
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("recommended_pitch", recommendedPitch);
                json.put("recommended_rate", recommendedRate);
                json.put("recommended_volume", recommendedVolume);
                json.put("recommended_tone", recommendedTone);
                json.put("recommended_formality", recommendedFormality);
                json.put("emotion_adaptation", emotionAdaptation.toJson());
                json.put("scene_adaptation", sceneAdaptation.toJson());
                json.put("social_adaptation", socialAdaptation.toJson());
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static VoiceDecisionContext fromJson(@NonNull JSONObject json) throws JSONException {
            VoiceDecisionContext ctx = new VoiceDecisionContext();
            if (json.has("recommended_pitch")) ctx.recommendedPitch = json.getDouble("recommended_pitch");
            if (json.has("recommended_rate")) ctx.recommendedRate = json.getDouble("recommended_rate");
            if (json.has("recommended_volume")) ctx.recommendedVolume = json.getDouble("recommended_volume");
            if (json.has("recommended_tone")) ctx.recommendedTone = json.getString("recommended_tone");
            if (json.has("recommended_formality")) ctx.recommendedFormality = json.getString("recommended_formality");
            if (json.has("emotion_adaptation")) ctx.emotionAdaptation = VoiceEmotionAdaptation.fromJson(json.getJSONObject("emotion_adaptation"));
            if (json.has("scene_adaptation")) ctx.sceneAdaptation = VoiceSceneAdaptation.fromJson(json.getJSONObject("scene_adaptation"));
            if (json.has("social_adaptation")) ctx.socialAdaptation = VoiceSocialAdaptation.fromJson(json.getJSONObject("social_adaptation"));
            return ctx;
        }
    }

    /**
     * 情绪适应
     */
    public static class VoiceEmotionAdaptation {
        public String currentEmotion;
        public double pitchShift;
        public double rateMultiplier;
        public double volumeShift;
        public String timbreAdjust;
        public double expressivenessLevel;

        public VoiceEmotionAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("current_emotion", currentEmotion);
                json.put("pitch_shift", pitchShift);
                json.put("rate_multiplier", rateMultiplier);
                json.put("volume_shift", volumeShift);
                json.put("timbre_adjust", timbreAdjust);
                json.put("expressiveness_level", expressivenessLevel);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static VoiceEmotionAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            VoiceEmotionAdaptation va = new VoiceEmotionAdaptation();
            if (json.has("current_emotion")) va.currentEmotion = json.getString("current_emotion");
            if (json.has("pitch_shift")) va.pitchShift = json.getDouble("pitch_shift");
            if (json.has("rate_multiplier")) va.rateMultiplier = json.getDouble("rate_multiplier");
            if (json.has("volume_shift")) va.volumeShift = json.getDouble("volume_shift");
            if (json.has("timbre_adjust")) va.timbreAdjust = json.getString("timbre_adjust");
            if (json.has("expressiveness_level")) va.expressivenessLevel = json.getDouble("expressiveness_level");
            return va;
        }
    }

    /**
     * 场景适应
     */
    public static class VoiceSceneAdaptation {
        public String scene;
        public double pitchAdjust;
        public double rateAdjust;
        public double volumeAdjust;
        public String formalityAdjust;
        public String articulationMode;

        public VoiceSceneAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("scene", scene);
                json.put("pitch_adjust", pitchAdjust);
                json.put("rate_adjust", rateAdjust);
                json.put("volume_adjust", volumeAdjust);
                json.put("formality_adjust", formalityAdjust);
                json.put("articulation_mode", articulationMode);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static VoiceSceneAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            VoiceSceneAdaptation va = new VoiceSceneAdaptation();
            if (json.has("scene")) va.scene = json.getString("scene");
            if (json.has("pitch_adjust")) va.pitchAdjust = json.getDouble("pitch_adjust");
            if (json.has("rate_adjust")) va.rateAdjust = json.getDouble("rate_adjust");
            if (json.has("volume_adjust")) va.volumeAdjust = json.getDouble("volume_adjust");
            if (json.has("formality_adjust")) va.formalityAdjust = json.getString("formality_adjust");
            if (json.has("articulation_mode")) va.articulationMode = json.getString("articulation_mode");
            return va;
        }
    }

    /**
     * 社交适应
     */
    public static class VoiceSocialAdaptation {
        public String socialContext;
        public double authorityLevel;
        public double warmthLevel;
        public double directness;
        public double pauseBeforeSpeech;

        public VoiceSocialAdaptation() {}

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("social_context", socialContext);
                json.put("authority_level", authorityLevel);
                json.put("warmth_level", warmthLevel);
                json.put("directness", directness);
                json.put("pause_before_speech", pauseBeforeSpeech);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static VoiceSocialAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            VoiceSocialAdaptation va = new VoiceSocialAdaptation();
            if (json.has("social_context")) va.socialContext = json.getString("social_context");
            if (json.has("authority_level")) va.authorityLevel = json.getDouble("authority_level");
            if (json.has("warmth_level")) va.warmthLevel = json.getDouble("warmth_level");
            if (json.has("directness")) va.directness = json.getDouble("directness");
            if (json.has("pause_before_speech")) va.pauseBeforeSpeech = json.getDouble("pause_before_speech");
            return va;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "VoiceState{" +
                "voiceType='" + voiceCharacteristics.voiceType + '\'' +
                ", voiceAge='" + voiceCharacteristics.voiceAge + '\'' +
                ", pitch=" + voiceCharacteristics.basePitch +
                ", rate=" + voiceCharacteristics.speakingRate +
                '}';
    }
}