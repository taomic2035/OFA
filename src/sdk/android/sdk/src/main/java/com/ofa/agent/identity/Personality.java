package com.ofa.agent.identity;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Personality - 性格特质模型（Big Five + MBTI 双模型）
 *
 * 用于描述用户性格特征，支持 AI 个性化响应。
 */
public class Personality {

    // === MBTI 性格类型（标签化，易理解）===
    private String mbtiType;         // INTJ/ENFP/ISTP... 16种类型
    private double mbtiEI;           // E-I 维度 (-1 到 1, 正=外向)
    private double mbtiSN;           // S-N 维度 (-1 到 1, 正=直觉)
    private double mbtiTF;           // T-F 维度 (-1 到 1, 正=情感)
    private double mbtiJP;           // J-P 维度 (-1 到 1, 正=感知)
    private double mbtiConfidence;   // MBTI 置信度 (0-1)

    // === Big Five 核心特质 (0-1) ===
    private double openness;           // 开放性
    private double conscientiousness;  // 尽责性
    private double extraversion;       // 外向性
    private double agreeableness;      // 宜人性
    private double neuroticism;        // 神经质

    // === 收敛控制 ===
    private double stabilityScore;     // 性格稳定度 (0-1, 越高越稳定)
    private int observedCount;         // 行为观察次数
    private long lastInferredAt;       // 最后推断时间

    // 自定义特质
    private Map<String, Double> customTraits;

    // 说话风格
    private String speakingTone;       // formal/casual/humorous
    private String responseLength;     // brief/moderate/detailed
    private double emojiUsage;         // 表情使用频率 (0-1)

    // 性格描述（AI 生成）
    private String summary;

    // 性格标签（用于快速匹配）
    private List<String> tags;

    // MBTI 类型常量
    public static final String MBTI_INTJ = "INTJ";  // 建筑师
    public static final String MBTI_INTP = "INTP";  // 逻辑学家
    public static final String MBTI_ENTJ = "ENTJ";  // 指挥官
    public static final String MBTI_ENTP = "ENTP";  // 辩论家
    public static final String MBTI_INFJ = "INFJ";  // 提倡者
    public static final String MBTI_INFP = "INFP";  // 调停者
    public static final String MBTI_ENFJ = "ENFJ";  // 主人公
    public static final String MBTI_ENFP = "ENFP";  // 竞选者
    public static final String MBTI_ISTJ = "ISTJ";  // 物流师
    public static final String MBTI_ISFJ = "ISFJ";  // 守卫者
    public static final String MBTI_ESTJ = "ESTJ";  // 总经理
    public static final String MBTI_ESFJ = "ESFJ";  // 执行官
    public static final String MBTI_ISTP = "ISTP";  // 鉴赏家
    public static final String MBTI_ISFP = "ISFP";  // 探险家
    public static final String MBTI_ESTP = "ESTP";  // 企业家
    public static final String MBTI_ESFP = "ESFP";  // 表演者

    // 语调类型
    public static final String TONE_FORMAL = "formal";
    public static final String TONE_CASUAL = "casual";
    public static final String TONE_HUMOROUS = "humorous";
    public static final String TONE_WARM = "warm";
    public static final String TONE_NEUTRAL = "neutral";

    // 回复长度
    public static final String LENGTH_BRIEF = "brief";
    public static final String LENGTH_MODERATE = "moderate";
    public static final String LENGTH_DETAILED = "detailed";

    /**
     * 创建默认性格
     */
    public Personality() {
        // Big Five 默认值（中性）
        this.openness = 0.5;
        this.conscientiousness = 0.5;
        this.extraversion = 0.5;
        this.agreeableness = 0.5;
        this.neuroticism = 0.5;

        // MBTI 默认值（待推断）
        this.mbtiType = "";
        this.mbtiEI = 0;
        this.mbtiSN = 0;
        this.mbtiTF = 0;
        this.mbtiJP = 0;
        this.mbtiConfidence = 0;

        // 收敛控制
        this.stabilityScore = 0;
        this.observedCount = 0;
        this.lastInferredAt = 0;

        // 其他
        this.customTraits = new HashMap<>();
        this.speakingTone = TONE_CASUAL;
        this.responseLength = LENGTH_MODERATE;
        this.emojiUsage = 0.3;
        this.tags = new ArrayList<>();
        this.summary = "";
    }

    // === 更新方法 ===

    /**
     * 更新性格特质
     */
    public void updateTrait(@NonNull String key, double value) {
        value = clamp01(value);

        switch (key) {
            case "openness":
                openness = value;
                break;
            case "conscientiousness":
                conscientiousness = value;
                break;
            case "extraversion":
                extraversion = value;
                break;
            case "agreeableness":
                agreeableness = value;
                break;
            case "neuroticism":
                neuroticism = value;
                break;
            default:
                customTraits.put(key, value);
        }
    }

    /**
     * 批量更新性格特质
     */
    public void updateTraits(@NonNull Map<String, Double> updates) {
        for (Map.Entry<String, Double> entry : updates.entrySet()) {
            updateTrait(entry.getKey(), entry.getValue());
        }
    }

    /**
     * 设置 MBTI 类型
     */
    public void setMBTIType(@NonNull String type) {
        this.mbtiType = type;
        this.mbtiConfidence = 0.9; // 用户明确指定，高置信度

        // 从 MBTI 反推 Big Five（作为参考）
        inferBigFiveFromMBTI(type);
    }

    /**
     * 从 MBTI 推断 Big Five
     */
    private void inferBigFiveFromMBTI(@NonNull String type) {
        if (type.isEmpty() || type.length() < 4) return;

        // E vs I: 影响 Extraversion
        if (type.charAt(0) == 'E') {
            extraversion = 0.7;
            mbtiEI = 0.4;
        } else {
            extraversion = 0.3;
            mbtiEI = -0.4;
        }

        // S vs N: 影响 Openness
        if (type.charAt(1) == 'N') {
            openness = 0.75;
            mbtiSN = 0.5;
        } else {
            openness = 0.35;
            mbtiSN = -0.5;
        }

        // T vs F: 影响 Agreeableness
        if (type.charAt(2) == 'F') {
            agreeableness = 0.7;
            mbtiTF = 0.4;
        } else {
            agreeableness = 0.4;
            mbtiTF = -0.4;
        }

        // J vs P: 影响 Conscientiousness
        if (type.charAt(3) == 'J') {
            conscientiousness = 0.7;
            mbtiJP = -0.4;
        } else {
            conscientiousness = 0.4;
            mbtiJP = 0.4;
        }

        // Neuroticism 与 MBTI 无直接对应，保持中等
        neuroticism = 0.5;
    }

    /**
     * 获取性格描述
     */
    @NonNull
    public String getDescription() {
        StringBuilder desc = new StringBuilder();

        // 开放性
        if (openness > 0.7) {
            desc.append("富有创造力和好奇心，喜欢尝试新事物。");
        } else if (openness < 0.3) {
            desc.append("务实稳重，偏好熟悉的事物。");
        }

        // 尽责性
        if (conscientiousness > 0.7) {
            desc.append("做事认真负责，善于规划和执行。");
        } else if (conscientiousness < 0.3) {
            desc.append("随性灵活，不拘泥于细节。");
        }

        // 外向性
        if (extraversion > 0.7) {
            desc.append("外向开朗，善于社交。");
        } else if (extraversion < 0.3) {
            desc.append("内向沉静，享受独处时光。");
        }

        // 宜人性
        if (agreeableness > 0.7) {
            desc.append("友善亲和，乐于助人。");
        } else if (agreeableness < 0.3) {
            desc.append("独立直接，注重效率。");
        }

        if (desc.length() == 0) {
            desc.append("性格待完善");
        }

        return desc.toString();
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }

    // === Getters/Setters ===

    public String getMbtiType() { return mbtiType; }
    public double getMbtiEI() { return mbtiEI; }
    public double getMbtiSN() { return mbtiSN; }
    public double getMbtiTF() { return mbtiTF; }
    public double getMbtiJP() { return mbtiJP; }
    public double getMbtiConfidence() { return mbtiConfidence; }

    public double getOpenness() { return openness; }
    public double getConscientiousness() { return conscientiousness; }
    public double getExtraversion() { return extraversion; }
    public double getAgreeableness() { return agreeableness; }
    public double getNeuroticism() { return neuroticism; }

    public double getStabilityScore() { return stabilityScore; }
    public int getObservedCount() { return observedCount; }
    public long getLastInferredAt() { return lastInferredAt; }

    public Map<String, Double> getCustomTraits() { return new HashMap<>(customTraits); }

    public String getSpeakingTone() { return speakingTone; }
    public void setSpeakingTone(String tone) { this.speakingTone = tone; }

    public String getResponseLength() { return responseLength; }
    public void setResponseLength(String length) { this.responseLength = length; }

    public double getEmojiUsage() { return emojiUsage; }
    public void setEmojiUsage(double usage) { this.emojiUsage = clamp01(usage); }

    public String getSummary() { return summary; }
    public void setSummary(String summary) { this.summary = summary; }

    public List<String> getTags() { return new ArrayList<>(tags); }
    public void setTags(List<String> tags) { this.tags = new ArrayList<>(tags); }

    public void addTag(String tag) {
        if (!tags.contains(tag)) {
            tags.add(tag);
        }
    }

    public void incrementObservedCount() {
        this.observedCount++;
    }

    public void setLastInferredAt(long timestamp) {
        this.lastInferredAt = timestamp;
    }

    public void setStabilityScore(double score) {
        this.stabilityScore = clamp01(score);
    }

    /**
     * 转换为 JSON 字符串
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"mbti_type\":\"").append(mbtiType).append("\",");
        sb.append("\"openness\":").append(openness).append(",");
        sb.append("\"conscientiousness\":").append(conscientiousness).append(",");
        sb.append("\"extraversion\":").append(extraversion).append(",");
        sb.append("\"agreeableness\":").append(agreeableness).append(",");
        sb.append("\"neuroticism\":").append(neuroticism).append(",");
        sb.append("\"speaking_tone\":\"").append(speakingTone).append("\",");
        sb.append("\"response_length\":\"").append(responseLength).append("\",");
        sb.append("\"emoji_usage\":").append(emojiUsage);
        sb.append("}");
        return sb.toString();
    }
}