package com.ofa.agent.core;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Task Request - represents a task to be executed.
 */
public class TaskRequest {

    // Task types
    public static final String TYPE_INTENT = "intent";
    public static final String TYPE_SKILL = "skill";
    public static final String TYPE_AUTOMATION = "automation";
    public static final String TYPE_SOCIAL = "social";
    public static final String TYPE_MEMORY = "memory";
    public static final String TYPE_NATURAL_LANGUAGE = "nl";
    public static final String TYPE_CLOUD_LLM = "cloud_llm";

    public final String taskId;
    public final String type;
    public final Map<String, Object> params;
    public final long timestamp;
    public final String source; // "local", "center", "peer"
    public final int priority;
    public final long timeoutMs;

    private boolean requiresCloudCapability = false;
    private boolean requiresLocalCapability = false;

    private TaskRequest(Builder builder) {
        this.taskId = builder.taskId != null ? builder.taskId : UUID.randomUUID().toString();
        this.type = builder.type;
        this.params = new HashMap<>(builder.params);
        this.timestamp = System.currentTimeMillis();
        this.source = builder.source;
        this.priority = builder.priority;
        this.timeoutMs = builder.timeoutMs;
        this.requiresCloudCapability = builder.requiresCloud;
        this.requiresLocalCapability = builder.requiresLocal;
    }

    public boolean requiresCloudCapability() {
        return requiresCloudCapability;
    }

    public boolean requiresLocalCapability() {
        return requiresLocalCapability;
    }

    public static class Builder {
        private String taskId;
        private String type = TYPE_NATURAL_LANGUAGE;
        private final Map<String, Object> params = new HashMap<>();
        private String source = "local";
        private int priority = 5;
        private long timeoutMs = 30000;
        private boolean requiresCloud = false;
        private boolean requiresLocal = false;

        public Builder taskId(String taskId) {
            this.taskId = taskId;
            return this;
        }

        public Builder type(String type) {
            this.type = type;
            return this;
        }

        public Builder param(String key, Object value) {
            this.params.put(key, value);
            return this;
        }

        public Builder params(Map<String, Object> params) {
            this.params.putAll(params);
            return this;
        }

        public Builder source(String source) {
            this.source = source;
            return this;
        }

        public Builder priority(int priority) {
            this.priority = priority;
            return this;
        }

        public Builder timeout(long timeoutMs) {
            this.timeoutMs = timeoutMs;
            return this;
        }

        public Builder requiresCloud(boolean requires) {
            this.requiresCloud = requires;
            return this;
        }

        public Builder requiresLocal(boolean requires) {
            this.requiresLocal = requires;
            return this;
        }

        public TaskRequest build() {
            return new TaskRequest(this);
        }
    }

    /**
     * Create intent request
     */
    public static TaskRequest intent(String input) {
        return new Builder()
            .type(TYPE_INTENT)
            .param("input", input)
            .build();
    }

    /**
     * Create skill request
     */
    public static TaskRequest skill(String skillId, Map<String, String> inputs) {
        return new Builder()
            .type(TYPE_SKILL)
            .param("skillId", skillId)
            .param("inputs", inputs)
            .build();
    }

    /**
     * Create automation request
     */
    public static TaskRequest automation(String operation, Map<String, String> params) {
        return new Builder()
            .type(TYPE_AUTOMATION)
            .param("operation", operation)
            .param("params", params)
            .build();
    }

    /**
     * Create social notification request
     */
    public static TaskRequest social(String message, String recipient, String phone) {
        return new Builder()
            .type(TYPE_SOCIAL)
            .param("message", message)
            .param("recipient", recipient)
            .param("phone", phone)
            .build();
    }

    /**
     * Create natural language request
     */
    public static TaskRequest naturalLanguage(String input) {
        return new Builder()
            .type(TYPE_NATURAL_LANGUAGE)
            .param("input", input)
            .build();
    }

    @NonNull
    @Override
    public String toString() {
        return String.format("TaskRequest{id=%s, type=%s, source=%s}",
            taskId, type, source);
    }
}