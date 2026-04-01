package com.ofa.agent.ai;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPServer;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Tool Calling Adapter - adapts OFA tools for AI agent consumption.
 * Provides OpenAI-compatible function calling interface.
 */
public class ToolCallingAdapter implements AIAgentInterface {

    private static final String TAG = "ToolCallingAdapter";

    private final MCPServer mcpServer;
    private final ExecutorService executor;
    private Map<String, Object> executionContext;

    /**
     * Create adapter with MCP Server
     */
    public ToolCallingAdapter(@NonNull MCPServer mcpServer) {
        this.mcpServer = mcpServer;
        this.executor = Executors.newCachedThreadPool();
        this.executionContext = new HashMap<>();
    }

    @NonNull
    @Override
    public List<ToolDefinition> getAvailableTools() {
        return mcpServer.listTools();
    }

    @NonNull
    @Override
    public ToolResult callTool(@NonNull String toolName, @NonNull Map<String, Object> args) {
        Log.d(TAG, "Calling tool: " + toolName);
        return mcpServer.callTool(toolName, args);
    }

    @Override
    public void callToolAsync(@NonNull String toolName, @NonNull Map<String, Object> args,
                              @NonNull ToolCallback callback) {
        executor.execute(() -> {
            try {
                ToolResult result = callTool(toolName, args);
                callback.onSuccess(result);
            } catch (Exception e) {
                callback.onError(e.getMessage());
            }
        });
    }

    @Override
    public boolean isToolAvailable(@NonNull String toolName) {
        ToolDefinition def = mcpServer.getTool(toolName);
        return def != null;
    }

    @Nullable
    @Override
    public ToolDefinition getToolDefinition(@NonNull String toolName) {
        return mcpServer.getTool(toolName);
    }

    @NonNull
    @Override
    public List<ToolSuggestion> suggestTools(@NonNull String context) {
        List<ToolSuggestion> suggestions = new ArrayList<>();
        String contextLower = context.toLowerCase();

        // Simple keyword-based suggestions
        if (contextLower.contains("photo") || contextLower.contains("camera") || contextLower.contains("picture")) {
            suggestions.add(new ToolSuggestion("camera.capture", 0.8,
                    "Context mentions camera/photo", null));
        }

        if (contextLower.contains("call") || contextLower.contains("phone") || contextLower.contains("contact")) {
            suggestions.add(new ToolSuggestion("contacts.query", 0.7,
                    "Context mentions contacts", null));
        }

        if (contextLower.contains("calendar") || contextLower.contains("event") || contextLower.contains("meeting")) {
            suggestions.add(new ToolSuggestion("calendar.query", 0.8,
                    "Context mentions calendar/events", null));
        }

        if (contextLower.contains("wifi") || contextLower.contains("network")) {
            suggestions.add(new ToolSuggestion("wifi.scan", 0.7,
                    "Context mentions WiFi/network", null));
        }

        if (contextLower.contains("bluetooth") || contextLower.contains("pair")) {
            suggestions.add(new ToolSuggestion("bluetooth.scan", 0.7,
                    "Context mentions Bluetooth", null));
        }

        if (contextLower.contains("battery") || contextLower.contains("power") || contextLower.contains("charge")) {
            suggestions.add(new ToolSuggestion("battery.status", 0.8,
                    "Context mentions battery/power", null));
        }

        if (contextLower.contains("speak") || contextLower.contains("say") || contextLower.contains("voice")) {
            suggestions.add(new ToolSuggestion("speech.synthesize", 0.7,
                    "Context mentions speech/voice", null));
        }

        if (contextLower.contains("notify") || contextLower.contains("alert")) {
            suggestions.add(new ToolSuggestion("notification.send", 0.8,
                    "Context mentions notification/alert", null));
        }

        if (contextLower.contains("file") || contextLower.contains("document")) {
            suggestions.add(new ToolSuggestion("file.list", 0.6,
                    "Context mentions files", null));
        }

        if (contextLower.contains("app") || contextLower.contains("application")) {
            suggestions.add(new ToolSuggestion("app.list", 0.6,
                    "Context mentions apps", null));
        }

        return suggestions;
    }

    @NonNull
    @Override
    public JSONArray getToolsAsFunctions() {
        JSONArray functions = new JSONArray();

        for (ToolDefinition tool : mcpServer.listTools()) {
            JSONObject function = new JSONObject();
            try {
                function.put("name", sanitizeFunctionName(tool.getName()));
                function.put("description", tool.getDescription());

                JSONObject parameters = convertSchemaToOpenAI(tool.getInputSchema());
                function.put("parameters", parameters);

                functions.put(function);
            } catch (Exception e) {
                Log.e(TAG, "Error converting tool: " + tool.getName(), e);
            }
        }

        return functions;
    }

    @Override
    public void setExecutionContext(@Nullable Map<String, Object> context) {
        this.executionContext = context != null ? new HashMap<>(context) : new HashMap<>();
    }

    /**
     * Convert tool result to OpenAI-compatible format
     */
    @NonNull
    public JSONObject convertResultToMessage(@NonNull ToolResult result) {
        JSONObject message = new JSONObject();
        try {
            message.put("role", "tool");
            message.put("name", sanitizeFunctionName(result.getToolName()));
            message.put("content", result.isSuccess()
                    ? result.getOutput().toString()
                    : "{\"error\": \"" + result.getError() + "\"}");
        } catch (Exception e) {
            Log.e(TAG, "Error converting result", e);
        }
        return message;
    }

    /**
     * Create function call message for OpenAI format
     */
    @NonNull
    public JSONObject createFunctionCall(@NonNull String toolName,
                                          @NonNull Map<String, Object> args) {
        JSONObject message = new JSONObject();
        try {
            JSONObject functionCall = new JSONObject();
            functionCall.put("name", sanitizeFunctionName(toolName));
            functionCall.put("arguments", new JSONObject(args).toString());

            message.put("role", "assistant");
            message.put("function_call", functionCall);
        } catch (Exception e) {
            Log.e(TAG, "Error creating function call", e);
        }
        return message;
    }

    /**
     * Parse OpenAI function call response
     */
    @Nullable
    public FunctionCallInfo parseFunctionCall(@NonNull JSONObject message) {
        try {
            if (!message.has("function_call")) return null;

            JSONObject fc = message.getJSONObject("function_call");
            String name = fc.getString("name");
            String argsStr = fc.getString("arguments");

            JSONObject argsJson = new JSONObject(argsStr);
            Map<String, Object> args = new HashMap<>();

            for (java.util.Iterator<String> it = argsJson.keys(); it.hasNext(); ) {
                String key = it.next();
                args.put(key, argsJson.get(key));
            }

            return new FunctionCallInfo(desanitizeFunctionName(name), args);
        } catch (Exception e) {
            Log.e(TAG, "Error parsing function call", e);
            return null;
        }
    }

    // ===== Helper Methods =====

    @NonNull
    private String sanitizeFunctionName(@NonNull String name) {
        // Replace dots with underscores for OpenAI compatibility
        return name.replace(".", "_");
    }

    @NonNull
    private String desanitizeFunctionName(@NonNull String name) {
        // Convert back to original format
        // This is a simple heuristic - tools like app.launch become app_launch
        // We try to restore the dot
        int lastUnderscore = name.lastIndexOf('_');
        if (lastUnderscore > 0) {
            String category = name.substring(0, lastUnderscore);
            String action = name.substring(lastUnderscore + 1);
            // Check if this looks like a category.action pattern
            if (isKnownCategory(category)) {
                return category + "." + action;
            }
        }
        return name;
    }

    private boolean isKnownCategory(@NonNull String category) {
        String[] knownCategories = {"app", "settings", "clipboard", "file", "notification",
                "camera", "bluetooth", "wifi", "nfc", "sensor", "battery",
                "contacts", "calendar", "media", "speech"};
        for (String c : knownCategories) {
            if (c.equals(category)) return true;
        }
        return false;
    }

    @NonNull
    private JSONObject convertSchemaToOpenAI(@NonNull JSONObject schema) {
        JSONObject openAISchema = new JSONObject();
        try {
            // Copy basic properties
            openAISchema.put("type", schema.optString("type", "object"));

            if (schema.has("properties")) {
                openAISchema.put("properties", schema.getJSONObject("properties"));
            }

            if (schema.has("required")) {
                openAISchema.put("required", schema.getJSONArray("required"));
            }
        } catch (Exception e) {
            Log.e(TAG, "Error converting schema", e);
        }
        return openAISchema;
    }

    /**
     * Shutdown the adapter
     */
    public void shutdown() {
        executor.shutdown();
    }

    /**
     * Function call info
     */
    public static class FunctionCallInfo {
        public final String toolName;
        public final Map<String, Object> arguments;

        public FunctionCallInfo(@NonNull String toolName, @NonNull Map<String, Object> arguments) {
            this.toolName = toolName;
            this.arguments = arguments;
        }
    }
}