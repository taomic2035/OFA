package com.ofa.agent.identity;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * PersonalIdentity - 个人身份核心
 *
 * 描述用户的个人身份，包含基本信息、性格、价值观、兴趣等。
 * 用于跨设备身份同步，实现"万物都是我"的核心理念。
 */
public class PersonalIdentity {

    // 基本信息
    private String id;
    private String name;
    private String nickname;
    private String avatar;
    private long birthday;       // 时间戳
    private String gender;
    private String location;
    private String occupation;
    private List<String> languages;
    private String timezone;

    // 核心特质
    private Personality personality;
    private ValueSystem valueSystem;
    private List<Interest> interests;

    // 数字资产
    private VoiceProfile voiceProfile;
    private WritingStyle writingStyle;

    // 元数据
    private long createdAt;
    private long updatedAt;
    private long version;        // 版本号（用于同步）

    /**
     * 创建默认身份
     */
    public PersonalIdentity() {
        this.id = generateId();
        this.languages = new ArrayList<>(Arrays.asList("zh-CN"));
        this.timezone = "Asia/Shanghai";

        this.personality = new Personality();
        this.valueSystem = new ValueSystem();
        this.interests = new ArrayList<>();

        this.voiceProfile = new VoiceProfile();
        this.writingStyle = new WritingStyle();

        this.createdAt = System.currentTimeMillis();
        this.updatedAt = System.currentTimeMillis();
        this.version = 1;
    }

    /**
     * 创建身份（指定 ID）
     */
    public PersonalIdentity(@NonNull String id) {
        this();
        this.id = id;
    }

    // === 更新方法 ===

    /**
     * 更新性格特质
     */
    public void updatePersonality(@NonNull Map<String, Double> updates) {
        if (personality == null) {
            personality = new Personality();
        }
        personality.updateTraits(updates);
        updatedAt = System.currentTimeMillis();
        version++;
    }

    /**
     * 更新价值观
     */
    public void updateValueSystem(@NonNull Map<String, Double> updates) {
        if (valueSystem == null) {
            valueSystem = new ValueSystem();
        }
        valueSystem.updateValues(updates);
        updatedAt = System.currentTimeMillis();
        version++;
    }

    /**
     * 添加兴趣
     */
    public void addInterest(@NonNull Interest interest) {
        // 检查是否已存在
        for (int i = 0; i < interests.size(); i++) {
            Interest existing = interests.get(i);
            if (existing.getId().equals(interest.getId()) ||
                (existing.getCategory().equals(interest.getCategory()) &&
                 existing.getName().equals(interest.getName()))) {
                // 更新现有兴趣
                interests.set(i, interest);
                updatedAt = System.currentTimeMillis();
                version++;
                return;
            }
        }

        // 添加新兴趣
        interests.add(interest);
        updatedAt = System.currentTimeMillis();
        version++;
    }

    /**
     * 移除兴趣
     */
    public boolean removeInterest(@NonNull String interestId) {
        for (int i = 0; i < interests.size(); i++) {
            if (interests.get(i).getId().equals(interestId)) {
                interests.remove(i);
                updatedAt = System.currentTimeMillis();
                version++;
                return true;
            }
        }
        return false;
    }

    /**
     * 按类别获取兴趣
     */
    @NonNull
    public List<Interest> getInterestByCategory(@NonNull String category) {
        List<Interest> result = new ArrayList<>();
        for (Interest interest : interests) {
            if (interest.getCategory().equals(category)) {
                result.add(interest);
            }
        }
        return result;
    }

    /**
     * 获取最热衷的兴趣
     */
    @NonNull
    public List<Interest> getTopInterests(int limit) {
        if (interests.size() <= limit) {
            return new ArrayList<>(interests);
        }

        // 按热衷程度排序
        List<Interest> sorted = new ArrayList<>(interests);
        // 简单排序（冒泡）
        for (int i = 0; i < sorted.size() - 1; i++) {
            for (int j = i + 1; j < sorted.size(); j++) {
                if (sorted.get(j).getLevel() > sorted.get(i).getLevel()) {
                    Interest temp = sorted.get(i);
                    sorted.set(i, sorted.get(j));
                    sorted.set(j, temp);
                }
            }
        }

        return sorted.subList(0, limit);
    }

    /**
     * 获取性格描述
     */
    @NonNull
    public String getPersonalityDescription() {
        if (personality == null) {
            return "性格待完善";
        }
        return personality.getDescription();
    }

    /**
     * 获取价值观优先级
     */
    @NonNull
    public String[] getValuePriority() {
        if (valueSystem == null) {
            return new String[0];
        }
        return valueSystem.getValuePriority();
    }

    // === 辅助方法 ===

    private String generateId() {
        return System.currentTimeMillis() + "_" + Integer.toHexString((int)(Math.random() * 10000));
    }

    // === Getters/Setters ===

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getName() { return name; }
    public void setName(String name) { this.name = name; updatedAt(); }

    public String getNickname() { return nickname; }
    public void setNickname(String nickname) { this.nickname = nickname; updatedAt(); }

    public String getAvatar() { return avatar; }
    public void setAvatar(String avatar) { this.avatar = avatar; updatedAt(); }

    public long getBirthday() { return birthday; }
    public void setBirthday(long birthday) { this.birthday = birthday; updatedAt(); }

    public String getGender() { return gender; }
    public void setGender(String gender) { this.gender = gender; updatedAt(); }

    public String getLocation() { return location; }
    public void setLocation(String location) { this.location = location; updatedAt(); }

    public String getOccupation() { return occupation; }
    public void setOccupation(String occupation) { this.occupation = occupation; updatedAt(); }

    public List<String> getLanguages() { return new ArrayList<>(languages); }
    public void setLanguages(List<String> languages) { this.languages = new ArrayList<>(languages); updatedAt(); }

    public String getTimezone() { return timezone; }
    public void setTimezone(String timezone) { this.timezone = timezone; updatedAt(); }

    public Personality getPersonality() { return personality; }
    public void setPersonality(Personality personality) { this.personality = personality; updatedAt(); }

    public ValueSystem getValueSystem() { return valueSystem; }
    public void setValueSystem(ValueSystem valueSystem) { this.valueSystem = valueSystem; updatedAt(); }

    public List<Interest> getInterests() { return new ArrayList<>(interests); }
    public void setInterests(List<Interest> interests) { this.interests = new ArrayList<>(interests); updatedAt(); }

    public VoiceProfile getVoiceProfile() { return voiceProfile; }
    public void setVoiceProfile(VoiceProfile voiceProfile) { this.voiceProfile = voiceProfile; updatedAt(); }

    public WritingStyle getWritingStyle() { return writingStyle; }
    public void setWritingStyle(WritingStyle writingStyle) { this.writingStyle = writingStyle; updatedAt(); }

    public long getCreatedAt() { return createdAt; }
    public long getUpdatedAt() { return updatedAt; }
    public long getVersion() { return version; }

    private void updatedAt() {
        this.updatedAt = System.currentTimeMillis();
        this.version++;
    }

    /**
     * 转换为 JSON 字符串
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"id\":\"").append(id).append("\",");
        sb.append("\"name\":\"").append(name != null ? name : "").append("\",");
        sb.append("\"nickname\":\"").append(nickname != null ? nickname : "").append("\",");
        sb.append("\"timezone\":\"").append(timezone).append("\",");
        sb.append("\"version\":").append(version).append(",");

        if (personality != null) {
            sb.append("\"personality\":").append(personality.toJson()).append(",");
        }
        if (valueSystem != null) {
            sb.append("\"value_system\":").append(valueSystem.toJson()).append(",");
        }

        sb.append("\"created_at\":").append(createdAt).append(",");
        sb.append("\"updated_at\":").append(updatedAt);
        sb.append("}");
        return sb.toString();
    }

    /**
     * 从 JSON 解析（完整版）
     */
    @Nullable
    public static PersonalIdentity fromJson(@NonNull String json) {
        try {
            org.json.JSONObject obj = new org.json.JSONObject(json);

            PersonalIdentity identity = new PersonalIdentity();

            // 基本信息
            if (obj.has("id")) {
                identity.id = obj.getString("id");
            }
            if (obj.has("name")) {
                identity.name = obj.getString("name");
            }
            if (obj.has("nickname")) {
                identity.nickname = obj.getString("nickname");
            }
            if (obj.has("avatar")) {
                identity.avatar = obj.getString("avatar");
            }
            if (obj.has("birthday")) {
                identity.birthday = obj.getLong("birthday");
            }
            if (obj.has("gender")) {
                identity.gender = obj.getString("gender");
            }
            if (obj.has("location")) {
                identity.location = obj.getString("location");
            }
            if (obj.has("occupation")) {
                identity.occupation = obj.getString("occupation");
            }
            if (obj.has("timezone")) {
                identity.timezone = obj.getString("timezone");
            }

            // 语言列表
            if (obj.has("languages")) {
                org.json.JSONArray langArr = obj.getJSONArray("languages");
                identity.languages = new ArrayList<>();
                for (int i = 0; i < langArr.length(); i++) {
                    identity.languages.add(langArr.getString(i));
                }
            }

            // 性格
            if (obj.has("personality")) {
                identity.personality = Personality.fromJsonObject(obj.getJSONObject("personality"));
            }

            // 价值观
            if (obj.has("value_system")) {
                identity.valueSystem = ValueSystem.fromJsonObject(obj.getJSONObject("value_system"));
            }

            // 兴趣列表
            if (obj.has("interests")) {
                org.json.JSONArray interestArr = obj.getJSONArray("interests");
                identity.interests = new ArrayList<>();
                for (int i = 0; i < interestArr.length(); i++) {
                    Interest interest = Interest.fromJsonObject(interestArr.getJSONObject(i));
                    if (interest != null) {
                        identity.interests.add(interest);
                    }
                }
            }

            // 语音配置
            if (obj.has("voice_profile")) {
                identity.voiceProfile = VoiceProfile.fromJsonObject(obj.getJSONObject("voice_profile"));
            }

            // 写作风格
            if (obj.has("writing_style")) {
                identity.writingStyle = WritingStyle.fromJsonObject(obj.getJSONObject("writing_style"));
            }

            // 元数据
            if (obj.has("version")) {
                identity.version = obj.getLong("version");
            }
            if (obj.has("created_at")) {
                identity.createdAt = obj.getLong("created_at");
            }
            if (obj.has("updated_at")) {
                identity.updatedAt = obj.getLong("updated_at");
            }

            return identity;
        } catch (Exception e) {
            return null;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return String.format("PersonalIdentity{id=%s, name=%s, version=%d, interests=%d}",
            id, name, version, interests.size());
    }
}