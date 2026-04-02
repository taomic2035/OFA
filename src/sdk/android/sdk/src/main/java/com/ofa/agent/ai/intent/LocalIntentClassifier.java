package com.ofa.agent.ai.intent;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.ai.LocalAIEngine;
import com.ofa.agent.ai.InferenceConfig;
import com.ofa.agent.intent.UserIntent;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Local Intent Classifier - classifies user intents using on-device model.
 * Provides fast, offline intent recognition.
 */
public class LocalIntentClassifier {

    private static final String TAG = "LocalIntentClassifier";

    private static final String MODEL_ID = "intent_classifier";

    private final Context context;
    private final LocalAIEngine engine;
    private final Map<String, IntentDefinition> intentDefinitions;

    private boolean initialized = false;
    private float confidenceThreshold = 0.6f;

    /**
     * Intent definition for local classification
     */
    public static class IntentDefinition {
        public final String name;
        public final String[] keywords;
        public final String[] patterns;
        public final String[] slots;

        public IntentDefinition(String name, String[] keywords, String[] patterns, String[] slots) {
            this.name = name;
            this.keywords = keywords;
            this.patterns = patterns;
            this.slots = slots;
        }
    }

    /**
     * Create local intent classifier
     */
    public LocalIntentClassifier(@NonNull Context context) {
        this.context = context;
        this.engine = new LocalAIEngine(context);
        this.intentDefinitions = new HashMap<>();

        registerDefaultIntents();
    }

    /**
     * Initialize classifier
     */
    public void initialize() {
        Log.i(TAG, "Initializing Local Intent Classifier...");

        InferenceConfig config = InferenceConfig.builder()
            .requiredModels(MODEL_ID)
            .strictMode(false)
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
     * Register intent definition
     */
    public void registerIntent(@NonNull IntentDefinition definition) {
        intentDefinitions.put(definition.name, definition);
        Log.d(TAG, "Registered intent: " + definition.name);
    }

    /**
     * Register default intents
     */
    private void registerDefaultIntents() {
        registerIntent(new IntentDefinition(
            "search",
            new String[]{"搜索", "查找", "找", "查"},
            new String[]{"搜索(.+)", "查找(.+)", "找(.+)"},
            new String[]{"query"}
        ));

        registerIntent(new IntentDefinition(
            "app_launch",
            new String[]{"打开", "启动", "运行"},
            new String[]{"打开(.+)", "启动(.+)"},
            new String[]{"app_name"}
        ));

        registerIntent(new IntentDefinition(
            "order_food",
            new String[]{"点餐", "外卖", "订餐", "点外卖"},
            new String[]{"点(.+)外卖", "订(.+)"},
            new String[]{"food_type", "restaurant"}
        ));

        registerIntent(new IntentDefinition(
            "shopping",
            new String[]{"购物", "买东西", "购买"},
            new String[]{"买(.+)", "购买(.+)"},
            new String[]{"product"}
        ));

        registerIntent(new IntentDefinition(
            "setting",
            new String[]{"设置", "修改", "更改"},
            new String[]{"设置(.+)", "修改(.+)"},
            new String[]{"setting_name", "value"}
        ));

        Log.d(TAG, "Registered " + intentDefinitions.size() + " default intents");
    }

    /**
     * Classify text and return intent
     */
    @Nullable
    public UserIntent classify(@NonNull String text) {
        if (!initialized) {
            Log.w(TAG, "Classifier not initialized");
            return fallbackClassify(text);
        }

        // Run inference
        LocalAIEngine.InferenceResult result = engine.inferText(MODEL_ID, text);

        if (!result.success) {
            Log.w(TAG, "Inference failed: " + result.error);
            return fallbackClassify(text);
        }

        // Check confidence
        if (result.confidence < confidenceThreshold) {
            Log.d(TAG, "Low confidence: " + result.confidence + " for: " + result.prediction);
            return fallbackClassify(text);
        }

        // Build UserIntent
        String intentName = result.prediction;
        Map<String, String> slots = extractSlots(text, intentName);

        return new UserIntent(
            intentName,
            text,
            slots,
            result.confidence
        );
    }

    /**
     * Fallback classification using rules
     */
    @Nullable
    private UserIntent fallbackClassify(@NonNull String text) {
        String lowerText = text.toLowerCase();

        for (IntentDefinition def : intentDefinitions.values()) {
            // Check keywords
            for (String keyword : def.keywords) {
                if (lowerText.contains(keyword.toLowerCase())) {
                    Map<String, String> slots = extractSlots(text, def.name);
                    return new UserIntent(def.name, text, slots, 0.7f);
                }
            }

            // Check patterns
            for (String pattern : def.patterns) {
                try {
                    java.util.regex.Pattern p = java.util.regex.Pattern.compile(pattern);
                    java.util.regex.Matcher m = p.matcher(text);
                    if (m.find()) {
                        Map<String, String> slots = extractSlots(text, def.name);
                        return new UserIntent(def.name, text, slots, 0.8f);
                    }
                } catch (Exception e) {
                    // Invalid pattern, skip
                }
            }
        }

        return null;
    }

    /**
     * Extract slots from text
     */
    @NonNull
    private Map<String, String> extractSlots(@NonNull String text, @NonNull String intentName) {
        Map<String, String> slots = new HashMap<>();

        IntentDefinition def = intentDefinitions.get(intentName);
        if (def == null) return slots;

        // Pattern-based extraction
        for (String pattern : def.patterns) {
            try {
                java.util.regex.Pattern p = java.util.regex.Pattern.compile(pattern);
                java.util.regex.Matcher m = p.matcher(text);

                if (m.find() && m.groupCount() > 0) {
                    // Use first captured group as primary slot
                    if (def.slots.length > 0) {
                        slots.put(def.slots[0], m.group(1));
                    }
                }
            } catch (Exception e) {
                // Skip invalid pattern
            }
        }

        // Named entity extraction
        extractEntities(text, slots);

        return slots;
    }

    /**
     * Extract named entities
     */
    private void extractEntities(@NonNull String text, @NonNull Map<String, String> slots) {
        // Time extraction
        java.util.regex.Pattern timePattern = java.util.regex.Pattern.compile("(\\d{1,2})[点:时](\\d{0,2})分?");
        java.util.regex.Matcher timeMatcher = timePattern.matcher(text);
        if (timeMatcher.find()) {
            slots.put("time", timeMatcher.group());
        }

        // Number extraction
        java.util.regex.Pattern numPattern = java.util.regex.Pattern.compile("(\\d+)(个|份|杯|件)");
        java.util.regex.Matcher numMatcher = numPattern.matcher(text);
        if (numMatcher.find()) {
            slots.put("quantity", numMatcher.group(1));
        }

        // Price extraction
        java.util.regex.Pattern pricePattern = java.util.regex.Pattern.compile("(\\d+)元");
        java.util.regex.Matcher priceMatcher = pricePattern.matcher(text);
        if (priceMatcher.find()) {
            slots.put("price", priceMatcher.group(1));
        }
    }

    /**
     * Get multiple intent hypotheses
     */
    @NonNull
    public List<UserIntent> classifyWithAlternatives(@NonNull String text, int topK) {
        List<UserIntent> results = new ArrayList<>();

        if (!initialized) {
            UserIntent fallback = fallbackClassify(text);
            if (fallback != null) {
                results.add(fallback);
            }
            return results;
        }

        LocalAIEngine.InferenceResult result = engine.inferText(MODEL_ID, text);

        if (result.success && result.probabilities != null) {
            // Sort by probability
            List<Map.Entry<String, Float>> sorted = new ArrayList<>();
            for (Map.Entry<String, Float> entry : result.probabilities.entrySet()) {
                sorted.add(entry);
            }
            sorted.sort((a, b) -> Float.compare(b.getValue(), a.getValue()));

            // Take top K
            int count = 0;
            for (Map.Entry<String, Float> entry : sorted) {
                if (count >= topK) break;
                if (entry.getValue() >= confidenceThreshold * 0.5f) {
                    Map<String, String> slots = extractSlots(text, entry.getKey());
                    results.add(new UserIntent(entry.getKey(), text, slots, entry.getValue()));
                    count++;
                }
            }
        }

        return results;
    }

    /**
     * Check if ready
     */
    public boolean isReady() {
        return initialized;
    }

    /**
     * Get registered intent names
     */
    @NonNull
    public String[] getIntentNames() {
        return intentDefinitions.keySet().toArray(new String[0]);
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        engine.shutdown();
        initialized = false;
    }
}