package com.ofa.agent.group;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;
import com.ofa.agent.state.StateSyncService;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 设备群组客户端 (v3.5.0)
 *
 * 负责：
 * - 群组创建和管理
 * - 群组成员操作
 * - 群组内消息广播
 * - 群组优先级管理
 */
public class DeviceGroupClient {

    private static final String TAG = "DeviceGroupClient";

    // 消息类型
    private static final String MSG_GROUP_CREATE = "group_create";
    private static final String MSG_GROUP_UPDATE = "group_update";
    private static final String MSG_GROUP_DELETE = "group_delete";
    private static final String MSG_GROUP_MEMBER_ADD = "group_member_add";
    private static final String MSG_GROUP_MEMBER_REMOVE = "group_member_remove";
    private static final String MSG_GROUP_BROADCAST = "group_broadcast";

    // 本地缓存
    private final Map<String, DeviceGroup> groupCache;
    private final Map<String, List<String>> agentGroupsCache;
    private final List<GroupListener> listeners;

    private MessageBus messageBus;
    private StateSyncService stateSyncService;
    private String agentId;
    private String identityId;

    // 配置
    private GroupClientConfig config;

    /**
     * 群组监听器
     */
    public interface GroupListener {
        void onGroupCreated(@NonNull DeviceGroup group);
        void onGroupUpdated(@NonNull DeviceGroup group);
        void onGroupDeleted(@NonNull String groupId);
        void onMemberAdded(@NonNull String groupId, @NonNull GroupMember member);
        void onMemberRemoved(@NonNull String groupId, @NonNull String agentId);
        void onGroupBroadcast(@NonNull String groupId, @NonNull Message message);
    }

    /**
     * 配置
     */
    public static class GroupClientConfig {
        public int maxCachedGroups = 50;
        public boolean autoSync = true;
        public long syncInterval = 30000; // 30秒
    }

    public DeviceGroupClient() {
        this.groupCache = new HashMap<>();
        this.agentGroupsCache = new HashMap<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.config = new GroupClientConfig();
    }

    /**
     * 初始化
     */
    public void initialize(@NonNull String agentId, @NonNull String identityId,
                           @Nullable MessageBus messageBus,
                           @Nullable StateSyncService stateSyncService) {
        this.agentId = agentId;
        this.identityId = identityId;
        this.messageBus = messageBus;
        this.stateSyncService = stateSyncService;

        if (messageBus != null) {
            messageBus.addListener(this::handleMessage);
        }

        Log.i(TAG, "DeviceGroupClient initialized for " + agentId);
    }

    // === 群组操作 ===

    /**
     * 创建群组
     */
    public void createGroup(@NonNull String name, @NonNull String type,
                            @Nullable CreateGroupCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_GROUP_CREATE);
        payload.put("identity_id", identityId);
        payload.put("name", name);
        payload.put("type", type);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        Log.d(TAG, "Create group request sent: " + name);

        if (callback != null) {
            callback.onSuccess(null); // 异步创建，结果通过监听器获取
        }
    }

    /**
     * 获取群组
     */
    @Nullable
    public DeviceGroup getGroup(@NonNull String groupId) {
        return groupCache.get(groupId);
    }

    /**
     * 获取所有群组
     */
    @NonNull
    public List<DeviceGroup> getAllGroups() {
        return new ArrayList<>(groupCache.values());
    }

    /**
     * 获取设备所属群组
     */
    @NonNull
    public List<DeviceGroup> getAgentGroups(@NonNull String targetAgentId) {
        List<String> groupIds = agentGroupsCache.get(targetAgentId);
        if (groupIds == null || groupIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<DeviceGroup> groups = new ArrayList<>();
        for (String groupId : groupIds) {
            DeviceGroup group = groupCache.get(groupId);
            if (group != null && group.isActive()) {
                groups.add(group);
            }
        }
        return groups;
    }

    /**
     * 更新群组
     */
    public void updateGroup(@NonNull String groupId, @NonNull Map<String, Object> updates,
                            @Nullable UpdateGroupCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_GROUP_UPDATE);
        payload.put("group_id", groupId);
        payload.put("updates", updates);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        Log.d(TAG, "Update group request sent: " + groupId);

        if (callback != null) {
            callback.onSuccess();
        }
    }

    /**
     * 删除群组
     */
    public void deleteGroup(@NonNull String groupId, @Nullable DeleteGroupCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_GROUP_DELETE);
        payload.put("group_id", groupId);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        // 本地移除
        groupCache.remove(groupId);

        Log.d(TAG, "Delete group request sent: " + groupId);

        if (callback != null) {
            callback.onSuccess();
        }
    }

    // === 成员操作 ===

    /**
     * 添加成员
     */
    public void addMember(@NonNull String groupId, @NonNull String targetAgentId,
                          @NonNull String deviceName, @NonNull String deviceType,
                          @NonNull String role, @Nullable AddMemberCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_GROUP_MEMBER_ADD);
        payload.put("group_id", groupId);
        payload.put("agent_id", targetAgentId);
        payload.put("device_name", deviceName);
        payload.put("device_type", deviceType);
        payload.put("role", role);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        Log.d(TAG, "Add member request sent: " + targetAgentId + " to " + groupId);

        if (callback != null) {
            callback.onSuccess();
        }
    }

    /**
     * 移除成员
     */
    public void removeMember(@NonNull String groupId, @NonNull String targetAgentId,
                             @Nullable RemoveMemberCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_GROUP_MEMBER_REMOVE);
        payload.put("group_id", groupId);
        payload.put("agent_id", targetAgentId);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        // 本地更新
        DeviceGroup group = groupCache.get(groupId);
        if (group != null) {
            group.removeMember(targetAgentId);
        }

        Log.d(TAG, "Remove member request sent: " + targetAgentId + " from " + groupId);

        if (callback != null) {
            callback.onSuccess();
        }
    }

    /**
     * 获取群组成员
     */
    @NonNull
    public List<GroupMember> getMembers(@NonNull String groupId) {
        DeviceGroup group = groupCache.get(groupId);
        if (group == null) {
            return new ArrayList<>();
        }
        return group.getMembers();
    }

    /**
     * 获取在线成员
     */
    @NonNull
    public List<GroupMember> getOnlineMembers(@NonNull String groupId) {
        DeviceGroup group = groupCache.get(groupId);
        if (group == null) {
            return new ArrayList<>();
        }
        return group.getOnlineMembers();
    }

    // === 群组广播 ===

    /**
     * 广播消息到群组
     */
    public void broadcastToGroup(@NonNull String groupId, @NonNull Map<String, Object> content,
                                  @Nullable BroadcastCallback callback) {
        if (messageBus == null) {
            if (callback != null) {
                callback.onError("MessageBus not initialized");
            }
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_GROUP_BROADCAST);
        payload.put("group_id", groupId);
        payload.put("content", content);

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;
        msg.payload = payload;

        messageBus.send(msg);

        Log.d(TAG, "Broadcast request sent to group: " + groupId);

        if (callback != null) {
            callback.onSuccess();
        }
    }

    /**
     * 广播到所有群组
     */
    public void broadcastToAllGroups(@NonNull Map<String, Object> content,
                                      @Nullable BroadcastCallback callback) {
        for (DeviceGroup group : groupCache.values()) {
            if (group.isActive()) {
                broadcastToGroup(group.getGroupId(), content, null);
            }
        }

        if (callback != null) {
            callback.onSuccess();
        }
    }

    // === 群组优先级 ===

    /**
     * 获取最高优先级群组
     */
    @Nullable
    public DeviceGroup getHighestPriorityGroup() {
        DeviceGroup highest = null;
        int maxPriority = -1;

        for (DeviceGroup group : groupCache.values()) {
            if (group.isActive() && group.getPriority() > maxPriority) {
                maxPriority = group.getPriority();
                highest = group;
            }
        }

        return highest;
    }

    /**
     * 根据场景获取群组
     */
    @Nullable
    public DeviceGroup getGroupByScene(@NonNull String scene) {
        // 先查找自动激活场景匹配
        for (DeviceGroup group : groupCache.values()) {
            if (group.isActive() && scene.equals(group.getAutoActivateScene())) {
                return group;
            }
        }

        // 根据场景类型匹配群组类型
        String targetType = mapSceneToGroupType(scene);
        if (targetType != null) {
            for (DeviceGroup group : groupCache.values()) {
                if (group.isActive() && targetType.equals(group.getType())) {
                    return group;
                }
            }
        }

        return null;
    }

    private String mapSceneToGroupType(String scene) {
        switch (scene) {
            case "home":
            case "relaxing":
                return DeviceGroup.TYPE_HOME;
            case "work":
            case "meeting":
                return DeviceGroup.TYPE_WORK;
            case "running":
            case "fitness":
                return DeviceGroup.TYPE_FITNESS;
            case "driving":
            case "traveling":
                return DeviceGroup.TYPE_TRAVEL;
            default:
                return null;
        }
    }

    // === 消息处理 ===

    private void handleMessage(@NonNull Message message) {
        if (message.payload == null) {
            return;
        }

        Object typeObj = message.payload.get("type");
        if (!"group_event".equals(typeObj)) {
            return;
        }

        try {
            Object groupObj = message.payload.get("group");
            if (groupObj instanceof JSONObject) {
                DeviceGroup group = DeviceGroup.fromJson((JSONObject) groupObj);
                handleGroupEvent(message.payload, group);
            } else if (groupObj instanceof Map) {
                JSONObject json = new JSONObject((Map) groupObj);
                DeviceGroup group = DeviceGroup.fromJson(json);
                handleGroupEvent(message.payload, group);
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to handle group event", e);
        }
    }

    private void handleGroupEvent(Map<String, Object> payload, DeviceGroup group) {
        String event = (String) payload.get("event");
        if (event == null) {
            return;
        }

        switch (event) {
            case "created":
                updateLocalCache(group);
                notifyGroupCreated(group);
                break;

            case "updated":
                updateLocalCache(group);
                notifyGroupUpdated(group);
                break;

            case "deleted":
                String groupId = (String) payload.get("group_id");
                if (groupId != null) {
                    groupCache.remove(groupId);
                    notifyGroupDeleted(groupId);
                }
                break;

            case "member_added":
                updateLocalCache(group);
                Object memberObj = payload.get("member");
                if (memberObj instanceof JSONObject) {
                    GroupMember member = GroupMember.fromJson((JSONObject) memberObj);
                    notifyMemberAdded(group.getGroupId(), member);
                } else if (memberObj instanceof Map) {
                    JSONObject json = new JSONObject((Map) memberObj);
                    GroupMember member = GroupMember.fromJson(json);
                    notifyMemberAdded(group.getGroupId(), member);
                }
                break;

            case "member_removed":
                updateLocalCache(group);
                String agentId = (String) payload.get("agent_id");
                if (agentId != null) {
                    notifyMemberRemoved(group.getGroupId(), agentId);
                }
                break;

            case "broadcast":
                Object contentObj = payload.get("content");
                if (contentObj instanceof Map) {
                    @SuppressWarnings("unchecked")
                    Map<String, Object> content = (Map<String, Object>) contentObj;
                    Message broadcastMsg = new Message();
                    broadcastMsg.payload = content;
                    notifyGroupBroadcast(group.getGroupId(), broadcastMsg);
                }
                break;
        }
    }

    private void updateLocalCache(DeviceGroup group) {
        groupCache.put(group.getGroupId(), group);

        // 更新设备群组映射
        for (GroupMember member : group.getMembers()) {
            List<String> groups = agentGroupsCache.get(member.getAgentId());
            if (groups == null) {
                groups = new ArrayList<>();
                agentGroupsCache.put(member.getAgentId(), groups);
            }
            if (!groups.contains(group.getGroupId())) {
                groups.add(group.getGroupId());
            }
        }
    }

    // === 缓存管理 ===

    /**
     * 更新群组缓存
     */
    public void updateGroupCache(@NonNull DeviceGroup group) {
        updateLocalCache(group);
    }

    /**
     * 清理缓存
     */
    public void clearCache() {
        groupCache.clear();
        agentGroupsCache.clear();
    }

    // === 监听器管理 ===

    public void addListener(@NonNull GroupListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull GroupListener listener) {
        listeners.remove(listener);
    }

    private void notifyGroupCreated(DeviceGroup group) {
        for (GroupListener l : listeners) {
            l.onGroupCreated(group);
        }
    }

    private void notifyGroupUpdated(DeviceGroup group) {
        for (GroupListener l : listeners) {
            l.onGroupUpdated(group);
        }
    }

    private void notifyGroupDeleted(String groupId) {
        for (GroupListener l : listeners) {
            l.onGroupDeleted(groupId);
        }
    }

    private void notifyMemberAdded(String groupId, GroupMember member) {
        for (GroupListener l : listeners) {
            l.onMemberAdded(groupId, member);
        }
    }

    private void notifyMemberRemoved(String groupId, String agentId) {
        for (GroupListener l : listeners) {
            l.onMemberRemoved(groupId, agentId);
        }
    }

    private void notifyGroupBroadcast(String groupId, Message message) {
        for (GroupListener l : listeners) {
            l.onGroupBroadcast(groupId, message);
        }
    }

    // === 统计信息 ===

    /**
     * 获取统计信息
     */
    @NonNull
    public GroupClientStats getStats() {
        GroupClientStats stats = new GroupClientStats();
        stats.totalGroups = groupCache.size();

        for (DeviceGroup group : groupCache.values()) {
            if (group.isActive()) {
                stats.activeGroups++;
            }
            stats.totalMembers += group.getMemberCount();

            String type = group.getType();
            stats.byType.put(type, stats.byType.getOrDefault(type, 0) + 1);
        }

        return stats;
    }

    // === 辅助方法 ===

    private String generateMessageId() {
        return "group_msg_" + System.currentTimeMillis() + "_" + agentId;
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        groupCache.clear();
        agentGroupsCache.clear();
        listeners.clear();

        Log.i(TAG, "DeviceGroupClient cleaned up");
    }

    // === 回调接口 ===

    public interface CreateGroupCallback {
        void onSuccess(@Nullable DeviceGroup group);
        void onError(@NonNull String error);
    }

    public interface UpdateGroupCallback {
        void onSuccess();
        void onError(@NonNull String error);
    }

    public interface DeleteGroupCallback {
        void onSuccess();
        void onError(@NonNull String error);
    }

    public interface AddMemberCallback {
        void onSuccess();
        void onError(@NonNull String error);
    }

    public interface RemoveMemberCallback {
        void onSuccess();
        void onError(@NonNull String error);
    }

    public interface BroadcastCallback {
        void onSuccess();
        void onError(@NonNull String error);
    }

    /**
     * 统计信息
     */
    public static class GroupClientStats {
        public int totalGroups;
        public int activeGroups;
        public int totalMembers;
        public Map<String, Integer> byType = new HashMap<>();

        @NonNull
        @Override
        public String toString() {
            return "GroupClientStats{" +
                    "total=" + totalGroups +
                    ", active=" + activeGroups +
                    ", members=" + totalMembers +
                    '}';
        }
    }
}