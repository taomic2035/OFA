package com.ofa.agent.automation.advanced;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;

import org.json.JSONObject;

import java.io.File;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * Action Replay - replays recorded automation actions.
 */
public class ActionReplay {

    private static final String TAG = "ActionReplay";

    private final AutomationEngine engine;
    private final ScreenCapture screenCapture;

    private long playbackDelay = 500; // ms between actions
    private float speedMultiplier = 1.0f; // 1.0 = normal speed
    private boolean stopOnError = true;
    private boolean verifyActions = false;

    private final AtomicBoolean playing = new AtomicBoolean(false);
    private final AtomicBoolean stopped = new AtomicBoolean(false);

    // Callbacks
    private PlaybackListener playbackListener;

    public interface PlaybackListener {
        void onPlaybackStart(int totalActions);
        void onActionStart(int index, ActionRecorder.RecordedAction action);
        void onActionComplete(int index, ActionRecorder.RecordedAction action, AutomationResult result);
        void onPlaybackComplete(int successCount, int failCount);
        void onPlaybackError(int index, String error);
    }

    public ActionReplay(@NonNull AutomationEngine engine, @Nullable ScreenCapture screenCapture) {
        this.engine = engine;
        this.screenCapture = screenCapture;
    }

    /**
     * Set playback delay between actions
     */
    public void setPlaybackDelay(long delayMs) {
        this.playbackDelay = delayMs;
    }

    /**
     * Set playback speed multiplier
     * 1.0 = normal, 2.0 = double speed, 0.5 = half speed
     */
    public void setSpeedMultiplier(float multiplier) {
        this.speedMultiplier = Math.max(0.1f, Math.min(10.0f, multiplier));
    }

    /**
     * Set whether to stop on error
     */
    public void setStopOnError(boolean stop) {
        this.stopOnError = stop;
    }

    /**
     * Set whether to verify actions before execution
     */
    public void setVerifyActions(boolean verify) {
        this.verifyActions = verify;
    }

    /**
     * Set playback listener
     */
    public void setPlaybackListener(@Nullable PlaybackListener listener) {
        this.playbackListener = listener;
    }

    /**
     * Check if currently playing
     */
    public boolean isPlaying() {
        return playing.get();
    }

    /**
     * Stop current playback
     */
    public void stop() {
        stopped.set(true);
    }

    /**
     * Play recorded actions
     */
    @NonNull
    public PlaybackResult play(@NonNull List<ActionRecorder.RecordedAction> actions) {
        return play(actions, false);
    }

    /**
     * Play recorded actions
     * @param actions List of recorded actions
     * @param respectTiming Whether to respect original timing
     */
    @NonNull
    public PlaybackResult play(@NonNull List<ActionRecorder.RecordedAction> actions, boolean respectTiming) {
        if (actions.isEmpty()) {
            return new PlaybackResult(0, 0, "No actions to play");
        }

        if (!playing.compareAndSet(false, true)) {
            return new PlaybackResult(0, 0, "Already playing");
        }

        stopped.set(false);
        int successCount = 0;
        int failCount = 0;

        Log.i(TAG, "Starting playback of " + actions.size() + " actions");

        if (playbackListener != null) {
            playbackListener.onPlaybackStart(actions.size());
        }

        long lastActionTime = 0;

        for (int i = 0; i < actions.size(); i++) {
            if (stopped.get()) {
                Log.i(TAG, "Playback stopped at action " + i);
                break;
            }

            ActionRecorder.RecordedAction action = actions.get(i);

            // Skip non-action events
            if (action.actionType.endsWith("_start") || action.actionType.equals("page_change")) {
                continue;
            }

            // Respect timing if requested
            if (respectTiming && lastActionTime > 0) {
                long timeDiff = (long) ((action.relativeTime - lastActionTime) / speedMultiplier);
                if (timeDiff > 0) {
                    sleep(timeDiff);
                }
            } else {
                // Use fixed delay
                sleep((long) (playbackDelay / speedMultiplier));
            }
            lastActionTime = action.relativeTime;

            // Notify start
            if (playbackListener != null) {
                playbackListener.onActionStart(i, action);
            }

            // Execute action
            AutomationResult result = executeAction(action);

            // Notify complete
            if (playbackListener != null) {
                playbackListener.onActionComplete(i, action, result);
            }

            if (result.isSuccess()) {
                successCount++;
            } else {
                failCount++;

                if (playbackListener != null) {
                    playbackListener.onPlaybackError(i, result.getError());
                }

                if (stopOnError) {
                    Log.w(TAG, "Stopping playback due to error at action " + i);
                    break;
                }
            }
        }

        playing.set(false);

        Log.i(TAG, "Playback complete: " + successCount + " success, " + failCount + " failed");

        if (playbackListener != null) {
            playbackListener.onPlaybackComplete(successCount, failCount);
        }

        return new PlaybackResult(successCount, failCount, null);
    }

    /**
     * Play from file
     */
    @NonNull
    public PlaybackResult playFromFile(@NonNull File file, boolean respectTiming) {
        List<ActionRecorder.RecordedAction> actions = ActionRecorder.loadRecording(file);
        if (actions == null || actions.isEmpty()) {
            return new PlaybackResult(0, 0, "Failed to load recording");
        }
        return play(actions, respectTiming);
    }

    /**
     * Execute a single action
     */
    @NonNull
    private AutomationResult executeAction(@NonNull ActionRecorder.RecordedAction action) {
        String actionType = action.actionType;
        JSONObject params = action.params;

        try {
            // Handle complete actions
            if (actionType.endsWith("_complete")) {
                actionType = actionType.substring(0, actionType.length() - 9); // Remove "_complete"
            }

            // Gesture actions
            if (actionType.startsWith("gesture_")) {
                return executeGestureAction(actionType, params);
            }

            // Click actions
            if (actionType.equals("click_coordinate") || actionType.equals("click")) {
                int x = params.optInt("x", 0);
                int y = params.optInt("y", 0);
                return engine.click(x, y);
            }

            if (actionType.equals("click_element")) {
                String text = params.optString("text", null);
                if (text != null) {
                    return engine.click(text);
                }
                return new AutomationResult("click_element", "No text provided");
            }

            // Long click
            if (actionType.equals("longClick_coordinate") || actionType.equals("longClick")) {
                int x = params.optInt("x", 0);
                int y = params.optInt("y", 0);
                return engine.longClick(x, y);
            }

            // Swipe
            if (actionType.equals("swipe")) {
                String directionStr = params.optString("direction", null);
                if (directionStr != null) {
                    Direction direction = Direction.fromString(directionStr);
                    return engine.swipe(direction, 0);
                }

                int fromX = params.optInt("fromX", 0);
                int fromY = params.optInt("fromY", 0);
                int toX = params.optInt("toX", 0);
                int toY = params.optInt("toY", 0);
                long duration = params.optLong("duration", 300);

                return engine.swipe(fromX, fromY, toX, toY, duration);
            }

            // Input
            if (actionType.equals("input_text") || actionType.equals("input")) {
                String text = params.optString("text", "");
                int x = params.optInt("x", -1);
                int y = params.optInt("y", -1);

                if (x >= 0 && y >= 0) {
                    return engine.inputText(x, y, text);
                }
                return engine.inputText(text);
            }

            // Scroll find
            if (actionType.equals("scroll_find")) {
                String text = params.optString("text", null);
                if (text != null) {
                    return engine.scrollFind(text, 10);
                }
                return new AutomationResult("scroll_find", "No text provided");
            }

            // Wait for element
            if (actionType.equals("wait_for_element")) {
                String text = params.optString("text", null);
                long timeout = params.optLong("timeout", 30000);
                if (text != null) {
                    return engine.waitForElement(BySelector.text(text), timeout);
                }
                return new AutomationResult("wait_for_element", "No text provided");
            }

            // Unknown action
            Log.w(TAG, "Unknown action type: " + actionType);
            return new AutomationResult(actionType, "Unknown action type");

        } catch (Exception e) {
            Log.e(TAG, "Error executing action: " + actionType, e);
            return new AutomationResult(actionType, "Error: " + e.getMessage());
        }
    }

    /**
     * Execute gesture action
     */
    @NonNull
    private AutomationResult executeGestureAction(@NonNull String actionType, @NonNull JSONObject params) {
        String gestureType = actionType.substring(8); // Remove "gesture_"
        int x = params.optInt("x", 0);
        int y = params.optInt("y", 0);

        switch (gestureType) {
            case "click":
            case "tap":
                return engine.click(x, y);
            case "longClick":
            case "longPress":
                return engine.longClick(x, y);
            case "swipe":
                String direction = params.optString("direction", "down");
                return engine.swipe(Direction.fromString(direction), 0);
            default:
                return engine.click(x, y);
        }
    }

    private void sleep(long ms) {
        try {
            Thread.sleep(ms);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * Playback result
     */
    public static class PlaybackResult {
        public final int successCount;
        public final int failCount;
        public final String error;

        public PlaybackResult(int successCount, int failCount, String error) {
            this.successCount = successCount;
            this.failCount = failCount;
            this.error = error;
        }

        public boolean isSuccess() {
            return error == null || failCount == 0;
        }
    }
}