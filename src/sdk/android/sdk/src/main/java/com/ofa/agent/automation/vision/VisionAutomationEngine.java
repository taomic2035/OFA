package com.ofa.agent.automation.vision;

import android.content.Context;
import android.graphics.Bitmap;
import android.graphics.Rect;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.util.Base64;
import java.util.List;

/**
 * Vision Automation Engine - unified interface for visual automation.
 * Integrates OCR, template matching, diff detection, and element tracking.
 */
public class VisionAutomationEngine {

    private static final String TAG = "VisionAutomationEngine";

    private final Context context;
    private final AutomationEngine baseEngine;

    // Vision components
    private MlKitOcrEngine ocrEngine;
    private TemplateMatcher templateMatcher;
    private VisualElementFinder elementFinder;
    private ScreenDiffDetector diffDetector;
    private DynamicElementTracker elementTracker;

    private boolean initialized = false;

    // Configuration
    private boolean enableVisualFallback = true;
    private float visualConfidenceThreshold = 0.8f;

    public VisionAutomationEngine(@NonNull Context context, @NonNull AutomationEngine baseEngine) {
        this.context = context;
        this.baseEngine = baseEngine;
    }

    /**
     * Initialize the vision engine
     */
    public void initialize() {
        if (initialized) return;

        Log.i(TAG, "Initializing Vision Automation Engine...");

        // Initialize components
        ocrEngine = new MlKitOcrEngine(context);
        ocrEngine.initialize();

        templateMatcher = new TemplateMatcher();
        templateMatcher.setMatchThreshold(visualConfidenceThreshold);

        elementFinder = new VisualElementFinder(context, baseEngine);

        diffDetector = new ScreenDiffDetector();

        elementTracker = new DynamicElementTracker(context);

        initialized = true;
        Log.i(TAG, "Vision Automation Engine initialized");
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        if (ocrEngine != null) ocrEngine.shutdown();
        if (elementFinder != null) elementFinder.shutdown();
        if (elementTracker != null) elementTracker.shutdown();

        initialized = false;
        Log.i(TAG, "Vision Automation Engine shutdown");
    }

    /**
     * Check if vision engine is available
     */
    public boolean isAvailable() {
        return initialized && ocrEngine != null && ocrEngine.isAvailable();
    }

    // ===== OCR Operations =====

    /**
     * Recognize all text on screen
     */
    @NonNull
    public MlKitOcrEngine.OcrResult recognizeText() {
        Bitmap screenshot = baseEngine.takeScreenshot();
        if (screenshot == null) {
            return MlKitOcrEngine.OcrResult.error("Failed to capture screenshot");
        }

        MlKitOcrEngine.OcrResult result = ocrEngine.recognizeText(screenshot);
        screenshot.recycle();

        return result;
    }

    /**
     * Find text on screen
     */
    @NonNull
    public List<MlKitOcrEngine.TextBlock> findText(@NonNull String text) {
        Bitmap screenshot = baseEngine.takeScreenshot();
        if (screenshot == null) {
            return new java.util.ArrayList<>();
        }

        List<MlKitOcrEngine.TextBlock> blocks = ocrEngine.findText(screenshot, text);
        screenshot.recycle();

        return blocks;
    }

    /**
     * Check if text exists on screen
     */
    public boolean textExists(@NonNull String text) {
        Bitmap screenshot = baseEngine.takeScreenshot();
        if (screenshot == null) return false;

        boolean exists = ocrEngine.containsText(screenshot, text);
        screenshot.recycle();

        return exists;
    }

    /**
     * Get text coordinates
     */
    @Nullable
    public int[] getTextCoordinates(@NonNull String text) {
        Bitmap screenshot = baseEngine.takeScreenshot();
        if (screenshot == null) return null;

        int[] coords = ocrEngine.findTextCoordinates(screenshot, text);
        screenshot.recycle();

        return coords;
    }

    // ===== Template Matching Operations =====

    /**
     * Find template on screen
     */
    @NonNull
    public TemplateMatcher.MatchResult findTemplate(@NonNull Bitmap template) {
        Bitmap screenshot = baseEngine.takeScreenshot();
        if (screenshot == null) {
            return new TemplateMatcher.MatchResult(false, new java.util.ArrayList<>(), 0);
        }

        TemplateMatcher.MatchResult result = templateMatcher.findTemplate(screenshot, template);
        screenshot.recycle();

        return result;
    }

    /**
     * Check if template exists on screen
     */
    public boolean templateExists(@NonNull Bitmap template) {
        return findTemplate(template).success;
    }

    /**
     * Wait for template to appear
     */
    @Nullable
    public TemplateMatcher.Match waitForTemplate(@NonNull Bitmap template, long timeoutMs) {
        return templateMatcher.waitForTemplate(this::captureScreen, template, timeoutMs, 500);
    }

    // ===== Visual Element Operations =====

    /**
     * Find visual element by selector
     */
    @NonNull
    public VisualElementFinder.SearchResult findElement(@NonNull BySelector selector) {
        Bitmap screenshot = baseEngine.takeScreenshot();
        VisualElementFinder.SearchResult result = elementFinder.find(selector, screenshot);
        if (screenshot != null) screenshot.recycle();

        return result;
    }

    /**
     * Find clickable elements
     */
    @NonNull
    public List<VisualElementFinder.VisualElement> findClickableElements() {
        Bitmap screenshot = baseEngine.takeScreenshot();
        List<VisualElementFinder.VisualElement> elements = elementFinder.findClickableElements(screenshot);
        if (screenshot != null) screenshot.recycle();

        return elements;
    }

    /**
     * Find all text elements
     */
    @NonNull
    public List<VisualElementFinder.VisualElement> findAllText() {
        Bitmap screenshot = baseEngine.takeScreenshot();
        List<VisualElementFinder.VisualElement> elements = elementFinder.findAllText(screenshot);
        if (screenshot != null) screenshot.recycle();

        return elements;
    }

    // ===== Visual Click Operations =====

    /**
     * Click on text (visual fallback)
     */
    @NonNull
    public AutomationResult clickText(@NonNull String text) {
        // Try accessibility first
        AutomationResult result = baseEngine.click(text);
        if (result.isSuccess()) {
            return result;
        }

        // Fallback to visual
        if (!enableVisualFallback) {
            return result;
        }

        int[] coords = getTextCoordinates(text);
        if (coords != null) {
            return baseEngine.click(coords[0], coords[1]);
        }

        return new AutomationResult("clickText", "Text not found: " + text);
    }

    /**
     * Click on template image
     */
    @NonNull
    public AutomationResult clickTemplate(@NonNull Bitmap template) {
        TemplateMatcher.Match match = findTemplate(template).matches.stream()
                .findFirst()
                .orElse(null);

        if (match != null) {
            return baseEngine.click(match.getCenterX(), match.getCenterY());
        }

        return new AutomationResult("clickTemplate", "Template not found");
    }

    /**
     * Click on visual element
     */
    @NonNull
    public AutomationResult clickElement(@NonNull VisualElementFinder.VisualElement element) {
        if (element.bounds != null) {
            return baseEngine.click(element.bounds.centerX(), element.bounds.centerY());
        }
        return new AutomationResult("clickElement", "No bounds for element");
    }

    // ===== Screen Diff Operations =====

    /**
     * Compare current screen with reference
     */
    @NonNull
    public ScreenDiffDetector.DiffResult compareWithLast(@Nullable Bitmap reference) {
        Bitmap current = baseEngine.takeScreenshot();
        if (current == null) {
            return ScreenDiffDetector.DiffResult.noChange();
        }

        ScreenDiffDetector.DiffResult result = diffDetector.compare(reference, current);
        current.recycle();

        return result;
    }

    /**
     * Check if screen has changed
     */
    public boolean hasScreenChanged(@Nullable Bitmap reference) {
        return compareWithLast(reference).hasChanges;
    }

    /**
     * Wait for screen to stabilize
     */
    @Nullable
    public Bitmap waitForStableScreen(long timeoutMs) {
        return diffDetector.waitForStable(this::captureScreen, timeoutMs, 300);
    }

    // ===== Element Tracking Operations =====

    /**
     * Start tracking a region
     */
    @NonNull
    public String startTracking(@NonNull Rect bounds) {
        return elementTracker.startTracking(bounds, "region");
    }

    /**
     * Stop tracking
     */
    public void stopTracking(@NonNull String id) {
        elementTracker.stopTracking(id);
    }

    /**
     * Wait for animation to complete
     */
    public void waitForAnimation(long timeoutMs) {
        elementTracker.waitForAnimationEnd(this::captureScreen, timeoutMs);
    }

    /**
     * Check if screen is animating
     */
    public boolean isAnimating() {
        elementTracker.update(baseEngine.takeScreenshot());
        return elementTracker.isAnimating();
    }

    // ===== Combined Operations =====

    /**
     * Wait for text to appear (visual or accessibility)
     */
    @Nullable
    public VisualElementFinder.VisualElement waitForText(@NonNull String text, long timeoutMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            // Try accessibility first
            if (baseEngine.findElement(BySelector.text(text)) != null) {
                return new VisualElementFinder.VisualElement("text", text, null, 1.0f, "accessibility");
            }

            // Try visual
            int[] coords = getTextCoordinates(text);
            if (coords != null) {
                Rect bounds = new Rect(coords[0] - 50, coords[1] - 20,
                        coords[0] + 50, coords[1] + 20);
                return new VisualElementFinder.VisualElement("text", text, bounds, 0.9f, "ocr");
            }

            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    /**
     * Wait for either text or template
     */
    @NonNull
    public AutomationResult waitForAny(@NonNull String text, @Nullable Bitmap template, long timeoutMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            // Check text
            if (textExists(text)) {
                return new AutomationResult("waitForAny", 0).setData("found", "text");
            }

            // Check template
            if (template != null && templateExists(template)) {
                return new AutomationResult("waitForAny", 0).setData("found", "template");
            }

            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return new AutomationResult("waitForAny", "Timeout waiting for element");
    }

    // ===== Utility Methods =====

    /**
     * Capture screen
     */
    @Nullable
    public Bitmap captureScreen() {
        return baseEngine.takeScreenshot();
    }

    /**
     * Get screen content as JSON (for AI analysis)
     */
    @NonNull
    public JSONObject getScreenContent() {
        JSONObject content = new JSONObject();

        try {
            // Get text content
            MlKitOcrEngine.OcrResult ocrResult = recognizeText();
            if (ocrResult.success) {
                JSONArray textArray = new JSONArray();
                for (MlKitOcrEngine.TextBlock block : ocrResult.textBlocks) {
                    JSONObject textObj = new JSONObject();
                    textObj.put("text", block.text);
                    if (block.boundingBox != null) {
                        textObj.put("x", block.boundingBox.left);
                        textObj.put("y", block.boundingBox.top);
                        textObj.put("width", block.boundingBox.width());
                        textObj.put("height", block.boundingBox.height());
                    }
                    textObj.put("confidence", block.confidence);
                    textArray.put(textObj);
                }
                content.put("text_blocks", textArray);
                content.put("full_text", ocrResult.fullText);
            }

            // Get clickable elements
            List<VisualElementFinder.VisualElement> clickables = findClickableElements();
            JSONArray clickableArray = new JSONArray();
            for (VisualElementFinder.VisualElement element : clickables) {
                JSONObject clickObj = new JSONObject();
                clickObj.put("text", element.text);
                clickObj.put("type", element.type);
                if (element.bounds != null) {
                    clickObj.put("center_x", element.bounds.centerX());
                    clickObj.put("center_y", element.bounds.centerY());
                }
                clickableArray.put(clickObj);
            }
            content.put("clickable_elements", clickableArray);

        } catch (Exception e) {
            Log.e(TAG, "Error getting screen content", e);
        }

        return content;
    }

    /**
     * Get screenshot as base64
     */
    @Nullable
    public String getScreenshotBase64() {
        Bitmap screenshot = captureScreen();
        if (screenshot == null) return null;

        ByteArrayOutputStream baos = new ByteArrayOutputStream();
        screenshot.compress(Bitmap.CompressFormat.PNG, 100, baos);
        screenshot.recycle();

        return Base64.getEncoder().encodeToString(baos.toByteArray());
    }

    // ===== Configuration =====

    public void setEnableVisualFallback(boolean enable) {
        this.enableVisualFallback = enable;
    }

    public void setVisualConfidenceThreshold(float threshold) {
        this.visualConfidenceThreshold = threshold;
        if (templateMatcher != null) {
            templateMatcher.setMatchThreshold(threshold);
        }
    }

    // ===== Getters =====

    public MlKitOcrEngine getOcrEngine() {
        return ocrEngine;
    }

    public TemplateMatcher getTemplateMatcher() {
        return templateMatcher;
    }

    public VisualElementFinder getElementFinder() {
        return elementFinder;
    }

    public ScreenDiffDetector getDiffDetector() {
        return diffDetector;
    }

    public DynamicElementTracker getElementTracker() {
        return elementTracker;
    }

    /**
     * Get status summary
     */
    @NonNull
    public String getStatusSummary() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== Vision Automation Engine ===\n");
        sb.append("Initialized: ").append(initialized).append("\n");
        sb.append("Available: ").append(isAvailable()).append("\n");
        sb.append("Visual Fallback: ").append(enableVisualFallback).append("\n");
        sb.append("Confidence Threshold: ").append(visualConfidenceThreshold).append("\n");

        if (ocrEngine != null) {
            sb.append("\nOCR: ").append(ocrEngine.getStats()).append("\n");
        }

        return sb.toString();
    }
}