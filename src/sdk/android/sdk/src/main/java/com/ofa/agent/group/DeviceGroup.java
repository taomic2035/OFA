package com.ofa.agent.group;

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
 * 设备群组模型 (v3.5.0)
 */
public class DeviceGroup {

    // 群组类型
    public static final String TYPE_HOME = "home";
    public static final String TYPE_WORK = "work";
    public static final String TYPE_PERSONAL = "personal";
    public static final String TYPE_TRAVEL = "travel";
    public static final String TYPE_FITNESS = "fitness";
    public static final String TYPE_CUSTOM = "custom";

    private String groupId;
    private String identityId;
    private String name;
    private String type;
    private String description;
    private List<GroupMember> members;
    private long createdAt;
    private long updatedAt;
    private boolean active;
    private int priority;
    private GroupSettings settings;
    private Map<String, String> metadata;

    public DeviceGroup() {
        this.members = new ArrayList<>();
        this.settings = new GroupSettings();
        this.metadata = new HashMap<>();
        this.active = true;
        this.priority = 50;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static DeviceGroup fromJson(@NonNull JSONObject json) throws JSONException {
        DeviceGroup group = new DeviceGroup();
        group.groupId = json.optString("group_id");
        group.identityId = json.optString("identity_id");
        group.name = json.optString("name");
        group.type = json.optString("type");
        group.description = json.optString("description", "");
        group.createdAt = json.optLong("created_at", 0);
        group.updatedAt = json.optLong("updated_at", 0);
        group.active = json.optBoolean("is_active", true);
        group.priority = json.optInt("priority", 50);

        // 解析成员
        JSONArray membersArray = json.optJSONArray("members");
        if (membersArray != null) {
            for (int i = 0; i < membersArray.length(); i++) {
                JSONObject memberJson = membersArray.getJSONObject(i);
                group.members.add(GroupMember.fromJson(memberJson));
            }
        }

        // 解析设置
        JSONObject settingsJson = json.optJSONObject("settings");
        if (settingsJson != null) {
            group.settings = GroupSettings.fromJson(settingsJson);
        }

        // 解析元数据
        JSONObject metadataJson = json.optJSONObject("metadata");
        if (metadataJson != null) {
            for (java.util.Iterator<String> it = metadataJson.keys(); it.hasNext(); ) {
                String key = it.next();
                group.metadata.put(key, metadataJson.optString(key));
            }
        }

        return group;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("group_id", groupId);
        json.put("identity_id", identityId);
        json.put("name", name);
        json.put("type", type);
        json.put("description", description);
        json.put("created_at", createdAt);
        json.put("updated_at", updatedAt);
        json.put("is_active", active);
        json.put("priority", priority);

        // 成员
        JSONArray membersArray = new JSONArray();
        for (GroupMember member : members) {
            membersArray.put(member.toJson());
        }
        json.put("members", membersArray);

        // 设置
        if (settings != null) {
            json.put("settings", settings.toJson());
        }

        // 元数据
        if (!metadata.isEmpty()) {
            JSONObject metadataJson = new JSONObject();
            for (Map.Entry<String, String> entry : metadata.entrySet()) {
                metadataJson.put(entry.getKey(), entry.getValue());
            }
            json.put("metadata", metadataJson);
        }

        return json;
    }

    /**
     * 转换为 Map
     */
    @NonNull
    public Map<String, Object> toMap() {
        Map<String, Object> map = new HashMap<>();
        map.put("group_id", groupId);
        map.put("identity_id", identityId);
        map.put("name", name);
        map.put("type", type);
        map.put("is_active", active);
        map.put("priority", priority);
        map.put("member_count", members.size());

        if (!description.isEmpty()) {
            map.put("description", description);
        }

        return map;
    }

    // === 成员操作 ===

    /**
     * 添加成员
     */
    public void addMember(@NonNull GroupMember member) {
        members.add(member);
    }

    /**
     * 移除成员
     */
    public boolean removeMember(@NonNull String agentId) {
        for (int i = 0; i < members.size(); i++) {
            if (members.get(i).getAgentId().equals(agentId)) {
                members.remove(i);
                return true;
            }
        }
        return false;
    }

    /**
     * 获取成员
     */
    @Nullable
    public GroupMember getMember(@NonNull String agentId) {
        for (GroupMember member : members) {
            if (member.getAgentId().equals(agentId)) {
                return member;
            }
        }
        return null;
    }

    /**
     * 获取所有成员
     */
    @NonNull
    public List<GroupMember> getMembers() {
        return new ArrayList<>(members);
    }

    /**
     * 获取在线成员
     */
    @NonNull
    public List<GroupMember> getOnlineMembers() {
        List<GroupMember> online = new ArrayList<>();
        for (GroupMember member : members) {
            if (member.isOnline()) {
                online.add(member);
            }
        }
        return online;
    }

    /**
     * 获取成员数量
     */
    public int getMemberCount() {
        return members.size();
    }

    /**
     * 获取在线成员数量
     */
    public int getOnlineMemberCount() {
        int count = 0;
        for (GroupMember member : members) {
            if (member.isOnline()) {
                count++;
            }
        }
        return count;
    }

    // === Getter/Setter ===

    public String getGroupId() { return groupId; }
    public void setGroupId(String groupId) { this.groupId = groupId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public long getUpdatedAt() { return updatedAt; }
    public void setUpdatedAt(long updatedAt) { this.updatedAt = updatedAt; }

    public boolean isActive() { return active; }
    public void setActive(boolean active) { this.active = active; }

    public int getPriority() { return priority; }
    public void setPriority(int priority) { this.priority = priority; }

    public GroupSettings getSettings() { return settings; }
    public void setSettings(GroupSettings settings) { this.settings = settings; }

    public String getAutoActivateScene() {
        return settings != null ? settings.getAutoActivateScene() : null;
    }

    public Map<String, String> getMetadata() { return metadata; }
    public void setMetadata(Map<String, String> metadata) { this.metadata = metadata; }

    @NonNull
    @Override
    public String toString() {
        return "DeviceGroup{" +
                "groupId='" + groupId + '\'' +
                ", name='" + name + '\'' +
                ", type='" + type + '\'' +
                ", memberCount=" + members.size() +
                ", active=" + active +
                '}';
    }
}