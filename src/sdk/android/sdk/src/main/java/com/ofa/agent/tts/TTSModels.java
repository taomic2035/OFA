// Package tts provides TTS models (v5.6.3).
package com.ofa.agent.tts;

import java.util.ArrayList;
import java.util.List;

/**
 * SynthesisRequest represents a TTS synthesis request.
 *
 * @since v5.6.3
 */
public class SynthesisRequest {
    private String text;
    private String identityId;
    private String voiceId;
    private String format;
    private int sampleRate;
    private double rate;
    private double pitch;
    private double volume;
    private String emotion;
    private String style;
    private boolean streaming;

    public SynthesisRequest() {
    }

    public SynthesisRequest(String text) {
        this.text = text;
    }

    public SynthesisRequest(String text, String voiceId) {
        this.text = text;
        this.voiceId = voiceId;
    }

    // Getters and setters

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }

    public String getIdentityId() {
        return identityId;
    }

    public void setIdentityId(String identityId) {
        this.identityId = identityId;
    }

    public String getVoiceId() {
        return voiceId;
    }

    public void setVoiceId(String voiceId) {
        this.voiceId = voiceId;
    }

    public String getFormat() {
        return format;
    }

    public void setFormat(String format) {
        this.format = format;
    }

    public int getSampleRate() {
        return sampleRate;
    }

    public void setSampleRate(int sampleRate) {
        this.sampleRate = sampleRate;
    }

    public double getRate() {
        return rate;
    }

    public void setRate(double rate) {
        this.rate = rate;
    }

    public double getPitch() {
        return pitch;
    }

    public void setPitch(double pitch) {
        this.pitch = pitch;
    }

    public double getVolume() {
        return volume;
    }

    public void setVolume(double volume) {
        this.volume = volume;
    }

    public String getEmotion() {
        return emotion;
    }

    public void setEmotion(String emotion) {
        this.emotion = emotion;
    }

    public String getStyle() {
        return style;
    }

    public void setStyle(String style) {
        this.style = style;
    }

    public boolean isStreaming() {
        return streaming;
    }

    public void setStreaming(boolean streaming) {
        this.streaming = streaming;
    }

    // Builder pattern

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private SynthesisRequest request = new SynthesisRequest();

        public Builder text(String text) {
            request.setText(text);
            return this;
        }

        public Builder identityId(String identityId) {
            request.setIdentityId(identityId);
            return this;
        }

        public Builder voiceId(String voiceId) {
            request.setVoiceId(voiceId);
            return this;
        }

        public Builder format(String format) {
            request.setFormat(format);
            return this;
        }

        public Builder sampleRate(int sampleRate) {
            request.setSampleRate(sampleRate);
            return this;
        }

        public Builder rate(double rate) {
            request.setRate(rate);
            return this;
        }

        public Builder pitch(double pitch) {
            request.setPitch(pitch);
            return this;
        }

        public Builder volume(double volume) {
            request.setVolume(volume);
            return this;
        }

        public Builder emotion(String emotion) {
            request.setEmotion(emotion);
            return this;
        }

        public Builder style(String style) {
            request.setStyle(style);
            return this;
        }

        public Builder streaming(boolean streaming) {
            request.setStreaming(streaming);
            return this;
        }

        public SynthesisRequest build() {
            return request;
        }
    }
}

/**
 * SynthesisResult represents a TTS synthesis result.
 *
 * @since v5.6.3
 */
public class SynthesisResult {
    private String sessionId;
    private byte[] audioData;
    private String audioUrl;
    private int durationMs;
    private String format;
    private String voiceUsed;
    private String provider;
    private int latencyMs;
    private boolean success;
    private String error;

    public SynthesisResult() {
    }

    // Getters and setters

    public String getSessionId() {
        return sessionId;
    }

    public void setSessionId(String sessionId) {
        this.sessionId = sessionId;
    }

    public byte[] getAudioData() {
        return audioData;
    }

    public void setAudioData(byte[] audioData) {
        this.audioData = audioData;
    }

    public String getAudioUrl() {
        return audioUrl;
    }

    public void setAudioUrl(String audioUrl) {
        this.audioUrl = audioUrl;
    }

    public int getDurationMs() {
        return durationMs;
    }

    public void setDurationMs(int durationMs) {
        this.durationMs = durationMs;
    }

    public String getFormat() {
        return format;
    }

    public void setFormat(String format) {
        this.format = format;
    }

    public String getVoiceUsed() {
        return voiceUsed;
    }

    public void setVoiceUsed(String voiceUsed) {
        this.voiceUsed = voiceUsed;
    }

    public String getProvider() {
        return provider;
    }

    public void setProvider(String provider) {
        this.provider = provider;
    }

    public int getLatencyMs() {
        return latencyMs;
    }

    public void setLatencyMs(int latencyMs) {
        this.latencyMs = latencyMs;
    }

    public boolean isSuccess() {
        return success;
    }

    public void setSuccess(boolean success) {
        this.success = success;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }
}

/**
 * VoiceInfo represents information about a voice.
 *
 * @since v5.6.3
 */
public class VoiceInfo {
    private String voiceId;
    private String name;
    private String language;
    private String gender;
    private String age;
    private String style;
    private List<String> emotions = new ArrayList<>();
    private String provider;
    private String description;

    public VoiceInfo() {
    }

    public VoiceInfo(String voiceId, String name) {
        this.voiceId = voiceId;
        this.name = name;
    }

    // Getters and setters

    public String getVoiceId() {
        return voiceId;
    }

    public void setVoiceId(String voiceId) {
        this.voiceId = voiceId;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getLanguage() {
        return language;
    }

    public void setLanguage(String language) {
        this.language = language;
    }

    public String getGender() {
        return gender;
    }

    public void setGender(String gender) {
        this.gender = gender;
    }

    public String getAge() {
        return age;
    }

    public void setAge(String age) {
        this.age = age;
    }

    public String getStyle() {
        return style;
    }

    public void setStyle(String style) {
        this.style = style;
    }

    public List<String> getEmotions() {
        return emotions;
    }

    public void setEmotions(List<String> emotions) {
        this.emotions = emotions;
    }

    public String getProvider() {
        return provider;
    }

    public void setProvider(String provider) {
        this.provider = provider;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }
}

/**
 * CloneRequest represents a voice cloning request.
 *
 * @since v5.6.3
 */
public class CloneRequest {
    private String identityId;
    private String voiceName;
    private String language;
    private List<ReferenceAudio> referenceAudios = new ArrayList<>();

    public CloneRequest() {
    }

    public CloneRequest(String identityId, String voiceName) {
        this.identityId = identityId;
        this.voiceName = voiceName;
    }

    // Getters and setters

    public String getIdentityId() {
        return identityId;
    }

    public void setIdentityId(String identityId) {
        this.identityId = identityId;
    }

    public String getVoiceName() {
        return voiceName;
    }

    public void setVoiceName(String voiceName) {
        this.voiceName = voiceName;
    }

    public String getLanguage() {
        return language;
    }

    public void setLanguage(String language) {
        this.language = language;
    }

    public List<ReferenceAudio> getReferenceAudios() {
        return referenceAudios;
    }

    public void setReferenceAudios(List<ReferenceAudio> referenceAudios) {
        this.referenceAudios = referenceAudios;
    }

    public void addReferenceAudio(String audioUrl, int durationMs) {
        referenceAudios.add(new ReferenceAudio(audioUrl, durationMs));
    }

    public void addReferenceAudio(String audioUrl, int durationMs, String transcription) {
        referenceAudios.add(new ReferenceAudio(audioUrl, durationMs, transcription));
    }
}

/**
 * ReferenceAudio represents reference audio for voice cloning.
 *
 * @since v5.6.3
 */
public class ReferenceAudio {
    private String audioUrl;
    private int durationMs;
    private String transcription;

    public ReferenceAudio() {
    }

    public ReferenceAudio(String audioUrl, int durationMs) {
        this.audioUrl = audioUrl;
        this.durationMs = durationMs;
    }

    public ReferenceAudio(String audioUrl, int durationMs, String transcription) {
        this.audioUrl = audioUrl;
        this.durationMs = durationMs;
        this.transcription = transcription;
    }

    // Getters and setters

    public String getAudioUrl() {
        return audioUrl;
    }

    public void setAudioUrl(String audioUrl) {
        this.audioUrl = audioUrl;
    }

    public int getDurationMs() {
        return durationMs;
    }

    public void setDurationMs(int durationMs) {
        this.durationMs = durationMs;
    }

    public String getTranscription() {
        return transcription;
    }

    public void setTranscription(String transcription) {
        this.transcription = transcription;
    }
}

/**
 * CloneResult represents a voice cloning result.
 *
 * @since v5.6.3
 */
public class CloneResult {
    private String voiceId;
    private String voiceName;
    private String status;
    private double quality;
    private String message;

    public CloneResult() {
    }

    public CloneResult(String status, String message) {
        this.status = status;
        this.message = message;
    }

    // Getters and setters

    public String getVoiceId() {
        return voiceId;
    }

    public void setVoiceId(String voiceId) {
        this.voiceId = voiceId;
    }

    public String getVoiceName() {
        return voiceName;
    }

    public void setVoiceName(String voiceName) {
        this.voiceName = voiceName;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public double getQuality() {
        return quality;
    }

    public void setQuality(double quality) {
        this.quality = quality;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public boolean isReady() {
        return "ready".equals(status);
    }

    public boolean isPending() {
        return "pending".equals(status);
    }

    public boolean isFailed() {
        return "failed".equals(status);
    }
}

/**
 * TTSListener for TTS events.
 *
 * @since v5.6.3
 */
public interface TTSListener {
    /**
     * Called when synthesis completes.
     */
    void onSynthesisComplete(SynthesisResult result);

    /**
     * Called when synthesis fails.
     */
    void onSynthesisError(Exception error);

    /**
     * Called when stream starts.
     */
    void onStreamStarted(String sessionId);

    /**
     * Called when stream chunk received.
     */
    void onStreamChunk(String sessionId, byte[] data, boolean isLast);

    /**
     * Called when stream ends.
     */
    void onStreamComplete(String sessionId);

    /**
     * Called when playback completes.
     */
    void onPlaybackComplete();

    /**
     * Called when voice clone completes.
     */
    void onCloneComplete(CloneResult result);

    /**
     * Called when voice clone fails.
     */
    void onCloneError(Exception error);
}