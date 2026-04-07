// Package tts provides TTS client for Android (v5.6.3).
package com.ofa.agent.tts;

import android.media.AudioAttributes;
import android.media.AudioFormat;
import android.media.AudioTrack;
import android.util.Log;

import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * TTSClient handles Text-to-Speech synthesis.
 *
 * Features:
 * - Real-time streaming synthesis
 * - Voice-to-identity mapping
 * - Audio playback
 * - Voice cloning
 *
 * @since v5.6.3
 */
public class TTSClient {
    private static final String TAG = "TTSClient";

    // Center connection
    private String centerAddress;
    private String authToken;

    // Audio playback
    private AudioTrack audioTrack;
    private boolean isPlaying = false;

    // Voice mappings
    private Map<String, String> identityVoiceMap = new ConcurrentHashMap<>();

    // Voice list cache
    private List<VoiceInfo> voiceList = new ArrayList<>();

    // Background executor
    private ExecutorService executor = Executors.newCachedThreadPool();

    // Listener
    private TTSListener listener;

    /**
     * Creates a new TTS client.
     */
    public TTSClient() {
        this(null, null);
    }

    /**
     * Creates a new TTS client with center connection.
     */
    public TTSClient(String centerAddress, String authToken) {
        this.centerAddress = centerAddress;
        this.authToken = authToken;
    }

    /**
     * Set the TTS listener.
     */
    public void setListener(TTSListener listener) {
        this.listener = listener;
    }

    /**
     * Set center connection.
     */
    public void setCenterConnection(String address, String token) {
        this.centerAddress = address;
        this.authToken = token;
    }

    /**
     * Synthesize text to speech.
     *
     * @param request Synthesis request
     * @return Synthesis result
     */
    public CompletableFuture<SynthesisResult> synthesize(SynthesisRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                // Build request JSON
                JSONObject json = new JSONObject();
                json.put("text", request.getText());
                json.put("identity_id", request.getIdentityId());

                if (request.getVoiceId() != null) {
                    json.put("voice_id", request.getVoiceId());
                }
                if (request.getFormat() != null) {
                    json.put("format", request.getFormat());
                }
                if (request.getSampleRate() > 0) {
                    json.put("sample_rate", request.getSampleRate());
                }
                if (request.getRate() > 0) {
                    json.put("rate", request.getRate());
                }
                if (request.getPitch() > 0) {
                    json.put("pitch", request.getPitch());
                }
                if (request.getVolume() > 0) {
                    json.put("volume", request.getVolume());
                }
                if (request.getEmotion() != null) {
                    json.put("emotion", request.getEmotion());
                }
                if (request.getStyle() != null) {
                    json.put("style", request.getStyle());
                }
                json.put("streaming", request.isStreaming());

                // Send to Center
                String url = centerAddress + "/api/v1/tts/synthesize";
                String response = HttpClient.post(url, json.toString(), authToken);

                // Parse response
                JSONObject respJson = new JSONObject(response);
                SynthesisResult result = new SynthesisResult();
                result.setSessionId(respJson.optString("session_id"));
                result.setDurationMs(respJson.optInt("duration_ms"));
                result.setFormat(respJson.optString("format"));
                result.setVoiceUsed(respJson.optString("voice_used"));
                result.setProvider(respJson.optString("provider"));
                result.setLatencyMs(respJson.optInt("latency_ms"));
                result.setSuccess(respJson.optBoolean("success"));
                result.setError(respJson.optString("error"));

                // Get audio data (base64 encoded)
                String audioBase64 = respJson.optString("audio_data");
                if (!audioBase64.isEmpty()) {
                    result.setAudioData(Base64.decode(audioBase64));
                }

                // Notify listener
                if (listener != null) {
                    listener.onSynthesisComplete(result);
                }

                return result;

            } catch (Exception e) {
                Log.e(TAG, "Synthesis failed", e);
                SynthesisResult result = new SynthesisResult();
                result.setSuccess(false);
                result.setError(e.getMessage());

                if (listener != null) {
                    listener.onSynthesisError(e);
                }

                return result;
            }
        }, executor);
    }

    /**
     * Synthesize and play audio.
     *
     * @param request Synthesis request
     * @return CompletableFuture for async playback
     */
    public CompletableFuture<Void> synthesizeAndPlay(SynthesisRequest request) {
        return synthesize(request).thenAccept(result -> {
            if (result.isSuccess() && result.getAudioData() != null) {
                playAudio(result.getAudioData(), request.getSampleRate());
            }
        });
    }

    /**
     * Stream synthesis in real-time.
     *
     * @param request Synthesis request
     * @return Stream receiver
     */
    public CompletableFuture<StreamReceiver> synthesizeStream(SynthesisRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                // WebSocket streaming
                String wsUrl = centerAddress.replace("http", "ws") + "/api/v1/tts/stream";
                StreamReceiver receiver = new StreamReceiver(wsUrl, authToken, request);

                // Start receiving
                receiver.start();

                if (listener != null) {
                    listener.onStreamStarted(receiver.getSessionId());
                }

                return receiver;

            } catch (Exception e) {
                Log.e(TAG, "Stream synthesis failed", e);
                if (listener != null) {
                    listener.onSynthesisError(e);
                }
                return null;
            }
        }, executor);
    }

    /**
     * List available voices.
     *
     * @param provider Provider name (optional)
     * @return List of voice info
     */
    public CompletableFuture<List<VoiceInfo>> listVoices(String provider) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String url = centerAddress + "/api/v1/tts/voices";
                if (provider != null) {
                    url += "?provider=" + provider;
                }

                String response = HttpClient.get(url, authToken);
                JSONObject respJson = new JSONObject(response);

                List<VoiceInfo> voices = new ArrayList<>();
                org.json.JSONArray arr = respJson.optJSONArray("voices");
                if (arr != null) {
                    for (int i = 0; i < arr.length(); i++) {
                        JSONObject voiceJson = arr.getJSONObject(i);
                        VoiceInfo voice = new VoiceInfo();
                        voice.setVoiceId(voiceJson.optString("voice_id"));
                        voice.setName(voiceJson.optString("name"));
                        voice.setLanguage(voiceJson.optString("language"));
                        voice.setGender(voiceJson.optString("gender"));
                        voice.setAge(voiceJson.optString("age"));
                        voice.setProvider(voiceJson.optString("provider"));
                        voice.setDescription(voiceJson.optString("description"));
                        voices.add(voice);
                    }
                }

                // Cache
                this.voiceList = voices;

                return voices;

            } catch (Exception e) {
                Log.e(TAG, "List voices failed", e);
                return voiceList; // Return cached
            }
        }, executor);
    }

    /**
     * Clone a voice.
     *
     * @param request Clone request
     * @return Clone result
     */
    public CompletableFuture<CloneResult> cloneVoice(CloneRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                JSONObject json = new JSONObject();
                json.put("identity_id", request.getIdentityId());
                json.put("voice_name", request.getVoiceName());
                json.put("language", request.getLanguage());

                org.json.JSONArray refs = new org.json.JSONArray();
                for (ReferenceAudio ref : request.getReferenceAudios()) {
                    JSONObject refJson = new JSONObject();
                    refJson.put("audio_url", ref.getAudioUrl());
                    refJson.put("duration_ms", ref.getDurationMs());
                    refJson.put("transcription", ref.getTranscription());
                    refs.put(refJson);
                }
                json.put("reference_audios", refs);

                String url = centerAddress + "/api/v1/tts/clone";
                String response = HttpClient.post(url, json.toString(), authToken);

                JSONObject respJson = new JSONObject(response);
                CloneResult result = new CloneResult();
                result.setVoiceId(respJson.optString("voice_id"));
                result.setVoiceName(respJson.optString("voice_name"));
                result.setStatus(respJson.optString("status"));
                result.setQuality(respJson.optDouble("quality", 0));
                result.setMessage(respJson.optString("message"));

                // Update identity voice mapping if ready
                if ("ready".equals(result.getStatus())) {
                    setVoiceForIdentity(request.getIdentityId(), result.getVoiceId());
                }

                if (listener != null) {
                    listener.onCloneComplete(result);
                }

                return result;

            } catch (Exception e) {
                Log.e(TAG, "Clone voice failed", e);
                if (listener != null) {
                    listener.onCloneError(e);
                }
                return new CloneResult("failed", e.getMessage());
            }
        }, executor);
    }

    /**
     * Set voice for identity.
     */
    public void setVoiceForIdentity(String identityId, String voiceId) {
        identityVoiceMap.put(identityId, voiceId);
    }

    /**
     * Get voice for identity.
     */
    public String getVoiceForIdentity(String identityId) {
        return identityVoiceMap.get(identityId);
    }

    /**
     * Play audio data.
     */
    public void playAudio(byte[] audioData, int sampleRate) {
        if (audioData == null || audioData.length == 0) {
            return;
        }

        stopPlayback();

        // Create audio track
        int sampleRateHz = sampleRate > 0 ? sampleRate : 24000;
        int channelConfig = AudioFormat.CHANNEL_OUT_MONO;
        int audioFormat = AudioFormat.ENCODING_PCM_16BIT;
        int bufferSize = AudioTrack.getMinBufferSize(sampleRateHz, channelConfig, audioFormat);

        audioTrack = new AudioTrack(
            new AudioAttributes.Builder()
                .setUsage(AudioAttributes.USAGE_MEDIA)
                .setContentType(AudioAttributes.CONTENT_TYPE_SPEECH)
                .build(),
            new AudioFormat.Builder()
                .setSampleRate(sampleRateHz)
                .setChannelMask(channelConfig)
                .setEncoding(audioFormat)
                .build(),
            bufferSize,
            AudioTrack.MODE_STREAM,
            AudioTrack.AUDIO_SESSION_ID_GENERATE
        );

        audioTrack.play();
        isPlaying = true;

        // Write audio data
        executor.execute(() -> {
            audioTrack.write(audioData, 0, audioData.length);

            if (listener != null) {
                listener.onPlaybackComplete();
            }

            isPlaying = false;
        });
    }

    /**
     * Stop playback.
     */
    public void stopPlayback() {
        if (audioTrack != null && isPlaying) {
            audioTrack.stop();
            audioTrack.release();
            audioTrack = null;
            isPlaying = false;
        }
    }

    /**
     * Check if playing.
     */
    public boolean isPlaying() {
        return isPlaying;
    }

    /**
     * Close the client.
     */
    public void close() {
        stopPlayback();
        executor.shutdown();
    }

    // Inner classes

    /**
     * Stream receiver for real-time synthesis.
     */
    public class StreamReceiver {
        private String wsUrl;
        private String authToken;
        private SynthesisRequest request;
        private String sessionId;
        private ByteArrayOutputStream audioBuffer = new ByteArrayOutputStream();
        private boolean isReceiving = false;

        public StreamReceiver(String wsUrl, String authToken, SynthesisRequest request) {
            this.wsUrl = wsUrl;
            this.authToken = authToken;
            this.request = request;
        }

        public void start() {
            isReceiving = true;
            // WebSocket connection would be implemented with OkHttp or similar
            // For now, placeholder
        }

        public void stop() {
            isReceiving = false;
        }

        public byte[] getAudioData() {
            return audioBuffer.toByteArray();
        }

        public String getSessionId() {
            return sessionId;
        }

        public boolean isReceiving() {
            return isReceiving;
        }
    }

    // HTTP utilities (simplified)
    private static class HttpClient {
        static String post(String url, String body, String token) throws IOException {
            // Placeholder - actual implementation would use OkHttp
            java.net.HttpURLConnection conn = (java.net.HttpURLConnection) new java.net.URL(url).openConnection();
            conn.setRequestMethod("POST");
            conn.setRequestProperty("Content-Type", "application/json");
            if (token != null) {
                conn.setRequestProperty("Authorization", "Bearer " + token);
            }
            conn.setDoOutput(true);
            conn.getOutputStream().write(body.getBytes());

            java.io.BufferedReader reader = new java.io.BufferedReader(
                new java.io.InputStreamReader(conn.getInputStream()));
            StringBuilder response = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                response.append(line);
            }
            reader.close();
            return response.toString();
        }

        static String get(String url, String token) throws IOException {
            java.net.HttpURLConnection conn = (java.net.HttpURLConnection) new java.net.URL(url).openConnection();
            conn.setRequestMethod("GET");
            if (token != null) {
                conn.setRequestProperty("Authorization", "Bearer " + token);
            }

            java.io.BufferedReader reader = new java.io.BufferedReader(
                new java.io.InputStreamReader(conn.getInputStream()));
            StringBuilder response = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                response.append(line);
            }
            reader.close();
            return response.toString();
        }
    }

    // Base64 decoder placeholder
    private static class Base64 {
        static byte[] decode(String str) {
            return android.util.Base64.decode(str, android.util.Base64.DEFAULT);
        }
    }
}