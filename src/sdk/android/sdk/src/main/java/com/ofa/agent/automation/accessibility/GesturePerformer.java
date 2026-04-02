package com.ofa.agent.automation.accessibility;

import android.accessibilityservice.AccessibilityService;
import android.accessibilityservice.GestureDescription;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.graphics.Path;
import android.graphics.PointF;
import android.os.Build;
import android.util.Log;
import android.view.accessibility.AccessibilityNodeInfo;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationConfig;
import com.ofa.agent.automation.AutomationResult;

import org.json.JSONObject;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * Gesture performer - executes gestures through AccessibilityService.
 */
public class GesturePerformer {

    private static final String TAG = "GesturePerformer";

    private final AccessibilityEngine engine;
    private final ClipboardManager clipboardManager;

    public GesturePerformer(@NonNull AccessibilityEngine engine) {
        this.engine = engine;
        this.clipboardManager = (ClipboardManager)
                engine.getContext().getSystemService(Context.CLIPBOARD_SERVICE);
    }

    /**
     * Perform a click gesture at coordinates
     */
    @NonNull
    public AutomationResult performClick(int x, int y) {
        return performClick(x, y, AutomationConfig.DEFAULT_CLICK_DURATION);
    }

    /**
     * Perform a click gesture with custom duration
     */
    @NonNull
    public AutomationResult performClick(int x, int y, long duration) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return new AutomationResult("click", "Accessibility service not available");
        }

        if (!service.canPerformGestures()) {
            Log.w(TAG, "Service cannot perform gestures");
            return new AutomationResult("click", "Gesture capability not enabled");
        }

        // Create gesture
        Path path = new Path();
        path.moveTo(x, y);

        GestureDescription.StrokeDescription stroke =
                new GestureDescription.StrokeDescription(path, 0, duration);

        GestureDescription.Builder builder = new GestureDescription.Builder();
        builder.addStroke(stroke);

        GestureDescription gesture = builder.build();

        // Execute gesture
        long startTime = System.currentTimeMillis();
        AtomicBoolean success = new AtomicBoolean(false);
        CountDownLatch latch = new CountDownLatch(1);

        service.dispatchGesture(gesture, new AccessibilityService.GestureResultCallback() {
            @Override
            public void onCompleted(GestureDescription gestureDescription) {
                success.set(true);
                latch.countDown();
                Log.d(TAG, "Click gesture completed at (" + x + ", " + y + ")");
            }

            @Override
            public void onCancelled(GestureDescription gestureDescription) {
                success.set(false);
                latch.countDown();
                Log.w(TAG, "Click gesture cancelled at (" + x + ", " + y + ")");
            }
        }, null);

        // Wait for completion
        long timeout = engine.getConfig() != null ?
                engine.getConfig().getClickTimeout() : AutomationConfig.DEFAULT_CLICK_TIMEOUT;
        try {
            boolean completed = latch.await(timeout, TimeUnit.MILLISECONDS);
            if (!completed) {
                return new AutomationResult("click", "Gesture timed out");
            }
        } catch (InterruptedException e) {
            return new AutomationResult("click", "Gesture interrupted");
        }

        long elapsed = System.currentTimeMillis() - startTime;

        if (success.get()) {
            JSONObject data = new JSONObject();
            try {
                data.put("x", x);
                data.put("y", y);
                data.put("duration", duration);
            } catch (Exception e) {}
            return new AutomationResult("click", data, elapsed);
        } else {
            return new AutomationResult("click", "Gesture failed", elapsed);
        }
    }

    /**
     * Perform a long click gesture
     */
    @NonNull
    public AutomationResult performLongClick(int x, int y, long duration) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return new AutomationResult("longClick", "Accessibility service not available");
        }

        if (!service.canPerformGestures()) {
            return new AutomationResult("longClick", "Gesture capability not enabled");
        }

        // Create gesture
        Path path = new Path();
        path.moveTo(x, y);

        GestureDescription.StrokeDescription stroke =
                new GestureDescription.StrokeDescription(path, 0, duration);

        GestureDescription.Builder builder = new GestureDescription.Builder();
        builder.addStroke(stroke);

        GestureDescription gesture = builder.build();

        // Execute
        long startTime = System.currentTimeMillis();
        AtomicBoolean success = new AtomicBoolean(false);
        CountDownLatch latch = new CountDownLatch(1);

        service.dispatchGesture(gesture, new AccessibilityService.GestureResultCallback() {
            @Override
            public void onCompleted(GestureDescription gestureDescription) {
                success.set(true);
                latch.countDown();
                Log.d(TAG, "Long click gesture completed");
            }

            @Override
            public void onCancelled(GestureDescription gestureDescription) {
                success.set(false);
                latch.countDown();
                Log.w(TAG, "Long click gesture cancelled");
            }
        }, null);

        long timeout = duration + 5000; // Extra time for long click
        try {
            boolean completed = latch.await(timeout, TimeUnit.MILLISECONDS);
            if (!completed) {
                return new AutomationResult("longClick", "Gesture timed out");
            }
        } catch (InterruptedException e) {
            return new AutomationResult("longClick", "Gesture interrupted");
        }

        long elapsed = System.currentTimeMillis() - startTime;

        if (success.get()) {
            JSONObject data = new JSONObject();
            try {
                data.put("x", x);
                data.put("y", y);
                data.put("duration", duration);
            } catch (Exception e) {}
            return new AutomationResult("longClick", data, elapsed);
        } else {
            return new AutomationResult("longClick", "Gesture failed", elapsed);
        }
    }

    /**
     * Perform a swipe gesture
     */
    @NonNull
    public AutomationResult performSwipe(int fromX, int fromY, int toX, int toY, long duration) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return new AutomationResult("swipe", "Accessibility service not available");
        }

        if (!service.canPerformGestures()) {
            return new AutomationResult("swipe", "Gesture capability not enabled");
        }

        // Create swipe path
        Path path = new Path();
        path.moveTo(fromX, fromY);
        path.lineTo(toX, toY);

        GestureDescription.StrokeDescription stroke =
                new GestureDescription.StrokeDescription(path, 0, duration);

        GestureDescription.Builder builder = new GestureDescription.Builder();
        builder.addStroke(stroke);

        GestureDescription gesture = builder.build();

        // Execute
        long startTime = System.currentTimeMillis();
        AtomicBoolean success = new AtomicBoolean(false);
        CountDownLatch latch = new CountDownLatch(1);

        service.dispatchGesture(gesture, new AccessibilityService.GestureResultCallback() {
            @Override
            public void onCompleted(GestureDescription gestureDescription) {
                success.set(true);
                latch.countDown();
                Log.d(TAG, "Swipe gesture completed: (" + fromX + "," + fromY + ") -> (" + toX + "," + toY + ")");
            }

            @Override
            public void onCancelled(GestureDescription gestureDescription) {
                success.set(false);
                latch.countDown();
                Log.w(TAG, "Swipe gesture cancelled");
            }
        }, null);

        long timeout = engine.getConfig() != null ?
                engine.getConfig().getSwipeTimeout() : AutomationConfig.DEFAULT_SWIPE_TIMEOUT;
        try {
            boolean completed = latch.await(duration + timeout, TimeUnit.MILLISECONDS);
            if (!completed) {
                return new AutomationResult("swipe", "Gesture timed out");
            }
        } catch (InterruptedException e) {
            return new AutomationResult("swipe", "Gesture interrupted");
        }

        long elapsed = System.currentTimeMillis() - startTime;

        if (success.get()) {
            JSONObject data = new JSONObject();
            try {
                data.put("fromX", fromX);
                data.put("fromY", fromY);
                data.put("toX", toX);
                data.put("toY", toY);
                data.put("duration", duration);
            } catch (Exception e) {}
            return new AutomationResult("swipe", data, elapsed);
        } else {
            return new AutomationResult("swipe", "Gesture failed", elapsed);
        }
    }

    /**
     * Perform multi-point swipe (for more complex gestures)
     */
    @NonNull
    public AutomationResult performMultiSwipe(@NonNull PointF[] points, long duration) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return new AutomationResult("multiSwipe", "Accessibility service not available");
        }

        if (points.length < 2) {
            return new AutomationResult("multiSwipe", "Need at least 2 points");
        }

        // Create path through all points
        Path path = new Path();
        path.moveTo(points[0].x, points[0].y);

        for (int i = 1; i < points.length; i++) {
            path.lineTo(points[i].x, points[i].y);
        }

        GestureDescription.StrokeDescription stroke =
                new GestureDescription.StrokeDescription(path, 0, duration);

        GestureDescription.Builder builder = new GestureDescription.Builder();
        builder.addStroke(stroke);

        GestureDescription gesture = builder.build();

        // Execute
        AtomicBoolean success = new AtomicBoolean(false);
        CountDownLatch latch = new CountDownLatch(1);

        service.dispatchGesture(gesture, new AccessibilityService.GestureResultCallback() {
            @Override
            public void onCompleted(GestureDescription gestureDescription) {
                success.set(true);
                latch.countDown();
            }

            @Override
            public void onCancelled(GestureDescription gestureDescription) {
                success.set(false);
                latch.countDown();
            }
        }, null);

        try {
            latch.await(duration + 5000, TimeUnit.MILLISECONDS);
        } catch (InterruptedException e) {
            return new AutomationResult("multiSwipe", "Gesture interrupted");
        }

        if (success.get()) {
            JSONObject data = new JSONObject();
            try {
                data.put("pointsCount", points.length);
                data.put("duration", duration);
            } catch (Exception e) {}
            return new AutomationResult("multiSwipe", data, 0);
        } else {
            return new AutomationResult("multiSwipe", "Gesture failed");
        }
    }

    /**
     * Perform pinch gesture (two-finger)
     */
    @NonNull
    public AutomationResult performPinch(int centerX, int centerY, float startSpread,
                                          float endSpread, long duration) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return new AutomationResult("pinch", "Accessibility service not available");
        }

        // Create two paths for pinch gesture
        Path path1 = new Path();
        Path path2 = new Path();

        // Finger 1: starts from left, moves right (pinch in) or left (pinch out)
        int startOffset1 = (int) (startSpread / 2);
        int endOffset1 = (int) (endSpread / 2);

        path1.moveTo(centerX - startOffset1, centerY);
        path1.lineTo(centerX - endOffset1, centerY);

        // Finger 2: starts from right, moves left (pinch in) or right (pinch out)
        path2.moveTo(centerX + startOffset1, centerY);
        path2.lineTo(centerX + endOffset1, centerY);

        GestureDescription.StrokeDescription stroke1 =
                new GestureDescription.StrokeDescription(path1, 0, duration);
        GestureDescription.StrokeDescription stroke2 =
                new GestureDescription.StrokeDescription(path2, 0, duration);

        GestureDescription.Builder builder = new GestureDescription.Builder();
        builder.addStroke(stroke1);
        builder.addStroke(stroke2);

        GestureDescription gesture = builder.build();

        // Execute
        AtomicBoolean success = new AtomicBoolean(false);
        CountDownLatch latch = new CountDownLatch(1);

        service.dispatchGesture(gesture, new AccessibilityService.GestureResultCallback() {
            @Override
            public void onCompleted(GestureDescription gestureDescription) {
                success.set(true);
                latch.countDown();
                Log.d(TAG, "Pinch gesture completed");
            }

            @Override
            public void onCancelled(GestureDescription gestureDescription) {
                success.set(false);
                latch.countDown();
                Log.w(TAG, "Pinch gesture cancelled");
            }
        }, null);

        try {
            latch.await(duration + 5000, TimeUnit.MILLISECONDS);
        } catch (InterruptedException e) {
            return new AutomationResult("pinch", "Gesture interrupted");
        }

        if (success.get()) {
            JSONObject data = new JSONObject();
            try {
                data.put("centerX", centerX);
                data.put("centerY", centerY);
                data.put("startSpread", startSpread);
                data.put("endSpread", endSpread);
            } catch (Exception e) {}
            return new AutomationResult("pinch", data, 0);
        } else {
            return new AutomationResult("pinch", "Gesture failed");
        }
    }

    /**
     * Input text using clipboard paste
     * Note: Direct text input via accessibility is limited, clipboard paste is more reliable
     */
    @NonNull
    public AutomationResult performInput(@NonNull String text) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return new AutomationResult("input", "Accessibility service not available");
        }

        AccessibilityNodeInfo focusNode = service.getRootNode();
        if (focusNode == null) {
            return new AutomationResult("input", "No focused element");
        }

        // Find the focused element
        AccessibilityNodeInfo focusedNode = findFocusedNode(focusNode);
        if (focusedNode == null) {
            return new AutomationResult("input", "No focused input element");
        }

        long startTime = System.currentTimeMillis();

        // Try direct text input if available
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            // Try ACTION_SET_TEXT
            if (focusedNode.getActions() != null &&
                    (focusedNode.getActions() & AccessibilityNodeInfo.ACTION_SET_TEXT) != 0) {
                boolean result = focusedNode.performAction(
                        AccessibilityNodeInfo.ACTION_SET_TEXT,
                        createSetTextBundle(text));
                if (result) {
                    long elapsed = System.currentTimeMillis() - startTime;
                    JSONObject data = new JSONObject();
                    try {
                        data.put("text", text);
                        data.put("method", "ACTION_SET_TEXT");
                    } catch (Exception e) {}
                    return new AutomationResult("input", data, elapsed);
                }
            }
        }

        // Fallback: Use clipboard paste
        if (clipboardManager != null) {
            ClipData clip = ClipData.newPlainText("automation_input", text);
            clipboardManager.setPrimaryClip(clip);

            // Perform paste action
            if (focusedNode.getActions() != null &&
                    (focusedNode.getActions() & AccessibilityNodeInfo.ACTION_PASTE) != 0) {
                boolean result = focusedNode.performAction(AccessibilityNodeInfo.ACTION_PASTE);
                if (result) {
                    long elapsed = System.currentTimeMillis() - startTime;
                    JSONObject data = new JSONObject();
                    try {
                        data.put("text", text);
                        data.put("method", "clipboard_paste");
                    } catch (Exception e) {}
                    return new AutomationResult("input", data, elapsed);
                }
            }
        }

        return new AutomationResult("input", "Could not input text - no supported method");
    }

    /**
     * Find focused node recursively
     */
    @Nullable
    private AccessibilityNodeInfo findFocusedNode(@NonNull AccessibilityNodeInfo node) {
        if (node.isFocused() && node.isEditable()) {
            return node;
        }

        // Check for accessibility focus
        if (node.isAccessibilityFocused()) {
            return node;
        }

        int childCount = node.getChildCount();
        for (int i = 0; i < childCount; i++) {
            AccessibilityNodeInfo child = node.getChild(i);
            if (child != null) {
                AccessibilityNodeInfo found = findFocusedNode(child);
                if (found != null) {
                    return found;
                }
            }
        }

        return null;
    }

    /**
     * Create bundle for ACTION_SET_TEXT
     */
    private android.os.Bundle createSetTextBundle(@NonNull String text) {
        android.os.Bundle bundle = new android.os.Bundle();
        bundle.putString(AccessibilityNodeInfo.ACTION_ARGUMENT_SET_TEXT_CHARSEQUENCE, text);
        return bundle;
    }

    // ===== Helper Methods =====

    /**
     * Get engine config
     */
    @Nullable
    private AutomationConfig getConfig() {
        return engine.getConfig();
    }
}