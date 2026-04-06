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

import java.util.ArrayList;
import java.util.List;

/**
 * Visual Element Finder - locates UI elements using visual analysis.
 * Combines OCR, template matching, and accessibility tree for robust element location.
 */
public class VisualElementFinder {

    private static final String TAG = "VisualElementFinder";

    private final Context context;
    private final AutomationEngine engine;
    private final MlKitOcrEngine ocrEngine;
    private final TemplateMatcher templateMatcher;

    // Configuration
    private boolean preferAccessibility = true; // Use accessibility first
    private float visualFallbackThreshold = 0.7f; // Use visual if accessibility confidence < 70%

    /**
     * Found visual element
     */
    public static class VisualElement {
        public final String type; // "text", "image", "icon", "button"
        public final String text;
        public final Rect bounds;
        public final float confidence;
        public final String source; // "accessibility", "ocr", "template"

        public VisualElement(String type, String text, Rect bounds, float confidence, String source) {
            this.type = type;
            this.text = text;
            this.bounds = bounds;
            this.confidence = confidence;
            this.source = source;
        }

        public int getCenterX() {
            return bounds != null ? bounds.centerX() : 0;
        }

        public int getCenterY() {
            return bounds != null ? bounds.centerY() : 0;
        }
    }

    /**
     * Search result
     */
    public static class SearchResult {
        public final boolean found;
        public final List<VisualElement> elements;
        public final String message;

        public SearchResult(boolean found, List<VisualElement> elements, String message) {
            this.found = found;
            this.elements = elements;
            this.message = message;
        }

        public static SearchResult notFound(String message) {
            return new SearchResult(false, new ArrayList<>(), message);
        }
    }

    public VisualElementFinder(@NonNull Context context, @NonNull AutomationEngine engine) {
        this.context = context;
        this.engine = engine;
        this.ocrEngine = new MlKitOcrEngine(context);
        this.templateMatcher = new TemplateMatcher();

        ocrEngine.initialize();
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        ocrEngine.shutdown();
    }

    /**
     * Find element by text using combined approaches
     */
    @NonNull
    public SearchResult findByText(@NonNull String text, @Nullable Bitmap screenshot) {
        List<VisualElement> elements = new ArrayList<>();

        // 1. Try accessibility tree first
        if (preferAccessibility) {
            try {
                AutomationResult result = engine.click(text);
                if (result.isSuccess()) {
                    // Found via accessibility
                    elements.add(new VisualElement("text", text, null, 1.0f, "accessibility"));
                    return new SearchResult(true, elements, "Found via accessibility");
                }
            } catch (Exception e) {
                Log.d(TAG, "Accessibility search failed: " + e.getMessage());
            }
        }

        // 2. Fall back to OCR
        if (screenshot != null && !elements.isEmpty()) {
            List<MlKitOcrEngine.TextBlock> ocrBlocks = ocrEngine.findText(screenshot, text);
            for (MlKitOcrEngine.TextBlock block : ocrBlocks) {
                if (block.boundingBox != null) {
                    elements.add(new VisualElement("text", block.text, block.boundingBox,
                            block.confidence, "ocr"));
                }
            }
        }

        if (!elements.isEmpty()) {
            return new SearchResult(true, elements, "Found " + elements.size() + " elements");
        }

        return SearchResult.notFound("Text not found: " + text);
    }

    /**
     * Find element by image template
     */
    @NonNull
    public SearchResult findByTemplate(@NonNull Bitmap template, @NonNull Bitmap screenshot) {
        TemplateMatcher.MatchResult matchResult = templateMatcher.findTemplate(screenshot, template);

        List<VisualElement> elements = new ArrayList<>();
        for (TemplateMatcher.Match match : matchResult.matches) {
            Rect bounds = new Rect(match.x, match.y, match.x + match.width, match.y + match.height);
            elements.add(new VisualElement("image", null, bounds, match.score, "template"));
        }

        if (!elements.isEmpty()) {
            return new SearchResult(true, elements, "Found " + elements.size() + " matches");
        }

        return SearchResult.notFound("Template not found");
    }

    /**
     * Find clickable elements
     */
    @NonNull
    public List<VisualElement> findClickableElements(@Nullable Bitmap screenshot) {
        List<VisualElement> elements = new ArrayList<>();

        // Try to get from accessibility tree
        String pageSource = engine.getPageSource();
        if (pageSource != null) {
            // Parse clickable nodes from accessibility tree
            // This is simplified - real implementation would parse XML
            // For now, we detect via OCR
        }

        // Use OCR to find potential buttons
        if (screenshot != null) {
            MlKitOcrEngine.OcrResult ocrResult = ocrEngine.recognizeText(screenshot);
            if (ocrResult.success && ocrResult.textBlocks != null) {
                for (MlKitOcrEngine.TextBlock block : ocrResult.textBlocks) {
                    // Heuristics for detecting clickable text
                    if (isLikelyClickable(block)) {
                        elements.add(new VisualElement("button", block.text, block.boundingBox,
                                block.confidence * 0.8f, "ocr")); // Lower confidence for heuristic
                    }
                }
            }
        }

        return elements;
    }

    /**
     * Heuristic to detect if text block is likely clickable
     */
    private boolean isLikelyClickable(MlKitOcrEngine.TextBlock block) {
        if (block.text == null || block.boundingBox == null) return false;

        String text = block.text.trim().toLowerCase();

        // Common button texts
        String[] buttonKeywords = {
            "确定", "取消", "提交", "确认", "支付", "购买", "下单", "结算",
            "登录", "注册", "搜索", "发送", "保存", "删除", "关闭",
            "ok", "cancel", "submit", "confirm", "pay", "buy", "order",
            "login", "register", "search", "send", "save", "delete", "close",
            "下一步", "上一步", "继续", "完成", "跳过", "next", "back", "skip", "done"
        };

        for (String keyword : buttonKeywords) {
            if (text.contains(keyword)) {
                return true;
            }
        }

        // Short text is often clickable
        if (text.length() <= 4) {
            return true;
        }

        return false;
    }

    /**
     * Find all text in screen
     */
    @NonNull
    public List<VisualElement> findAllText(@Nullable Bitmap screenshot) {
        List<VisualElement> elements = new ArrayList<>();

        if (screenshot == null) {
            screenshot = engine.takeScreenshot();
        }

        if (screenshot != null) {
            MlKitOcrEngine.OcrResult ocrResult = ocrEngine.recognizeText(screenshot);
            if (ocrResult.success && ocrResult.textBlocks != null) {
                for (MlKitOcrEngine.TextBlock block : ocrResult.textBlocks) {
                    elements.add(new VisualElement("text", block.text, block.boundingBox,
                            block.confidence, "ocr"));
                }
            }
        }

        return elements;
    }

    /**
     * Find element near coordinates
     */
    @Nullable
    public VisualElement findNearPoint(int x, int y, int radius, @Nullable Bitmap screenshot) {
        List<VisualElement> allElements = findAllText(screenshot);

        VisualElement nearest = null;
        float nearestDist = Float.MAX_VALUE;

        for (VisualElement element : allElements) {
            if (element.bounds == null) continue;

            int cx = element.bounds.centerX();
            int cy = element.bounds.centerY();

            float dist = (float) Math.sqrt(Math.pow(cx - x, 2) + Math.pow(cy - y, 2));

            if (dist <= radius && dist < nearestDist) {
                nearestDist = dist;
                nearest = element;
            }
        }

        return nearest;
    }

    /**
     * Find element matching selector
     */
    @NonNull
    public SearchResult find(@NonNull BySelector selector, @Nullable Bitmap screenshot) {
        // Check text-based selector
        if (selector.getText() != null) {
            return findByText(selector.getText(), screenshot);
        }

        // Check text contains
        if (selector.getTextContains() != null) {
            String searchText = selector.getTextContains();
            List<VisualElement> elements = new ArrayList<>();

            if (screenshot != null) {
                List<MlKitOcrEngine.TextBlock> blocks = ocrEngine.findText(screenshot, searchText);
                for (MlKitOcrEngine.TextBlock block : blocks) {
                    if (block.boundingBox != null) {
                        elements.add(new VisualElement("text", block.text, block.boundingBox,
                                block.confidence, "ocr"));
                    }
                }
            }

            if (!elements.isEmpty()) {
                return new SearchResult(true, elements, "Found " + elements.size() + " elements");
            }
        }

        // Check description-based selector
        if (selector.getContentDescription() != null) {
            // Try accessibility tree first
            // Then fall back to OCR
            return findByText(selector.getContentDescription(), screenshot);
        }

        return SearchResult.notFound("No matching element for selector");
    }

    /**
     * Wait for element to appear
     */
    @Nullable
    public VisualElement waitForElement(@NonNull BySelector selector, long timeoutMs) {
        long startTime = System.currentTimeMillis();
        long intervalMs = 500;

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            Bitmap screenshot = engine.takeScreenshot();
            if (screenshot != null) {
                SearchResult result = find(selector, screenshot);
                if (result.found && !result.elements.isEmpty()) {
                    screenshot.recycle();
                    return result.elements.get(0);
                }
                screenshot.recycle();
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

    /**
     * Get element at point
     */
    @Nullable
    public VisualElement getElementAtPoint(int x, int y, @Nullable Bitmap screenshot) {
        if (screenshot == null) {
            screenshot = engine.takeScreenshot();
        }

        if (screenshot == null) return null;

        List<VisualElement> elements = findAllText(screenshot);

        for (VisualElement element : elements) {
            if (element.bounds != null && element.bounds.contains(x, y)) {
                return element;
            }
        }

        return null;
    }

    // ===== Configuration =====

    public void setPreferAccessibility(boolean prefer) {
        this.preferAccessibility = prefer;
    }

    public void setVisualFallbackThreshold(float threshold) {
        this.visualFallbackThreshold = threshold;
    }

    public MlKitOcrEngine getOcrEngine() {
        return ocrEngine;
    }

    public TemplateMatcher getTemplateMatcher() {
        return templateMatcher;
    }
}