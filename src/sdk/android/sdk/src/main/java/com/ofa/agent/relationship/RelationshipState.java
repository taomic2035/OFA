package com.ofa.agent.relationship;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * 人际关系状态模型 (v4.6.0)
 *
 * 端侧只接收 Center 推送的人际关系状态，用于调整社交决策。
 * 深层人际关系管理在 Center 端 RelationshipEngine 完成。
 */
public class RelationshipState {

    // === 网络摘要 ===
    private NetworkSummary networkSummary;

    // === 关系倾向 ===
    private RelationshipOrientation orientation;

    // === 社交风格 ===
    private SocialStyle socialStyle;

    // === 依恋风格 ===
    private AttachmentStyle attachmentStyle;

    // === 决策影响 ===
    private RelationshipDecisionInfluence decisionInfluence;

    // === 社交建议 ===
    private SocialGuidance socialGuidance;

    // === 亲密关系 ===
    private List<RelationshipSummary> closeRelationships;

    // === 时间属性 ===
    private long timestamp;

    public RelationshipState() {
        this.networkSummary = new NetworkSummary();
        this.orientation = new RelationshipOrientation();
        this.socialStyle = new SocialStyle();
        this.attachmentStyle = new AttachmentStyle();
        this.decisionInfluence = new RelationshipDecisionInfluence();
        this.socialGuidance = new SocialGuidance();
        this.closeRelationships = new ArrayList<>();
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static RelationshipState fromJson(@NonNull JSONObject json) throws JSONException {
        RelationshipState state = new RelationshipState();

        // 网络摘要
        JSONObject networkJson = json.optJSONObject("network_summary");
        if (networkJson != null) {
            state.networkSummary = NetworkSummary.fromJson(networkJson);
        }

        // 关系倾向
        JSONObject orientationJson = json.optJSONObject("relationship_orientation");
        if (orientationJson != null) {
            state.orientation = RelationshipOrientation.fromJson(orientationJson);
        }

        // 社交风格
        JSONObject socialJson = json.optJSONObject("social_style");
        if (socialJson != null) {
            state.socialStyle = SocialStyle.fromJson(socialJson);
        }

        // 依恋风格
        JSONObject attachJson = json.optJSONObject("attachment_style");
        if (attachJson != null) {
            state.attachmentStyle = AttachmentStyle.fromJson(attachJson);
        }

        // 决策影响
        JSONObject decisionJson = json.optJSONObject("decision_influence");
        if (decisionJson != null) {
            state.decisionInfluence = RelationshipDecisionInfluence.fromJson(decisionJson);
        }

        // 社交建议
        JSONObject guidanceJson = json.optJSONObject("social_guidance");
        if (guidanceJson != null) {
            state.socialGuidance = SocialGuidance.fromJson(guidanceJson);
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
            if (networkSummary != null) {
                json.put("network_summary", networkSummary.toJson());
            }
            if (orientation != null) {
                json.put("relationship_orientation", orientation.toJson());
            }
            if (socialStyle != null) {
                json.put("social_style", socialStyle.toJson());
            }
            if (attachmentStyle != null) {
                json.put("attachment_style", attachmentStyle.toJson());
            }
            if (decisionInfluence != null) {
                json.put("decision_influence", decisionInfluence.toJson());
            }
            if (socialGuidance != null) {
                json.put("social_guidance", socialGuidance.toJson());
            }
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    // === 便捷方法 ===

    /**
     * 是否应社交
     */
    public boolean shouldSocialize() {
        return socialGuidance != null && socialGuidance.shouldSocialize;
    }

    /**
     * 是否需要自我关怀
     */
    public boolean needsSelfCare() {
        return socialGuidance != null && socialGuidance.selfCareNeeded;
    }

    /**
     * 是否安全依恋
     */
    public boolean isSecurelyAttached() {
        return attachmentStyle != null && attachmentStyle.isSecurelyAttached();
    }

    /**
     * 是否焦虑依恋
     */
    public boolean isAnxiouslyAttached() {
        return attachmentStyle != null && attachmentStyle.isAnxiouslyAttached();
    }

    /**
     * 是否回避依恋
     */
    public boolean isAvoidantlyAttached() {
        return attachmentStyle != null && attachmentStyle.isAvoidantlyAttached();
    }

    /**
     * 是否有亲密关系
     */
    public boolean hasCloseRelationships() {
        return networkSummary != null && networkSummary.closeContacts > 0;
    }

    /**
     * 获取社交能量水平
     */
    public double getSocialEnergyLevel() {
        return socialGuidance != null ? socialGuidance.socialEnergyLevel : 0.5;
    }

    /**
     * 获取冲突风险
     */
    public double getConflictRisk() {
        return socialGuidance != null ? socialGuidance.conflictRisk : 0;
    }

    // === Getter/Setter ===

    public NetworkSummary getNetworkSummary() { return networkSummary; }
    public void setNetworkSummary(NetworkSummary networkSummary) { this.networkSummary = networkSummary; }

    public RelationshipOrientation getOrientation() { return orientation; }
    public void setOrientation(RelationshipOrientation orientation) { this.orientation = orientation; }

    public SocialStyle getSocialStyle() { return socialStyle; }
    public void setSocialStyle(SocialStyle socialStyle) { this.socialStyle = socialStyle; }

    public AttachmentStyle getAttachmentStyle() { return attachmentStyle; }
    public void setAttachmentStyle(AttachmentStyle attachmentStyle) { this.attachmentStyle = attachmentStyle; }

    public RelationshipDecisionInfluence getDecisionInfluence() { return decisionInfluence; }
    public void setDecisionInfluence(RelationshipDecisionInfluence decisionInfluence) { this.decisionInfluence = decisionInfluence; }

    public SocialGuidance getSocialGuidance() { return socialGuidance; }
    public void setSocialGuidance(SocialGuidance socialGuidance) { this.socialGuidance = socialGuidance; }

    public List<RelationshipSummary> getCloseRelationships() { return closeRelationships; }
    public void setCloseRelationships(List<RelationshipSummary> closeRelationships) { this.closeRelationships = closeRelationships; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "RelationshipState{" +
                "contacts=" + (networkSummary != null ? networkSummary.totalContacts : 0) +
                ", closeContacts=" + (networkSummary != null ? networkSummary.closeContacts : 0) +
                ", attachment=" + (attachmentStyle != null ? attachmentStyle.primaryStyle : "unknown") +
                '}';
    }

    /**
     * 网络摘要
     */
    public static class NetworkSummary {
        public int totalContacts;
        public int closeContacts;
        public int supportContacts;
        public double socialCapital;
        public int strongTies;
        public int weakTies;
        public double networkHealth;

        public NetworkSummary() {}

        public static NetworkSummary fromJson(JSONObject json) {
            NetworkSummary n = new NetworkSummary();
            n.totalContacts = json.optInt("total_contacts", 0);
            n.closeContacts = json.optInt("close_contacts", 0);
            n.supportContacts = json.optInt("support_contacts", 0);
            n.socialCapital = json.optDouble("social_capital", 0);
            n.strongTies = json.optInt("strong_ties", 0);
            n.weakTies = json.optInt("weak_ties", 0);
            n.networkHealth = json.optDouble("network_health", 0.5);
            return n;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("total_contacts", totalContacts);
                json.put("close_contacts", closeContacts);
                json.put("support_contacts", supportContacts);
                json.put("social_capital", socialCapital);
                json.put("strong_ties", strongTies);
                json.put("weak_ties", weakTies);
                json.put("network_health", networkHealth);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean hasSupport() { return supportContacts > 0; }
        public boolean hasStrongNetwork() { return strongTies >= 3; }
    }

    /**
     * 关系倾向
     */
    public static class RelationshipOrientation {
        public String socialOrientation; // extraverted/introverted/ambiverted
        public double relationshipSeeking;
        public double independencePreference;
        public double commitmentReadiness;
        public double intimacyComfort;
        public double vulnerabilityWillingness;
        public int idealRelationshipCount;
        public double qualityVsQuantity;

        public RelationshipOrientation() {
            this.socialOrientation = "ambiverted";
            this.relationshipSeeking = 0.5;
            this.independencePreference = 0.5;
        }

        public static RelationshipOrientation fromJson(JSONObject json) {
            RelationshipOrientation o = new RelationshipOrientation();
            o.socialOrientation = json.optString("social_orientation", "ambiverted");
            o.relationshipSeeking = json.optDouble("relationship_seeking", 0.5);
            o.independencePreference = json.optDouble("independence_preference", 0.5);
            o.commitmentReadiness = json.optDouble("commitment_readiness", 0.5);
            o.intimacyComfort = json.optDouble("intimacy_comfort", 0.5);
            o.vulnerabilityWillingness = json.optDouble("vulnerability_willingness", 0.5);
            o.idealRelationshipCount = json.optInt("ideal_relationship_count", 5);
            o.qualityVsQuantity = json.optDouble("quality_vs_quantity", 0.6);
            return o;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("social_orientation", socialOrientation);
                json.put("relationship_seeking", relationshipSeeking);
                json.put("independence_preference", independencePreference);
                json.put("commitment_readiness", commitmentReadiness);
                json.put("intimacy_comfort", intimacyComfort);
                json.put("vulnerability_willingness", vulnerabilityWillingness);
                json.put("ideal_relationship_count", idealRelationshipCount);
                json.put("quality_vs_quantity", qualityVsQuantity);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isExtraverted() { return "extraverted".equals(socialOrientation); }
        public boolean isIntroverted() { return "introverted".equals(socialOrientation); }
    }

    /**
     * 社交风格
     */
    public static class SocialStyle {
        public double directness;
        public double expressiveness;
        public String listeningStyle;
        public String initiationStyle;
        public double groupPreference;
        public double smallTalkComfort;
        public double deepTalkPreference;
        public double socialEnergy;
        public String rechargeStyle;

        public SocialStyle() {
            this.listeningStyle = "active";
            this.initiationStyle = "proactive";
            this.rechargeStyle = "alone";
        }

        public static SocialStyle fromJson(JSONObject json) {
            SocialStyle s = new SocialStyle();
            s.directness = json.optDouble("directness", 0.5);
            s.expressiveness = json.optDouble("expressiveness", 0.5);
            s.listeningStyle = json.optString("listening_style", "active");
            s.initiationStyle = json.optString("initiation_style", "proactive");
            s.groupPreference = json.optDouble("group_preference", 0.5);
            s.smallTalkComfort = json.optDouble("small_talk_comfort", 0.5);
            s.deepTalkPreference = json.optDouble("deep_talk_preference", 0.6);
            s.socialEnergy = json.optDouble("social_energy", 0.5);
            s.rechargeStyle = json.optString("recharge_style", "alone");
            return s;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("directness", directness);
                json.put("expressiveness", expressiveness);
                json.put("listening_style", listeningStyle);
                json.put("initiation_style", initiationStyle);
                json.put("group_preference", groupPreference);
                json.put("small_talk_comfort", smallTalkComfort);
                json.put("deep_talk_preference", deepTalkPreference);
                json.put("social_energy", socialEnergy);
                json.put("recharge_style", rechargeStyle);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean prefersOneOnOne() { return groupPreference < 0.4; }
        public boolean prefersGroup() { return groupPreference > 0.6; }
        public boolean hasHighSocialEnergy() { return socialEnergy > 0.6; }
        public boolean hasLowSocialEnergy() { return socialEnergy < 0.4; }
    }

    /**
     * 依恋风格
     */
    public static class AttachmentStyle {
        public String primaryStyle; // secure/anxious/avoidant/disorganized
        public double anxietyLevel;
        public double avoidanceLevel;
        public double separationAnxiety;
        public double proximitySeeking;
        public double safeHavenUse;
        public double secureBaseUse;

        public AttachmentStyle() {
            this.primaryStyle = "secure";
            this.anxietyLevel = 0.3;
            this.avoidanceLevel = 0.3;
        }

        public static AttachmentStyle fromJson(JSONObject json) {
            AttachmentStyle a = new AttachmentStyle();
            a.primaryStyle = json.optString("primary_style", "secure");
            a.anxietyLevel = json.optDouble("anxiety_level", 0.3);
            a.avoidanceLevel = json.optDouble("avoidance_level", 0.3);
            a.separationAnxiety = json.optDouble("separation_anxiety", 0.3);
            a.proximitySeeking = json.optDouble("proximity_seeking", 0.5);
            a.safeHavenUse = json.optDouble("safe_haven_use", 0.5);
            a.secureBaseUse = json.optDouble("secure_base_use", 0.5);
            return a;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("primary_style", primaryStyle);
                json.put("anxiety_level", anxietyLevel);
                json.put("avoidance_level", avoidanceLevel);
                json.put("separation_anxiety", separationAnxiety);
                json.put("proximity_seeking", proximitySeeking);
                json.put("safe_haven_use", safeHavenUse);
                json.put("secure_base_use", secureBaseUse);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isSecurelyAttached() { return "secure".equals(primaryStyle) && anxietyLevel < 0.4 && avoidanceLevel < 0.4; }
        public boolean isAnxiouslyAttached() { return anxietyLevel > 0.6; }
        public boolean isAvoidantlyAttached() { return avoidanceLevel > 0.6; }
    }

    /**
     * 关系决策影响
     */
    public static class RelationshipDecisionInfluence {
        public double socialApproach;
        public double trustInOthers;
        public double cooperationReadiness;
        public double vulnerabilityComfort;
        public double commitmentWillingness;
        public double supportWillingness;
        public double intimacySeeking;
        public String conflictStyle;
        public double confrontationReadiness;
        public double forgivenessReadiness;
        public double securityLevel;
        public double separationAnxiety;
        public double reassuranceNeed;

        public RelationshipDecisionInfluence() {
            this.conflictStyle = "collaborating";
        }

        public static RelationshipDecisionInfluence fromJson(JSONObject json) {
            RelationshipDecisionInfluence d = new RelationshipDecisionInfluence();
            d.socialApproach = json.optDouble("social_approach", 0.5);
            d.trustInOthers = json.optDouble("trust_in_others", 0.5);
            d.cooperationReadiness = json.optDouble("cooperation_readiness", 0.5);
            d.vulnerabilityComfort = json.optDouble("vulnerability_comfort", 0.5);
            d.commitmentWillingness = json.optDouble("commitment_willingness", 0.5);
            d.supportWillingness = json.optDouble("support_willingness", 0.5);
            d.intimacySeeking = json.optDouble("intimacy_seeking", 0.5);
            d.conflictStyle = json.optString("conflict_style", "collaborating");
            d.confrontationReadiness = json.optDouble("confrontation_readiness", 0.5);
            d.forgivenessReadiness = json.optDouble("forgiveness_readiness", 0.6);
            d.securityLevel = json.optDouble("security_level", 0.5);
            d.separationAnxiety = json.optDouble("separation_anxiety", 0.3);
            d.reassuranceNeed = json.optDouble("reassurance_need", 0.3);
            return d;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("social_approach", socialApproach);
                json.put("trust_in_others", trustInOthers);
                json.put("cooperation_readiness", cooperationReadiness);
                json.put("vulnerability_comfort", vulnerabilityComfort);
                json.put("commitment_willingness", commitmentWillingness);
                json.put("support_willingness", supportWillingness);
                json.put("intimacy_seeking", intimacySeeking);
                json.put("conflict_style", conflictStyle);
                json.put("confrontation_readiness", confrontationReadiness);
                json.put("forgiveness_readiness", forgivenessReadiness);
                json.put("security_level", securityLevel);
                json.put("separation_anxiety", separationAnxiety);
                json.put("reassurance_need", reassuranceNeed);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isSociallyApproach() { return socialApproach > 0.6; }
        public boolean isTrusting() { return trustInOthers > 0.6; }
        public boolean isCooperative() { return cooperationReadiness > 0.6; }
        public boolean hasHighSecurity() { return securityLevel > 0.6; }
    }

    /**
     * 社交建议
     */
    public static class SocialGuidance {
        public boolean shouldSocialize;
        public double socialEnergyLevel;
        public String preferredGroupSize;
        public String preferredDepth;
        public List<String> maintenanceNeeded;
        public List<String> outreachSuggestions;
        public double conflictRisk;
        public List<String> tensionRelationships;
        public List<String> boundaryReminders;
        public boolean selfCareNeeded;
        public List<String> growthOpportunities;
        public List<String> newConnectionSuggestions;

        public SocialGuidance() {
            this.maintenanceNeeded = new ArrayList<>();
            this.outreachSuggestions = new ArrayList<>();
            this.tensionRelationships = new ArrayList<>();
            this.boundaryReminders = new ArrayList<>();
            this.growthOpportunities = new ArrayList<>();
            this.newConnectionSuggestions = new ArrayList<>();
        }

        public static SocialGuidance fromJson(JSONObject json) {
            SocialGuidance g = new SocialGuidance();
            g.shouldSocialize = json.optBoolean("should_socialize", true);
            g.socialEnergyLevel = json.optDouble("social_energy_level", 0.5);
            g.preferredGroupSize = json.optString("preferred_group_size", "small");
            g.preferredDepth = json.optString("preferred_depth", "mixed");
            g.conflictRisk = json.optDouble("conflict_risk", 0);
            g.selfCareNeeded = json.optBoolean("self_care_needed", false);

            g.maintenanceNeeded = parseStringList(json, "maintenance_needed");
            g.outreachSuggestions = parseStringList(json, "outreach_suggestions");
            g.tensionRelationships = parseStringList(json, "tension_relationships");
            g.boundaryReminders = parseStringList(json, "boundary_reminders");
            g.growthOpportunities = parseStringList(json, "growth_opportunities");
            g.newConnectionSuggestions = parseStringList(json, "new_connection_suggestions");

            return g;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("should_socialize", shouldSocialize);
                json.put("social_energy_level", socialEnergyLevel);
                json.put("preferred_group_size", preferredGroupSize);
                json.put("preferred_depth", preferredDepth);
                json.put("conflict_risk", conflictRisk);
                json.put("self_care_needed", selfCareNeeded);
                json.put("maintenance_needed", listToJson(maintenanceNeeded));
                json.put("outreach_suggestions", listToJson(outreachSuggestions));
                json.put("tension_relationships", listToJson(tensionRelationships));
                json.put("boundary_reminders", listToJson(boundaryReminders));
                json.put("growth_opportunities", listToJson(growthOpportunities));
                json.put("new_connection_suggestions", listToJson(newConnectionSuggestions));
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean hasConflictRisk() { return conflictRisk > 0.5; }
        public boolean needsMaintenance() { return maintenanceNeeded.size() > 0; }
        public boolean hasTension() { return tensionRelationships.size() > 0; }

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
    }

    /**
     * 关系摘要
     */
    public static class RelationshipSummary {
        public String relationshipId;
        public String personName;
        public String relationshipType;
        public double intimacy;
        public double trust;
        public String stage;
        public String trend;

        public RelationshipSummary() {}

        public static RelationshipSummary fromJson(JSONObject json) {
            RelationshipSummary r = new RelationshipSummary();
            r.relationshipId = json.optString("relationship_id", "");
            r.personName = json.optString("person_name", "");
            r.relationshipType = json.optString("relationship_type", "other");
            r.intimacy = json.optDouble("intimacy", 0.5);
            r.trust = json.optDouble("trust", 0.5);
            r.stage = json.optString("stage", "casual");
            r.trend = json.optString("trend", "stable");
            return r;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("relationship_id", relationshipId);
                json.put("person_name", personName);
                json.put("relationship_type", relationshipType);
                json.put("intimacy", intimacy);
                json.put("trust", trust);
                json.put("stage", stage);
                json.put("trend", trend);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public boolean isClose() { return intimacy > 0.7; }
        public boolean isTrusted() { return trust > 0.7; }
    }
}