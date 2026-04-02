package com.ofa.agent.automation.advanced;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationListener;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Locale;

/**
 * Action Recorder - records and replays UI automation actions.
 * Supports recording user interactions and playing them back.
 */
public class ActionRecorder implements AutomationListener {

    private static final String TAG = "ActionRecorder";

    private final AutomationEngine engine;
    private final List<RecordedAction> recordedActions = new ArrayList<>();
    private final ScreenCapture screenCapture;

    private volatile boolean recording = false;
    private volatile boolean paused = false;
    private String recordingName;
    private long recordingStartTime;
    private File recordingDirectory;

    // Configuration
    private boolean captureScreenshots = false;
    private boolean captureOnAction = true;
    private long defaultPlaybackDelay = 500; // ms between actions during playback

    /**
     * Recorded action data class
     */
    public static class RecordedAction {
        public final long timestamp;
        public final long relativeTime; // ms from start of recording
        public final String actionType;
        public final JSONObject params;
        public String screenshotPath;

        public RecordedAction(long timestamp, long relativeTime, String actionType, JSONObject params) {
            this.timestamp = timestamp;
            this.relativeTime = relativeTime;
            this.actionType = actionType;
            this.params = params;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("timestamp", timestamp);
                json.put("relativeTime", relativeTime);
                json.put("actionType", actionType);
                json.put("params", params);
                if (screenshotPath != null) {
                    json.put("screenshot", screenshotPath);
                }
            } catch (Exception e) {
                // Ignore
            }
            return json;
        }

        @Nullable
        public static RecordedAction fromJson(@NonNull JSONObject json) {
            try {
                long timestamp = json.getLong("timestamp");
                long relativeTime = json.getLong("relativeTime");
                String actionType = json.getString("actionType");
                JSONObject params = json.getJSONObject("params");

                RecordedAction action = new RecordedAction(timestamp, relativeTime, actionType, params);
                action.screenshotPath = json.optString("screenshot", null);
                return action;
            } catch (Exception e) {
                return null;
            }
        }
    }

    public ActionRecorder(@NonNull AutomationEngine engine, @Nullable ScreenCapture screenCapture) {
        this.engine = engine;
        this.screenCapture = screenCapture;
    }

    /**
     * Set recording directory
     */
    public void setRecordingDirectory(@NonNull File directory) {
        this.recordingDirectory = directory;
        if (!directory.exists()) {
            directory.mkdirs();
        }
    }

    /**
     * Enable/disable screenshot capture during recording
     */
    public void setCaptureScreenshots(boolean enabled) {
        this.captureScreenshots = enabled;
    }

    /**
     * Set delay between actions during playback
     */
    public void setPlaybackDelay(long delayMs) {
        this.defaultPlaybackDelay = delayMs;
    }

    /**
     * Start recording
     */
    public void startRecording(@Nullable String name) {
        if (recording) {
            Log.w(TAG, "Already recording");
            return;
        }

        recordingName = name != null ? name :
            "recording_" + new SimpleDateFormat("yyyyMMdd_HHmmss", Locale.getDefault()).format(new Date());

        recordedActions.clear();
        recordingStartTime = System.currentTimeMillis();
        recording = true;
        paused = false;

        engine.setListener(this);

        Log.i(TAG, "Started recording: " + recordingName);
    }

    /**
     * Start recording with default name
     */
    public void startRecording() {
        startRecording(null);
    }

    /**
     * Pause recording
     */
    public void pauseRecording() {
        if (recording && !paused) {
            paused = true;
            Log.i(TAG, "Recording paused");
        }
    }

    /**
     * Resume recording
     */
    public void resumeRecording() {
        if (recording && paused) {
            paused = false;
            Log.i(TAG, "Recording resumed");
        }
    }

    /**
     * Stop recording
     */
    @NonNull
    public List<RecordedAction> stopRecording() {
        if (!recording) {
            Log.w(TAG, "Not recording");
            return new ArrayList<>();
        }

        recording = false;
        paused = false;
        engine.setListener(null);

        Log.i(TAG, "Stopped recording: " + recordedActions.size() + " actions");
        return new ArrayList<>(recordedActions);
    }

    /**
     * Check if currently recording
     */
    public boolean isRecording() {
        return recording;
    }

    /**
     * Check if recording is paused
     */
    public boolean isPaused() {
        return paused;
    }

    /**
     * Get recorded actions
     */
    @NonNull
    public List<RecordedAction> getRecordedActions() {
        return new ArrayList<>(recordedActions);
    }

    /**
     * Get number of recorded actions
     */
    public int getActionCount() {
        return recordedActions.size();
    }

    /**
     * Save recording to file
     */
    @Nullable
    public String saveRecording() {
        if (recordedActions.isEmpty()) {
            Log.w(TAG, "No actions to save");
            return null;
        }

        try {
            File dir = recordingDirectory;
            if (dir == null) {
                dir = new File("/sdcard/ofa/recordings");
            }
            if (!dir.exists()) {
                dir.mkdirs();
            }

            File file = new File(dir, recordingName + ".json");

            JSONObject recording = new JSONObject();
            recording.put("name", recordingName);
            recording.put("startTime", recordingStartTime);
            recording.put("actionCount", recordedActions.size());

            JSONArray actions = new JSONArray();
            for (RecordedAction action : recordedActions) {
                actions.put(action.toJson());
            }
            recording.put("actions", actions);

            try (FileOutputStream fos = new FileOutputStream(file)) {
                fos.write(recording.toString(2).getBytes());
            }

            Log.i(TAG, "Recording saved to: " + file.getAbsolutePath());
            return file.getAbsolutePath();

        } catch (Exception e) {
            Log.e(TAG, "Error saving recording", e);
            return null;
        }
    }

    /**
     * Load recording from file
     */
    @Nullable
    public static List<RecordedAction> loadRecording(@NonNull File file) {
        try {
            StringBuilder content = new StringBuilder();
            try (FileInputStream fis = new FileInputStream(file)) {
                byte[] buffer = new byte[1024];
                int len;
                while ((len = fis.read(buffer)) > 0) {
                    content.append(new String(buffer, 0, len));
                }
            }

            JSONObject recording = new JSONObject(content.toString());
            JSONArray actions = recording.getJSONArray("actions");

            List<RecordedAction> result = new ArrayList<>();
            for (int i = 0; i < actions.length(); i++) {
                RecordedAction action = RecordedAction.fromJson(actions.getJSONObject(i));
                if (action != null) {
                    result.add(action);
                }
            }

            Log.i(TAG, "Loaded " + result.size() + " actions from " + file.getName());
            return result;

        } catch (Exception e) {
            Log.e(TAG, "Error loading recording", e);
            return null;
        }
    }

    // ===== AutomationListener Implementation =====

    @Override
    public void onOperationStart(@NonNull String operation, @Nullable String target) {
        if (!recording || paused) return;

        JSONObject params = new JSONObject();
        try {
            params.put("target", target);
        } catch (Exception e) {}

        long now = System.currentTimeMillis();
        RecordedAction action = new RecordedAction(
            now,
            now - recordingStartTime,
            operation + "_start",
            params
        );

        recordedActions.add(action);
        Log.d(TAG, "Recorded: " + operation + "_start");

        // Capture screenshot if enabled
        if (captureScreenshots && screenCapture != null && captureOnAction) {
            action.screenshotPath = screenCapture.captureToDefaultLocation();
        }
    }

    @Override
    public void onOperationComplete(@NonNull String operation, @NonNull AutomationResult result) {
        if (!recording || paused) return;

        JSONObject params = new JSONObject();
        try {
            params.put("success", result.isSuccess());
            params.put("executionTimeMs", result.getExecutionTimeMs());
            if (result.getFoundNode() != null) {
                params.put("node", result.getFoundNode().toJson());
            }
        } catch (Exception e) {}

        long now = System.currentTimeMillis();
        RecordedAction action = new RecordedAction(
            now,
            now - recordingStartTime,
            operation + "_complete",
            params
        );

        recordedActions.add(action);
        Log.d(TAG, "Recorded: " + operation + "_complete");
    }

    @Override
    public void onGesturePerformed(@NonNull String gestureType, int x, int y) {
        if (!recording || paused) return;

        JSONObject params = new JSONObject();
        try {
            params.put("gestureType", gestureType);
            params.put("x", x);
            params.put("y", y);
        } catch (Exception e) {}

        long now = System.currentTimeMillis();
        RecordedAction action = new RecordedAction(
            now,
            now - recordingStartTime,
            "gesture_" + gestureType,
            params
        );

        recordedActions.add(action);
        Log.d(TAG, "Recorded gesture: " + gestureType + " at (" + x + ", " + y + ")");
    }

    @Override
    public void onPageChange(@Nullable String packageName, @Nullable String activityName) {
        if (!recording || paused) return;

        JSONObject params = new JSONObject();
        try {
            params.put("packageName", packageName);
            params.put("activityName", activityName);
        } catch (Exception e) {}

        long now = System.currentTimeMillis();
        RecordedAction action = new RecordedAction(
            now,
            now - recordingStartTime,
            "page_change",
            params
        );

        recordedActions.add(action);
        Log.d(TAG, "Recorded page change: " + packageName);
    }

    // Unused listener methods
    @Override
    public void onEngineAvailable(@NonNull com.ofa.agent.automation.AutomationCapability capability) {}
    @Override
    public void onEngineUnavailable(@NonNull String reason) {}
    @Override
    public void onOperationError(@NonNull String operation, @NonNull String error, boolean willRetry) {}
    @Override
    public void onElementFound(@NonNull BySelector selector, @NonNull AutomationNode node) {}
    @Override
    public void onElementNotFound(@NonNull BySelector selector, boolean timedOut) {}
    @Override
    public void onAccessibilityServiceStateChanged(boolean enabled) {}
    @Override
    public void onScreenshotCaptured(@Nullable String screenshotPath) {}
}