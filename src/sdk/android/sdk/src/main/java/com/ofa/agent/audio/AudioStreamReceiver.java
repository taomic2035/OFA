package com.ofa.agent.audio;

import android.util.Log;

import com.ofa.agent.websocket.WebSocketClient;
import com.ofa.agent.websocket.WebSocketMessage;

import org.json.JSONObject;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * AudioStreamReceiver handles receiving audio streams from Center (v8.0.1).
 *
 * Integrates with:
 * - WebSocket for audio stream messages
 * - AudioPlayer for playback
 * - TTS engine responses
 * - Chat response audio generation
 */
public class AudioStreamReceiver {
    private static final String TAG = "AudioStreamReceiver";

    // Audio player for playback
    private AudioPlayer audioPlayer;

    // WebSocket client for receiving streams
    private WebSocketClient webSocketClient;

    // Audio buffer queue
    private BlockingQueue<byte[]> streamBuffer;

    // State management
    private AtomicBoolean isReceiving;
    private AtomicBoolean autoPlay;
    private AtomicBoolean stopRequested;

    // Current stream session
    private String currentStreamId;
    private ByteArrayOutputStream currentStreamData;

    // Stream listener
    private StreamListener streamListener;

    /**
     * Creates a new AudioStreamReceiver.
     */
    public AudioStreamReceiver() {
        this.streamBuffer = new LinkedBlockingQueue<>();
        this.audioPlayer = new AudioPlayer();
        this.isReceiving = new AtomicBoolean(false);
        this.autoPlay = new AtomicBoolean(true);
        this.stopRequested = new AtomicBoolean(false);
        this.currentStreamData = new ByteArrayOutputStream();
    }

    /**
     * Sets the WebSocket client.
     *
     * @param client WebSocket client
     */
    public void setWebSocketClient(WebSocketClient client) {
        this.webSocketClient = client;

        // Register audio stream handler
        if (client != null) {
            client.registerHandler("audio_stream", this::handleAudioStreamMessage);
            client.registerHandler("audio_chunk", this::handleAudioChunkMessage);
            client.registerHandler("audio_end", this::handleAudioEndMessage);
        }
    }

    /**
     * Sets auto-play mode.
     *
     * @param autoPlay true to automatically play received audio
     */
    public void setAutoPlay(boolean autoPlay) {
        this.autoPlay.set(autoPlay);
    }

    /**
     * Sets the stream listener.
     *
     * @param listener Stream listener
     */
    public void setStreamListener(StreamListener listener) {
        this.streamListener = listener;
    }

    /**
     * Handles audio stream start message.
     *
     * @param message WebSocket message
     */
    private void handleAudioStreamMessage(WebSocketMessage message) {
        try {
            JSONObject payload = message.getPayload();
            currentStreamId = payload.optString("stream_id");
            String format = payload.optString("format", "pcm");
            int sampleRate = payload.optInt("sample_rate", 24000);

            Log.d(TAG, "Audio stream started: " + currentStreamId);

            // Reset stream data
            currentStreamData = new ByteArrayOutputStream();
            isReceiving.set(true);
            stopRequested.set(false);
            streamBuffer.clear();

            // Initialize audio player with correct sample rate
            audioPlayer.initialize();

            // Notify listener
            if (streamListener != null) {
                streamListener.onStreamStart(currentStreamId, format, sampleRate);
            }

        } catch (Exception e) {
            Log.e(TAG, "Error handling audio stream message", e);
        }
    }

    /**
     * Handles audio chunk message.
     *
     * @param message WebSocket message
     */
    private void handleAudioChunkMessage(WebSocketMessage message) {
        try {
            JSONObject payload = message.getPayload();
            String streamId = payload.optString("stream_id");
            String audioDataStr = payload.optString("data");

            // Decode base64 audio data
            byte[] audioData = decodeBase64(audioDataStr);

            Log.d(TAG, "Received audio chunk: " + audioData.length + " bytes");

            // Add to stream buffer
            currentStreamData.write(audioData);
            streamBuffer.offer(audioData);

            // Auto-play if enabled
            if (autoPlay.get() && audioPlayer.isPlaying()) {
                audioPlayer.queueAudio(audioData);
            }

            // Notify listener
            if (streamListener != null) {
                streamListener.onStreamChunk(streamId, audioData);
            }

        } catch (Exception e) {
            Log.e(TAG, "Error handling audio chunk message", e);
        }
    }

    /**
     * Handles audio stream end message.
     *
     * @param message WebSocket message
     */
    private void handleAudioEndMessage(WebSocketMessage message) {
        try {
            JSONObject payload = message.getPayload();
            String streamId = payload.optString("stream_id");
            int totalSize = payload.optInt("total_size");

            Log.d(TAG, "Audio stream ended: " + streamId + ", total size: " + totalSize);

            isReceiving.set(false);

            // Get complete stream data
            byte[] completeAudio = currentStreamData.toByteArray();

            // Play if auto-play enabled and not already playing
            if (autoPlay.get() && !audioPlayer.isPlaying()) {
                audioPlayer.play(completeAudio);
            }

            // Notify listener
            if (streamListener != null) {
                streamListener.onStreamEnd(streamId, completeAudio);
            }

        } catch (Exception e) {
            Log.e(TAG, "Error handling audio end message", e);
        }
    }

    /**
     * Request TTS audio stream from Center.
     *
     * @param text Text to synthesize
     * @param voiceId Voice ID (optional)
     * @param speed Speech speed (optional)
     */
    public void requestTTSStream(String text, String voiceId, float speed) {
        if (webSocketClient == null || !webSocketClient.isConnected()) {
            Log.e(TAG, "WebSocket not connected");
            return;
        }

        try {
            JSONObject payload = new JSONObject();
            payload.put("type", "tts_request");
            payload.put("text", text);
            if (voiceId != null) {
                payload.put("voice_id", voiceId);
            }
            if (speed > 0) {
                payload.put("speed", speed);
            }
            payload.put("stream", true);

            WebSocketMessage message = new WebSocketMessage("tts_request", payload);
            webSocketClient.send(message);

            Log.d(TAG, "TTS stream requested: " + text.substring(0, Math.min(50, text.length())));

        } catch (Exception e) {
            Log.e(TAG, "Error requesting TTS stream", e);
        }
    }

    /**
     * Request chat response audio stream.
     *
     * @param conversationId Conversation ID
     * @param response Chat response text
     */
    public void requestChatAudio(String conversationId, String response) {
        requestTTSStream(response, null, 1.0f);
    }

    /**
     * Plays received audio data.
     */
    public void playReceivedAudio() {
        if (!isReceiving.get() && currentStreamData.size() > 0) {
            byte[] audioData = currentStreamData.toByteArray();
            audioPlayer.play(audioData);
        }
    }

    /**
     * Stops receiving and playback.
     */
    public void stop() {
        stopRequested.set(true);
        isReceiving.set(false);
        audioPlayer.stop();
        streamBuffer.clear();
        currentStreamData.reset();

        Log.d(TAG, "Audio stream receiver stopped");
    }

    /**
     * Pauses playback.
     */
    public void pause() {
        audioPlayer.pause();
    }

    /**
     * Resumes playback.
     */
    public void resume() {
        audioPlayer.resume();
    }

    /**
     * Sets playback volume.
     *
     * @param volume Volume (0-1)
     */
    public void setVolume(float volume) {
        audioPlayer.setVolume(volume);
    }

    /**
     * Gets playback volume.
     *
     * @return Volume (0-1)
     */
    public float getVolume() {
        return audioPlayer.getVolume();
    }

    /**
     * Checks if receiving audio.
     *
     * @return true if receiving
     */
    public boolean isReceiving() {
        return isReceiving.get();
    }

    /**
     * Checks if playing audio.
     *
     * @return true if playing
     */
    public boolean isPlaying() {
        return audioPlayer.isPlaying();
    }

    /**
     * Gets the audio player.
     *
     * @return AudioPlayer instance
     */
    public AudioPlayer getAudioPlayer() {
        return audioPlayer;
    }

    /**
     * Gets current stream data.
     *
     * @return Stream data as InputStream
     */
    public InputStream getStreamData() {
        return new ByteArrayInputStream(currentStreamData.toByteArray());
    }

    /**
     * Releases resources.
     */
    public void release() {
        stop();

        if (audioPlayer != null) {
            audioPlayer.release();
        }

        if (webSocketClient != null) {
            webSocketClient.unregisterHandler("audio_stream");
            webSocketClient.unregisterHandler("audio_chunk");
            webSocketClient.unregisterHandler("audio_end");
        }

        Log.d(TAG, "Audio stream receiver released");
    }

    /**
     * Decodes base64 string to bytes.
     */
    private byte[] decodeBase64(String base64) {
        return java.util.Base64.getDecoder().decode(base64);
    }

    /**
     * Stream listener interface.
     */
    public interface StreamListener {
        /**
         * Called when stream starts.
         */
        void onStreamStart(String streamId, String format, int sampleRate);

        /**
         * Called when chunk received.
         */
        void onStreamChunk(String streamId, byte[] audioData);

        /**
         * Called when stream ends.
         */
        void onStreamEnd(String streamId, byte[] completeAudio);
    }
}