package com.ofa.agent.intent;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.regex.Pattern;

/**
 * 意图定义
 * 定义一个意图的模式匹配规则和槽位
 */
public class IntentDefinition {

    private final String id;
    private final String category;
    private final String action;
    private final String description;
    private final List<String> keywords;
    private final List<Pattern> patterns;
    private final List<SlotDefinition> slots;
    private final List<String> requiredSlots;
    private final String confirmationTemplate;
    private final double defaultConfidence;

    private IntentDefinition(Builder builder) {
        this.id = builder.id;
        this.category = builder.category;
        this.action = builder.action;
        this.description = builder.description;
        this.keywords = new ArrayList<>(builder.keywords);
        this.patterns = new ArrayList<>(builder.patterns);
        this.slots = new ArrayList<>(builder.slots);
        this.requiredSlots = new ArrayList<>(builder.requiredSlots);
        this.confirmationTemplate = builder.confirmationTemplate;
        this.defaultConfidence = builder.defaultConfidence;
    }

    @NonNull
    public String getId() { return id; }

    @NonNull
    public String getCategory() { return category; }

    @NonNull
    public String getAction() { return action; }

    @NonNull
    public String getDescription() { return description; }

    @NonNull
    public List<String> getKeywords() { return new ArrayList<>(keywords); }

    @NonNull
    public List<Pattern> getPatterns() { return new ArrayList<>(patterns); }

    @NonNull
    public List<SlotDefinition> getSlots() { return new ArrayList<>(slots); }

    @NonNull
    public List<String> getRequiredSlots() { return new ArrayList<>(requiredSlots); }

    @Nullable
    public String getConfirmationTemplate() { return confirmationTemplate; }

    public double getDefaultConfidence() { return defaultConfidence; }

    /**
     * 检查输入是否匹配此意图
     */
    public double matchScore(@NonNull String input) {
        String normalized = input.toLowerCase().trim();

        // 模式匹配
        for (Pattern pattern : patterns) {
            if (pattern.matcher(normalized).find()) {
                return Math.min(1.0, defaultConfidence + 0.2);
            }
        }

        // 关键词匹配
        int matchCount = 0;
        for (String keyword : keywords) {
            if (normalized.contains(keyword.toLowerCase())) {
                matchCount++;
            }
        }

        if (matchCount > 0) {
            double score = defaultConfidence * ((double) matchCount / keywords.size());
            return Math.min(1.0, score + 0.3);
        }

        return 0;
    }

    /**
     * 槽位定义
     */
    public static class SlotDefinition {
        public final String name;
        public final String type;
        public final String description;
        public final boolean required;
        public final Pattern extractPattern;
        public final List<String> possibleValues;

        public SlotDefinition(@NonNull String name, @NonNull String type,
                              @Nullable String description, boolean required,
                              @Nullable Pattern extractPattern,
                              @Nullable List<String> possibleValues) {
            this.name = name;
            this.type = type;
            this.description = description != null ? description : "";
            this.required = required;
            this.extractPattern = extractPattern;
            this.possibleValues = possibleValues != null ? new ArrayList<>(possibleValues) : null;
        }

        /**
         * 从输入中提取槽位值
         */
        @Nullable
        public Object extract(@NonNull String input) {
            if (extractPattern != null) {
                java.util.regex.Matcher matcher = extractPattern.matcher(input);
                if (matcher.find()) {
                    return matcher.groupCount() > 0 ? matcher.group(1) : matcher.group();
                }
            }
            return null;
        }
    }

    /**
     * Builder
     */
    public static class Builder {
        private String id;
        private String category;
        private String action;
        private String description = "";
        private List<String> keywords = new ArrayList<>();
        private List<Pattern> patterns = new ArrayList<>();
        private List<SlotDefinition> slots = new ArrayList<>();
        private List<String> requiredSlots = new ArrayList<>();
        private String confirmationTemplate;
        private double defaultConfidence = 0.7;

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

        public Builder description(@NonNull String description) {
            this.description = description;
            return this;
        }

        public Builder keywords(@NonNull String... keywords) {
            this.keywords.addAll(Arrays.asList(keywords));
            return this;
        }

        public Builder pattern(@NonNull String regex) {
            this.patterns.add(Pattern.compile(regex, Pattern.CASE_INSENSITIVE));
            return this;
        }

        public Builder slot(@NonNull String name, @NonNull String type,
                            @Nullable String description, boolean required) {
            this.slots.add(new SlotDefinition(name, type, description, required, null, null));
            if (required) {
                this.requiredSlots.add(name);
            }
            return this;
        }

        public Builder slotWithPattern(@NonNull String name, @NonNull String type,
                                       @NonNull String pattern, boolean required) {
            Pattern p = Pattern.compile(pattern, Pattern.CASE_INSENSITIVE);
            SlotDefinition slot = new SlotDefinition(name, type, null, required, p, null);
            this.slots.add(slot);
            if (required) {
                this.requiredSlots.add(name);
            }
            return this;
        }

        public Builder confirmationTemplate(@NonNull String template) {
            this.confirmationTemplate = template;
            return this;
        }

        public Builder defaultConfidence(double confidence) {
            this.defaultConfidence = confidence;
            return this;
        }

        public IntentDefinition build() {
            if (id == null) {
                id = category + "_" + action;
            }
            return new IntentDefinition(this);
        }
    }
}