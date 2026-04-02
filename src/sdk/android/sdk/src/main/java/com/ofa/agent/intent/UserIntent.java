package com.ofa.agent.intent;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.HashMap;
import java.util.Map;

/**
 * 用户意图
 * 表示解析后的用户意图
 */
public class UserIntent {

    private final String id;
    private final String category;
    private final String action;
    private final double confidence;
    private final Map<String, Object> slots;
    private final String rawInput;
    private final String normalizedInput;

    public UserIntent(@NonNull String id, @NonNull String category, @NonNull String action,
                      double confidence, @Nullable Map<String, Object> slots,
                      @NonNull String rawInput) {
        this.id = id;
        this.category = category;
        this.action = action;
        this.confidence = confidence;
        this.slots = slots != null ? new HashMap<>(slots) : new HashMap<>();
        this.rawInput = rawInput;
        this.normalizedInput = rawInput.toLowerCase().trim();
    }

    @NonNull
    public String getId() { return id; }

    @NonNull
    public String getCategory() { return category; }

    @NonNull
    public String getAction() { return action; }

    public double getConfidence() { return confidence; }

    @NonNull
    public Map<String, Object> getSlots() { return new HashMap<>(slots); }

    @Nullable
    public Object getSlot(@NonNull String name) { return slots.get(name); }

    @Nullable
    public String getSlotAsString(@NonNull String name) {
        Object val = slots.get(name);
        return val != null ? val.toString() : null;
    }

    @NonNull
    public String getRawInput() { return rawInput; }

    @NonNull
    public String getNormalizedInput() { return normalizedInput; }

    /**
     * 获取完整意图名称 (category.action)
     */
    @NonNull
    public String getFullName() {
        return category + "." + action;
    }

    /**
     * 是否是高置信度意图
     */
    public boolean isHighConfidence() {
        return confidence >= 0.7;
    }

    /**
     * 是否是中等置信度意图
     */
    public boolean isMediumConfidence() {
        return confidence >= 0.4 && confidence < 0.7;
    }

    /**
     * 是否是低置信度意图
     */
    public boolean isLowConfidence() {
        return confidence < 0.4;
    }

    @NonNull
    @Override
    public String toString() {
        return "UserIntent{" + getFullName() +
                ", confidence=" + confidence +
                ", slots=" + slots + "}";
    }

    /**
     * Builder for creating intents
     */
    public static class Builder {
        private String id;
        private String category;
        private String action;
        private double confidence = 1.0;
        private Map<String, Object> slots = new HashMap<>();
        private String rawInput = "";

        public Builder id(@NonNull String id) {
            this.id = id;
            return this;
        }

        public Builder category(@NonNull String category) {
            this.category = category;
            return this;
        }

        public Builder action(@NonNull String action) {
            this.action = action;
            return this;
        }

        public Builder confidence(double confidence) {
            this.confidence = Math.max(0, Math.min(1, confidence));
            return this;
        }

        public Builder slot(@NonNull String name, @Nullable Object value) {
            this.slots.put(name, value);
            return this;
        }

        public Builder slots(@NonNull Map<String, Object> slots) {
            this.slots.putAll(slots);
            return this;
        }

        public Builder rawInput(@NonNull String input) {
            this.rawInput = input;
            return this;
        }

        public UserIntent build() {
            if (id == null) id = category + "_" + action;
            return new UserIntent(id, category, action, confidence, slots, rawInput);
        }
    }
}