package com.ofa.agent.mcp;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.List;
import java.util.Map;

/**
 * MCP Server interface - provides standard MCP protocol support.
 * Implements Model Context Protocol for AI agent tool interactions.
 */
public interface MCPServer {

    // ===== Tool Management =====

    /**
     * List all available tools
     * @return list of tool definitions
     */
    @NonNull
    List<ToolDefinition> listTools();

    /**
     * Get specific tool definition
     * @param name tool name
     * @return tool definition or null if not found
     */
    @Nullable
    ToolDefinition getTool(@NonNull String name);

    /**
     * Execute a tool with given arguments
     * @param name tool name
     * @param args input arguments as key-value map
     * @return execution result
     */
    @NonNull
    ToolResult callTool(@NonNull String name, @NonNull Map<String, Object> args);

    /**
     * Execute a tool asynchronously
     * @param name tool name
     * @param args input arguments
     * @param callback result callback
     * @return execution ID for tracking/cancellation
     */
    @NonNull
    String callToolAsync(@NonNull String name, @NonNull Map<String, Object> args,
                         @NonNull ToolCallback callback);

    /**
     * Cancel an async tool execution
     * @param executionId execution ID returned from callToolAsync
     * @return true if cancelled successfully
     */
    boolean cancelExecution(@NonNull String executionId);

    // ===== Resource Access =====

    /**
     * List all available resources
     * @return list of resource definitions
     */
    @NonNull
    List<ResourceDefinition> listResources();

    /**
     * Get specific resource definition
     * @param uri resource URI
     * @return resource definition or null
     */
    @Nullable
    ResourceDefinition getResource(@NonNull String uri);

    /**
     * Read resource content
     * @param uri resource URI
     * @return resource content
     */
    @NonNull
    ResourceContent readResource(@NonNull String uri);

    // ===== Prompt Templates =====

    /**
     * List all available prompt templates
     * @return list of prompt definitions
     */
    @NonNull
    List<PromptDefinition> listPrompts();

    /**
     * Get specific prompt template
     * @param name prompt name
     * @return prompt definition or null
     */
    @Nullable
    PromptDefinition getPrompt(@NonNull String name);

    /**
     * Render a prompt with given arguments
     * @param name prompt name
     * @param args template arguments
     * @return rendered prompt result
     */
    @NonNull
    PromptResult getPrompt(@NonNull String name, @Nullable Map<String, Object> args);

    // ===== Server Info =====

    /**
     * Get server information
     * @return server info JSON
     */
    @NonNull
    JSONObject getServerInfo();

    /**
     * Check if server is ready
     * @return true if server is initialized and ready
     */
    boolean isReady();

    /**
     * Shutdown the server
     */
    void shutdown();

    // ===== Callbacks =====

    /**
     * Tool execution callback
     */
    interface ToolCallback {
        void onSuccess(@NonNull ToolResult result);
        void onError(@NonNull String error);
        void onCancelled();
    }
}