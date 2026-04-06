package com.ofa.agent.automation.vision;

import android.graphics.Bitmap;
import android.graphics.Rect;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * Dynamic Element Tracker - tracks moving/changing UI elements.
 * Useful for animations, scroll content, loading indicators, etc.
 */
public class DynamicElementTracker {

    private static final String TAG = "DynamicElementTracker";

    private final ScreenDiffDetector diffDetector;
    private final MlKitOcrEngine ocrEngine;

    // Tracking state
    private final Map<String, TrackedElement> trackedElements = new HashMap<>();
    private final List<MovementHistory> movementHistory = new ArrayList<>();

    // Configuration
    private long trackingIntervalMs = 100; // Check interval
    private int historySize = 10; // Keep last 10 positions
    private float movementThreshold = 5.0f; // Minimum pixels to consider movement

    /**
     * Tracked element
     */
    public static class TrackedElement {
        public final String id;
        public final String type;
        public final Rect initialBounds;
        public Rect currentBounds;
        public final List<Rect> positionHistory;
        public long createdAt;
        public long lastSeenAt;
        public int visibilityCount; // Times seen
        public int movementCount; // Times moved

        public TrackedElement(String id, String type, Rect bounds) {
            this.id = id;
            this.type = type;
            this.initialBounds = new Rect(bounds);
            this.currentBounds = new Rect(bounds);
            this.positionHistory = new ArrayList<>();
            this.positionHistory.add(new Rect(bounds));
            this.createdAt = System.currentTimeMillis();
            this.lastSeenAt = this.createdAt;
            this.visibilityCount = 1;
            this.movementCount = 0;
        }

        public void updatePosition(Rect newBounds) {
            positionHistory.add(new Rect(newBounds));
            if (positionHistory.size() > 20) {
                positionHistory.remove(0);
            }
            currentBounds = new Rect(newBounds);
            lastSeenAt = System.currentTimeMillis();
            visibilityCount++;
        }

        public boolean hasMoved() {
            return movementCount > 0;
        }

        public float getVelocityX() {
            if (positionHistory.size() < 2) return 0;
            Rect prev = positionHistory.get(positionHistory.size() - 2);
            Rect curr = positionHistory.get(positionHistory.size() - 1);
            long timeDiff = 100; // Approximate interval
            return (curr.centerX() - prev.centerX()) / (float) timeDiff;
        }

        public float getVelocityY() {
            if (positionHistory.size() < 2) return 0;
            Rect prev = positionHistory.get(positionHistory.size() - 2);
            Rect curr = positionHistory.get(positionHistory.size() - 1);
            long timeDiff = 100; // Approximate interval
            return (curr.centerY() - prev.centerY()) / (float) timeDiff;
        }

        public String getMovementDirection() {
            if (positionHistory.size() < 2) return "none";

            Rect first = positionHistory.get(0);
            Rect last = positionHistory.get(positionHistory.size() - 1);

            int dx = last.centerX() - first.centerX();
            int dy = last.centerY() - first.centerY();

            if (Math.abs(dx) < 10 && Math.abs(dy) < 10) return "none";
            if (Math.abs(dx) > Math.abs(dy)) {
                return dx > 0 ? "right" : "left";
            } else {
                return dy > 0 ? "down" : "up";
            }
        }
    }

    /**
     * Movement history entry
     */
    public static class MovementHistory {
        public final String elementId;
        public final long timestamp;
        public final Rect fromBounds;
        public final Rect toBounds;
        public final float distance;

        public MovementHistory(String elementId, Rect from, Rect to) {
            this.elementId = elementId;
            this.timestamp = System.currentTimeMillis();
            this.fromBounds = new Rect(from);
            this.toBounds = new Rect(to);
            this.distance = calculateDistance(from, to);
        }

        private float calculateDistance(Rect a, Rect b) {
            float dx = a.centerX() - b.centerX();
            float dy = a.centerY() - b.centerY();
            return (float) Math.sqrt(dx * dx + dy * dy);
        }
    }

    public DynamicElementTracker(@NonNull android.content.Context context) {
        this.diffDetector = new ScreenDiffDetector();
        this.ocrEngine = new MlKitOcrEngine(context);
        this.ocrEngine.initialize();
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        ocrEngine.shutdown();
        trackedElements.clear();
        movementHistory.clear();
    }

    /**
     * Start tracking an element
     */
    @NonNull
    public String startTracking(@NonNull Rect bounds, @NonNull String type) {
        String id = UUID.randomUUID().toString().substring(0, 8);
        TrackedElement element = new TrackedElement(id, type, bounds);
        trackedElements.put(id, element);
        Log.d(TAG, "Started tracking element: " + id + " at " + bounds);
        return id;
    }

    /**
     * Stop tracking an element
     */
    public void stopTracking(@NonNull String id) {
        trackedElements.remove(id);
        Log.d(TAG, "Stopped tracking element: " + id);
    }

    /**
     * Update tracking with new screenshot
     */
    public void update(@Nullable Bitmap screenshot) {
        if (screenshot == null) return;

        long currentTime = System.currentTimeMillis();

        for (TrackedElement element : trackedElements.values()) {
            // Try to find the element in current screenshot
            Rect newBounds = locateElement(screenshot, element);

            if (newBounds != null) {
                // Check if element has moved
                float distance = calculateDistance(element.currentBounds, newBounds);

                if (distance > movementThreshold) {
                    // Element moved
                    MovementHistory history = new MovementHistory(
                            element.id, element.currentBounds, newBounds);
                    movementHistory.add(history);
                    element.movementCount++;

                    // Limit history size
                    if (movementHistory.size() > 100) {
                        movementHistory.remove(0);
                    }
                }

                element.updatePosition(newBounds);
            }
        }
    }

    /**
     * Locate element in screenshot
     */
    @Nullable
    private Rect locateElement(@NonNull Bitmap screenshot, @NonNull TrackedElement element) {
        switch (element.type) {
            case "text":
                // Use OCR to find text
                // Would need to store original text
                break;

            case "region":
                // Use template matching or color analysis
                return locateByRegion(screenshot, element.currentBounds);

            case "color":
                // Locate by dominant color
                break;
        }

        // Default: assume element is still in same position
        return new Rect(element.currentBounds);
    }

    /**
     * Locate by checking if region has changed
     */
    @Nullable
    private Rect locateByRegion(@NonNull Bitmap screenshot, @NonNull Rect bounds) {
        // Simple implementation - check if region is still visible
        if (bounds.left < 0 || bounds.top < 0 ||
            bounds.right > screenshot.getWidth() || bounds.bottom > screenshot.getHeight()) {
            return null;
        }

        // Check if content at region is still there
        // For now, just return current bounds
        return new Rect(bounds);
    }

    /**
     * Calculate distance between two rects
     */
    private float calculateDistance(@NonNull Rect a, @NonNull Rect b) {
        float dx = a.centerX() - b.centerX();
        float dy = a.centerY() - b.centerY();
        return (float) Math.sqrt(dx * dx + dy * dy);
    }

    /**
     * Get tracked element by ID
     */
    @Nullable
    public TrackedElement getElement(@NonNull String id) {
        return trackedElements.get(id);
    }

    /**
     * Get all tracked elements
     */
    @NonNull
    public List<TrackedElement> getAllTrackedElements() {
        return new ArrayList<>(trackedElements.values());
    }

    /**
     * Get moving elements
     */
    @NonNull
    public List<TrackedElement> getMovingElements() {
        List<TrackedElement> moving = new ArrayList<>();
        for (TrackedElement element : trackedElements.values()) {
            if (element.hasMoved()) {
                moving.add(element);
            }
        }
        return moving;
    }

    /**
     * Get recent movements
     */
    @NonNull
    public List<MovementHistory> getRecentMovements(int count) {
        int fromIndex = Math.max(0, movementHistory.size() - count);
        return new ArrayList<>(movementHistory.subList(fromIndex, movementHistory.size()));
    }

    /**
     * Check if any element is animating
     */
    public boolean isAnimating() {
        for (TrackedElement element : trackedElements.values()) {
            if (element.positionHistory.size() >= 3) {
                // Check recent movement
                Rect prev = element.positionHistory.get(element.positionHistory.size() - 2);
                Rect curr = element.positionHistory.get(element.positionHistory.size() - 1);
                if (calculateDistance(prev, curr) > movementThreshold) {
                    return true;
                }
            }
        }
        return false;
    }

    /**
     * Wait for animation to complete
     */
    public void waitForAnimationEnd(@NonNull BitmapProvider provider, long timeoutMs) {
        long startTime = System.currentTimeMillis();
        long noMovementStart = 0;
        long stableThreshold = 300; // 300ms of no movement = stable

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            Bitmap screenshot = provider.capture();
            if (screenshot != null) {
                update(screenshot);
                screenshot.recycle();
            }

            if (!isAnimating()) {
                if (noMovementStart == 0) {
                    noMovementStart = System.currentTimeMillis();
                } else if (System.currentTimeMillis() - noMovementStart > stableThreshold) {
                    Log.d(TAG, "Animation ended after " + (System.currentTimeMillis() - startTime) + "ms");
                    return;
                }
            } else {
                noMovementStart = 0;
            }

            try {
                Thread.sleep(trackingIntervalMs);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        Log.w(TAG, "Animation wait timed out");
    }

    /**
     * Predict element position
     */
    @Nullable
    public Rect predictPosition(@NonNull String elementId, long timeMs) {
        TrackedElement element = trackedElements.get(elementId);
        if (element == null || element.positionHistory.size() < 2) {
            return null;
        }

        // Calculate velocity
        float vx = element.getVelocityX();
        float vy = element.getVelocityY();

        // Predict new position
        int newX = Math.round(element.currentBounds.centerX() + vx * timeMs);
        int newY = Math.round(element.currentBounds.centerY() + vy * timeMs);

        int width = element.currentBounds.width();
        int height = element.currentBounds.height();

        return new Rect(newX - width / 2, newY - height / 2,
                        newX + width / 2, newY + height / 2);
    }

    /**
     * Detect loading indicator
     */
    @Nullable
    public TrackedElement detectLoadingIndicator(@NonNull BitmapProvider provider, int checkCount) {
        String id = null;
        Rect lastPos = null;

        for (int i = 0; i < checkCount; i++) {
            Bitmap screenshot = provider.capture();
            if (screenshot == null) continue;

            // Look for spinning/rotating elements
            // This is simplified - real implementation would use computer vision
            ScreenDiffDetector.DiffResult diff = diffDetector.compare(lastPos != null ?
                    provider.capture() : null, screenshot);

            // Check for consistent small movements in same area

            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }

            if (screenshot != null) screenshot.recycle();
        }

        return id != null ? trackedElements.get(id) : null;
    }

    // ===== Configuration =====

    public void setTrackingInterval(long intervalMs) {
        this.trackingIntervalMs = intervalMs;
    }

    public void setHistorySize(int size) {
        this.historySize = size;
    }

    public void setMovementThreshold(float threshold) {
        this.movementThreshold = threshold;
    }

    /**
     * Interface for bitmap capture
     */
    public interface BitmapProvider {
        @Nullable
        Bitmap capture();
    }
}