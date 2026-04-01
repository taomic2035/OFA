package com.ofa.agent.tool;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.constraint.ConstraintType;
import com.ofa.agent.mcp.ToolDefinition;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Tool Registry - central registry for all tools.
 * Manages tool definitions and their executors.
 */
public class ToolRegistry {

    private static final String TAG = "ToolRegistry";

    private final Context context;
    private final Map<String, ToolDefinition> definitions = new ConcurrentHashMap<>();
    private final Map<String, ToolExecutor> executors = new ConcurrentHashMap<>();
    private final Map<String, String> fallbacks = new ConcurrentHashMap<>();

    public ToolRegistry(@NonNull Context context) {
        this.context = context.getApplicationContext();
    }

    /**
     * Register a tool with definition and executor
     */
    public void register(@NonNull ToolDefinition definition, @NonNull ToolExecutor executor) {
        String name = definition.getName();
        definitions.put(name, definition);
        executors.put(name, executor);
        Log.i(TAG, "Registered tool: " + name);
    }

    /**
     * Register a tool with fallback
     */
    public void register(@NonNull ToolDefinition definition, @NonNull ToolExecutor executor,
                          @Nullable String fallbackTool) {
        register(definition, executor);
        if (fallbackTool != null) {
            fallbacks.put(definition.getName(), fallbackTool);
        }
    }

    /**
     * Unregister a tool
     */
    public void unregister(@NonNull String name) {
        definitions.remove(name);
        executors.remove(name);
        fallbacks.remove(name);
        Log.i(TAG, "Unregistered tool: " + name);
    }

    /**
     * Check if tool is registered
     */
    public boolean isRegistered(@NonNull String name) {
        return definitions.containsKey(name);
    }

    /**
     * Get tool definition
     */
    @Nullable
    public ToolDefinition getDefinition(@NonNull String name) {
        return definitions.get(name);
    }

    /**
     * Get tool executor
     */
    @Nullable
    public ToolExecutor getExecutor(@NonNull String name) {
        return executors.get(name);
    }

    /**
     * Get fallback tool name
     */
    @Nullable
    public String getFallback(@NonNull String name) {
        return fallbacks.get(name);
    }

    /**
     * List all registered tools
     */
    @NonNull
    public List<ToolDefinition> listAll() {
        return new ArrayList<>(definitions.values());
    }

    /**
     * List tools by category
     */
    @NonNull
    public List<ToolDefinition> listByCategory(@NonNull String category) {
        List<ToolDefinition> result = new ArrayList<>();
        for (ToolDefinition def : definitions.values()) {
            if (def.getCategory().equals(category)) {
                result.add(def);
            }
        }
        return result;
    }

    /**
     * List offline-capable tools
     */
    @NonNull
    public List<ToolDefinition> listOfflineCapable() {
        List<ToolDefinition> result = new ArrayList<>();
        for (ToolDefinition def : definitions.values()) {
            if (def.isOfflineCapable()) {
                result.add(def);
            }
        }
        return result;
    }

    /**
     * List tools requiring specific constraint
     */
    @NonNull
    public List<ToolDefinition> listByConstraint(@NonNull ConstraintType type) {
        List<ToolDefinition> result = new ArrayList<>();
        for (ToolDefinition def : definitions.values()) {
            if (def.hasConstraint(type)) {
                result.add(def);
            }
        }
        return result;
    }

    /**
     * List tools requiring specific permission
     */
    @NonNull
    public List<ToolDefinition> listByPermission(@NonNull String permission) {
        List<ToolDefinition> result = new ArrayList<>();
        for (ToolDefinition def : definitions.values()) {
            for (String perm : def.getRequiredPermissions()) {
                if (perm.equals(permission)) {
                    result.add(def);
                    break;
                }
            }
        }
        return result;
    }

    /**
     * Get count of registered tools
     */
    public int getCount() {
        return definitions.size();
    }

    /**
     * Get count of offline-capable tools
     */
    public int getOfflineCapableCount() {
        int count = 0;
        for (ToolDefinition def : definitions.values()) {
            if (def.isOfflineCapable()) count++;
        }
        return count;
    }

    /**
     * Check if tool is available (executor exists and is available)
     */
    public boolean isAvailable(@NonNull String name) {
        ToolExecutor executor = executors.get(name);
        return executor != null && executor.isAvailable();
    }

    /**
     * Check if tool can execute offline
     */
    public boolean canExecuteOffline(@NonNull String name) {
        ToolDefinition def = definitions.get(name);
        ToolExecutor executor = executors.get(name);
        if (def == null || executor == null) return false;
        return def.isOfflineCapable() && executor.supportsOffline();
    }

    /**
     * Clear all registrations
     */
    public void clear() {
        definitions.clear();
        executors.clear();
        fallbacks.clear();
        Log.i(TAG, "Tool registry cleared");
    }

    /**
     * Get registry statistics
     */
    @NonNull
    public Stats getStats() {
        int total = definitions.size();
        int offlineCapable = 0;
        int requiringAuth = 0;
        int unavailable = 0;

        for (ToolDefinition def : definitions.values()) {
            if (def.isOfflineCapable()) offlineCapable++;
            if (def.requiresAuth()) requiringAuth++;

            ToolExecutor executor = executors.get(def.getName());
            if (executor == null || !executor.isAvailable()) {
                unavailable++;
            }
        }

        return new Stats(total, offlineCapable, requiringAuth, unavailable);
    }

    /**
     * Build tool list as JSON array
     */
    @NonNull
    public org.json.JSONArray toJsonArray() {
        org.json.JSONArray array = new org.json.JSONArray();
        for (ToolDefinition def : definitions.values()) {
            array.put(def.toJson());
        }
        return array;
    }

    /**
     * Registry statistics
     */
    public static class Stats {
        public final int total;
        public final int offlineCapable;
        public final int requiringAuth;
        public final int unavailable;

        public Stats(int total, int offlineCapable, int requiringAuth, int unavailable) {
            this.total = total;
            this.offlineCapable = offlineCapable;
            this.requiringAuth = requiringAuth;
            this.unavailable = unavailable;
        }
    }
}