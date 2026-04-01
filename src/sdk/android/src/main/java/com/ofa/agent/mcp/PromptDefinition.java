package com.ofa.agent.mcp;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * Prompt Definition - describes a prompt template for MCP.
 * Templates can be parameterized and rendered with arguments.
 */
public class PromptDefinition {

    private final String name;
    private final String description;
    private final List<PromptArgument> arguments;
    private final String template;

    /**
     * Create a prompt definition
     * @param name Prompt name
     * @param description Prompt description
     * @param template Template string with {{arg}} placeholders
     */
    public PromptDefinition(@NonNull String name, @NonNull String description,
                            @Nullable String template) {
        this.name = name;
        this.description = description;
        this.template = template != null ? template : "";
        this.arguments = new ArrayList<>();
    }

    /**
     * Full constructor with arguments
     */
    public PromptDefinition(@NonNull String name, @NonNull String description,
                            @NonNull List<PromptArgument> arguments, @NonNull String template) {
        this.name = name;
        this.description = description;
        this.arguments = arguments;
        this.template = template;
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
    public List<PromptArgument> getArguments() {
        return new ArrayList<>(arguments);
    }

    @NonNull
    public String getTemplate() {
        return template;
    }

    /**
     * Add an argument definition
     */
    public void addArgument(@NonNull PromptArgument arg) {
        arguments.add(arg);
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
            json.put("template", template);

            if (!arguments.isEmpty()) {
                JSONArray argsArray = new JSONArray();
                for (PromptArgument arg : arguments) {
                    argsArray.put(arg.toJson());
                }
                json.put("arguments", argsArray);
            }
        } catch (Exception e) {
            // JSON creation should not fail
        }
        return json;
    }

    @Nullable
    public static PromptDefinition fromJson(@NonNull JSONObject json) {
        try {
            String name = json.getString("name");
            String description = json.getString("description");
            String template = json.optString("template", "");

            List<PromptArgument> arguments = new ArrayList<>();
            if (json.has("arguments")) {
                JSONArray argsArray = json.getJSONArray("arguments");
                for (int i = 0; i < argsArray.length(); i++) {
                    PromptArgument arg = PromptArgument.fromJson(argsArray.getJSONObject(i));
                    if (arg != null) {
                        arguments.add(arg);
                    }
                }
            }

            return new PromptDefinition(name, description, arguments, template);
        } catch (Exception e) {
            return null;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "PromptDefinition{" + name + "}";
    }

    /**
     * Prompt argument definition
     */
    public static class PromptArgument {
        private final String name;
        private final String description;
        private final boolean required;

        public PromptArgument(@NonNull String name, @Nullable String description, boolean required) {
            this.name = name;
            this.description = description != null ? description : "";
            this.required = required;
        }

        @NonNull
        public String getName() {
            return name;
        }

        @NonNull
        public String getDescription() {
            return description;
        }

        public boolean isRequired() {
            return required;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("name", name);
                json.put("description", description);
                json.put("required", required);
            } catch (Exception e) {
                // JSON creation should not fail
            }
            return json;
        }

        @Nullable
        public static PromptArgument fromJson(@NonNull JSONObject json) {
            try {
                String name = json.getString("name");
                String description = json.optString("description", "");
                boolean required = json.optBoolean("required", false);
                return new PromptArgument(name, description, required);
            } catch (Exception e) {
                return null;
            }
        }
    }
}