package com.ofa.agent.automation.integration;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.intent.IntentEngine;
import com.ofa.agent.intent.UserIntent;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Intent-Triggered Automation - triggers automation based on intent recognition.
 * Bridges natural language intents to automation actions.
 */
public class IntentTriggeredAutomation {

    private static final String TAG = "IntentTriggeredAutomation";

    private final Context context;
    private final AutomationEngine engine;
    private final IntentEngine intentEngine;
    private final ExecutorService executor;
    private final Map<String, IntentHandler> handlers;

    private AutomationCallback callback;

    /**
     * Callback for automation events
     */
    public interface AutomationCallback {
        void onIntentRecognized(@NonNull UserIntent intent);
        void onAutomationStarted(@NonNull String action);
        void onAutomationComplete(@NonNull String action, @NonNull AutomationResult result);
        void onAutomationError(@NonNull String action, @NonNull String error);
    }

    /**
     * Intent handler interface
     */
    public interface IntentHandler {
        @NonNull
        AutomationResult handle(@NonNull AutomationEngine engine, @NonNull UserIntent intent);
    }

    /**
     * Create intent-triggered automation
     */
    public IntentTriggeredAutomation(@NonNull Context context,
                                      @NonNull AutomationEngine engine,
                                      @NonNull IntentEngine intentEngine) {
        this.context = context;
        this.engine = engine;
        this.intentEngine = intentEngine;
        this.executor = Executors.newSingleThreadExecutor();
        this.handlers = new HashMap<>();

        registerDefaultHandlers();
    }

    /**
     * Set callback
     */
    public void setCallback(@Nullable AutomationCallback callback) {
        this.callback = callback;
    }

    /**
     * Register intent handler
     */
    public void registerHandler(@NonNull String intentName, @NonNull IntentHandler handler) {
        handlers.put(intentName, handler);
        Log.i(TAG, "Registered handler for intent: " + intentName);
    }

    /**
     * Process natural language input and trigger automation
     */
    @NonNull
    public AutomationResult processAndExecute(@NonNull String input) {
        Log.i(TAG, "Processing input: " + input);

        // Recognize intent
        UserIntent intent = intentEngine.recognize(input);

        if (intent == null) {
            Log.w(TAG, "No intent recognized for input: " + input);
            return new AutomationResult("intent", "Could not understand: " + input);
        }

        Log.i(TAG, "Recognized intent: " + intent.getIntentName() + " (confidence: " + intent.getConfidence() + ")");

        if (callback != null) {
            callback.onIntentRecognized(intent);
        }

        // Execute automation based on intent
        return executeIntent(intent);
    }

    /**
     * Execute automation for recognized intent
     */
    @NonNull
    public AutomationResult executeIntent(@NonNull UserIntent intent) {
        String intentName = intent.getIntentName();

        Log.i(TAG, "Executing automation for intent: " + intentName);

        IntentHandler handler = handlers.get(intentName);
        if (handler == null) {
            Log.w(TAG, "No handler registered for intent: " + intentName);
            return new AutomationResult(intentName, "No automation handler for: " + intentName);
        }

        if (callback != null) {
            callback.onAutomationStarted(intentName);
        }

        try {
            AutomationResult result = handler.handle(engine, intent);

            if (callback != null) {
                if (result.isSuccess()) {
                    callback.onAutomationComplete(intentName, result);
                } else {
                    callback.onAutomationError(intentName, result.getMessage());
                }
            }

            return result;
        } catch (Exception e) {
            Log.e(TAG, "Automation error: " + e.getMessage(), e);

            if (callback != null) {
                callback.onAutomationError(intentName, e.getMessage());
            }

            return new AutomationResult(intentName, "Automation error: " + e.getMessage());
        }
    }

    /**
     * Process input asynchronously
     */
    public void processAsync(@NonNull String input) {
        executor.execute(() -> processAndExecute(input));
    }

    /**
     * Register default intent handlers
     */
    private void registerDefaultHandlers() {
        // App launch handler
        registerHandler("app_launch", (engine, intent) -> {
            String appName = intent.getSlot("app_name");
            if (appName != null) {
                // Use app launch tool
                return new AutomationResult("app_launch", 0);
            }
            return new AutomationResult("app_launch", "Missing app name");
        });

        // Search handler
        registerHandler("search_query", (engine, intent) -> {
            String query = intent.getSlot("query");
            if (query != null) {
                return engine.inputText(query);
            }
            return new AutomationResult("search_query", "Missing query");
        });

        // Settings handler
        registerHandler("setting_change", (engine, intent) -> {
            String setting = intent.getSlot("setting_name");
            String value = intent.getSlot("setting_value");
            if (setting != null && value != null) {
                // Handle setting change
                return new AutomationResult("setting_change", 0);
            }
            return new AutomationResult("setting_change", "Missing setting parameters");
        });

        Log.i(TAG, "Registered " + handlers.size() + " default handlers");
    }

    /**
     * Get registered intent names
     */
    @NonNull
    public String[] getRegisteredIntents() {
        return handlers.keySet().toArray(new String[0]);
    }

    /**
     * Check if intent has handler
     */
    public boolean hasHandler(@NonNull String intentName) {
        return handlers.containsKey(intentName);
    }

    /**
     * Unregister handler
     */
    public void unregisterHandler(@NonNull String intentName) {
        handlers.remove(intentName);
        Log.i(TAG, "Unregistered handler for intent: " + intentName);
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        executor.shutdown();
        handlers.clear();
        Log.i(TAG, "IntentTriggeredAutomation shutdown");
    }

    /**
     * Get intent suggestions for partial input
     */
    @NonNull
    public String[] getIntentSuggestions(@NonNull String partialInput) {
        // Use intent engine to get suggestions
        return intentEngine.getSuggestions(partialInput);
    }

    /**
     * Validate intent before execution
     */
    public boolean validateIntent(@NonNull UserIntent intent) {
        if (intent.getConfidence() < 0.5f) {
            Log.w(TAG, "Intent confidence too low: " + intent.getConfidence());
            return false;
        }

        return hasHandler(intent.getIntentName());
    }

    /**
     * Build automation context from intent
     */
    @NonNull
    public Map<String, String> buildContext(@NonNull UserIntent intent) {
        Map<String, String> context = new HashMap<>();
        context.put("intent", intent.getIntentName());
        context.put("confidence", String.valueOf(intent.getConfidence()));

        for (Map.Entry<String, String> slot : intent.getSlots().entrySet()) {
            context.put("slot_" + slot.getKey(), slot.getValue());
        }

        return context;
    }
}