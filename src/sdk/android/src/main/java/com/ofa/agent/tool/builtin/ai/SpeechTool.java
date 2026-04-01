package com.ofa.agent.tool.builtin.ai;

import android.content.Context;
import android.speech.tts.TextToSpeech;
import android.speech.tts.UtteranceProgressListener;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.Locale;
import java.util.Map;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;

/**
 * Speech Tool - text-to-speech synthesis.
 */
public class SpeechTool implements ToolExecutor {

    private static final String TAG = "SpeechTool";
    private static final int TTS_TIMEOUT = 30000; // 30 seconds

    private final Context context;
    private TextToSpeech textToSpeech;
    private volatile boolean ttsReady = false;

    public SpeechTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        initTTS();
    }

    private void initTTS() {
        textToSpeech = new TextToSpeech(context, status -> {
            ttsReady = (status == TextToSpeech.SUCCESS);
            Log.i(TAG, "TTS initialized: " + ttsReady);
        });
    }

    @NonNull
    @Override
    public String getToolId() {
        return "speech";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Text-to-speech synthesis";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "speak");

        switch (operation.toLowerCase()) {
            case "speak":
                return executeSpeak(args, ctx);
            case "stop":
                return executeStop(ctx);
            case "status":
                return executeStatus(ctx);
            case "languages":
                return executeLanguages(ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return textToSpeech != null && ttsReady;
    }

    @Override
    public boolean requiresAuth() {
        return false;
    }

    @Override
    public boolean supportsOffline() {
        return true;
    }

    @Nullable
    @Override
    public String[] getRequiredPermissions() {
        return null;
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        if (operation == null) return false;

        if ("speak".equalsIgnoreCase(operation)) {
            return args.containsKey("text");
        }

        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 3000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getSpeakDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'speak'");
            operation.put("default", "speak");
            props.put("operation", operation);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to speak");
            props.put("text", text);

            JSONObject language = new JSONObject();
            language.put("type", "string");
            language.put("description", "Language code (e.g., 'en-US', 'zh-CN')");
            language.put("default", "en-US");
            props.put("language", language);

            JSONObject pitch = new JSONObject();
            pitch.put("type", "number");
            pitch.put("description", "Speech pitch (0.5 - 2.0)");
            pitch.put("default", 1.0);
            props.put("pitch", pitch);

            JSONObject rate = new JSONObject();
            rate.put("type", "number");
            rate.put("description", "Speech rate (0.5 - 2.0)");
            rate.put("default", 1.0);
            props.put("rate", rate);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"text"});
        return new ToolDefinition("speech.synthesize", "Synthesize speech from text",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getStopDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'stop'");
        return new ToolDefinition("speech.stop", "Stop current speech",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getStatusDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'status'");
        return new ToolDefinition("speech.status", "Get TTS status",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeSpeak(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (!ttsReady) {
            return new ToolResult(getToolId(), "TTS not ready");
        }

        String text = getStringArg(args, "text", null);
        if (text == null || text.isEmpty()) {
            return new ToolResult(getToolId(), "Missing text parameter");
        }

        String language = getStringArg(args, "language", Locale.getDefault().toString());
        float pitch = getFloatArg(args, "pitch", 1.0f);
        float rate = getFloatArg(args, "rate", 1.0f);

        try {
            // Set language
            Locale locale = Locale.forLanguageTag(language.replace("_", "-"));
            int langResult = textToSpeech.setLanguage(locale);
            if (langResult == TextToSpeech.LANG_MISSING_DATA || langResult == TextToSpeech.LANG_NOT_SUPPORTED) {
                locale = Locale.getDefault();
                textToSpeech.setLanguage(locale);
            }

            // Set pitch and rate
            textToSpeech.setPitch(pitch);
            textToSpeech.setSpeechRate(rate);

            // Use semaphore to wait for completion
            Semaphore semaphore = new Semaphore(0);
            final boolean[] success = {false};
            final String[] error = {null};

            String utteranceId = "utterance_" + System.currentTimeMillis();

            textToSpeech.setOnUtteranceProgressListener(new UtteranceProgressListener() {
                @Override
                public void onStart(String utteranceId) {
                    // Speaking started
                }

                @Override
                public void onDone(String utteranceId) {
                    success[0] = true;
                    semaphore.release();
                }

                @Override
                public void onError(String utteranceId) {
                    error[0] = "TTS error";
                    semaphore.release();
                }
            });

            // Speak
            textToSpeech.speak(text, TextToSpeech.QUEUE_FLUSH, null, utteranceId);

            // Wait for completion (with timeout)
            boolean completed = semaphore.tryAcquire(TTS_TIMEOUT, TimeUnit.MILLISECONDS);

            JSONObject output = new JSONObject();
            output.put("success", completed && success[0]);
            output.put("text", text);
            output.put("language", locale.toString());
            output.put("pitch", pitch);
            output.put("rate", rate);
            output.put("completed", completed);

            if (error[0] != null) {
                output.put("error", error[0]);
            }

            return new ToolResult(getToolId(), output, 1000);

        } catch (Exception e) {
            Log.e(TAG, "Speak failed", e);
            return new ToolResult(getToolId(), "Speak failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeStop(@NonNull ExecutionContext ctx) {
        if (textToSpeech != null) {
            textToSpeech.stop();
        }

        JSONObject output = new JSONObject();
        output.put("success", true);
        output.put("action", "stopped");

        return new ToolResult(getToolId(), output, 50);
    }

    @NonNull
    private ToolResult executeStatus(@NonNull ExecutionContext ctx) {
        JSONObject output = new JSONObject();
        output.put("success", true);
        output.put("available", ttsReady);

        if (textToSpeech != null) {
            Locale locale = textToSpeech.getVoice() != null
                    ? textToSpeech.getVoice().getLocale()
                    : Locale.getDefault();
            output.put("language", locale.toString());
            output.put("engines", textToSpeech.getEngines().size());
        }

        return new ToolResult(getToolId(), output, 20);
    }

    @NonNull
    private ToolResult executeLanguages(@NonNull ExecutionContext ctx) {
        if (textToSpeech == null) {
            return new ToolResult(getToolId(), "TTS not available");
        }

        org.json.JSONArray languagesArray = new org.json.JSONArray();

        for (TextToSpeech.EngineInfo engine : textToSpeech.getEngines()) {
            try {
                for (Locale locale : Locale.getAvailableLocales()) {
                    int result = textToSpeech.isLanguageAvailable(locale);
                    if (result >= TextToSpeech.LANG_AVAILABLE) {
                        JSONObject langInfo = new JSONObject();
                        langInfo.put("code", locale.toString());
                        langInfo.put("name", locale.getDisplayName());
                        langInfo.put("engine", engine.name);
                        languagesArray.put(langInfo);
                    }
                }
            } catch (Exception e) {
                // Skip
            }
        }

        JSONObject output = new JSONObject();
        output.put("success", true);
        output.put("languages", languagesArray);

        return new ToolResult(getToolId(), output, 500);
    }

    // ===== Helper Methods =====

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }

    private float getFloatArg(@NonNull Map<String, Object> args, @NonNull String key, float defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        if (value instanceof Number) return ((Number) value).floatValue();
        try {
            return Float.parseFloat(value.toString());
        } catch (NumberFormatException e) {
            return defaultVal;
        }
    }

    /**
     * Cleanup resources
     */
    public void shutdown() {
        if (textToSpeech != null) {
            textToSpeech.stop();
            textToSpeech.shutdown();
            textToSpeech = null;
            ttsReady = false;
        }
    }
}