package com.ofa.agent.emotion;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * 欲望状态模型 (v4.0.0)
 *
 * 端侧只接收 Center 推送的欲望状态，用于行为驱动调整。
 * 深层欲望管理和计算在 Center 端 EmotionEngine 完成。
 */
public class DesireState {

    // === 马斯洛需求层次 (0-1 满足度) ===
    private double physiological;      // 生理需求 (食、睡、安全)
    private double safety;             // 安全需求 (稳定、保障)
    private double loveBelonging;      // 爱与归属 (友谊、爱情)
    private double esteem;             // 尊重需求 (自尊、认可)
    private double selfActualization;  // 自我实现 (潜能、意义)

    // === 当前驱动力 ===
    private String primaryDesire;      // 当前主要欲望
    private double desireStrength;     // 欲望强度
    private String desireTarget;       // 欲望目标
    private String desireAction;       // 欲望驱动的行为

    // === 满足度指标 ===
    private double satisfactionLevel;  // 整体满足度
    private double frustrationLevel;   // 挫折程度

    // === 时间属性 ===
    private long timestamp;

    public DesireState() {
        // 默认状态
        this.physiological = 0.7;
        this.safety = 0.6;
        this.loveBelonging = 0.5;
        this.esteem = 0.4;
        this.selfActualization = 0.3;
        this.primaryDesire = "esteem";
        this.desireStrength = 0.5;
        this.desireTarget = "个人成就";
        this.desireAction = "";
        this.satisfactionLevel = 0.5;
        this.frustrationLevel = 0.3;
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析 (Center 推送的状态)
     */
    @NonNull
    public static DesireState fromJson(@NonNull JSONObject json) throws JSONException {
        DesireState state = new DesireState();

        // 需求层次
        state.physiological = json.optDouble("physiological", 0.7);
        state.safety = json.optDouble("safety", 0.6);
        state.loveBelonging = json.optDouble("love_belonging", 0.5);
        state.esteem = json.optDouble("esteem", 0.4);
        state.selfActualization = json.optDouble("self_actualization", 0.3);

        // 当前驱动力
        state.primaryDesire = json.optString("primary_desire", "esteem");
        state.desireStrength = json.optDouble("desire_strength", 0.5);
        state.desireTarget = json.optString("desire_target", "");
        state.desireAction = json.optString("desire_action", "");

        // 满足度
        state.satisfactionLevel = json.optDouble("satisfaction_level", 0.5);
        state.frustrationLevel = json.optDouble("frustration_level", 0.3);

        // 时间
        state.timestamp = json.optLong("timestamp", System.currentTimeMillis());

        return state;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("physiological", physiological);
            json.put("safety", safety);
            json.put("love_belonging", loveBelonging);
            json.put("esteem", esteem);
            json.put("self_actualization", selfActualization);
            json.put("primary_desire", primaryDesire);
            json.put("desire_strength", desireStrength);
            json.put("desire_target", desireTarget);
            json.put("desire_action", desireAction);
            json.put("satisfaction_level", satisfactionLevel);
            json.put("frustration_level", frustrationLevel);
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 获取最紧迫的需求（满足度最低）
     */
    @NonNull
    public String getMostUrgentNeed() {
        double min = physiological;
        String urgent = "physiological";

        if (safety < min) { min = safety; urgent = "safety"; }
        if (loveBelonging < min) { min = loveBelonging; urgent = "love_belonging"; }
        if (esteem < min) { min = esteem; urgent = "esteem"; }
        if (selfActualization < min) { urgent = "self_actualization"; }

        return urgent;
    }

    /**
     * 获取需求层次名称（中文）
     */
    @NonNull
    public String getNeedName(@NonNull String needType) {
        switch (needType) {
            case "physiological":
                return "生理需求";
            case "safety":
                return "安全需求";
            case "love_belonging":
                return "爱与归属";
            case "esteem":
                return "尊重需求";
            case "self_actualization":
                return "自我实现";
            default:
                return needType;
        }
    }

    /**
     * 获取需求描述
     */
    @NonNull
    public String getNeedDescription(@NonNull String needType) {
        switch (needType) {
            case "physiological":
                return "食物、睡眠、安全感";
            case "safety":
                return "稳定、保障、秩序";
            case "love_belonging":
                return "友谊、爱情、家庭";
            case "esteem":
                return "自尊、认可、成就";
            case "self_actualization":
                return "潜能、创造力、意义";
            default:
                return "";
        }
    }

    /**
     * 是否底层需求紧迫（生理或安全）
     */
    public boolean isBasicNeedsUrgent() {
        return physiological < 0.4 || safety < 0.4;
    }

    /**
     * 是否高层次需求驱动（尊重或自我实现）
     */
    public boolean isHighLevelNeedDriven() {
        return esteem < 0.5 || selfActualization < 0.4;
    }

    /**
     * 是否需要行动（高挫折度）
     */
    public boolean needsAction() {
        return frustrationLevel > 0.5;
    }

    /**
     * 是否整体满足良好
     */
    public boolean isSatisfied() {
        return satisfactionLevel > 0.6;
    }

    /**
     * 是否处于匮乏状态
     */
    public boolean isDeficient() {
        return satisfactionLevel < 0.4;
    }

    /**
     * 获取当前驱动力描述
     */
    @NonNull
    public String getDriveDescription() {
        String needName = getNeedName(primaryDesire);
        String level;

        if (desireStrength > 0.7) {
            level = "强烈";
        } else if (desireStrength > 0.4) {
            level = "中等";
        } else {
            level = "轻微";
        }

        return level + "的" + needName + "驱动";
    }

    /**
     * 获取建议行为类型
     */
    @NonNull
    public String getRecommendedBehaviorType() {
        String urgent = getMostUrgentNeed();
        switch (urgent) {
            case "physiological":
                return "rest";  // 休息补充
            case "safety":
                return "secure"; // 寻求安全感
            case "love_belonging":
                return "connect"; // 连接他人
            case "esteem":
                return "achieve"; // 追求成就
            case "self_actualization":
                return "grow"; // 自我成长
            default:
                return "neutral";
        }
    }

    /**
     * 获取满足度层级（用于可视化）
     */
    @NonNull
    public List<NeedLevel> getNeedLevels() {
        List<NeedLevel> levels = new ArrayList<>();

        levels.add(new NeedLevel(1, "生理需求", physiological, "食物、睡眠、安全感"));
        levels.add(new NeedLevel(2, "安全需求", safety, "稳定、保障、秩序"));
        levels.add(new NeedLevel(3, "爱与归属", loveBelonging, "友谊、爱情、家庭"));
        levels.add(new NeedLevel(4, "尊重需求", esteem, "自尊、认可、成就"));
        levels.add(new NeedLevel(5, "自我实现", selfActualization, "潜能、创造力、意义"));

        return levels;
    }

    /**
     * 需求层级（用于可视化）
     */
    public static class NeedLevel {
        public int level;
        public String name;
        public double satisfaction;
        public String description;

        public NeedLevel(int level, String name, double satisfaction, String description) {
            this.level = level;
            this.name = name;
            this.satisfaction = satisfaction;
            this.description = description;
        }

        public double getUrgency() {
            return 1.0 - satisfaction;
        }

        public String getColorCode() {
            if (satisfaction > 0.7) return "#4CAF50";  // 绿色 - 满足
            if (satisfaction > 0.5) return "#FFC107";  // 黄色 - 中等
            if (satisfaction > 0.3) return "#FF9800";  // 橙色 - 偏低
            return "#F44336";                          // 红色 - 紧迫
        }
    }

    // === Getter/Setter ===

    public double getPhysiological() { return physiological; }
    public void setPhysiological(double physiological) { this.physiological = clamp01(physiological); }

    public double getSafety() { return safety; }
    public void setSafety(double safety) { this.safety = clamp01(safety); }

    public double getLoveBelonging() { return loveBelonging; }
    public void setLoveBelonging(double loveBelonging) { this.loveBelonging = clamp01(loveBelonging); }

    public double getEsteem() { return esteem; }
    public void setEsteem(double esteem) { this.esteem = clamp01(esteem); }

    public double getSelfActualization() { return selfActualization; }
    public void setSelfActualization(double selfActualization) { this.selfActualization = clamp01(selfActualization); }

    public String getPrimaryDesire() { return primaryDesire; }
    public void setPrimaryDesire(String primaryDesire) { this.primaryDesire = primaryDesire; }

    public double getDesireStrength() { return desireStrength; }
    public void setDesireStrength(double desireStrength) { this.desireStrength = clamp01(desireStrength); }

    public String getDesireTarget() { return desireTarget; }
    public void setDesireTarget(String desireTarget) { this.desireTarget = desireTarget; }

    public String getDesireAction() { return desireAction; }
    public void setDesireAction(String desireAction) { this.desireAction = desireAction; }

    public double getSatisfactionLevel() { return satisfactionLevel; }
    public void setSatisfactionLevel(double satisfactionLevel) { this.satisfactionLevel = clamp01(satisfactionLevel); }

    public double getFrustrationLevel() { return frustrationLevel; }
    public void setFrustrationLevel(double frustrationLevel) { this.frustrationLevel = clamp01(frustrationLevel); }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "DesireState{" +
                "primary='" + primaryDesire + '\'' +
                ", strength=" + desireStrength +
                ", satisfaction=" + satisfactionLevel +
                ", frustration=" + frustrationLevel +
                '}';
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }
}