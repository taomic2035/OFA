package com.ofa.agent.skill;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 技能定义
 * 定义一个完整的技能，包含多个步骤
 */
public class SkillDefinition {

    private final String id;
    private final String name;
    private final String description;
    private final String category;
    private final String version;
    private final String author;
    private final List<String> tags;
    private final List<SkillStep> steps;
    private final Map<String, Object> inputs;     // 输入参数定义
    private final Map<String, Object> outputs;    // 输出参数定义
    private final String startStep;               // 起始步骤ID
    private final List<Trigger> triggers;         // 触发条件
    private final List<String> requiredPermissions;
    private final boolean supportsOffline;
    private final int estimatedTimeMs;
    private final JSONObject metadata;

    /**
     * 触发条件
     */
    public static class Trigger {
        public final String type;      // intent, schedule, event, voice
        public final String pattern;   // 匹配模式
        public final Map<String, Object> config;

        public Trigger(@NonNull String type, @Nullable String pattern) {
            this.type = type;
            this.pattern = pattern;
            this.config = new HashMap<>();
        }

        public Trigger(@NonNull String type, @Nullable String pattern,
                       @Nullable Map<String, Object> config) {
            this.type = type;
            this.pattern = pattern;
            this.config = config != null ? new HashMap<>(config) : new HashMap<>();
        }
    }

    private SkillDefinition(Builder builder) {
        this.id = builder.id;
        this.name = builder.name;
        this.description = builder.description;
        this.category = builder.category;
        this.version = builder.version;
        this.author = builder.author;
        this.tags = new ArrayList<>(builder.tags);
        this.steps = new ArrayList<>(builder.steps);
        this.inputs = new HashMap<>(builder.inputs);
        this.outputs = new HashMap<>(builder.outputs);
        this.startStep = builder.startStep;
        this.triggers = new ArrayList<>(builder.triggers);
        this.requiredPermissions = new ArrayList<>(builder.requiredPermissions);
        this.supportsOffline = builder.supportsOffline;
        this.estimatedTimeMs = builder.estimatedTimeMs;
        this.metadata = builder.metadata;
    }

    @NonNull
    public String getId() { return id; }

    @NonNull
    public String getName() { return name; }

    @NonNull
    public String getDescription() { return description; }

    @NonNull
    public String getCategory() { return category; }

    @NonNull
    public String getVersion() { return version; }

    @Nullable
    public String getAuthor() { return author; }

    @NonNull
    public List<String> getTags() { return new ArrayList<>(tags); }

    @NonNull
    public List<SkillStep> getSteps() { return new ArrayList<>(steps); }

    @NonNull
    public Map<String, Object> getInputs() { return new HashMap<>(inputs); }

    @NonNull
    public Map<String, Object> getOutputs() { return new HashMap<>(outputs); }

    @Nullable
    public String getStartStep() { return startStep; }

    @NonNull
    public List<Trigger> getTriggers() { return new ArrayList<>(triggers); }

    @NonNull
    public List<String> getRequiredPermissions() { return new ArrayList<>(requiredPermissions); }

    public boolean supportsOffline() { return supportsOffline; }

    public int getEstimatedTimeMs() { return estimatedTimeMs; }

    @Nullable
    public JSONObject getMetadata() { return metadata; }

    /**
     * 根据ID获取步骤
     */
    @Nullable
    public SkillStep getStep(@NonNull String stepId) {
        for (SkillStep step : steps) {
            if (step.getId().equals(stepId)) {
                return step;
            }
        }
        return null;
    }

    /**
     * Builder
     */
    public static class Builder {
        private String id;
        private String name;
        private String description = "";
        private String category = "custom";
        private String version = "1.0.0";
        private String author;
        private List<String> tags = new ArrayList<>();
        private List<SkillStep> steps = new ArrayList<>();
        private Map<String, Object> inputs = new HashMap<>();
        private Map<String, Object> outputs = new HashMap<>();
        private String startStep;
        private List<Trigger> triggers = new ArrayList<>();
        private List<String> requiredPermissions = new ArrayList<>();
        private boolean supportsOffline = false;
        private int estimatedTimeMs = 60000;
        private JSONObject metadata;

        public Builder id(@NonNull String id) {
            this.id = id;
            return this;
        }

        public Builder name(@NonNull String name) {
            this.name = name;
            return this;
        }

        public Builder description(@NonNull String description) {
            this.description = description;
            return this;
        }

        public Builder category(@NonNull String category) {
            this.category = category;
            return this;
        }

        public Builder version(@NonNull String version) {
            this.version = version;
            return this;
        }

        public Builder author(@Nullable String author) {
            this.author = author;
            return this;
        }

        public Builder tag(@NonNull String tag) {
            this.tags.add(tag);
            return this;
        }

        public Builder step(@NonNull SkillStep step) {
            this.steps.add(step);
            if (startStep == null) {
                startStep = step.getId();
            }
            return this;
        }

        public Builder input(@NonNull String name, @NonNull Object definition) {
            this.inputs.put(name, definition);
            return this;
        }

        public Builder output(@NonNull String name, @NonNull Object definition) {
            this.outputs.put(name, definition);
            return this;
        }

        public Builder startStep(@NonNull String stepId) {
            this.startStep = stepId;
            return this;
        }

        public Builder trigger(@NonNull String type, @Nullable String pattern) {
            this.triggers.add(new Trigger(type, pattern));
            return this;
        }

        public Builder trigger(@NonNull Trigger trigger) {
            this.triggers.add(trigger);
            return this;
        }

        public Builder requiredPermission(@NonNull String permission) {
            this.requiredPermissions.add(permission);
            return this;
        }

        public Builder supportsOffline(boolean supports) {
            this.supportsOffline = supports;
            return this;
        }

        public Builder estimatedTimeMs(int ms) {
            this.estimatedTimeMs = ms;
            return this;
        }

        public Builder metadata(@Nullable JSONObject metadata) {
            this.metadata = metadata;
            return this;
        }

        public SkillDefinition build() {
            if (id == null || name == null) {
                throw new IllegalArgumentException("Skill id and name are required");
            }
            if (steps.isEmpty()) {
                throw new IllegalArgumentException("Skill must have at least one step");
            }
            return new SkillDefinition(this);
        }
    }
}