package com.ofa.agent.automation.vision;

import android.content.Context;
import android.graphics.Bitmap;
import android.graphics.Point;
import android.graphics.Rect;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.google.mlkit.vision.common.InputImage;
import com.google.mlkit.vision.text.Text;
import com.google.mlkit.vision.text.TextRecognition;
import com.google.mlkit.vision.text.TextRecognizer;
import com.google.mlkit.vision.text.chinese.ChineseTextRecognizerOptions;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * ML Kit OCR Engine - real OCR using Google ML Kit.
 * Supports Chinese and Latin text recognition with high accuracy.
 */
public class MlKitOcrEngine {

    private static final String TAG = "MlKitOcrEngine";

    private final Context context;
    private TextRecognizer textRecognizer;
    private boolean initialized = false;

    // Configuration
    private long timeoutMs = 10000; // 10 seconds default timeout

    public MlKitOcrEngine(@NonNull Context context) {
        this.context = context.getApplicationContext();
    }

    /**
     * Initialize the OCR engine
     */
    public void initialize() {
        if (initialized) return;

        try {
            // Use Chinese text recognizer for better Chinese support
            textRecognizer = TextRecognition.getClient(new ChineseTextRecognizerOptions.Builder().build());
            initialized = true;
            Log.i(TAG, "ML Kit OCR Engine initialized with Chinese support");
        } catch (Exception e) {
            Log.e(TAG, "Failed to initialize ML Kit OCR: " + e.getMessage());
            initialized = false;
        }
    }

    /**
     * Shutdown the OCR engine
     */
    public void shutdown() {
        if (textRecognizer != null) {
            textRecognizer.close();
            textRecognizer = null;
        }
        initialized = false;
    }

    /**
     * Check if OCR is available
     */
    public boolean isAvailable() {
        return initialized && textRecognizer != null;
    }

    /**
     * Text block detected in image
     */
    public static class TextBlock {
        public final String text;
        public final Rect boundingBox;
        public final List<Point> cornerPoints;
        public final float confidence;
        public final String recognizedLanguage;

        public TextBlock(String text, Rect boundingBox, List<Point> cornerPoints,
                         float confidence, String language) {
            this.text = text;
            this.boundingBox = boundingBox;
            this.cornerPoints = cornerPoints;
            this.confidence = confidence;
            this.recognizedLanguage = language;
        }

        public int getCenterX() {
            return boundingBox != null ? boundingBox.centerX() : 0;
        }

        public int getCenterY() {
            return boundingBox != null ? boundingBox.centerY() : 0;
        }

        public int getWidth() {
            return boundingBox != null ? boundingBox.width() : 0;
        }

        public int getHeight() {
            return boundingBox != null ? boundingBox.height() : 0;
        }
    }

    /**
     * OCR result containing detected text
     */
    public static class OcrResult {
        public final boolean success;
        public final String error;
        public final List<TextBlock> textBlocks;
        public final String fullText;
        public final long processingTimeMs;

        public OcrResult(boolean success, String error, List<TextBlock> textBlocks,
                         String fullText, long processingTimeMs) {
            this.success = success;
            this.error = error;
            this.textBlocks = textBlocks;
            this.fullText = fullText;
            this.processingTimeMs = processingTimeMs;
        }

        public static OcrResult error(String error) {
            return new OcrResult(false, error, null, null, 0);
        }
    }

    /**
     * Recognize text in bitmap synchronously
     */
    @NonNull
    public OcrResult recognizeText(@NonNull Bitmap bitmap) {
        if (!isAvailable()) {
            return OcrResult.error("OCR engine not initialized");
        }

        long startTime = System.currentTimeMillis();

        try {
            InputImage image = InputImage.fromBitmap(bitmap, 0);

            CountDownLatch latch = new CountDownLatch(1);
            AtomicReference<Text> textResultRef = new AtomicReference<>();
            AtomicReference<Exception> errorRef = new AtomicReference<>();

            textRecognizer.process(image)
                    .addOnSuccessListener(text -> {
                        textResultRef.set(text);
                        latch.countDown();
                    })
                    .addOnFailureListener(e -> {
                        errorRef.set(e);
                        latch.countDown();
                    });

            // Wait with timeout
            if (!latch.await(timeoutMs, TimeUnit.MILLISECONDS)) {
                return OcrResult.error("OCR timeout");
            }

            // Check for error
            Exception error = errorRef.get();
            if (error != null) {
                return OcrResult.error(error.getMessage());
            }

            // Process result
            Text textResult = textResultRef.get();
            if (textResult == null) {
                return OcrResult.error("No text result");
            }

            List<TextBlock> textBlocks = extractTextBlocks(textResult);
            String fullText = textResult.getText();

            long processingTime = System.currentTimeMillis() - startTime;
            Log.d(TAG, "OCR completed in " + processingTime + "ms, found " + textBlocks.size() + " blocks");

            return new OcrResult(true, null, textBlocks, fullText, processingTime);

        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return OcrResult.error("OCR interrupted");
        } catch (Exception e) {
            return OcrResult.error(e.getMessage());
        }
    }

    /**
     * Extract text blocks from ML Kit result
     */
    @NonNull
    private List<TextBlock> extractTextBlocks(@NonNull Text text) {
        List<TextBlock> blocks = new ArrayList<>();

        for (Text.TextBlock block : text.getTextBlocks()) {
            String blockText = block.getText();
            Rect boundingBox = block.getBoundingBox();
            List<Point> cornerPoints = block.getCornerPoints();

            // Get confidence if available
            float confidence = 1.0f; // ML Kit doesn't provide confidence per block
            String language = block.getRecognizedLanguage();

            blocks.add(new TextBlock(blockText, boundingBox, cornerPoints, confidence, language));

            // Also extract lines within block for more granular positioning
            for (Text.Line line : block.getLines()) {
                String lineText = line.getText();
                Rect lineBox = line.getBoundingBox();

                if (lineBox != null && lineText != null) {
                    blocks.add(new TextBlock(lineText, lineBox,
                            line.getCornerPoints(), confidence, language));
                }
            }
        }

        return blocks;
    }

    /**
     * Find text in bitmap and return matching blocks
     */
    @NonNull
    public List<TextBlock> findText(@NonNull Bitmap bitmap, @NonNull String searchText) {
        List<TextBlock> result = new ArrayList<>();

        OcrResult ocrResult = recognizeText(bitmap);
        if (!ocrResult.success || ocrResult.textBlocks == null) {
            return result;
        }

        String searchLower = searchText.toLowerCase();
        for (TextBlock block : ocrResult.textBlocks) {
            if (block.text.toLowerCase().contains(searchLower)) {
                result.add(block);
            }
        }

        Log.d(TAG, "Found " + result.size() + " blocks matching: " + searchText);
        return result;
    }

    /**
     * Check if text exists in bitmap
     */
    public boolean containsText(@NonNull Bitmap bitmap, @NonNull String text) {
        OcrResult result = recognizeText(bitmap);
        return result.success && result.fullText != null &&
                result.fullText.toLowerCase().contains(text.toLowerCase());
    }

    /**
     * Find coordinates of text
     */
    @Nullable
    public int[] findTextCoordinates(@NonNull Bitmap bitmap, @NonNull String text) {
        List<TextBlock> blocks = findText(bitmap, text);
        if (blocks.isEmpty()) {
            return null;
        }

        TextBlock first = blocks.get(0);
        return new int[] { first.getCenterX(), first.getCenterY() };
    }

    /**
     * Find exact text match (whole word)
     */
    @NonNull
    public List<TextBlock> findExactText(@NonNull Bitmap bitmap, @NonNull String searchText) {
        List<TextBlock> result = new ArrayList<>();

        OcrResult ocrResult = recognizeText(bitmap);
        if (!ocrResult.success || ocrResult.textBlocks == null) {
            return result;
        }

        String searchLower = searchText.toLowerCase().trim();
        for (TextBlock block : ocrResult.textBlocks) {
            String blockText = block.text.toLowerCase().trim();
            if (blockText.equals(searchLower)) {
                result.add(block);
            }
        }

        return result;
    }

    /**
     * Find text in specific region
     */
    @Nullable
    public TextBlock findTextInRegion(@NonNull Bitmap bitmap, @NonNull String text,
                                       @NonNull Rect region) {
        List<TextBlock> blocks = findText(bitmap, text);

        for (TextBlock block : blocks) {
            if (block.boundingBox != null && Rect.intersects(region, block.boundingBox)) {
                return block;
            }
        }

        return null;
    }

    /**
     * Get all text blocks sorted by position (top to bottom, left to right)
     */
    @NonNull
    public List<TextBlock> getTextBlocksSorted(@NonNull Bitmap bitmap) {
        OcrResult result = recognizeText(bitmap);
        if (!result.success || result.textBlocks == null) {
            return new ArrayList<>();
        }

        List<TextBlock> blocks = new ArrayList<>(result.textBlocks);
        // Sort by Y first, then by X
        blocks.sort((a, b) -> {
            if (a.boundingBox == null && b.boundingBox == null) return 0;
            if (a.boundingBox == null) return 1;
            if (b.boundingBox == null) return -1;

            int yDiff = a.boundingBox.top - b.boundingBox.top;
            if (Math.abs(yDiff) > 20) { // Different lines
                return yDiff;
            }
            return a.boundingBox.left - b.boundingBox.left;
        });

        return blocks;
    }

    /**
     * Set timeout for OCR operations
     */
    public void setTimeout(long timeoutMs) {
        this.timeoutMs = timeoutMs;
    }

    /**
     * Get processing statistics
     */
    @NonNull
    public String getStats() {
        return "ML Kit OCR Engine: initialized=" + initialized +
                ", timeout=" + timeoutMs + "ms";
    }
}