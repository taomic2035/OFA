package com.ofa.agent.tool;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.constraint.ConstraintResult;
import com.ofa.agent.offline.OfflineLevel;

import java.util.Map;

/**
 * Tool Executor interface - implement to create custom tools.
 */
public interface ToolExecutor {

    /**
     * Get the tool ID/name
     */
    @NonNull
    String getToolId();

    /**
     * Get tool description
     */
    @NonNull
    String getDescription();

    /**
     * Execute the tool with given arguments
     * @param args Input arguments as key-value map
     * @param ctx Execution context (offline state, permissions, etc.)
     * @return Execution result
     */
    @NonNull
    ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx);

    /**
     * Check if tool is currently available
     * @return true if tool can be executed
     */
    boolean isAvailable();

    /**
     * Check if tool requires authentication
     * @return true if auth is required
     */
    boolean requiresAuth();

    /**
     * Check if tool supports offline execution
     * @return true if tool can run offline
     */
    boolean supportsOffline();

    /**
     * Get required permissions for this tool
     * @return array of permission strings
     */
    @Nullable
    String[] getRequiredPermissions();

    /**
     * Validate arguments before execution
     * @param args Input arguments
     * @return true if arguments are valid
     */
    boolean validateArgs(@NonNull Map<String, Object> args);

    /**
     * Get estimated execution time (for timeout purposes)
     * @return estimated time in milliseconds
     */
    int getEstimatedTimeMs();
}