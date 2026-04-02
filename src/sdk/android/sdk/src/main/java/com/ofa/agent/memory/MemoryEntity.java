package com.ofa.agent.memory;

import androidx.room.Entity;
import androidx.room.Index;
import androidx.room.PrimaryKey;

/**
 * Memory Entity - Room数据库实体
 */
@Entity(tableName = "memories",
        indices = {
            @Index(value = {"key"}),
            @Index(value = {"category"}),
            @Index(value = {"timestamp"}),
            @Index(value = {"score"}),
            @Index(value = {"key", "value"}, unique = true)
        })
public class MemoryEntity {

    @PrimaryKey(autoGenerate = true)
    public long id;

    public String key;          // 记忆键 (如 "bubble_tea.drink_name")
    public String category;     // 分类 (如 "food", "settings")
    public String value;        // 值 (如 "珍珠奶茶")
    public String attributes;   // JSON格式的附加属性
    public long timestamp;      // 时间戳
    public int count;           // 使用次数
    public float score;         // 重要性评分
    public String context;      // 上下文
    public long lastAccessed;   // 最后访问时间

    public MemoryEntity() {
        this.timestamp = System.currentTimeMillis();
        this.lastAccessed = this.timestamp;
        this.count = 1;
        this.score = 1.0f;
    }

    /**
     * 从MemoryEntry创建Entity
     */
    public static MemoryEntity fromEntry(MemoryEntry entry) {
        MemoryEntity entity = new MemoryEntity();
        entity.key = entry.getKey();
        entity.category = entry.getCategory();
        entity.value = entry.getValue();
        entity.attributes = attributesToJson(entry.getAttributes());
        entity.timestamp = entry.getTimestamp();
        entity.count = entry.getCount();
        entity.score = entry.getScore();
        entity.context = entry.getContext();
        entity.lastAccessed = System.currentTimeMillis();
        return entity;
    }

    /**
     * 转换为MemoryEntry
     */
    public MemoryEntry toEntry() {
        return new MemoryEntry.Builder()
                .key(key)
                .category(category != null ? category : "general")
                .value(value)
                .attributes(attributesFromJson(attributes))
                .timestamp(timestamp)
                .count(count)
                .score(score)
                .context(context)
                .build();
    }

    /**
     * 计算推荐分数
     */
    public float calculateRecommendationScore() {
        // 使用次数权重 (最大1.0)
        float countWeight = Math.min(count * 0.1f, 1.0f);

        // 时间衰减
        long ageHours = (System.currentTimeMillis() - timestamp) / (1000 * 60 * 60);
        float timeDecay = (float) Math.exp(-ageHours * 0.01);

        // 最近访问加成
        long accessAgeHours = (System.currentTimeMillis() - lastAccessed) / (1000 * 60 * 60);
        float accessDecay = (float) Math.exp(-accessAgeHours * 0.02);

        return score * 0.3f + countWeight * 0.4f + timeDecay * 0.2f + accessDecay * 0.1f;
    }

    private static String attributesToJson(java.util.Map<String, String> attrs) {
        if (attrs == null || attrs.isEmpty()) return "{}";
        try {
            org.json.JSONObject json = new org.json.JSONObject();
            for (java.util.Map.Entry<String, String> e : attrs.entrySet()) {
                json.put(e.getKey(), e.getValue());
            }
            return json.toString();
        } catch (Exception e) {
            return "{}";
        }
    }

    private static java.util.Map<String, String> attributesFromJson(String json) {
        java.util.Map<String, String> attrs = new java.util.HashMap<>();
        if (json == null || json.isEmpty()) return attrs;
        try {
            org.json.JSONObject obj = new org.json.JSONObject(json);
            for (java.util.Iterator<String> it = obj.keys(); it.hasNext(); ) {
                String key = it.next();
                attrs.put(key, obj.optString(key));
            }
        } catch (Exception e) {
            // Ignore
        }
        return attrs;
    }
}