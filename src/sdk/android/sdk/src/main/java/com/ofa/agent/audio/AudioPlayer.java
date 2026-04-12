package com.ofa.agent.audio;

import android.media.AudioAttributes;
import android.media.AudioFormat;
import android.media.AudioManager;
import android.media.AudioTrack;
import android.util.Log;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * AudioPlayer handles real-time audio playback for TTS and voice streaming (v8.0.1).
 *
 * Supports:
 * - PCM audio streaming playback
 * - Real-time audio buffer queue
 * - Volume control
 * - Play/Pause/Stop controls
 * - Integration with TTS engine and chat responses
 */
public class AudioPlayer {
    private static final String TAG = "AudioPlayer";

    // Audio configuration
    private static final int SAMPLE_RATE = 24000; // Standard TTS sample rate
    private static final int CHANNEL_CONFIG = AudioFormat.CHANNEL_OUT_MONO;
    private static final int AUDIO_FORMAT = AudioFormat.ENCODING_PCM_16BIT;
    private static final int BUFFER_SIZE = AudioTrack.getMinBufferSize(
            SAMPLE_RATE, CHANNEL_CONFIG, AUDIO_FORMAT) * 2;

    // Audio track for playback
    private AudioTrack audioTrack;

    // Buffer queue for streaming audio
    private BlockingQueue<byte[]> audioQueue;

    // State management
    private AtomicBoolean isPlaying;
    private AtomicBoolean isPaused;
    private AtomicBoolean stopRequested;

    // Playback thread
    private Thread playbackThread;

    // Volume (0-1)
    private float volume = 1.0f;

    // Current playback position
    private long totalBytesPlayed;
    private long startTime;

    /**
     * Creates a new AudioPlayer.
     */
    public AudioPlayer() {
        this.audioQueue = new LinkedBlockingQueue<>();
        this.isPlaying = new AtomicBoolean(false);
        this.isPaused = new AtomicBoolean(false);
        this.stopRequested = new AtomicBoolean(false);
        this.totalBytesPlayed = 0;
    }

    /**
     * Initializes the audio track.
     */
    public void initialize() {
        if (audioTrack != null) {
            release();
        }

        AudioAttributes attributes = new AudioAttributes.Builder()
                .setUsage(AudioAttributes.USAGE_MEDIA)
                .setContentType(AudioAttributes.CONTENT_TYPE_SPEECH)
                .build();

        audioTrack = new AudioTrack(
                attributes,
                new AudioFormat.Builder()
                        .setSampleRate(SAMPLE_RATE)
                        .setChannelMask(CHANNEL_CONFIG)
                        .setEncoding(AUDIO_FORMAT)
                        .build(),
                BUFFER_SIZE,
                AudioTrack.MODE_STREAM,
                AudioManager.AUDIO_SESSION_ID_GENERATE
        );

        audioTrack.setVolume(volume);

        Log.d(TAG, "AudioPlayer initialized with buffer size: " + BUFFER_SIZE);
    }

    /**
     * Starts playback from an audio stream.
     *
     * @param inputStream The audio stream to play
     */
    public void playStream(InputStream inputStream) {
        if (isPlaying.get()) {
            Log.w(TAG, "Already playing, stopping previous playback");
            stop();
        }

        initialize();
        stopRequested.set(false);
        isPlaying.set(true);
        isPaused.set(false);
        totalBytesPlayed = 0;
        startTime = System.currentTimeMillis();

        // Clear any queued audio
        audioQueue.clear();

        // Start playback thread
        playbackThread = new Thread(() -> {
            audioTrack.play();

            try {
                byte[] buffer = new byte[4096];
                int bytesRead;

                while (!stopRequested.get() && (bytesRead = inputStream.read(buffer)) != -1) {
                    if (isPaused.get()) {
                        // Wait while paused
                        while (isPaused.get() && !stopRequested.get()) {
                            Thread.sleep(50);
                        }
                        if (stopRequested.get()) break;
                    }

                    // Write to audio track
                    int written = audioTrack.write(buffer, 0, bytesRead);
                    if (written > 0) {
                        totalBytesPlayed += written;
                    }
                }

            } catch (IOException | InterruptedException e) {
                Log.e(TAG, "Error during playback", e);
            } finally {
                audioTrack.stop();
                isPlaying.set(false);

                try {
                    inputStream.close();
                } catch (IOException e) {
                    Log.e(TAG, "Error closing stream", e);
                }

                Log.d(TAG, "Playback completed. Total bytes: " + totalBytesPlayed);
            }
        });

        playbackThread.start();
    }

    /**
     * Queues audio data for streaming playback.
     *
     * @param audioData PCM audio data
     */
    public void queueAudio(byte[] audioData) {
        if (!isPlaying.get()) {
            initialize();
            audioTrack.play();
            isPlaying.set(true);
            isPaused.set(false);
            startTime = System.currentTimeMillis();

            // Start playback thread
            playbackThread = new Thread(this::playbackLoop);
            playbackThread.start();
        }

        audioQueue.offer(audioData);
    }

    /**
     * Playback loop for queued audio.
     */
    private void playbackLoop() {
        try {
            while (!stopRequested.get() && isPlaying.get()) {
                if (isPaused.get()) {
                    Thread.sleep(50);
                    continue;
                }

                byte[] data = audioQueue.poll();
                if (data != null) {
                    int written = audioTrack.write(data, 0, data.length);
                    if (written > 0) {
                        totalBytesPlayed += written;
                    }
                } else {
                    // No data available, brief pause
                    Thread.sleep(10);
                }
            }
        } catch (InterruptedException e) {
            Log.e(TAG, "Playback loop interrupted", e);
        } finally {
            if (audioTrack != null) {
                audioTrack.stop();
            }
            isPlaying.set(false);
        }
    }

    /**
     * Plays audio data directly (non-streaming).
     *
     * @param audioData PCM audio data
     */
    public void play(byte[] audioData) {
        if (isPlaying.get()) {
            stop();
        }

        initialize();
        audioTrack.play();
        isPlaying.set(true);

        // Write all data
        int written = audioTrack.write(audioData, 0, audioData.length);
        totalBytesPlayed = written;

        // Wait for playback to complete
        try {
            Thread.sleep(audioData.length * 1000 / (SAMPLE_RATE * 2)); // Calculate duration
        } catch (InterruptedException e) {
            Log.e(TAG, "Interrupted during playback", e);
        }

        audioTrack.stop();
        isPlaying.set(false);
    }

    /**
     * Pauses playback.
     */
    public void pause() {
        if (isPlaying.get() && !isPaused.get()) {
            isPaused.set(true);
            audioTrack.pause();
            Log.d(TAG, "Playback paused");
        }
    }

    /**
     * Resumes playback.
     */
    public void resume() {
        if (isPlaying.get() && isPaused.get()) {
            isPaused.set(false);
            audioTrack.play();
            Log.d(TAG, "Playback resumed");
        }
    }

    /**
     * Stops playback.
     */
    public void stop() {
        stopRequested.set(true);
        isPaused.set(false);

        if (audioTrack != null) {
            audioTrack.stop();
            audioTrack.flush();
        }

        if (playbackThread != null) {
            try {
                playbackThread.join(1000);
            } catch (InterruptedException e) {
                Log.e(TAG, "Error stopping playback thread", e);
            }
        }

        audioQueue.clear();
        isPlaying.set(false);

        Log.d(TAG, "Playback stopped");
    }

    /**
     * Sets the volume (0-1).
     *
     * @param volume Volume level
     */
    public void setVolume(float volume) {
        this.volume = Math.max(0f, Math.min(1f, volume));
        if (audioTrack != null) {
            audioTrack.setVolume(this.volume);
        }
    }

    /**
     * Gets the current volume.
     *
     * @return Volume level (0-1)
     */
    public float getVolume() {
        return volume;
    }

    /**
     * Checks if playback is active.
     *
     * @return true if playing
     */
    public boolean isPlaying() {
        return isPlaying.get() && !isPaused.get();
    }

    /**
     * Checks if playback is paused.
     *
     * @return true if paused
     */
    public boolean isPaused() {
        return isPaused.get();
    }

    /**
     * Gets the playback duration in seconds.
     *
     * @return Duration in seconds
     */
    public float getDuration() {
        // Calculate duration from bytes played
        // For 16-bit mono at sample rate: duration = bytes / (sampleRate * 2)
        return totalBytesPlayed / (SAMPLE_RATE * 2f);
    }

    /**
     * Gets the elapsed playback time.
     *
     * @return Elapsed time in seconds
     */
    public float getElapsedTime() {
        if (startTime == 0) return 0;
        return (System.currentTimeMillis() - startTime) / 1000f;
    }

    /**
     * Releases resources.
     */
    public void release() {
        stop();

        if (audioTrack != null) {
            audioTrack.release();
            audioTrack = null;
        }

        audioQueue.clear();

        Log.d(TAG, "AudioPlayer released");
    }

    /**
     * Clears the audio queue.
     */
    public void clearQueue() {
        audioQueue.clear();
        Log.d(TAG, "Audio queue cleared");
    }

    /**
     * Gets the queue size.
     *
     * @return Number of audio chunks in queue
     */
    public int getQueueSize() {
        return audioQueue.size();
    }
}