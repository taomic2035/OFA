package com.ofa.agent.culture;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 地域文化状态模型 (v4.3.0)
 *
 * 端侧只接收 Center 推送的地域文化状态，用于调整决策倾向。
 * 深层地域文化管理在 Center 端 RegionalCultureEngine 完成。
 */
public class RegionalCultureState {

    // === 基本信息 ===
    private String province;          // 省份
    private String city;              // 城市
    private String cityTier;          // 城市等级
    private String region;            // 大区域

    // === 文化特征 ===
    private String dialect;           // 方言
    private double dialectProficiency; // 方言熟练度
    private List<String> regionalTraits; // 地域性格特征
    private List<String> customs;     // 习俗

    // === 文化维度 ===
    private double collectivism;      // 集体主义倾向
    private double traditionOriented; // 传统导向
    private double innovationOriented; // 创新导向
    private double powerDistance;     // 权力距离
    private double uncertaintyAvoidance; // 不确定性规避
    private double longTermOrientation; // 长期导向

    // === 风格特征 ===
    private String communicationStyle; // 沟通风格
    private double expressionLevel;    // 表达开放度
    private String socialStyle;        // 社交风格
    private double hospitality;        // 好客程度
    private double faceConscious;      // 面子意识
    private String lifePace;           // 生活节奏
    private double timeUrgency;        // 时间紧迫感

    // === 迁移信息 ===
    private boolean nativeBorn;        // 是否本地人
    private double culturalAdaptation; // 文化适应能力
    private int migrationCount;        // 迁移次数

    // === 决策倾向 ===
    private CulturalTendencies tendencies;

    // === 时间属性 ===
    private long timestamp;

    public RegionalCultureState() {
        this.cityTier = "tier2";
        this.region = "east";
        this.dialectProficiency = 0.5;
        this.regionalTraits = new ArrayList<>();
        this.customs = new ArrayList<>();
        this.collectivism = 0.5;
        this.traditionOriented = 0.4;
        this.innovationOriented = 0.6;
        this.powerDistance = 0.5;
        this.uncertaintyAvoidance = 0.5;
        this.longTermOrientation = 0.6;
        this.communicationStyle = "mixed";
        this.expressionLevel = 0.5;
        this.socialStyle = "pragmatic";
        this.hospitality = 0.6;
        this.faceConscious = 0.5;
        this.lifePace = "moderate";
        this.timeUrgency = 0.5;
        this.nativeBorn = true;
        this.culturalAdaptation = 0.5;
        this.migrationCount = 0;
        this.tendencies = new CulturalTendencies();
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static RegionalCultureState fromJson(@NonNull JSONObject json) throws JSONException {
        RegionalCultureState state = new RegionalCultureState();

        // 基本信息
        state.province = json.optString("province", "");
        state.city = json.optString("city", "");
        state.cityTier = json.optString("city_tier", "tier2");
        state.region = json.optString("region", "east");

        // 文化特征
        state.dialect = json.optString("dialect", "");
        state.dialectProficiency = json.optDouble("dialect_proficiency", 0.5);
        state.regionalTraits = parseStringList(json, "regional_traits");
        state.customs = parseStringList(json, "customs");

        // 文化维度
        state.collectivism = json.optDouble("collectivism", 0.5);
        state.traditionOriented = json.optDouble("tradition_oriented", 0.4);
        state.innovationOriented = json.optDouble("innovation_oriented", 0.6);
        state.powerDistance = json.optDouble("power_distance", 0.5);
        state.uncertaintyAvoidance = json.optDouble("uncertainty_avoidance", 0.5);
        state.longTermOrientation = json.optDouble("long_term_orientation", 0.6);

        // 风格特征
        state.communicationStyle = json.optString("communication_style", "mixed");
        state.expressionLevel = json.optDouble("expression_level", 0.5);
        state.socialStyle = json.optString("social_style", "pragmatic");
        state.hospitality = json.optDouble("hospitality", 0.6);
        state.faceConscious = json.optDouble("face_conscious", 0.5);
        state.lifePace = json.optString("life_pace", "moderate");
        state.timeUrgency = json.optDouble("time_urgency", 0.5);

        // 迁移信息
        state.nativeBorn = json.optBoolean("native", true);
        state.culturalAdaptation = json.optDouble("cultural_adaptation", 0.5);
        state.migrationCount = json.optInt("migration_count", 0);

        // 决策倾向
        JSONObject tendencies = json.optJSONObject("cultural_tendencies");
        if (tendencies != null) {
            state.tendencies = CulturalTendencies.fromJson(tendencies);
        }

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
            json.put("province", province);
            json.put("city", city);
            json.put("city_tier", cityTier);
            json.put("region", region);
            json.put("dialect", dialect);
            json.put("dialect_proficiency", dialectProficiency);
            json.put("regional_traits", listToJson(regionalTraits));
            json.put("customs", listToJson(customs));
            json.put("collectivism", collectivism);
            json.put("tradition_oriented", traditionOriented);
            json.put("innovation_oriented", innovationOriented);
            json.put("power_distance", powerDistance);
            json.put("uncertainty_avoidance", uncertaintyAvoidance);
            json.put("long_term_orientation", longTermOrientation);
            json.put("communication_style", communicationStyle);
            json.put("expression_level", expressionLevel);
            json.put("social_style", socialStyle);
            json.put("hospitality", hospitality);
            json.put("face_conscious", faceConscious);
            json.put("life_pace", lifePace);
            json.put("time_urgency", timeUrgency);
            json.put("native", nativeBorn);
            json.put("cultural_adaptation", culturalAdaptation);
            json.put("migration_count", migrationCount);
            if (tendencies != null) {
                json.put("cultural_tendencies", tendencies.toJson());
            }
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    // === 描述方法 ===

    /**
     * 获取城市等级名称
     */
    @NonNull
    public String getCityTierName() {
        switch (cityTier) {
            case "tier1": return "一线城市";
            case "new_tier1": return "新一线城市";
            case "tier2": return "二线城市";
            case "tier3": return "三线城市";
            case "tier4": return "四线城市";
            case "tier5": return "五线及以下";
            default: return cityTier;
        }
    }

    /**
     * 获取大区域名称
     */
    @NonNull
    public String getRegionName() {
        switch (region) {
            case "north": return "华北";
            case "northeast": return "东北";
            case "east": return "华东";
            case "south": return "华南";
            case "central": return "华中";
            case "southwest": return "西南";
            case "northwest": return "西北";
            default: return region;
        }
    }

    /**
     * 获取沟通风格名称
     */
    @NonNull
    public String getCommunicationStyleName() {
        switch (communicationStyle) {
            case "direct": return "直接表达";
            case "indirect": return "委婉含蓄";
            case "mixed": return "因地制宜";
            default: return communicationStyle;
        }
    }

    /**
     * 获取社交风格名称
     */
    @NonNull
    public String getSocialStyleName() {
        switch (socialStyle) {
            case "reserved": return "内敛含蓄";
            case "open": return "开放热情";
            case "warm": return "热情好客";
            case "pragmatic": return "务实理性";
            default: return socialStyle;
        }
    }

    /**
     * 是否大都市
     */
    public boolean isMetropolitan() {
        return "tier1".equals(cityTier) || "new_tier1".equals(cityTier);
    }

    /**
     * 是否集体主义倾向
     */
    public boolean isCollectivist() {
        return collectivism > 0.6;
    }

    /**
     * 是否传统导向
     */
    public boolean isTraditional() {
        return traditionOriented > innovationOriented;
    }

    /**
     * 是否直接沟通
     */
    public boolean isDirectCommunication() {
        return "direct".equals(communicationStyle);
    }

    /**
     * 是否高面子意识
     */
    public boolean hasHighFaceConsciousness() {
        return faceConscious > 0.6;
    }

    /**
     * 是否快节奏
     */
    public boolean isFastPaced() {
        return "fast".equals(lifePace) || timeUrgency > 0.7;
    }

    /**
     * 是否有迁移经历
     */
    public boolean hasMigrationExperience() {
        return migrationCount > 0;
    }

    /**
     * 获取文化描述
     */
    @NonNull
    public String getCultureDescription() {
        StringBuilder desc = new StringBuilder();

        if (province != null && !province.isEmpty()) {
            desc.append("来自").append(province);
            if (city != null && !city.isEmpty()) {
                desc.append(city);
            }
            desc.append("，");
        }

        desc.append(getRegionName()).append("地区。");

        if (collectivism > 0.6) {
            desc.append("注重集体和人际关系，");
        } else if (collectivism < 0.4) {
            desc.append("倾向于个人独立，");
        }

        if (traditionOriented > innovationOriented) {
            desc.append("重视传统价值。");
        } else if (innovationOriented > traditionOriented) {
            desc.append("开放接受新事物。");
        } else {
            desc.append("在传统与创新间保持平衡。");
        }

        return desc.toString();
    }

    // === Getter/Setter ===

    public String getProvince() { return province; }
    public void setProvince(String province) { this.province = province; }

    public String getCity() { return city; }
    public void setCity(String city) { this.city = city; }

    public String getCityTier() { return cityTier; }
    public void setCityTier(String cityTier) { this.cityTier = cityTier; }

    public String getRegion() { return region; }
    public void setRegion(String region) { this.region = region; }

    public String getDialect() { return dialect; }
    public void setDialect(String dialect) { this.dialect = dialect; }

    public double getDialectProficiency() { return dialectProficiency; }
    public void setDialectProficiency(double dialectProficiency) { this.dialectProficiency = clamp01(dialectProficiency); }

    public List<String> getRegionalTraits() { return regionalTraits; }
    public void setRegionalTraits(List<String> regionalTraits) { this.regionalTraits = regionalTraits; }

    public List<String> getCustoms() { return customs; }
    public void setCustoms(List<String> customs) { this.customs = customs; }

    public double getCollectivism() { return collectivism; }
    public void setCollectivism(double collectivism) { this.collectivism = clamp01(collectivism); }

    public double getTraditionOriented() { return traditionOriented; }
    public void setTraditionOriented(double traditionOriented) { this.traditionOriented = clamp01(traditionOriented); }

    public double getInnovationOriented() { return innovationOriented; }
    public void setInnovationOriented(double innovationOriented) { this.innovationOriented = clamp01(innovationOriented); }

    public double getPowerDistance() { return powerDistance; }
    public void setPowerDistance(double powerDistance) { this.powerDistance = clamp01(powerDistance); }

    public double getUncertaintyAvoidance() { return uncertaintyAvoidance; }
    public void setUncertaintyAvoidance(double uncertaintyAvoidance) { this.uncertaintyAvoidance = clamp01(uncertaintyAvoidance); }

    public double getLongTermOrientation() { return longTermOrientation; }
    public void setLongTermOrientation(double longTermOrientation) { this.longTermOrientation = clamp01(longTermOrientation); }

    public String getCommunicationStyle() { return communicationStyle; }
    public void setCommunicationStyle(String communicationStyle) { this.communicationStyle = communicationStyle; }

    public double getExpressionLevel() { return expressionLevel; }
    public void setExpressionLevel(double expressionLevel) { this.expressionLevel = clamp01(expressionLevel); }

    public String getSocialStyle() { return socialStyle; }
    public void setSocialStyle(String socialStyle) { this.socialStyle = socialStyle; }

    public double getHospitality() { return hospitality; }
    public void setHospitality(double hospitality) { this.hospitality = clamp01(hospitality); }

    public double getFaceConscious() { return faceConscious; }
    public void setFaceConscious(double faceConscious) { this.faceConscious = clamp01(faceConscious); }

    public String getLifePace() { return lifePace; }
    public void setLifePace(String lifePace) { this.lifePace = lifePace; }

    public double getTimeUrgency() { return timeUrgency; }
    public void setTimeUrgency(double timeUrgency) { this.timeUrgency = clamp01(timeUrgency); }

    public boolean isNativeBorn() { return nativeBorn; }
    public void setNativeBorn(boolean nativeBorn) { this.nativeBorn = nativeBorn; }

    public double getCulturalAdaptation() { return culturalAdaptation; }
    public void setCulturalAdaptation(double culturalAdaptation) { this.culturalAdaptation = clamp01(culturalAdaptation); }

    public int getMigrationCount() { return migrationCount; }
    public void setMigrationCount(int migrationCount) { this.migrationCount = migrationCount; }

    public CulturalTendencies getTendencies() { return tendencies; }
    public void setTendencies(CulturalTendencies tendencies) { this.tendencies = tendencies; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "RegionalCultureState{" +
                "province='" + province + '\'' +
                ", region='" + getRegionName() + '\'' +
                ", style='" + getCommunicationStyleName() + '\'' +
                '}';
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }

    private static List<String> parseStringList(JSONObject json, String key) {
        List<String> result = new ArrayList<>();
        JSONArray array = json.optJSONArray(key);
        if (array != null) {
            for (int i = 0; i < array.length(); i++) {
                result.add(array.optString(i));
            }
        }
        return result;
    }

    private JSONArray listToJson(List<String> list) {
        JSONArray array = new JSONArray();
        for (String s : list) {
            array.put(s);
        }
        return array;
    }

    /**
     * 文化决策倾向
     */
    public static class CulturalTendencies {
        private double groupDecisionPreference;
        private double socialConformity;
        private double traditionRespect;
        private double innovationOpenness;
        private double directness;
        private double harmonyMaintenance;
        private double hospitalityTendency;
        private double facePreservation;
        private double pacePreference;
        private double culturalFlexibility;

        public CulturalTendencies() {
            this.groupDecisionPreference = 0.5;
            this.socialConformity = 0.5;
            this.traditionRespect = 0.5;
            this.innovationOpenness = 0.5;
            this.directness = 0.5;
            this.harmonyMaintenance = 0.5;
            this.hospitalityTendency = 0.5;
            this.facePreservation = 0.5;
            this.pacePreference = 0.5;
            this.culturalFlexibility = 0.5;
        }

        public static CulturalTendencies fromJson(JSONObject json) {
            CulturalTendencies t = new CulturalTendencies();
            t.groupDecisionPreference = json.optDouble("group_decision_preference", 0.5);
            t.socialConformity = json.optDouble("social_conformity", 0.5);
            t.traditionRespect = json.optDouble("tradition_respect", 0.5);
            t.innovationOpenness = json.optDouble("innovation_openness", 0.5);
            t.directness = json.optDouble("directness", 0.5);
            t.harmonyMaintenance = json.optDouble("harmony_maintenance", 0.5);
            t.hospitalityTendency = json.optDouble("hospitality_tendency", 0.5);
            t.facePreservation = json.optDouble("face_preservation", 0.5);
            t.pacePreference = json.optDouble("pace_preference", 0.5);
            t.culturalFlexibility = json.optDouble("cultural_flexibility", 0.5);
            return t;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("group_decision_preference", groupDecisionPreference);
                json.put("social_conformity", socialConformity);
                json.put("tradition_respect", traditionRespect);
                json.put("innovation_openness", innovationOpenness);
                json.put("directness", directness);
                json.put("harmony_maintenance", harmonyMaintenance);
                json.put("hospitality_tendency", hospitalityTendency);
                json.put("face_preservation", facePreservation);
                json.put("pace_preference", pacePreference);
                json.put("cultural_flexibility", culturalFlexibility);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public double getGroupDecisionPreference() { return groupDecisionPreference; }
        public double getSocialConformity() { return socialConformity; }
        public double getTraditionRespect() { return traditionRespect; }
        public double getInnovationOpenness() { return innovationOpenness; }
        public double getDirectness() { return directness; }
        public double getHarmonyMaintenance() { return harmonyMaintenance; }
        public double getHospitalityTendency() { return hospitalityTendency; }
        public double getFacePreservation() { return facePreservation; }
        public double getPacePreference() { return pacePreference; }
        public double getCulturalFlexibility() { return culturalFlexibility; }
    }
}