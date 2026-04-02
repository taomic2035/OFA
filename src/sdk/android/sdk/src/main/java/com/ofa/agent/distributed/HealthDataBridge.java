package com.ofa.agent.distributed;

import android.content.Context;
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import com.ofa.agent.core.AgentProfile;
import com.ofa.agent.core.PeerNetwork;

/**
 * Health Data Bridge - bridges health data from wearables to phone.
 *
 * Health Data Types:
 * - Heart rate (实时心率)
 * - Temperature (体温)
 * - Blood oxygen (血氧)
 * - Steps (步数)
 * - Calories (卡路里)
 * - Distance (距离)
 * - Sleep quality (睡眠质量)
 * - Stress level (压力水平)
 *
 * Alert Thresholds:
 * - Heart rate: >120 bpm (running), >180 bpm (dangerous), <50 bpm (abnormal)
 * - Temperature: >37.5°C (fever), <35°C (hypothermia)
 * - Blood oxygen: <95% (abnormal), <90% (dangerous)
 *
 * The bridge:
 * 1. Collects local health data (if available)
 * 2. Receives health data from wearable via EventBus
 * 3. Monitors for anomalies and triggers alerts
 * 4. Stores health history for analysis
 */
public class HealthDataBridge implements SensorEventListener {

    private static final String TAG = "HealthDataBridge";

    private final Context context;
    private final AgentProfile localProfile;
    private final EventBus eventBus;
    private final SensorManager sensorManager;

    // Sensors
    private Sensor heartRateSensor;
    private Sensor temperatureSensor;
    private Sensor bloodOxygenSensor;
    private Sensor stepCounter;

    // Current readings
    private float currentHeartRate = 0;
    private float currentTemperature = 0;
    private float currentBloodOxygen = 0;
    private int currentSteps = 0;

    // Alert thresholds
    private static final float HEART_RATE_HIGH_DANGER = 180f;
    private static final float HEART_RATE_HIGH_WARNING = 120f;
    private static final float HEART_RATE_LOW_WARNING = 50f;
    private static final float TEMPERATURE_HIGH_WARNING = 37.5f;
    private static final float TEMPERATURE_LOW_WARNING = 35f;
    private static final float BLOOD_OXYGEN_WARNING = 95f;
    private static final float BLOOD_OXYGEN_DANGER = 90f;

    // Alert cooldown (prevent repeated alerts)
    private static final long ALERT_COOLDOWN_MS = 60000; // 1 minute
    private long lastHeartRateAlertTime = 0;
    private long lastTemperatureAlertTime = 0;
    private long lastBloodOxygenAlertTime = 0;

    // Health history
    private final List<HealthRecord> healthHistory = new ArrayList<>();
    private static final int MAX_HISTORY_SIZE = 1000;

    // Listeners
    private final List<HealthAlertListener> alertListeners = new ArrayList<>();

    // Executor
    private final ExecutorService executor;

    // State
    private volatile boolean monitoring = false;

    /**
     * Health record
     */
    public static class HealthRecord {
        public final String dataType;
        public final float value;
        public final long timestamp;
        public final String sourceDeviceId;
        public final Map<String, Object> metadata;

        public HealthRecord(String dataType, float value, long timestamp,
                            String sourceDeviceId, Map<String, Object> metadata) {
            this.dataType = dataType;
            this.value = value;
            this.timestamp = timestamp;
            this.sourceDeviceId = sourceDeviceId;
            this.metadata = new HashMap<>(metadata);
        }
    }

    /**
     * Health alert
     */
    public static class HealthAlert {
        public final String alertType;
        public final float value;
        public final float threshold;
        public final String severity;
        public final String recommendation;
        public final long timestamp;
        public final String sourceDeviceId;

        public HealthAlert(String alertType, float value, float threshold,
                           String severity, String recommendation,
                           long timestamp, String sourceDeviceId) {
            this.alertType = alertType;
            this.value = value;
            this.threshold = threshold;
            this.severity = severity;
            this.recommendation = recommendation;
            this.timestamp = timestamp;
            this.sourceDeviceId = sourceDeviceId;
        }
    }

    /**
     * Health alert listener
     */
    public interface HealthAlertListener {
        void onHealthAlert(@NonNull HealthAlert alert);
    }

    /**
     * Create health data bridge
     */
    public HealthDataBridge(@NonNull Context context, @NonNull AgentProfile localProfile,
                            @NonNull EventBus eventBus) {
        this.context = context;
        this.localProfile = localProfile;
        this.eventBus = eventBus;
        this.sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        this.executor = Executors.newSingleThreadExecutor();

        // Get available sensors
        if (sensorManager != null) {
            heartRateSensor = sensorManager.getDefaultSensor(Sensor.TYPE_HEART_RATE);
            temperatureSensor = sensorManager.getDefaultSensor(Sensor.TYPE_AMBIENT_TEMPERATURE);
            // Note: Blood oxygen sensor may not be available on all devices
            stepCounter = sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER);
        }

        // Subscribe to health events from EventBus
        eventBus.subscribe(EventBus.HEALTH_DATA, this::handleHealthDataEvent);
        eventBus.subscribe(EventBus.HEALTH_ALERT, this::handleHealthAlertEvent);
    }

    /**
     * Start monitoring health data
     */
    public void startMonitoring() {
        if (monitoring) return;

        Log.i(TAG, "Starting health monitoring...");
        monitoring = true;

        // Register local sensors if available
        if (sensorManager != null) {
            if (heartRateSensor != null) {
                sensorManager.registerListener(this, heartRateSensor,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
            if (temperatureSensor != null) {
                sensorManager.registerListener(this, temperatureSensor,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
            if (stepCounter != null) {
                sensorManager.registerListener(this, stepCounter,
                    SensorManager.SENSOR_DELAY_NORMAL);
            }
        }

        Log.i(TAG, "Health monitoring started");
    }

    /**
     * Stop monitoring
     */
    public void stopMonitoring() {
        if (!monitoring) return;

        Log.i(TAG, "Stopping health monitoring...");
        monitoring = false;

        if (sensorManager != null) {
            sensorManager.unregisterListener(this);
        }

        executor.shutdown();

        Log.i(TAG, "Health monitoring stopped");
    }

    /**
     * Get current health readings
     */
    @NonNull
    public Map<String, Float> getCurrentReadings() {
        Map<String, Float> readings = new HashMap<>();
        if (currentHeartRate > 0) readings.put("heart_rate", currentHeartRate);
        if (currentTemperature > 0) readings.put("temperature", currentTemperature);
        if (currentBloodOxygen > 0) readings.put("blood_oxygen", currentBloodOxygen);
        readings.put("steps", (float) currentSteps);
        return readings;
    }

    /**
     * Add health alert listener
     */
    public void addAlertListener(@NonNull HealthAlertListener listener) {
        alertListeners.add(listener);
    }

    /**
     * Remove health alert listener
     */
    public void removeAlertListener(@NonNull HealthAlertListener listener) {
        alertListeners.remove(listener);
    }

    /**
     * Get health history
     */
    @NonNull
    public List<HealthRecord> getHealthHistory(String dataType, int limit) {
        List<HealthRecord> result = new ArrayList<>();
        for (int i = healthHistory.size() - 1; i >= 0 && result.size() < limit; i--) {
            HealthRecord record = healthHistory.get(i);
            if (dataType == null || record.dataType.equals(dataType)) {
                result.add(record);
            }
        }
        return result;
    }

    // ===== SensorEventListener =====

    @Override
    public void onSensorChanged(SensorEvent event) {
        if (!monitoring) return;

        switch (event.sensor.getType()) {
            case Sensor.TYPE_HEART_RATE:
                currentHeartRate = event.values[0];
                processHealthData("heart_rate", currentHeartRate);
                break;

            case Sensor.TYPE_AMBIENT_TEMPERATURE:
                currentTemperature = event.values[0];
                // Note: Ambient temperature is different from body temperature
                // Would need proper body temperature sensor
                break;

            case Sensor.TYPE_STEP_COUNTER:
                currentSteps = (int) event.values[0];
                processHealthData("steps", currentSteps);
                break;
        }
    }

    @Override
    public void onAccuracyChanged(Sensor sensor, int accuracy) {
        // Ignore
    }

    // ===== Health Data Processing =====

    /**
     * Process health data from local sensors
     */
    private void processHealthData(String dataType, float value) {
        // Store in history
        HealthRecord record = new HealthRecord(
            dataType, value, System.currentTimeMillis(),
            localProfile.getAgentId(), new HashMap<>()
        );
        addToHistory(record);

        // Publish to event bus
        eventBus.publishHealthData(dataType, value, null);

        // Check for alerts
        checkForAlerts(dataType, value, localProfile.getAgentId());
    }

    /**
     * Handle health data event from EventBus (from other devices)
     */
    private void handleHealthDataEvent(@NonNull EventBus.Event event) {
        String dataType = (String) event.data.get("data_type");
        Float value = (Float) event.data.get("value");
        String sourceDevice = event.sourceDeviceId;

        if (dataType != null && value != null) {
            // Update current readings based on source
            switch (dataType) {
                case "heart_rate":
                    currentHeartRate = value;
                    break;
                case "temperature":
                    currentTemperature = value;
                    break;
                case "blood_oxygen":
                    currentBloodOxygen = value;
                    break;
            }

            // Store in history
            HealthRecord record = new HealthRecord(
                dataType, value, event.timestamp, sourceDevice, event.data
            );
            addToHistory(record);

            // Check for alerts
            checkForAlerts(dataType, value, sourceDevice);

            Log.d(TAG, "Health data from " + sourceDevice + ": " + dataType + "=" + value);
        }
    }

    /**
     * Handle health alert event from EventBus
     */
    private void handleHealthAlertEvent(@NonNull EventBus.Event event) {
        String alertType = (String) event.data.get("alert_type");
        Float value = (Float) event.data.get("value");
        Float threshold = (Float) event.data.get("threshold");
        String severity = (String) event.data.get("severity");
        String recommendation = (String) event.data.get("recommendation");

        if (alertType != null && value != null) {
            HealthAlert alert = new HealthAlert(
                alertType, value, threshold != null ? threshold : 0,
                severity != null ? severity : "unknown",
                recommendation,
                event.timestamp, event.sourceDeviceId
            );

            notifyAlertListeners(alert);
            Log.i(TAG, "Health alert: " + alertType + "=" + value + " (" + severity + ")");
        }
    }

    /**
     * Check for health anomalies and trigger alerts
     */
    private void checkForAlerts(String dataType, float value, String sourceDevice) {
        long now = System.currentTimeMillis();

        // Heart rate alerts
        if (dataType.equals("heart_rate")) {
            if (value > HEART_RATE_HIGH_DANGER && now - lastHeartRateAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("heart_rate_high_danger", value, HEART_RATE_HIGH_DANGER,
                    "critical", "心率过高！请立即停止运动并休息", sourceDevice);
                lastHeartRateAlertTime = now;
            } else if (value > HEART_RATE_HIGH_WARNING && now - lastHeartRateAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("heart_rate_high", value, HEART_RATE_HIGH_WARNING,
                    "warning", "心率偏高，建议适当放缓", sourceDevice);
                lastHeartRateAlertTime = now;
            } else if (value < HEART_RATE_LOW_WARNING && now - lastHeartRateAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("heart_rate_low", value, HEART_RATE_LOW_WARNING,
                    "warning", "心率偏低，如有不适请就医", sourceDevice);
                lastHeartRateAlertTime = now;
            }
        }

        // Temperature alerts
        if (dataType.equals("temperature")) {
            if (value > TEMPERATURE_HIGH_WARNING && now - lastTemperatureAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("temperature_high", value, TEMPERATURE_HIGH_WARNING,
                    "warning", "体温偏高，可能发烧，建议测量确认", sourceDevice);
                lastTemperatureAlertTime = now;
            } else if (value < TEMPERATURE_LOW_WARNING && now - lastTemperatureAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("temperature_low", value, TEMPERATURE_LOW_WARNING,
                    "warning", "体温偏低，请注意保暖", sourceDevice);
                lastTemperatureAlertTime = now;
            }
        }

        // Blood oxygen alerts
        if (dataType.equals("blood_oxygen")) {
            if (value < BLOOD_OXYGEN_DANGER && now - lastBloodOxygenAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("blood_oxygen_low_danger", value, BLOOD_OXYGEN_DANGER,
                    "critical", "血氧过低！请立即停止运动并休息", sourceDevice);
                lastBloodOxygenAlertTime = now;
            } else if (value < BLOOD_OXYGEN_WARNING && now - lastBloodOxygenAlertTime > ALERT_COOLDOWN_MS) {
                triggerAlert("blood_oxygen_low", value, BLOOD_OXYGEN_WARNING,
                    "warning", "血氧偏低，建议适当休息", sourceDevice);
                lastBloodOxygenAlertTime = now;
            }
        }
    }

    /**
     * Trigger a health alert
     */
    private void triggerAlert(String alertType, float value, float threshold,
                              String severity, String recommendation, String sourceDevice) {
        HealthAlert alert = new HealthAlert(
            alertType, value, threshold, severity, recommendation,
            System.currentTimeMillis(), sourceDevice
        );

        // Publish to event bus
        eventBus.publishHealthAlert(alertType, value, threshold, recommendation);

        // Notify local listeners
        notifyAlertListeners(alert);

        Log.w(TAG, "Health alert triggered: " + alertType);
    }

    /**
     * Notify alert listeners
     */
    private void notifyAlertListeners(@NonNull HealthAlert alert) {
        for (HealthAlertListener listener : alertListeners) {
            try {
                listener.onHealthAlert(alert);
            } catch (Exception e) {
                Log.w(TAG, "Alert listener error: " + e.getMessage());
            }
        }
    }

    /**
     * Add to health history
     */
    private void addToHistory(@NonNull HealthRecord record) {
        healthHistory.add(record);
        if (healthHistory.size() > MAX_HISTORY_SIZE) {
            healthHistory.remove(0);
        }
    }

    /**
     * Check if local sensors are available
     */
    public boolean hasHeartRateSensor() {
        return heartRateSensor != null;
    }

    public boolean hasTemperatureSensor() {
        return temperatureSensor != null;
    }

    public boolean hasStepCounter() {
        return stepCounter != null;
    }

    /**
     * Request health data from wearable
     */
    public void requestHealthDataFromWearable(@NonNull String wearableId) {
        // Would send a request via peer network
        Log.i(TAG, "Requesting health data from: " + wearableId);
    }
}