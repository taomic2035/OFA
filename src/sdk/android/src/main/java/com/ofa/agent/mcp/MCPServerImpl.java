package com.ofa.agent.mcp;

import android.content.Context;
import android.os.Handler;
import android.os.HandlerThread;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.constraint.ConstraintChecker;
import com.ofa.agent.constraint.ConstraintResult;
import com.ofa.agent.constraint.ConstraintType;
import com.ofa.agent.offline.OfflineLevel;
import com.ofa.agent.offline.OfflineManager;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.PermissionManager;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolRegistry;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

/**
 * MCP Server Implementation - provides MCP protocol support for OFA Agent.
 */
public class MCPServerImpl implements MCPServer {

    private static final String TAG = "MCPServerImpl";

    private final Context context;
    private final ToolRegistry registry;
    private final PermissionManager permissionManager;
    private final ConstraintChecker constraintChecker;
    private final OfflineManager offlineManager;

    private final ExecutorService executor;
    private final Handler mainHandler;
    private final HandlerThread workerThread;
    private final Handler workerHandler;

    private final Map<String, Future<?>> activeExecutions = new ConcurrentHashMap<>();
    private final Map<String, ToolResult> executionResults = new ConcurrentHashMap<>();

    private volatile boolean ready = false;
    private final Map<String, PromptDefinition> prompts = new ConcurrentHashMap<>();
    private final Map<String, ResourceDefinition> resources = new ConcurrentHashMap<>();

    /**
     * Create MCP Server with all components
     */
    public MCPServerImpl(@NonNull Context context, @NonNull ToolRegistry registry,
                         @Nullable PermissionManager permissionManager,
                         @Nullable ConstraintChecker constraintChecker,
                         @Nullable OfflineManager offlineManager) {
        this.context = context.getApplicationContext();
        this.registry = registry;
        this.permissionManager = permissionManager != null
                ? permissionManager : new PermissionManager(context);
        this.constraintChecker = constraintChecker != null
                ? constraintChecker : new ConstraintChecker();
        this.offlineManager = offlineManager;

        this.executor = Executors.newCachedThreadPool();
        this.mainHandler = new Handler(Looper.getMainLooper());
        this.workerThread = new HandlerThread("MCPServerWorker");
        this.workerThread.start();
        this.workerHandler = new Handler(workerThread.getLooper());

        loadDefaultPrompts();
        loadDefaultResources();
    }

    /**
     * Simple constructor
     */
    public MCPServerImpl(@NonNull Context context) {
        this(context, new ToolRegistry(context), null, null, null);
    }

    /**
     * Start the server
     */
    public void start() {
        ready = true;
        Log.i(TAG, "MCP Server started");
    }

    // ===== Tool Management =====

    @NonNull
    @Override
    public List<ToolDefinition> listTools() {
        return registry.listAll();
    }

    @Nullable
    @Override
    public ToolDefinition getTool(@NonNull String name) {
        return registry.getDefinition(name);
    }

    @NonNull
    @Override
    public ToolResult callTool(@NonNull String name, @NonNull Map<String, Object> args) {
        return executeToolInternal(name, args, null);
    }

    @NonNull
    @Override
    public String callToolAsync(@NonNull String name, @NonNull Map<String, Object> args,
                                @NonNull ToolCallback callback) {
        String executionId = "exec-" + UUID.randomUUID().toString().substring(0, 8);

        Future<?> future = executor.submit(() -> {
            try {
                ToolResult result = executeToolInternal(name, args, executionId);
                executionResults.put(executionId, result);
                mainHandler.post(() -> callback.onSuccess(result));
            } catch (Exception e) {
                mainHandler.post(() -> callback.onError(e.getMessage()));
            }
        });

        activeExecutions.put(executionId, future);
        return executionId;
    }

    @Override
    public boolean cancelExecution(@NonNull String executionId) {
        Future<?> future = activeExecutions.remove(executionId);
        if (future != null) {
            boolean cancelled = future.cancel(true);
            if (cancelled) {
                Log.i(TAG, "Execution cancelled: " + executionId);
            }
            return cancelled;
        }
        return false;
    }

    @Nullable
    public ToolResult getExecutionResult(@NonNull String executionId) {
        return executionResults.remove(executionId);
    }

    // ===== Internal Execution =====

    @NonNull
    private ToolResult executeToolInternal(@NonNull String name, @NonNull Map<String, Object> args,
                                            @Nullable String executionId) {
        long startTime = System.currentTimeMillis();

        // 1. Get tool definition
        ToolDefinition definition = registry.getDefinition(name);
        if (definition == null) {
            return new ToolResult(name, "Tool not found: " + name, 0);
        }

        // 2. Get executor
        ToolExecutor executor = registry.getExecutor(name);
        if (executor == null) {
            return new ToolResult(name, "No executor for tool: " + name, 0);
        }

        // 3. Check availability
        if (!executor.isAvailable()) {
            return new ToolResult(name, "Tool not available: " + name, 0);
        }

        // 4. Check offline compatibility
        boolean isOffline = offlineManager != null && offlineManager.isOfflineMode();
        if (isOffline && !definition.isOfflineCapable()) {
            // Check for fallback
            String fallback = registry.getFallback(name);
            if (fallback != null) {
                ToolDefinition fallbackDef = registry.getDefinition(fallback);
                if (fallbackDef != null && fallbackDef.isOfflineCapable()) {
                    Log.i(TAG, "Using fallback tool: " + fallback);
                    return executeToolInternal(fallback, args, executionId);
                }
            }
            return new ToolResult(name, "Tool not available offline", 0);
        }

        // 5. Check permissions
        String[] requiredPerms = definition.getRequiredPermissions();
        if (requiredPerms.length > 0) {
            if (!permissionManager.checkPermissions(requiredPerms)) {
                return new ToolResult(name, "Permission denied for tool: " + name, 0);
            }
        }

        // 6. Check constraints
        JSONObject argsJson = mapToJson(args);
        ConstraintResult constraintResult = constraintChecker.check(name, argsJson.toString());
        if (!constraintResult.allowed) {
            return new ToolResult(name, "Constraint violation: " + constraintResult.reason, 0);
        }

        // 7. Validate arguments
        if (!executor.validateArgs(args)) {
            return new ToolResult(name, "Invalid arguments for tool: " + name, 0);
        }

        // 8. Build execution context
        OfflineLevel offlineLevel = offlineManager != null
                ? offlineManager.getLevel() : OfflineLevel.L4;
        ExecutionContext ctx = ExecutionContext.builder(context)
                .offline(isOffline)
                .offlineLevel(offlineLevel)
                .autoPopulateDeviceInfo()
                .constraints(constraintResult)
                .timeoutMs(definition.getTimeoutMs())
                .executionId(executionId)
                .permissionGranted(requiredPerms.length == 0 || permissionManager.checkPermissions(requiredPerms))
                .build();

        // 9. Execute
        try {
            ToolResult result = executor.execute(args, ctx);
            long execTime = System.currentTimeMillis() - startTime;
            return new ToolResult(name,
                    result.isSuccess() ? result.getOutput() : new JSONObject(),
                    execTime, result.isCached());
        } catch (Exception e) {
            long execTime = System.currentTimeMillis() - startTime;
            return new ToolResult(name, "Execution error: " + e.getMessage(), execTime);
        }
    }

    // ===== Resource Access =====

    @NonNull
    @Override
    public List<ResourceDefinition> listResources() {
        return new java.util.ArrayList<>(resources.values());
    }

    @Nullable
    @Override
    public ResourceDefinition getResource(@NonNull String uri) {
        return resources.get(uri);
    }

    @NonNull
    @Override
    public ResourceContent readResource(@NonNull String uri) {
        // TODO: Implement actual resource reading
        ResourceDefinition def = resources.get(uri);
        if (def == null) {
            return new ResourceContent(uri, "text/plain", "Resource not found", System.currentTimeMillis());
        }
        return new ResourceContent(uri, def.getMimeType(), "Content placeholder", System.currentTimeMillis());
    }

    public void registerResource(@NonNull ResourceDefinition resource) {
        resources.put(resource.getUri(), resource);
    }

    // ===== Prompt Templates =====

    @NonNull
    @Override
    public List<PromptDefinition> listPrompts() {
        return new java.util.ArrayList<>(prompts.values());
    }

    @Nullable
    @Override
    public PromptDefinition getPrompt(@NonNull String name) {
        return prompts.get(name);
    }

    @NonNull
    @Override
    public PromptResult getPrompt(@NonNull String name, @Nullable Map<String, Object> args) {
        PromptDefinition def = prompts.get(name);
        if (def == null) {
            return new PromptResult(name, "Prompt not found: " + name);
        }

        String template = def.getTemplate();
        if (args != null) {
            template = renderTemplate(template, args);
        }

        return new PromptResult(name, template, args);
    }

    public void registerPrompt(@NonNull PromptDefinition prompt) {
        prompts.put(prompt.getName(), prompt);
    }

    // ===== Server Info =====

    @NonNull
    @Override
    public JSONObject getServerInfo() {
        JSONObject info = new JSONObject();
        try {
            info.put("protocolVersion", MCPProtocol.VERSION);
            info.put("name", "OFA Agent MCP Server");
            info.put("version", "1.0.0");

            JSONObject capabilities = new JSONObject();
            capabilities.put("tools", true);
            capabilities.put("resources", true);
            capabilities.put("prompts", true);
            info.put("capabilities", capabilities);

            ToolRegistry.Stats stats = registry.getStats();
            JSONObject toolStats = new JSONObject();
            toolStats.put("total", stats.total);
            toolStats.put("offlineCapable", stats.offlineCapable);
            info.put("tools", toolStats);
        } catch (Exception e) {
            // Should not fail
        }
        return info;
    }

    @Override
    public boolean isReady() {
        return ready;
    }

    @Override
    public void shutdown() {
        ready = false;

        // Cancel all active executions
        for (Future<?> future : activeExecutions.values()) {
            future.cancel(true);
        }
        activeExecutions.clear();

        executor.shutdown();
        workerThread.quit();

        Log.i(TAG, "MCP Server shutdown");
    }

    // ===== Helper Methods =====

    private void loadDefaultPrompts() {
        // Default prompts for common scenarios
        PromptDefinition taskPrompt = new PromptDefinition(
                "task_analysis",
                "Analyze a task and suggest appropriate tools",
                "Analyze the following task and suggest which tools should be used:\n\nTask: {{task}}\n\nAvailable tools: {{tools}}"
        );
        taskPrompt.addArgument(new PromptDefinition.PromptArgument("task", "The task description", true));
        taskPrompt.addArgument(new PromptDefinition.PromptArgument("tools", "Available tool list", false));
        prompts.put("task_analysis", taskPrompt);

        PromptDefinition errorPrompt = new PromptDefinition(
                "error_handling",
                "Generate error handling suggestions",
                "The tool '{{tool}}' failed with error: {{error}}\n\nSuggest alternative approaches or solutions."
        );
        errorPrompt.addArgument(new PromptDefinition.PromptArgument("tool", "The failed tool name", true));
        errorPrompt.addArgument(new PromptDefinition.PromptArgument("error", "The error message", true));
        prompts.put("error_handling", errorPrompt);
    }

    private void loadDefaultResources() {
        // Default resources
        resources.put("ofa://tools", new ResourceDefinition(
                "ofa://tools", "Tool List",
                "List of all registered tools", "application/json"
        ));
        resources.put("ofa://status", new ResourceDefinition(
                "ofa://status", "Agent Status",
                "Current agent status and statistics", "application/json"
        ));
    }

    @NonNull
    private String renderTemplate(@NonNull String template, @NonNull Map<String, Object> args) {
        String result = template;
        for (Map.Entry<String, Object> entry : args.entrySet()) {
            String placeholder = "{{" + entry.getKey() + "}}";
            String value = entry.getValue() != null ? entry.getValue().toString() : "";
            result = result.replace(placeholder, value);
        }
        return result;
    }

    @NonNull
    private JSONObject mapToJson(@NonNull Map<String, Object> map) {
        JSONObject json = new JSONObject();
        try {
            for (Map.Entry<String, Object> entry : map.entrySet()) {
                json.put(entry.getKey(), entry.getValue());
            }
        } catch (Exception e) {
            // Should not fail
        }
        return json;
    }
}