package com.ofa.agent.memory;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.HashMap;
import java.util.Map;

/**
 * 记忆条目
 * 存储单条用户行为或偏好记录
 */
public class MemoryEntry {

    private final String key;           // 记忆键（如 "bubble_tea.drink_name"）
    private final String category;      // 分类（如 "food", "settings"）
    private final String value;         // 值（如 "珍珠奶茶"）
    private final Map<String, String> attributes;  // 附加属性（如 {"sweetness": "五分糖"}）
    private final long timestamp;       // 时间戳
    private final int count;            // 使用次数
    private final float score;          // 重要性评分
    private final String context;       // 上下文（如 "点奶茶技能"）

    private MemoryEntry(Builder builder) {
        this.key = builder.key;
        this.category = builder.category;
        this.value = builder.value;
        this.attributes = new HashMap<>(builder.attributes);
        this.timestamp = builder.timestamp;
        this.count = builder.count;
        this.score = builder.score;
        this.context = builder.context;
    }

    @NonNull
    public String getKey() { return key; }

    @NonNull
    public String getCategory() { return category; }

    @NonNull
    public String getValue() { return value; }

    @NonNull
    public Map<String, String> getAttributes() { return new HashMap<>(attributes); }

    @Nullable
    public String getAttribute(@NonNull String name) { return attributes.get(name); }

    public long getTimestamp() { return timestamp; }

    public int getCount() { return count; }

    public float getScore() { return score; }

    @Nullable
    public String getContext() { return context; }

    /**
     * 是否过期
     * @param maxAgeMs 最大存活时间（毫秒）
     */
    public boolean isExpired(long maxAgeMs) {
        return System.currentTimeMillis() - timestamp > maxAgeMs;
    }

    /**
     * 计算推荐分数
     * 考虑使用次数、最近使用时间、属性匹配度
     */
    public float calculateRecommendationScore() {
        // 基础分数 = 使用次数权重 + 时间衰减 + 原始分数
        float countWeight = Math.min(count * 0.1f, 1.0f);

        // 时间衰减：最近使用的权重更高
        long ageHours = (System.currentTimeMillis() - timestamp) / (1000 * 60 * 60);
        float timeDecay = (float) Math.exp(-ageHours * 0.01); // 每小时衰减1%

        return score * 0.3f + countWeight * 0.4f + timeDecay * 0.3f;
    }

    @NonNull
    @Override
    public String toString() {
        return "MemoryEntry{" + key + "=" + value +
                ", count=" + count +
                ", score=" + score + "}";
    }

    /**
     * Builder
     */
    public static class Builder {
        private String key;
        private String category = "general";
        private String value;
        private Map<String, String> attributes = new HashMap<>();
        private long timestamp = System.currentTimeMillis();
        private int count = 1;
        private float score = 1.0f;
        private String context;

        public Builder key(@NonNull String key) {
            this.key = key;
            return this;
        }

        public Builder category(@NonNull String category) {
            this.category = category;
            return this;
        }

        public Builder value(@NonNull String value) {
            this.value = value;
            return this;
        }

        public Builder attribute(@NonNull String name, @Nullable String value) {
            if (value != null) {
                this.attributes.put(name, value);
            }
            return this;
        }

        public Builder attributes(@NonNull Map<String, String> attributes) {
            this.attributes.putAll(attributes);
            return this;
        }

        public Builder timestamp(long timestamp) {
            this.timestamp = timestamp;
            return this;
        }

        public Builder count(int count) {
            this.count = count;
            return this;
        }

        public Builder score(float score) {
            this.score = score;
            return this;
        }

        public Builder context(@Nullable String context) {
            this.context = context;
            return this;
        }

        public MemoryEntry build() {
            if (key == null || value == null) {
                throw new IllegalArgumentException("Key and value are required");
            }
            return new MemoryEntry(this);
        }
    }
}