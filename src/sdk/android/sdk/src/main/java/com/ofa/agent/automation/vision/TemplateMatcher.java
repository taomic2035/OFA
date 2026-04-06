package com.ofa.agent.automation.vision;

import android.graphics.Bitmap;
import android.graphics.Color;
import android.graphics.Rect;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;

/**
 * Template Matcher - image template matching for visual automation.
 * Finds template images within a larger screenshot using normalized cross-correlation.
 */
public class TemplateMatcher {

    private static final String TAG = "TemplateMatcher";

    // Configuration
    private float matchThreshold = 0.8f; // 80% similarity threshold
    private int maxMatches = 10;
    private boolean useGrayscale = true; // Faster matching with grayscale

    /**
     * Match result
     */
    public static class Match {
        public final int x;
        public final int y;
        public final int width;
        public final int height;
        public final float score;
        public final Rect rect;

        public Match(int x, int y, int width, int height, float score) {
            this.x = x;
            this.y = y;
            this.width = width;
            this.height = height;
            this.score = score;
            this.rect = new Rect(x, y, x + width, y + height);
        }

        public int getCenterX() {
            return x + width / 2;
        }

        public int getCenterY() {
            return y + height / 2;
        }

        @Override
        public String toString() {
            return String.format("Match(%.2f%% at %d,%d)", score * 100, x, y);
        }
    }

    /**
     * Match result with multiple matches
     */
    public static class MatchResult {
        public final boolean success;
        public final List<Match> matches;
        public final long processingTimeMs;

        public MatchResult(boolean success, List<Match> matches, long processingTimeMs) {
            this.success = success;
            this.matches = matches;
            this.processingTimeMs = processingTimeMs;
        }
    }

    /**
     * Find template in source image
     */
    @NonNull
    public MatchResult findTemplate(@NonNull Bitmap source, @NonNull Bitmap template) {
        return findTemplate(source, template, matchThreshold, maxMatches);
    }

    /**
     * Find template with custom threshold
     */
    @NonNull
    public MatchResult findTemplate(@NonNull Bitmap source, @NonNull Bitmap template,
                                     float threshold, int maxMatches) {
        long startTime = System.currentTimeMillis();

        int sourceWidth = source.getWidth();
        int sourceHeight = source.getHeight();
        int templateWidth = template.getWidth();
        int templateHeight = template.getHeight();

        // Validate dimensions
        if (templateWidth > sourceWidth || templateHeight > sourceHeight) {
            return new MatchResult(false, Collections.emptyList(), 0);
        }

        List<Match> matches = new ArrayList<>();

        // Extract pixels
        int[] sourcePixels = new int[sourceWidth * sourceHeight];
        int[] templatePixels = new int[templateWidth * templateHeight];
        source.getPixels(sourcePixels, 0, sourceWidth, 0, 0, sourceWidth, sourceHeight);
        template.getPixels(templatePixels, 0, templateWidth, 0, 0, templateWidth, templateHeight);

        // Convert to grayscale if enabled
        double[] sourceGray = null;
        double[] templateGray = null;
        if (useGrayscale) {
            sourceGray = toGrayscale(sourcePixels);
            templateGray = toGrayscale(templatePixels);
        }

        // Calculate template mean and std for NCC
        double templateMean = 0;
        double templateStd = 0;
        if (useGrayscale) {
            double[] stats = calculateStats(templateGray, templateWidth, templateHeight, 0, 0);
            templateMean = stats[0];
            templateStd = stats[1];
        }

        // Slide template over source
        int searchWidth = sourceWidth - templateWidth;
        int searchHeight = sourceHeight - templateHeight;

        // Use step for faster search (skip pixels)
        int step = 2; // Check every 2nd pixel initially

        for (int y = 0; y <= searchHeight; y += step) {
            for (int x = 0; x <= searchWidth; x += step) {
                double score = useGrayscale ?
                        calculateNCC(sourceGray, templateGray, sourceWidth, templateWidth, templateHeight,
                                x, y, templateMean, templateStd) :
                        calculateNCCColor(sourcePixels, templatePixels, sourceWidth, templateWidth, templateHeight,
                                x, y);

                if (score >= threshold) {
                    // Refine with full precision around this point
                    Match refinedMatch = refineMatch(source, template, x, y, step, threshold);
                    if (refinedMatch != null) {
                        // Check if this match overlaps with existing matches
                        if (!isOverlapping(refinedMatch, matches)) {
                            matches.add(refinedMatch);
                        }
                    }
                }
            }
        }

        // Sort by score descending
        matches.sort((a, b) -> Float.compare(b.score, a.score));

        // Limit matches
        if (matches.size() > maxMatches) {
            matches = matches.subList(0, maxMatches);
        }

        long processingTime = System.currentTimeMillis() - startTime;
        Log.d(TAG, "Template matching completed in " + processingTime + "ms, found " + matches.size() + " matches");

        return new MatchResult(!matches.isEmpty(), matches, processingTime);
    }

    /**
     * Refine match around initial position
     */
    @Nullable
    private Match refineMatch(@NonNull Bitmap source, @NonNull Bitmap template,
                               int startX, int startY, int step, float threshold) {
        int sourceWidth = source.getWidth();
        int templateWidth = template.getWidth();
        int templateHeight = template.getHeight();

        float bestScore = 0;
        int bestX = startX;
        int bestY = startY;

        // Search in neighborhood
        int refineStartX = Math.max(0, startX - step);
        int refineStartY = Math.max(0, startY - step);
        int refineEndX = Math.min(sourceWidth - templateWidth, startX + step);
        int refineEndY = Math.min(source.getHeight() - templateHeight, startY + step);

        int[] sourcePixels = new int[sourceWidth * source.getHeight()];
        int[] templatePixels = new int[templateWidth * templateHeight];
        source.getPixels(sourcePixels, 0, sourceWidth, 0, 0, sourceWidth, source.getHeight());
        template.getPixels(templatePixels, 0, templateWidth, 0, 0, templateWidth, templateHeight);

        double[] sourceGray = useGrayscale ? toGrayscale(sourcePixels) : null;
        double[] templateGray = useGrayscale ? toGrayscale(templatePixels) : null;

        double templateMean = 0;
        double templateStd = 0;
        if (useGrayscale) {
            double[] stats = calculateStats(templateGray, templateWidth, templateHeight, 0, 0);
            templateMean = stats[0];
            templateStd = stats[1];
        }

        for (int y = refineStartY; y <= refineEndY; y++) {
            for (int x = refineStartX; x <= refineEndX; x++) {
                double score = useGrayscale ?
                        calculateNCC(sourceGray, templateGray, sourceWidth, templateWidth, templateHeight,
                                x, y, templateMean, templateStd) :
                        calculateNCCColor(sourcePixels, templatePixels, sourceWidth, templateWidth, templateHeight,
                                x, y);

                if (score > bestScore) {
                    bestScore = (float) score;
                    bestX = x;
                    bestY = y;
                }
            }
        }

        if (bestScore >= threshold) {
            return new Match(bestX, bestY, templateWidth, templateHeight, bestScore);
        }

        return null;
    }

    /**
     * Calculate Normalized Cross-Correlation for grayscale
     */
    private double calculateNCC(double[] sourceGray, double[] templateGray,
                                 int sourceWidth, int templateWidth, int templateHeight,
                                 int offsetX, int offsetY, double templateMean, double templateStd) {
        if (templateStd == 0) return 0;

        double sum = 0;
        double sourceSum = 0;

        int count = templateWidth * templateHeight;

        for (int ty = 0; ty < templateHeight; ty++) {
            for (int tx = 0; tx < templateWidth; tx++) {
                int sx = offsetX + tx;
                int sy = offsetY + ty;
                int sourceIdx = sy * sourceWidth + sx;
                int templateIdx = ty * templateWidth + tx;

                sourceSum += sourceGray[sourceIdx];
            }
        }

        double sourceMean = sourceSum / count;

        double numerator = 0;
        double sourceVar = 0;

        for (int ty = 0; ty < templateHeight; ty++) {
            for (int tx = 0; tx < templateWidth; tx++) {
                int sx = offsetX + tx;
                int sy = offsetY + ty;
                int sourceIdx = sy * sourceWidth + sx;
                int templateIdx = ty * templateWidth + tx;

                double sourceDiff = sourceGray[sourceIdx] - sourceMean;
                double templateDiff = templateGray[templateIdx] - templateMean;

                numerator += sourceDiff * templateDiff;
                sourceVar += sourceDiff * sourceDiff;
            }
        }

        double denominator = Math.sqrt(sourceVar) * templateStd * Math.sqrt(count);

        return denominator > 0 ? numerator / denominator : 0;
    }

    /**
     * Calculate NCC for color images
     */
    private double calculateNCCColor(int[] sourcePixels, int[] templatePixels,
                                      int sourceWidth, int templateWidth, int templateHeight,
                                      int offsetX, int offsetY) {
        int count = templateWidth * templateHeight * 3; // RGB channels

        // Calculate means
        double[] sourceSums = new double[3];
        double[] templateSums = new double[3];

        for (int ty = 0; ty < templateHeight; ty++) {
            for (int tx = 0; tx < templateWidth; tx++) {
                int sx = offsetX + tx;
                int sy = offsetY + ty;
                int sourcePixel = sourcePixels[sy * sourceWidth + sx];
                int templatePixel = templatePixels[ty * templateWidth + tx];

                sourceSums[0] += Color.red(sourcePixel);
                sourceSums[1] += Color.green(sourcePixel);
                sourceSums[2] += Color.blue(sourcePixel);

                templateSums[0] += Color.red(templatePixel);
                templateSums[1] += Color.green(templatePixel);
                templateSums[2] += Color.blue(templatePixel);
            }
        }

        int pixelCount = templateWidth * templateHeight;
        double[] sourceMeans = {sourceSums[0] / pixelCount, sourceSums[1] / pixelCount, sourceSums[2] / pixelCount};
        double[] templateMeans = {templateSums[0] / pixelCount, templateSums[1] / pixelCount, templateSums[2] / pixelCount};

        // Calculate NCC
        double numerator = 0;
        double sourceVar = 0;
        double templateVar = 0;

        for (int ty = 0; ty < templateHeight; ty++) {
            for (int tx = 0; tx < templateWidth; tx++) {
                int sx = offsetX + tx;
                int sy = offsetY + ty;
                int sourcePixel = sourcePixels[sy * sourceWidth + sx];
                int templatePixel = templatePixels[ty * templateWidth + tx];

                for (int c = 0; c < 3; c++) {
                    int sourceVal = c == 0 ? Color.red(sourcePixel) : (c == 1 ? Color.green(sourcePixel) : Color.blue(sourcePixel));
                    int templateVal = c == 0 ? Color.red(templatePixel) : (c == 1 ? Color.green(templatePixel) : Color.blue(templatePixel));

                    double sourceDiff = sourceVal - sourceMeans[c];
                    double templateDiff = templateVal - templateMeans[c];

                    numerator += sourceDiff * templateDiff;
                    sourceVar += sourceDiff * sourceDiff;
                    templateVar += templateDiff * templateDiff;
                }
            }
        }

        double denominator = Math.sqrt(sourceVar * templateVar);
        return denominator > 0 ? numerator / denominator : 0;
    }

    /**
     * Convert pixels to grayscale
     */
    @NonNull
    private double[] toGrayscale(int[] pixels) {
        double[] gray = new double[pixels.length];
        for (int i = 0; i < pixels.length; i++) {
            gray[i] = (Color.red(pixels[i]) + Color.green(pixels[i]) + Color.blue(pixels[i])) / 3.0;
        }
        return gray;
    }

    /**
     * Calculate mean and std for template
     */
    @NonNull
    private double[] calculateStats(double[] gray, int width, int height, int offsetX, int offsetY) {
        double sum = 0;
        int count = width * height;

        for (int y = 0; y < height; y++) {
            for (int x = 0; x < width; x++) {
                sum += gray[(offsetY + y) * width + (offsetX + x)];
            }
        }

        double mean = sum / count;

        double variance = 0;
        for (int y = 0; y < height; y++) {
            for (int x = 0; x < width; x++) {
                double diff = gray[(offsetY + y) * width + (offsetX + x)] - mean;
                variance += diff * diff;
            }
        }

        double std = Math.sqrt(variance / count);

        return new double[] { mean, std };
    }

    /**
     * Check if match overlaps with existing matches
     */
    private boolean isOverlapping(Match match, List<Match> existing) {
        for (Match m : existing) {
            if (Rect.intersects(match.rect, m.rect)) {
                return true;
            }
        }
        return false;
    }

    /**
     * Find best match
     */
    @Nullable
    public Match findBestMatch(@NonNull Bitmap source, @NonNull Bitmap template) {
        MatchResult result = findTemplate(source, template);
        if (result.success && !result.matches.isEmpty()) {
            return result.matches.get(0);
        }
        return null;
    }

    /**
     * Check if template exists in source
     */
    public boolean templateExists(@NonNull Bitmap source, @NonNull Bitmap template) {
        MatchResult result = findTemplate(source, template);
        return result.success;
    }

    /**
     * Wait for template to appear
     */
    @Nullable
    public Match waitForTemplate(@NonNull BitmapProvider bitmapProvider,
                                  @NonNull Bitmap template,
                                  long timeoutMs,
                                  long intervalMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            Bitmap current = bitmapProvider.capture();
            if (current != null) {
                Match match = findBestMatch(current, template);
                if (match != null) {
                    return match;
                }
                current.recycle();
            }

            try {
                Thread.sleep(intervalMs);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    // ===== Configuration =====

    public void setMatchThreshold(float threshold) {
        this.matchThreshold = Math.max(0.1f, Math.min(1.0f, threshold));
    }

    public void setMaxMatches(int maxMatches) {
        this.maxMatches = maxMatches;
    }

    public void setUseGrayscale(boolean useGrayscale) {
        this.useGrayscale = useGrayscale;
    }

    /**
     * Interface for bitmap capture
     */
    public interface BitmapProvider {
        @Nullable
        Bitmap capture();
    }
}