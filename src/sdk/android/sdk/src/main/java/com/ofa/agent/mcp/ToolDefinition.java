package com.ofa.agent.mcp;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.constraint.ConstraintType;

import org.json.JSONObject;

import java.util.Arrays;

/**
 * Tool Definition - describes a tool's interface and constraints.
 * Used for MCP protocol tool registration and discovery.
 */
public class ToolDefinition {

    private final String name;
    private final String description;
    private final JSONObject inputSchema;
    private final boolean offlineCapable;
    private final ConstraintType[] constraints;
    private final String category;
    private final String[] requiredPermissions;
    private final boolean requiresAuth;
    private final int timeoutMs;

    /**
     * Create a tool definition
     * @param name Tool name (e.g., "app.launch", "camera.capture")
     * @param description Human-readable description
     * @param inputSchema JSON Schema defining input parameters
     * @param offlineCapable Whether tool can execute offline
     * @param constraints Constraint types that apply to this tool
     */
    public ToolDefinition(@NonNull String name, @NonNull String description,
                          @NonNull JSONObject inputSchema, boolean offlineCapable,
                          @Nullable ConstraintType[] constraints) {
        this.name = name;
        this.description = description;
        this.inputSchema = inputSchema;
        this.offlineCapable = offlineCapable;
        this.constraints = constraints != null ? constraints : new ConstraintType[0];
        this.category = extractCategory(name);
        this.requiredPermissions = new String[0];
        this.requiresAuth = hasAuthConstraint(constraints);
        this.timeoutMs = 30000; // default 30s
    }

    /**
     * Simple constructor with string input schema
     */
    public ToolDefinition(@NonNull String name, @NonNull String description,
                          @NonNull String inputSchemaJson) {
        this.name = name;
        this.description = description;
        JSONObject schema = new JSONObject();
        try {
            schema = new JSONObject(inputSchemaJson);
        } catch (Exception e) {
            // Use empty schema on parse error
        }
        this.inputSchema = schema;
        this.offlineCapable = true;
        this.constraints = new ConstraintType[0];
        this.category = extractCategory(name);
        this.requiredPermissions = new String[0];
        this.requiresAuth = false;
        this.timeoutMs = 30000;
    }

    /**
     * Full constructor with all options
     */
    public ToolDefinition(@NonNull String name, @NonNull String description,
                          @NonNull JSONObject inputSchema, boolean offlineCapable,
                          @Nullable ConstraintType[] constraints,
                          @Nullable String[] requiredPermissions,
                          boolean requiresAuth, int timeoutMs) {
        this.name = name;
        this.description = description;
        this.inputSchema = inputSchema;
        this.offlineCapable = offlineCapable;
        this.constraints = constraints != null ? constraints : new ConstraintType[0];
        this.category = extractCategory(name);
        this.requiredPermissions = requiredPermissions != null ? requiredPermissions : new String[0];
        this.requiresAuth = requiresAuth;
        this.timeoutMs = timeoutMs;
    }

    @NonNull
    public String getName() {
        return name;
    }

    @NonNull
    public String getDescription() {
        return description;
    }

    @NonNull
    public JSONObject getInputSchema() {
        return inputSchema;
    }

    public boolean isOfflineCapable() {
        return offlineCapable;
    }

    @NonNull
    public ConstraintType[] getConstraints() {
        return constraints;
    }

    @NonNull
    public String getCategory() {
        return category;
    }

    @NonNull
    public String[] getRequiredPermissions() {
        return requiredPermissions;
    }

    public boolean requiresAuth() {
        return requiresAuth;
    }

    public int getTimeoutMs() {
        return timeoutMs;
    }

    /**
     * Check if tool has a specific constraint
     */
    public boolean hasConstraint(@NonNull ConstraintType type) {
        for (ConstraintType c : constraints) {
            if (c == type) return true;
        }
        return false;
    }

    /**
     * Extract category from tool name (e.g., "app.launch" -> "app")
     */
    private static String extractCategory(@NonNull String name) {
        int dotIndex = name.indexOf('.');
        if (dotIndex > 0) {
            return name.substring(0, dotIndex);
        }
        return "general";
    }

    /**
     * Check if constraints include AUTH_REQUIRED
     */
    private static boolean hasAuthConstraint(@Nullable ConstraintType[] constraints) {
        if (constraints == null) return false;
        for (ConstraintType c : constraints) {
            if (c == ConstraintType.AUTH_REQUIRED) return true;
        }
        return false;
    }

    /**
     * Convert to JSON for MCP protocol
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("name", name);
            json.put("description", description);
            json.put("inputSchema", inputSchema);
            json.put("offlineCapable", offlineCapable);
            json.put("category", category);

            if (constraints.length > 0) {
                JSONObject constraintsJson = new JSONObject();
                for (ConstraintType c : constraints) {
                    constraintsJson.put(c.name().toLowerCase(), true);
                }
                json.put("constraints", constraintsJson);
            }
        } catch (Exception e) {
            // JSON creation should not fail
        }
        return json;
    }

    /**
     * Create from JSON
     */
    @Nullable
    public static ToolDefinition fromJson(@NonNull JSONObject json) {
        try {
            String name = json.getString("name");
            String description = json.getString("description");
            JSONObject inputSchema = json.getJSONObject("inputSchema");
            boolean offlineCapable = json.optBoolean("offlineCapable", true);

            ConstraintType[] constraints = null;
            if (json.has("constraints")) {
                JSONObject cJson = json.getJSONObject("constraints");
                constraints = parseConstraints(cJson);
            }

            return new ToolDefinition(name, description, inputSchema, offlineCapable, constraints);
        } catch (Exception e) {
            return null;
        }
    }

    @Nullable
    private static ConstraintType[] parseConstraints(@NonNull JSONObject json) {
        try {
            ConstraintType[] result = new ConstraintType[ConstraintType.values().length];
            int count = 0;

            for (ConstraintType type : ConstraintType.values()) {
                if (json.optBoolean(type.name().toLowerCase(), false)) {
                    result[count++] = type;
                }
            }

            return Arrays.copyOf(result, count);
        } catch (Exception e) {
            return null;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "ToolDefinition{" + name + ", offline=" + offlineCapable + "}";
    }
}