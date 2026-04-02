package com.ofa.agent.ai.vision;

import android.content.Context;
import android.graphics.Bitmap;
import android.graphics.Rect;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.ai.LocalAIEngine;
import com.ofa.agent.ai.InferenceConfig;
import com.ofa.agent.automation.AutomationNode;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * UI Element Recognizer - identifies UI elements from screenshots.
 * Uses on-device vision models to understand screen content.
 */
public class UIElementRecognizer {

    private static final String TAG = "UIElementRecognizer";

    private static final String MODEL_ID = "ui_element";

    private final Context context;
    private final LocalAIEngine engine;

    private boolean initialized = false;
    private float confidenceThreshold = 0.5f;

    /**
     * Recognized UI element
     */
    public static class RecognizedElement {
        public final String type;
        public final String text;
        public final Rect bounds;
        public final float confidence;
        public final boolean isClickable;
        public final boolean isEditable;

        public RecognizedElement(String type, String text, Rect bounds, float confidence,
                                  boolean isClickable, boolean isEditable) {
            this.type = type;
            this.text = text;
            this.bounds = bounds;
            this.confidence = confidence;
            this.isClickable = isClickable;
            this.isEditable = isEditable;
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("Element{type=%s, text='%s', bounds=%s, conf=%.2f}",
                type, text, bounds, confidence);
        }
    }

    /**
     * Element type constants
     */
    public static final String TYPE_BUTTON = "button";
    public static final String TYPE_TEXT = "text";
    public static final String TYPE_IMAGE = "image";
    public static final String TYPE_INPUT = "input";
    public static final String TYPE_LIST = "list";
    public static final String TYPE_ICON = "icon";

    /**
     * Create UI element recognizer
     */
    public UIElementRecognizer(@NonNull Context context) {
        this.context = context;
        this.engine = new LocalAIEngine(context);
    }

    /**
     * Initialize recognizer
     */
    public void initialize() {
        Log.i(TAG, "Initializing UI Element Recognizer...");

        InferenceConfig config = InferenceConfig.builder()
            .requiredModels(MODEL_ID)
            .strictMode(false)
            .imageSize(320, 320)
            .build();

        engine.initialize(config);
        initialized = engine.isReady();

        Log.i(TAG, "Initialized: " + initialized);
    }

    /**
     * Set confidence threshold
     */
    public void setConfidenceThreshold(float threshold) {
        this.confidenceThreshold = Math.max(0, Math.min(1, threshold));
    }

    /**
     * Recognize elements from screenshot
     */
    @NonNull
    public List<RecognizedElement> recognize(@NonNull Bitmap screenshot) {
        List<RecognizedElement> elements = new ArrayList<>();

        if (!initialized) {
            Log.w(TAG, "Recognizer not initialized");
            return elements;
        }

        // Run inference
        LocalAIEngine.InferenceResult result = engine.inferImage(MODEL_ID, screenshot);

        if (!result.success) {
            Log.w(TAG, "Inference failed: " + result.error);
            return elements;
        }

        // Parse results
        // This is a simplified version - real implementation would parse detection output
        elements.addAll(parseInferenceResult(result, screenshot.getWidth(), screenshot.getHeight()));

        Log.d(TAG, "Recognized " + elements.size() + " elements");
        return elements;
    }

    /**
     * Parse inference result into elements
     */
    @NonNull
    private List<RecognizedElement> parseInferenceResult(@NonNull LocalAIEngine.InferenceResult result,
                                                          int imageWidth, int imageHeight) {
        List<RecognizedElement> elements = new ArrayList<>();

        // Placeholder: create mock elements based on prediction
        // Real implementation would parse bounding boxes from model output

        if (result.extras != null && result.extras.containsKey("boxes")) {
            try {
                JSONArray boxes = new JSONArray(result.extras.get("boxes"));
                for (int i = 0; i < boxes.length(); i++) {
                    JSONObject box = boxes.getJSONObject(i);
                    RecognizedElement element = parseElement(box, imageWidth, imageHeight);
                    if (element != null && element.confidence >= confidenceThreshold) {
                        elements.add(element);
                    }
                }
            } catch (Exception e) {
                Log.w(TAG, "Failed to parse boxes: " + e.getMessage());
            }
        }

        return elements;
    }

    /**
     * Parse element from JSON
     */
    @Nullable
    private RecognizedElement parseElement(@NonNull JSONObject json, int width, int height) {
        try {
            String type = json.optString("type", TYPE_TEXT);
            String text = json.optString("text", "");

            int left = (int) (json.optDouble("left", 0) * width);
            int top = (int) (json.optDouble("top", 0) * height);
            int right = (int) (json.optDouble("right", 0.1) * width);
            int bottom = (int) (json.optDouble("bottom", 0.1) * height);

            float confidence = (float) json.optDouble("confidence", 0.5);
            boolean clickable = json.optBoolean("clickable", false);
            boolean editable = json.optBoolean("editable", false);

            return new RecognizedElement(type, text, new Rect(left, top, right, bottom),
                confidence, clickable, editable);

        } catch (Exception e) {
            return null;
        }
    }

    /**
     * Find element by type
     */
    @NonNull
    public List<RecognizedElement> findByType(@NonNull List<RecognizedElement> elements,
                                               @NonNull String type) {
        List<RecognizedElement> filtered = new ArrayList<>();
        for (RecognizedElement element : elements) {
            if (element.type.equals(type)) {
                filtered.add(element);
            }
        }
        return filtered;
    }

    /**
     * Find clickable elements
     */
    @NonNull
    public List<RecognizedElement> findClickable(@NonNull List<RecognizedElement> elements) {
        List<RecognizedElement> clickable = new ArrayList<>();
        for (RecognizedElement element : elements) {
            if (element.isClickable) {
                clickable.add(element);
            }
        }
        return clickable;
    }

    /**
     * Find element containing text
     */
    @Nullable
    public RecognizedElement findContainingText(@NonNull List<RecognizedElement> elements,
                                                 @NonNull String text) {
        for (RecognizedElement element : elements) {
            if (element.text != null && element.text.toLowerCase().contains(text.toLowerCase())) {
                return element;
            }
        }
        return null;
    }

    /**
     * Find element at position
     */
    @Nullable
    public RecognizedElement findAtPosition(@NonNull List<RecognizedElement> elements,
                                             int x, int y) {
        for (RecognizedElement element : elements) {
            if (element.bounds.contains(x, y)) {
                return element;
            }
        }
        return null;
    }

    /**
     * Convert AutomationNode to RecognizedElement
     */
    @NonNull
    public static RecognizedElement fromNode(@NonNull AutomationNode node) {
        String type = inferType(node);
        boolean clickable = node.isClickable();
        boolean editable = node.isEditable();

        Rect bounds = new Rect(node.getBounds());

        return new RecognizedElement(type, node.getText(), bounds, 1.0f, clickable, editable);
    }

    /**
     * Infer element type from node
     */
    @NonNull
    private static String inferType(@NonNull AutomationNode node) {
        String className = node.getClassName();
        if (className == null) return TYPE_TEXT;

        className = className.toLowerCase();

        if (className.contains("button")) return TYPE_BUTTON;
        if (className.contains("edittext") || className.contains("input")) return TYPE_INPUT;
        if (className.contains("image")) return TYPE_IMAGE;
        if (className.contains("list") || className.contains("recycler")) return TYPE_LIST;
        if (className.contains("imagebutton") || className.contains("icon")) return TYPE_ICON;

        return TYPE_TEXT;
    }

    /**
     * Check if ready
     */
    public boolean isReady() {
        return initialized;
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        engine.shutdown();
        initialized = false;
    }
}