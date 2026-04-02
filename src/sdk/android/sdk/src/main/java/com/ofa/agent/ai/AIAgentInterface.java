package com.ofa.agent.ai;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.List;
import java.util.Map;

/**
 * AI Agent Interface - interface for AI agents to interact with OFA tools.
 * Provides a standardized way for AI models to call tools.
 */
public interface AIAgentInterface {

    /**
     * Get all available tools
     * @return list of tool definitions
     */
    @NonNull
    List<ToolDefinition> getAvailableTools();

    /**
     * Call a tool
     * @param toolName Name of the tool to call
     * @param args Arguments for the tool
     * @return Tool execution result
     */
    @NonNull
    ToolResult callTool(@NonNull String toolName, @NonNull Map<String, Object> args);

    /**
     * Call a tool asynchronously
     * @param toolName Tool name
     * @param args Arguments
     * @param callback Result callback
     */
    void callToolAsync(@NonNull String toolName, @NonNull Map<String, Object> args,
                       @NonNull ToolCallback callback);

    /**
     * Check if a tool is available
     * @param toolName Tool name
     * @return true if available
     */
    boolean isToolAvailable(@NonNull String toolName);

    /**
     * Get tool definition
     * @param toolName Tool name
     * @return Tool definition or null
     */
    @Nullable
    ToolDefinition getToolDefinition(@NonNull String toolName);

    /**
     * Generate tool call suggestions for a given context
     * @param context Context description
     * @return List of suggested tool calls
     */
    @NonNull
    List<ToolSuggestion> suggestTools(@NonNull String context);

    /**
     * Convert tools to AI-friendly format (e.g., OpenAI function calling format)
     * @return JSON array of tool definitions
     */
    @NonNull
    org.json.JSONArray getToolsAsFunctions();

    /**
     * Set execution context (user ID, session info, etc.)
     * @param context Execution context
     */
    void setExecutionContext(@Nullable Map<String, Object> context);

    /**
     * Tool callback interface
     */
    interface ToolCallback {
        void onSuccess(@NonNull ToolResult result);
        void onError(@NonNull String error);
    }

    /**
     * Tool suggestion
     */
    class ToolSuggestion {
        public final String toolName;
        public final double confidence;
        public final String reason;
        public final Map<String, Object> suggestedArgs;

        public ToolSuggestion(@NonNull String toolName, double confidence,
                              @Nullable String reason, @Nullable Map<String, Object> suggestedArgs) {
            this.toolName = toolName;
            this.confidence = confidence;
            this.reason = reason;
            this.suggestedArgs = suggestedArgs;
        }
    }
}