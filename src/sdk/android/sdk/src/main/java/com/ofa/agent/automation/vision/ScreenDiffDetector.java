package com.ofa.agent.automation.vision;

import android.graphics.Bitmap;
import android.graphics.Color;
import android.graphics.Rect;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.List;

/**
 * Screen Diff Detector - detects changes between screenshots.
 * Useful for detecting page changes, animations, and dynamic content.
 */
public class ScreenDiffDetector {

    private static final String TAG = "ScreenDiffDetector";

    // Configuration
    private float diffThreshold = 0.05f; // 5% pixels different = significant change
    private int blockSize = 8; // Block size for change detection
    private boolean ignoreSmallChanges = true;

    /**
     * Diff region
     */
    public static class DiffRegion {
        public final Rect bounds;
        public final float changeRatio; // 0.0 - 1.0
        public final int changedPixels;
        public final int totalPixels;

        public DiffRegion(Rect bounds, float changeRatio, int changedPixels, int totalPixels) {
            this.bounds = bounds;
            this.changeRatio = changeRatio;
            this.changedPixels = changedPixels;
            this.totalPixels = totalPixels;
        }

        public boolean isSignificant() {
            return changeRatio > 0.1f; // More than 10% changed
        }
    }

    /**
     * Diff result
     */
    public static class DiffResult {
        public final boolean hasChanges;
        public final float overallChangeRatio;
        public final List<DiffRegion> regions;
        public final long processingTimeMs;

        public DiffResult(boolean hasChanges, float overallChangeRatio,
                          List<DiffRegion> regions, long processingTimeMs) {
            this.hasChanges = hasChanges;
            this.overallChangeRatio = overallChangeRatio;
            this.regions = regions;
            this.processingTimeMs = processingTimeMs;
        }

        public static DiffResult noChange() {
            return new DiffResult(false, 0, new ArrayList<>(), 0);
        }
    }

    /**
     * Compare two screenshots
     */
    @NonNull
    public DiffResult compare(@Nullable Bitmap before, @Nullable Bitmap after) {
        if (before == null || after == null) {
            return DiffResult.noChange();
        }

        if (before.getWidth() != after.getWidth() || before.getHeight() != after.getHeight()) {
            // Different sizes - consider as major change
            return new DiffResult(true, 1.0f, new ArrayList<>(), 0);
        }

        long startTime = System.currentTimeMillis();

        int width = before.getWidth();
        int height = before.getHeight();

        int[] beforePixels = new int[width * height];
        int[] afterPixels = new int[width * height];

        before.getPixels(beforePixels, 0, width, 0, 0, width, height);
        after.getPixels(afterPixels, 0, width, 0, 0, width, height);

        // Count overall differences
        int totalChanged = 0;

        // Detect changed regions using block-based approach
        int blocksX = width / blockSize;
        int blocksY = height / blockSize;

        List<DiffRegion> regions = new ArrayList<>();

        for (int by = 0; by < blocksY; by++) {
            for (int bx = 0; bx < blocksX; bx++) {
                int startX = bx * blockSize;
                int startY = by * blockSize;
                int endX = Math.min(startX + blockSize, width);
                int endY = Math.min(startY + blockSize, height);

                int blockChanged = 0;
                int blockTotal = (endX - startX) * (endY - startY);

                for (int y = startY; y < endY; y++) {
                    for (int x = startX; x < endX; x++) {
                        int idx = y * width + x;
                        if (!pixelsSimilar(beforePixels[idx], afterPixels[idx])) {
                            blockChanged++;
                            totalChanged++;
                        }
                    }
                }

                // If block has significant changes
                float blockRatio = (float) blockChanged / blockTotal;
                if (blockRatio > diffThreshold) {
                    Rect bounds = new Rect(startX, startY, endX, endY);
                    regions.add(new DiffRegion(bounds, blockRatio, blockChanged, blockTotal));
                }
            }
        }

        float overallRatio = (float) totalChanged / (width * height);
        boolean hasChanges = overallRatio > diffThreshold;

        long processingTime = System.currentTimeMillis() - startTime;

        // Merge adjacent regions
        regions = mergeAdjacentRegions(regions);

        Log.d(TAG, "Diff: " + (hasChanges ? "CHANGED" : "NO CHANGE") +
                " (" + (overallRatio * 100) + "%), " + regions.size() + " regions, " +
                processingTime + "ms");

        return new DiffResult(hasChanges, overallRatio, regions, processingTime);
    }

    /**
     * Check if two pixels are similar
     */
    private boolean pixelsSimilar(int p1, int p2) {
        // Allow some tolerance for compression artifacts
        int r1 = Color.red(p1), g1 = Color.green(p1), b1 = Color.blue(p1);
        int r2 = Color.red(p2), g2 = Color.green(p2), b2 = Color.blue(p2);

        int tolerance = 10; // Allow small differences

        return Math.abs(r1 - r2) <= tolerance &&
               Math.abs(g1 - g2) <= tolerance &&
               Math.abs(b1 - b2) <= tolerance;
    }

    /**
     * Merge adjacent diff regions
     */
    @NonNull
    private List<DiffRegion> mergeAdjacentRegions(@NonNull List<DiffRegion> regions) {
        if (regions.size() <= 1) return regions;

        List<DiffRegion> merged = new ArrayList<>();
        boolean[] used = new boolean[regions.size()];

        for (int i = 0; i < regions.size(); i++) {
            if (used[i]) continue;

            DiffRegion current = regions.get(i);
            Rect mergedBounds = new Rect(current.bounds);

            for (int j = i + 1; j < regions.size(); j++) {
                if (used[j]) continue;

                DiffRegion other = regions.get(j);
                // Check if adjacent (within 2 block sizes)
                Rect expanded = new Rect(mergedBounds);
                expanded.inset(-blockSize * 2, -blockSize * 2);

                if (Rect.intersects(expanded, other.bounds)) {
                    mergedBounds.union(other.bounds);
                    used[j] = true;
                }
            }

            merged.add(new DiffRegion(mergedBounds, current.changeRatio,
                    current.changedPixels, current.totalPixels));
            used[i] = true;
        }

        return merged;
    }

    /**
     * Check if page has changed
     */
    public boolean hasPageChanged(@Nullable Bitmap before, @Nullable Bitmap after) {
        DiffResult result = compare(before, after);
        return result.hasChanges;
    }

    /**
     * Wait for page to stabilize
     */
    @Nullable
    public Bitmap waitForStable(@NonNull BitmapProvider provider,
                                 long timeoutMs,
                                 long intervalMs) {
        long startTime = System.currentTimeMillis();

        Bitmap lastBitmap = provider.capture();
        if (lastBitmap == null) return null;

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            try {
                Thread.sleep(intervalMs);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }

            Bitmap currentBitmap = provider.capture();
            if (currentBitmap == null) continue;

            DiffResult diff = compare(lastBitmap, currentBitmap);

            if (!diff.hasChanges) {
                // Page is stable
                return currentBitmap;
            }

            // Page still changing, update reference
            lastBitmap.recycle();
            lastBitmap = currentBitmap;
        }

        // Timeout - return current state
        return lastBitmap;
    }

    /**
     * Wait for specific change
     */
    public boolean waitForChange(@NonNull BitmapProvider provider,
                                  @NonNull ChangeDetector detector,
                                  long timeoutMs,
                                  long intervalMs) {
        long startTime = System.currentTimeMillis();
        Bitmap lastBitmap = provider.capture();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            try {
                Thread.sleep(intervalMs);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }

            Bitmap currentBitmap = provider.capture();
            if (currentBitmap == null) continue;

            DiffResult diff = compare(lastBitmap, currentBitmap);

            if (detector.hasChanged(diff)) {
                lastBitmap.recycle();
                currentBitmap.recycle();
                return true;
            }

            lastBitmap.recycle();
            lastBitmap = currentBitmap;
        }

        if (lastBitmap != null) lastBitmap.recycle();
        return false;
    }

    /**
     * Detect specific region change
     */
    @Nullable
    public DiffRegion getRegionChange(@Nullable Bitmap before, @Nullable Bitmap after,
                                        @NonNull Rect region) {
        if (before == null || after == null) return null;

        // Extract region from both bitmaps
        Bitmap beforeRegion = Bitmap.createBitmap(before,
                region.left, region.top, region.width(), region.height());
        Bitmap afterRegion = Bitmap.createBitmap(after,
                region.left, region.top, region.width(), region.height());

        DiffResult diff = compare(beforeRegion, afterRegion);

        beforeRegion.recycle();
        afterRegion.recycle();

        return diff.hasChanges && !diff.regions.isEmpty() ? diff.regions.get(0) : null;
    }

    /**
     * Get change mask as bitmap
     */
    @Nullable
    public Bitmap getChangeMask(@Nullable Bitmap before, @Nullable Bitmap after) {
        if (before == null || after == null) return null;

        int width = before.getWidth();
        int height = before.getHeight();

        Bitmap mask = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888);

        int[] beforePixels = new int[width * height];
        int[] afterPixels = new int[width * height];
        int[] maskPixels = new int[width * height];

        before.getPixels(beforePixels, 0, width, 0, 0, width, height);
        after.getPixels(afterPixels, 0, width, 0, 0, width, height);

        for (int i = 0; i < width * height; i++) {
            if (pixelsSimilar(beforePixels[i], afterPixels[i])) {
                maskPixels[i] = Color.TRANSPARENT;
            } else {
                maskPixels[i] = Color.RED; // Changed pixels in red
            }
        }

        mask.setPixels(maskPixels, 0, width, 0, 0, width, height);
        return mask;
    }

    // ===== Configuration =====

    public void setDiffThreshold(float threshold) {
        this.diffThreshold = threshold;
    }

    public void setBlockSize(int blockSize) {
        this.blockSize = blockSize;
    }

    public void setIgnoreSmallChanges(boolean ignore) {
        this.ignoreSmallChanges = ignore;
    }

    /**
     * Interface for bitmap capture
     */
    public interface BitmapProvider {
        @Nullable
        Bitmap capture();
    }

    /**
     * Interface for change detection
     */
    public interface ChangeDetector {
        boolean hasChanged(DiffResult diff);
    }
}