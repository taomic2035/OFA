package com.ofa.agent.automation.template;

import androidx.annotation.NonNull;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Operation Template - defines a reusable sequence of operations.
 * Templates can be used to perform common tasks across different apps.
 */
public class OperationTemplate {

    private final String id;
    private final String name;
    private final String description;
    private final String category;
    private final List<TemplateStep> steps;
    private final Map<String, String> defaultParams;
    private final List<String> requiredParams;

    /**
     * Create a new operation template
     */
    public OperationTemplate(@NonNull String id,
                             @NonNull String name,
                             @NonNull String description,
                             @NonNull String category) {
        this.id = id;
        this.name = name;
        this.description = description;
        this.category = category;
        this.steps = new ArrayList<>();
        this.defaultParams = new HashMap<>();
        this.requiredParams = new ArrayList<>();
    }

    /**
     * Get template ID
     */
    @NonNull
    public String getId() {
        return id;
    }

    /**
     * Get template name
     */
    @NonNull
    public String getName() {
        return name;
    }

    /**
     * Get template description
     */
    @NonNull
    public String getDescription() {
        return description;
    }

    /**
     * Get template category
     */
    @NonNull
    public String getCategory() {
        return category;
    }

    /**
     * Get all steps in this template
     */
    @NonNull
    public List<TemplateStep> getSteps() {
        return new ArrayList<>(steps);
    }

    /**
     * Add a step to this template
     */
    public OperationTemplate addStep(@NonNull TemplateStep step) {
        steps.add(step);
        return this;
    }

    /**
     * Add a step with operation and parameters
     */
    public OperationTemplate addStep(@NonNull String operation,
                                      @NonNull Map<String, String> params) {
        steps.add(new TemplateStep(operation, params));
        return this;
    }

    /**
     * Add a simple step with operation name
     */
    public OperationTemplate addStep(@NonNull String operation) {
        steps.add(new TemplateStep(operation, new HashMap<>()));
        return this;
    }

    /**
     * Get default parameters
     */
    @NonNull
    public Map<String, String> getDefaultParams() {
        return new HashMap<>(defaultParams);
    }

    /**
     * Set a default parameter
     */
    public OperationTemplate setDefaultParam(@NonNull String key, @NonNull String value) {
        defaultParams.put(key, value);
        return this;
    }

    /**
     * Get required parameters
     */
    @NonNull
    public List<String> getRequiredParams() {
        return new ArrayList<>(requiredParams);
    }

    /**
     * Add a required parameter
     */
    public OperationTemplate addRequiredParam(@NonNull String param) {
        requiredParams.add(param);
        return this;
    }

    /**
     * Check if all required parameters are provided
     */
    public boolean hasAllRequiredParams(@NonNull Map<String, String> params) {
        for (String required : requiredParams) {
            if (!params.containsKey(required) && !defaultParams.containsKey(required)) {
                return false;
            }
        }
        return true;
    }

    /**
     * Merge provided params with defaults
     */
    @NonNull
    public Map<String, String> mergeParams(@NonNull Map<String, String> providedParams) {
        Map<String, String> merged = new HashMap<>(defaultParams);
        merged.putAll(providedParams);
        return merged;
    }

    /**
     * Template step definition
     */
    public static class TemplateStep {
        private final String operation;
        private final Map<String, String> params;
        private final int waitAfterMs;
        private final boolean optional;
        private final String conditionParam;

        public TemplateStep(@NonNull String operation,
                            @NonNull Map<String, String> params) {
            this.operation = operation;
            this.params = new HashMap<>(params);
            this.waitAfterMs = 0;
            this.optional = false;
            this.conditionParam = null;
        }

        public TemplateStep(@NonNull String operation,
                            @NonNull Map<String, String> params,
                            int waitAfterMs,
                            boolean optional,
                            String conditionParam) {
            this.operation = operation;
            this.params = new HashMap<>(params);
            this.waitAfterMs = waitAfterMs;
            this.optional = optional;
            this.conditionParam = conditionParam;
        }

        @NonNull
        public String getOperation() {
            return operation;
        }

        @NonNull
        public Map<String, String> getParams() {
            return new HashMap<>(params);
        }

        public int getWaitAfterMs() {
            return waitAfterMs;
        }

        public boolean isOptional() {
            return optional;
        }

        public String getConditionParam() {
            return conditionParam;
        }

        /**
         * Create a builder for this step
         */
        public static Builder builder(@NonNull String operation) {
            return new Builder(operation);
        }

        /**
         * Builder for TemplateStep
         */
        public static class Builder {
            private final String operation;
            private final Map<String, String> params = new HashMap<>();
            private int waitAfterMs = 0;
            private boolean optional = false;
            private String conditionParam = null;

            public Builder(@NonNull String operation) {
                this.operation = operation;
            }

            public Builder param(@NonNull String key, @NonNull String value) {
                params.put(key, value);
                return this;
            }

            public Builder waitAfter(int ms) {
                this.waitAfterMs = ms;
                return this;
            }

            public Builder optional() {
                this.optional = true;
                return this;
            }

            public Builder condition(@NonNull String paramName) {
                this.conditionParam = paramName;
                return this;
            }

            public TemplateStep build() {
                return new TemplateStep(operation, params, waitAfterMs, optional, conditionParam);
            }
        }
    }

    /**
     * Builder for OperationTemplate
     */
    public static class Builder {
        private final String id;
        private final String name;
        private final String description;
        private final String category;
        private final List<TemplateStep> steps = new ArrayList<>();
        private final Map<String, String> defaultParams = new HashMap<>();
        private final List<String> requiredParams = new ArrayList<>();

        public Builder(@NonNull String id,
                       @NonNull String name,
                       @NonNull String description,
                       @NonNull String category) {
            this.id = id;
            this.name = name;
            this.description = description;
            this.category = category;
        }

        public Builder addStep(@NonNull TemplateStep step) {
            steps.add(step);
            return this;
        }

        public Builder addStep(@NonNull String operation) {
            steps.add(new TemplateStep(operation, new HashMap<>()));
            return this;
        }

        public Builder addStep(@NonNull String operation, @NonNull Map<String, String> params) {
            steps.add(new TemplateStep(operation, params));
            return this;
        }

        public Builder defaultParam(@NonNull String key, @NonNull String value) {
            defaultParams.put(key, value);
            return this;
        }

        public Builder requiredParam(@NonNull String param) {
            requiredParams.add(param);
            return this;
        }

        public OperationTemplate build() {
            OperationTemplate template = new OperationTemplate(id, name, description, category);
            for (TemplateStep step : steps) {
                template.addStep(step);
            }
            for (Map.Entry<String, String> entry : defaultParams.entrySet()) {
                template.setDefaultParam(entry.getKey(), entry.getValue());
            }
            for (String param : requiredParams) {
                template.addRequiredParam(param);
            }
            return template;
        }
    }
}