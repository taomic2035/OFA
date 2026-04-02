package com.ofa.agent.automation.vision;

import android.content.Context;
import android.graphics.Bitmap;
import android.graphics.Color;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.List;

/**
 * Simple OCR Helper - basic text detection in images.
 * This is a simplified implementation that provides basic functionality.
 * For production use, integrate with ML Kit or Tesseract.
 */
public class SimpleOcrHelper {

    private static final String TAG = "SimpleOcrHelper";

    private final Context context;

    // Configuration
    private int minTextWidth = 10;
    private int minTextHeight = 10;
    private int contrastThreshold = 128;

    public SimpleOcrHelper(@NonNull Context context) {
        this.context = context.getApplicationContext();
    }

    /**
     * Text region detected in image
     */
    public static class TextRegion {
        public final int x;
        public final int y;
        public final int width;
        public final int height;
        public final String text;
        public final float confidence;

        public TextRegion(int x, int y, int width, int height, String text, float confidence) {
            this.x = x;
            this.y = y;
            this.width = width;
            this.height = height;
            this.text = text;
            this.confidence = confidence;
        }

        public int getCenterX() {
            return x + width / 2;
        }

        public int getCenterY() {
            return y + height / 2;
        }
    }

    /**
     * OCR result containing detected text
     */
    public static class OcrResult {
        public final boolean success;
        public final List<TextRegion> regions;
        public final String fullText;
        public final long processingTimeMs;

        public OcrResult(boolean success, List<TextRegion> regions, long processingTimeMs) {
            this.success = success;
            this.regions = regions;
            this.processingTimeMs = processingTimeMs;

            // Build full text
            StringBuilder sb = new StringBuilder();
            for (TextRegion region : regions) {
                if (sb.length() > 0) sb.append("\n");
                sb.append(region.text);
            }
            this.fullText = sb.toString();
        }
    }

    /**
     * Recognize text in bitmap
     * Note: This is a simplified placeholder implementation.
     * For real OCR, integrate with ML Kit or Tesseract.
     */
    @NonNull
    public OcrResult recognizeText(@NonNull Bitmap bitmap) {
        long startTime = System.currentTimeMillis();

        // Note: This is a simplified implementation
        // Real OCR should use ML Kit or Tesseract
        Log.w(TAG, "SimpleOcrHelper is a placeholder. For real OCR, integrate ML Kit.");

        List<TextRegion> regions = new ArrayList<>();

        // Detect text regions using edge detection (simplified)
        List<int[]> potentialRegions = detectTextRegions(bitmap);

        for (int[] region : potentialRegions) {
            // Placeholder - in real implementation, this would be actual OCR
            String detectedText = "text_" + region[0] + "_" + region[1];
            regions.add(new TextRegion(
                region[0], region[1], region[2], region[3],
                detectedText, 0.5f
            ));
        }

        long processingTime = System.currentTimeMillis() - startTime;

        return new OcrResult(true, regions, processingTime);
    }

    /**
     * Find text in bitmap and return regions
     */
    @NonNull
    public List<TextRegion> findText(@NonNull Bitmap bitmap, @NonNull String searchText) {
        List<TextRegion> result = new ArrayList<>();

        OcrResult ocrResult = recognizeText(bitmap);
        for (TextRegion region : ocrResult.regions) {
            if (region.text.toLowerCase().contains(searchText.toLowerCase())) {
                result.add(region);
            }
        }

        return result;
    }

    /**
     * Check if text exists in bitmap
     */
    public boolean containsText(@NonNull Bitmap bitmap, @NonNull String text) {
        OcrResult result = recognizeText(bitmap);
        return result.fullText.toLowerCase().contains(text.toLowerCase());
    }

    /**
     * Find coordinates of text
     */
    @Nullable
    public int[] findTextCoordinates(@NonNull Bitmap bitmap, @NonNull String text) {
        List<TextRegion> regions = findText(bitmap, text);
        if (regions.isEmpty()) {
            return null;
        }

        TextRegion first = regions.get(0);
        return new int[] { first.getCenterX(), first.getCenterY() };
    }

    // ===== Text Region Detection =====

    /**
     * Detect potential text regions using image analysis
     * This is a simplified implementation using edge detection
     */
    @NonNull
    private List<int[]> detectTextRegions(@NonNull Bitmap bitmap) {
        List<int[]> regions = new ArrayList<>();

        int width = bitmap.getWidth();
        int height = bitmap.getHeight();

        // Convert to grayscale and detect edges
        int[] pixels = new int[width * height];
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height);

        // Simple edge detection (Sobel-like)
        boolean[][] edges = detectEdges(pixels, width, height);

        // Find connected regions (text candidates)
        regions = findConnectedRegions(edges, width, height);

        return regions;
    }

    /**
     * Simple edge detection
     */
    private boolean[][] detectEdges(int[] pixels, int width, int height) {
        boolean[][] edges = new boolean[height][width];

        for (int y = 1; y < height - 1; y++) {
            for (int x = 1; x < width - 1; x++) {
                int gray = getGrayscale(pixels[y * width + x]);

                // Check if this is an edge pixel
                int leftGray = getGrayscale(pixels[y * width + (x - 1)]);
                int rightGray = getGrayscale(pixels[y * width + (x + 1)]);
                int topGray = getGrayscale(pixels[(y - 1) * width + x]);
                int bottomGray = getGrayscale(pixels[(y + 1) * width + x]);

                int gx = Math.abs(rightGray - leftGray);
                int gy = Math.abs(bottomGray - topGray);

                edges[y][x] = (gx + gy) > contrastThreshold;
            }
        }

        return edges;
    }

    /**
     * Find connected regions in edge map
     */
    @NonNull
    private List<int[]> findConnectedRegions(boolean[][] edges, int width, int height) {
        List<int[]> regions = new ArrayList<>();
        boolean[][] visited = new boolean[height][width];

        for (int y = 0; y < height; y++) {
            for (int x = 0; x < width; x++) {
                if (edges[y][x] && !visited[y][x]) {
                    int[] bounds = floodFill(edges, visited, x, y, width, height);
                    if (bounds != null && bounds[2] >= minTextWidth && bounds[3] >= minTextHeight) {
                        regions.add(bounds);
                    }
                }
            }
        }

        return regions;
    }

    /**
     * Flood fill to find region bounds
     */
    @Nullable
    private int[] floodFill(boolean[][] edges, boolean[][] visited, int startX, int startY, int width, int height) {
        if (startX < 0 || startX >= width || startY < 0 || startY >= height) {
            return null;
        }
        if (visited[startY][startX] || !edges[startY][startX]) {
            return null;
        }

        int minX = startX, maxX = startX;
        int minY = startY, maxY = startY;

        // Simple BFS
        List<int[]> queue = new ArrayList<>();
        queue.add(new int[] { startX, startY });
        visited[startY][startX] = true;

        int[][] directions = { { -1, 0 }, { 1, 0 }, { 0, -1 }, { 0, 1 } };

        while (!queue.isEmpty()) {
            int[] current = queue.remove(0);
            int cx = current[0];
            int cy = current[1];

            minX = Math.min(minX, cx);
            maxX = Math.max(maxX, cx);
            minY = Math.min(minY, cy);
            maxY = Math.max(maxY, cy);

            for (int[] dir : directions) {
                int nx = cx + dir[0];
                int ny = cy + dir[1];

                if (nx >= 0 && nx < width && ny >= 0 && ny < height &&
                    !visited[ny][nx] && edges[ny][nx]) {
                    visited[ny][nx] = true;
                    queue.add(new int[] { nx, ny });
                }
            }
        }

        return new int[] { minX, minY, maxX - minX, maxY - minY };
    }

    /**
     * Convert pixel to grayscale
     */
    private int getGrayscale(int pixel) {
        int r = Color.red(pixel);
        int g = Color.green(pixel);
        int b = Color.blue(pixel);
        return (r + g + b) / 3;
    }

    /**
     * Set minimum text region size
     */
    public void setMinTextSize(int minWidth, int minHeight) {
        this.minTextWidth = minWidth;
        this.minTextHeight = minHeight;
    }

    /**
     * Set contrast threshold for edge detection
     */
    public void setContrastThreshold(int threshold) {
        this.contrastThreshold = threshold;
    }

    /**
     * Check if OCR is available
     * Real implementation should check for ML Kit availability
     */
    public boolean isOcrAvailable() {
        // Placeholder - always returns false
        // Real implementation should check for ML Kit or Tesseract
        return false;
    }
}