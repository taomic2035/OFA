package com.ofa.agent.distributed;

import android.content.Context;
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.location.Location;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Scene Detector - detects user context/scene from sensors and data.
 *
 * Detection sources:
 * 1. Motion sensors (accelerometer, gyroscope) -> activity recognition
 * 2. Heart rate sensor -> physical intensity
 * 3. GPS/location -> location context
 * 4. Time of day -> schedule-based context
 * 5. Calendar events -> meeting/working context
 * 6. Bluetooth devices -> driving/commuting context
 *
 * The detector runs continuously and broadcasts scene changes to listeners.
 */
public class SceneDetector implements SensorEventListener {

    private static final String TAG = "SceneDetector";

    private final Context context;
    private final SensorManager sensorManager;
    private final ExecutorService executor;
    private final Handler mainHandler;

    // Sensors
    private Sensor accelerometer;
    private Sensor gyroscope;
    private Sensor heartRateSensor;
    private Sensor stepCounter;

    // Detection state
    private volatile boolean detecting = false;
    private SceneContext currentScene;
    private final List<SceneChangeListener> listeners = new ArrayList<>();

    // Sensor data buffers
    private final float[] accelerometerData = new float[3];
    private final float[] gyroscopeData = new float[3];
    private float currentHeartRate = 0;
    private int stepCount = 0;

    // Activity recognition state
    private String detectedActivity = SceneContext.UNKNOWN;
    private float activityConfidence = 0f;
    private long lastActivityUpdate = 0;

    // Detection parameters
    private static final long DETECTION_INTERVAL_MS = 5000; // 5 seconds
    private static final float RUNNING_HEART_RATE_THRESHOLD = 120f;
    private static final float WALKING_HEART_RATE_THRESHOLD = 80f;

    /**
     * Scene change listener
     */
    public interface SceneChangeListener {
        void onSceneChanged(@NonNull SceneContext oldScene, @NonNull SceneContext newScene);
    }

    /**
     * Create scene detector
     */
    public SceneDetector(@NonNull Context context) {
        this.context = context;
        this.sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        this.executor = Executors.newSingleThreadExecutor();
        this.mainHandler = new Handler(Looper.getMainLooper());
        this.currentScene = SceneContext.unknown();

        // Get available sensors
        if (sensorManager != null) {
            accelerometer = sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER);
            gyroscope = sensorManager.getDefaultSensor(Sensor.TYPE_GYROSCOPE);
            heartRateSensor = sensorManager.getDefaultSensor(Sensor.TYPE_HEART_RATE);
            stepCounter = sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER);
        }
    }

    /**
     * Start scene detection
     */
    public void startDetection() {
        if (detecting) return;

        Log.i(TAG, "Starting scene detection...");
        detecting = true;

        // Register sensor listeners
        if (sensorManager != null) {
            if (accelerometer != null) {
                sensorManager.registerListener(this, accelerometer,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
            if (gyroscope != null) {
                sensorManager.registerListener(this, gyroscope,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
            if (heartRateSensor != null) {
                sensorManager.registerListener(this, heartRateSensor,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
            if (stepCounter != null) {
                sensorManager.registerListener(this, stepCounter,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
        }

        // Start periodic analysis
        startPeriodicAnalysis();

        Log.i(TAG, "Scene detection started");
    }

    /**
     * Stop scene detection
     */
    public void stopDetection() {
        if (!detecting) return;

        Log.i(TAG, "Stopping scene detection...");
        detecting = false;

        if (sensorManager != null) {
            sensorManager.unregisterListener(this);
        }

        executor.shutdown();

        Log.i(TAG, "Scene detection stopped");
    }

    /**
     * Get current scene
     */
    @NonNull
    public SceneContext getCurrentScene() {
        return currentScene;
    }

    /**
     * Add scene change listener
     */
    public void addListener(@NonNull SceneChangeListener listener) {
        listeners.add(listener);
    }

    /**
     * Remove scene change listener
     */
    public void removeListener(@NonNull SceneChangeListener listener) {
        listeners.remove(listener);
    }

    // ===== SensorEventListener Implementation =====

    @Override
    public void onSensorChanged(SensorEvent event) {
        if (!detecting) return;

        switch (event.sensor.getType()) {
            case Sensor.TYPE_ACCELEROMETER:
                System.arraycopy(event.values, 0, accelerometerData, 0, 3);
                break;

            case Sensor.TYPE_GYROSCOPE:
                System.arraycopy(event.values, 0, gyroscopeData, 0, 3);
                break;

            case Sensor.TYPE_HEART_RATE:
                currentHeartRate = event.values[0];
                break;

            case Sensor.TYPE_STEP_COUNTER:
                stepCount = (int) event.values[0];
                break;
        }
    }

    @Override
    public void onAccuracyChanged(Sensor sensor, int accuracy) {
        // Ignore accuracy changes
    }

    // ===== Analysis Methods =====

    private void startPeriodicAnalysis() {
        executor.execute(() -> {
            while (detecting) {
                try {
                    analyzeCurrentState();
                    Thread.sleep(DETECTION_INTERVAL_MS);
                } catch (InterruptedException e) {
                    break;
                }
            }
        });
    }

    /**
     * Analyze current sensor state and detect scene
     */
    private void analyzeCurrentState() {
        // Calculate motion intensity from accelerometer
        float motionIntensity = calculateMotionIntensity();

        // Detect activity based on motion and heart rate
        String activity = detectActivity(motionIntensity, currentHeartRate);

        // Build metadata
        Map<String, Object> metadata = new HashMap<>();
        metadata.put("motion_intensity", motionIntensity);
        metadata.put("heart_rate", currentHeartRate);
        metadata.put("step_count", stepCount);

        // Create new scene context
        SceneContext newScene = new SceneContext(
            activity,
            activityConfidence,
            System.currentTimeMillis(),
            null, // Local detection
            metadata
        );

        // Check if scene changed
        if (!newScene.getSceneType().equals(currentScene.getSceneType())) {
            notifySceneChange(currentScene, newScene);
        }

        currentScene = newScene;
    }

    /**
     * Calculate motion intensity from accelerometer data
     */
    private float calculateMotionIntensity() {
        // Calculate magnitude of acceleration vector
        float magnitude = (float) Math.sqrt(
            accelerometerData[0] * accelerometerData[0] +
            accelerometerData[1] * accelerometerData[1] +
            accelerometerData[2] * accelerometerData[2]
        );

        // Subtract gravity (approximately 9.8 m/s²)
        float dynamicAccel = Math.abs(magnitude - SensorManager.GRAVITY_EARTH);

        return dynamicAccel;
    }

    /**
     * Detect activity type based on motion and heart rate
     */
    @NonNull
    private String detectActivity(float motionIntensity, float heartRate) {
        // High heart rate + significant motion = running
        if (heartRate > RUNNING_HEART_RATE_THRESHOLD && motionIntensity > 3.0f) {
            activityConfidence = 0.85f;
            return SceneContext.RUNNING;
        }

        // Moderate heart rate + moderate motion = walking
        if (heartRate > WALKING_HEART_RATE_THRESHOLD && motionIntensity > 1.5f) {
            activityConfidence = 0.75f;
            return SceneContext.WALKING;
        }

        // Significant motion but normal heart rate = cycling (or driving)
        if (motionIntensity > 2.0f && heartRate < WALKING_HEART_RATE_THRESHOLD) {
            activityConfidence = 0.6f;
            return SceneContext.CYCLING;
        }

        // Low motion, low heart rate = resting
        if (motionIntensity < 0.5f && heartRate < 70f) {
            activityConfidence = 0.7f;
            return SceneContext.RESTING;
        }

        // Default unknown
        activityConfidence = 0.3f;
        return SceneContext.UNKNOWN;
    }

    /**
     * Notify listeners of scene change
     */
    private void notifySceneChange(@NonNull SceneContext oldScene, @NonNull SceneContext newScene) {
        Log.i(TAG, "Scene changed: " + oldScene.getSceneType() + " -> " + newScene.getSceneType());

        mainHandler.post(() -> {
            for (SceneChangeListener listener : listeners) {
                listener.onSceneChanged(oldScene, newScene);
            }
        });
    }

    /**
     * Update scene from external source (e.g., watch)
     */
    public void updateSceneFromExternal(@NonNull SceneContext externalScene) {
        if (externalScene.getConfidence() > currentScene.getConfidence()) {
            Log.i(TAG, "Updating scene from external: " + externalScene.getSceneType());
            notifySceneChange(currentScene, externalScene);
            currentScene = externalScene;
        }
    }

    /**
     * Check if specific sensors are available
     */
    public boolean hasHeartRateSensor() {
        return heartRateSensor != null;
    }

    public boolean hasAccelerometer() {
        return accelerometer != null;
    }

    public boolean hasStepCounter() {
        return stepCounter != null;
    }
}