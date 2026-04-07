package com.ofa.agent.identity;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.List;

/**
 * Interest - 兴趣爱好模型
 *
 * 描述用户的兴趣爱好，用于个性化推荐和内容匹配。
 */
public class Interest {

    private String id;
    private String category;       // sports/tech/art/music/food/travel...
    private String name;           // 具体名称
    private double level;          // 热衷程度 (0-1)
    private List<String> keywords; // 关键词
    private String description;    // 描述
    private long since;            // 开始时间（时间戳）
    private long lastActive;       // 最近活跃（时间戳）

    // 兴趣类别常量
    public static final String CATEGORY_SPORTS = "sports";
    public static final String CATEGORY_TECH = "tech";
    public static final String CATEGORY_ART = "art";
    public static final String CATEGORY_MUSIC = "music";
    public static final String CATEGORY_FOOD = "food";
    public static final String CATEGORY_TRAVEL = "travel";
    public static final String CATEGORY_READING = "reading";
    public static final String CATEGORY_GAMING = "gaming";
    public static final String CATEGORY_FITNESS = "fitness";
    public static final String CATEGORY_MOVIES = "movies";
    public static final String CATEGORY_FASHION = "fashion";
    public static final String CATEGORY_FINANCE = "finance";
    public static final String CATEGORY_SOCIAL = "social";
    public static final String CATEGORY_OTHER = "other";

    /**
     * 创建兴趣
     */
    public Interest(@NonNull String category, @NonNull String name) {
        this.id = generateId();
        this.category = category;
        this.name = name;
        this.level = 0.5;
        this.keywords = new ArrayList<>();
        this.description = "";
        this.since = System.currentTimeMillis();
        this.lastActive = System.currentTimeMillis();
    }

    /**
     * 创建兴趣（完整参数）
     */
    public Interest(@NonNull String id, @NonNull String category, @NonNull String name,
                    double level, @Nullable List<String> keywords, @Nullable String description,
                    long since, long lastActive) {
        this.id = id;
        this.category = category;
        this.name = name;
        this.level = clamp01(level);
        this.keywords = keywords != null ? new ArrayList<>(keywords) : new ArrayList<>();
        this.description = description != null ? description : "";
        this.since = since;
        this.lastActive = lastActive;
    }

    // === 更新方法 ===

    /**
     * 更新热衷程度
     */
    public void updateLevel(double newLevel) {
        this.level = clamp01(newLevel);
        this.lastActive = System.currentTimeMillis();
    }

    /**
     * 添加关键词
     */
    public void addKeyword(@NonNull String keyword) {
        if (!keywords.contains(keyword)) {
            keywords.add(keyword);
        }
    }

    /**
     * 移除关键词
     */
    public void removeKeyword(@NonNull String keyword) {
        keywords.remove(keyword);
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }

    private String generateId() {
        return System.currentTimeMillis() + "_" + Integer.toHexString((int)(Math.random() * 10000));
    }

    // === Getters ===

    public String getId() { return id; }
    public String getCategory() { return category; }
    public String getName() { return name; }
    public double getLevel() { return level; }
    public List<String> getKeywords() { return new ArrayList<>(keywords); }
    public String getDescription() { return description; }
    public long getSince() { return since; }
    public long getLastActive() { return lastActive; }

    /**
     * 转换为 JSON 字符串
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"id\":\"").append(id).append("\",");
        sb.append("\"category\":\"").append(category).append("\",");
        sb.append("\"name\":\"").append(name).append("\",");
        sb.append("\"level\":").append(level).append(",");
        sb.append("\"keywords\":[");

        for (int i = 0; i < keywords.size(); i++) {
            if (i > 0) sb.append(",");
            sb.append("\"").append(keywords.get(i)).append("\"");
        }

        sb.append("]");
        sb.append("}");
        return sb.toString();
    }

    @NonNull
    @Override
    public String toString() {
        return String.format("Interest{id=%s, category=%s, name=%s, level=%.2f}",
            id, category, name, level);
    }
}