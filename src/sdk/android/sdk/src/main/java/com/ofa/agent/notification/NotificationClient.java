package com.ofa.agent.notification;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;
import com.ofa.agent.state.StateSyncService;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 跨设备通知客户端 (v3.4.0)
 *
 * 负责：
 * - 接收来自 Center 的通知
 * - 本地通知展示和管理
 * - 通知状态回传
 */
public class NotificationClient {

    private static final String TAG = "NotificationClient";

    // 消息类型
    private static final String MSG_NOTIFICATION = "notification";
    private static final String MSG_NOTIFICATION_READ = "notification_read";
    private static final String MSG_NOTIFICATION_DISMISSED = "notification_dismissed";

    private final List<CrossDeviceNotification> activeNotifications;
    private final Map<String, CrossDeviceNotification> notificationCache;
    private final List<NotificationListener> listeners;

    private MessageBus messageBus;
    private StateSyncService stateSyncService;
    private String agentId;
    private String identityId;

    // 配置
    private NotificationClientConfig config;

    // 本地通知处理器
    private LocalNotificationHandler localHandler;

    /**
     * 通知监听器
     */
    public interface NotificationListener {
        void onNotificationReceived(@NonNull CrossDeviceNotification notification);
        void onNotificationUpdated(@NonNull CrossDeviceNotification notification);
        void onNotificationRemoved(@NonNull String notificationId);
    }

    /**
     * 本地通知处理器接口
     */
    public interface LocalNotificationHandler {
        void showNotification(@NonNull CrossDeviceNotification notification);
        void cancelNotification(@NonNull String notificationId);
        void updateNotification(@NonNull CrossDeviceNotification notification);
    }

    /**
     * 配置
     */
    public static class NotificationClientConfig {
        public int maxCachedNotifications = 100;
        public boolean autoShowLocal = true;
        public boolean respectQuietHours = true;
        public int quietHourStart = 22; // 22:00
        public int quietHourEnd = 7;    // 07:00
    }

    public NotificationClient() {
        this.activeNotifications = new ArrayList<>();
        this.notificationCache = new HashMap<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.config = new NotificationClientConfig();
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

        Log.i(TAG, "NotificationClient initialized for " + agentId);
    }

    /**
     * 设置本地通知处理器
     */
    public void setLocalHandler(@Nullable LocalNotificationHandler handler) {
        this.localHandler = handler;
    }

    /**
     * 设置配置
     */
    public void setConfig(@NonNull NotificationClientConfig config) {
        this.config = config;
    }

    // === 通知接收 ===

    private void handleMessage(@NonNull Message message) {
        if (message.payload == null) {
            return;
        }

        Object typeObj = message.payload.get("type");
        if (!MSG_NOTIFICATION.equals(typeObj)) {
            return;
        }

        try {
            Object notifObj = message.payload.get("notification");
            CrossDeviceNotification notification = null;

            if (notifObj instanceof JSONObject) {
                notification = CrossDeviceNotification.fromJson((JSONObject) notifObj);
            } else if (notifObj instanceof Map) {
                JSONObject json = new JSONObject((Map) notifObj);
                notification = CrossDeviceNotification.fromJson(json);
            }

            if (notification != null) {
                receiveNotification(notification);
            }

        } catch (Exception e) {
            Log.e(TAG, "Failed to parse notification", e);
        }
    }

    /**
     * 接收通知
     */
    public void receiveNotification(@NonNull CrossDeviceNotification notification) {
        // 检查是否已过期
        if (notification.isExpired()) {
            Log.d(TAG, "Notification expired, ignoring: " + notification.getNotificationId());
            return;
        }

        // 勿扰检查
        if (config.respectQuietHours && isQuietHours() &&
                notification.getPriority() < CrossDeviceNotification.PRIORITY_HIGH) {
            Log.d(TAG, "Quiet hours, ignoring low priority notification");
            return;
        }

        // 缓存通知
        notificationCache.put(notification.getNotificationId(), notification);
        notification.markDelivered(agentId);

        // 添加到活动列表
        activeNotifications.add(notification);

        // 通知监听器
        notifyReceived(notification);

        // 本地展示
        if (config.autoShowLocal && localHandler != null) {
            localHandler.showNotification(notification);
        }

        // 发送送达确认
        reportDelivered(notification.getNotificationId());

        Log.d(TAG, "Notification received: " + notification.getNotificationId());
    }

    // === 通知操作 ===

    /**
     * 标记通知已读
     */
    public void markAsRead(@NonNull String notificationId) {
        CrossDeviceNotification notification = notificationCache.get(notificationId);
        if (notification == null) {
            return;
        }

        notification.markRead(agentId);

        // 通知监听器
        notifyUpdated(notification);

        // 发送已读状态
        reportRead(notificationId);

        // 从活动列表移除
        activeNotifications.removeIf(n -> n.getNotificationId().equals(notificationId));

        // 取消本地通知
        if (localHandler != null) {
            localHandler.cancelNotification(notificationId);
        }
    }

    /**
     * 标记通知忽略
     */
    public void markAsDismissed(@NonNull String notificationId) {
        CrossDeviceNotification notification = notificationCache.get(notificationId);
        if (notification == null) {
            return;
        }

        notification.markDismissed(agentId);

        // 通知监听器
        notifyUpdated(notification);

        // 发送忽略状态
        reportDismissed(notificationId);

        // 从活动列表移除
        activeNotifications.removeIf(n -> n.getNotificationId().equals(notificationId));

        // 取消本地通知
        if (localHandler != null) {
            localHandler.cancelNotification(notificationId);
        }
    }

    /**
     * 标记所有通知已读
     */
    public void markAllAsRead() {
        for (CrossDeviceNotification notification : new ArrayList<>(activeNotifications)) {
            markAsRead(notification.getNotificationId());
        }
    }

    /**
     * 执行通知动作
     */
    public void executeAction(@NonNull String notificationId, @NonNull String actionId) {
        CrossDeviceNotification notification = notificationCache.get(notificationId);
        if (notification == null) {
            return;
        }

        NotificationAction action = findAction(notification, actionId);
        if (action == null) {
            return;
        }

        // 处理动作
        switch (action.getType()) {
            case NotificationAction.TYPE_OPEN:
                // 打开相关内容
                markAsRead(notificationId);
                break;

            case NotificationAction.TYPE_REPLY:
                // 回复动作
                markAsRead(notificationId);
                break;

            case NotificationAction.TYPE_DISMISS:
                // 忽略
                markAsDismissed(notificationId);
                break;

            case NotificationAction.TYPE_CUSTOM:
                // 自定义动作，通知监听器处理
                break;
        }

        Log.d(TAG, "Action executed: " + actionId + " for " + notificationId);
    }

    private NotificationAction findAction(CrossDeviceNotification notification, String actionId) {
        if (notification.getActions() == null) {
            return null;
        }
        for (NotificationAction action : notification.getActions()) {
            if (action.getActionId().equals(actionId)) {
                return action;
            }
        }
        return null;
    }

    // === 状态上报 ===

    private void reportDelivered(@NonNull String notificationId) {
        if (messageBus == null) {
            return;
        }

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_NOTIFICATION);
        payload.put("notification_id", notificationId);
        payload.put("status", "delivered");
        payload.put("delivered_at", System.currentTimeMillis());
        msg.payload = payload;

        messageBus.send(msg);
    }

    private void reportRead(@NonNull String notificationId) {
        if (messageBus == null) {
            return;
        }

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_NOTIFICATION_READ);
        payload.put("notification_id", notificationId);
        payload.put("read_at", System.currentTimeMillis());
        msg.payload = payload;

        messageBus.send(msg);
    }

    private void reportDismissed(@NonNull String notificationId) {
        if (messageBus == null) {
            return;
        }

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_NORMAL;

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_NOTIFICATION_DISMISSED);
        payload.put("notification_id", notificationId);
        msg.payload = payload;

        messageBus.send(msg);
    }

    // === 通知查询 ===

    /**
     * 获取活动通知
     */
    @NonNull
    public List<CrossDeviceNotification> getActiveNotifications() {
        return new ArrayList<>(activeNotifications);
    }

    /**
     * 获取未读通知数
     */
    public int getUnreadCount() {
        int count = 0;
        for (CrossDeviceNotification notification : activeNotifications) {
            if (notification.requiresAttention()) {
                count++;
            }
        }
        return count;
    }

    /**
     * 获取指定通知
     */
    @Nullable
    public CrossDeviceNotification getNotification(@NonNull String notificationId) {
        return notificationCache.get(notificationId);
    }

    /**
     * 按分组获取通知
     */
    @NonNull
    public List<CrossDeviceNotification> getNotificationsByGroup(@NonNull String groupId) {
        List<CrossDeviceNotification> result = new ArrayList<>();
        for (CrossDeviceNotification notification : activeNotifications) {
            if (groupId.equals(notification.getGroupId())) {
                result.add(notification);
            }
        }
        return result;
    }

    /**
     * 按类型获取通知
     */
    @NonNull
    public List<CrossDeviceNotification> getNotificationsByType(@NonNull String type) {
        List<CrossDeviceNotification> result = new ArrayList<>();
        for (CrossDeviceNotification notification : activeNotifications) {
            if (type.equals(notification.getType())) {
                result.add(notification);
            }
        }
        return result;
    }

    // === 监听器管理 ===

    public void addListener(@NonNull NotificationListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull NotificationListener listener) {
        listeners.remove(listener);
    }

    private void notifyReceived(@NonNull CrossDeviceNotification notification) {
        for (NotificationListener l : listeners) {
            l.onNotificationReceived(notification);
        }
    }

    private void notifyUpdated(@NonNull CrossDeviceNotification notification) {
        for (NotificationListener l : listeners) {
            l.onNotificationUpdated(notification);
        }
    }

    private void notifyRemoved(@NonNull String notificationId) {
        for (NotificationListener l : listeners) {
            l.onNotificationRemoved(notificationId);
        }
    }

    // === 辅助方法 ===

    private boolean isQuietHours() {
        if (!config.respectQuietHours) {
            return false;
        }

        java.util.Calendar cal = java.util.Calendar.getInstance();
        int hour = cal.get(java.util.Calendar.HOUR_OF_DAY);

        if (config.quietHourStart > config.quietHourEnd) {
            // 跨午夜: 22:00 - 07:00
            return hour >= config.quietHourStart || hour < config.quietHourEnd;
        }

        return hour >= config.quietHourStart && hour < config.quietHourEnd;
    }

    private String generateMessageId() {
        return "notif_msg_" + System.currentTimeMillis() + "_" + agentId;
    }

    /**
     * 获取统计信息
     */
    @NonNull
    public NotificationClientStats getStats() {
        NotificationClientStats stats = new NotificationClientStats();
        stats.activeCount = activeNotifications.size();
        stats.cachedCount = notificationCache.size();
        stats.unreadCount = getUnreadCount();

        for (CrossDeviceNotification notification : activeNotifications) {
            String type = notification.getType();
            stats.byType.put(type, stats.byType.getOrDefault(type, 0) + 1);
        }

        return stats;
    }

    /**
     * 统计信息
     */
    public static class NotificationClientStats {
        public int activeCount;
        public int cachedCount;
        public int unreadCount;
        public Map<String, Integer> byType = new HashMap<>();

        @NonNull
        @Override
        public String toString() {
            return "NotificationClientStats{" +
                    "active=" + activeCount +
                    ", unread=" + unreadCount +
                    '}';
        }
    }

    /**
     * 清理过期通知
     */
    public void cleanupExpired() {
        List<String> expired = new ArrayList<>();
        for (CrossDeviceNotification notification : activeNotifications) {
            if (notification.isExpired()) {
                expired.add(notification.getNotificationId());
            }
        }

        for (String id : expired) {
            activeNotifications.removeIf(n -> n.getNotificationId().equals(id));
            notificationCache.remove(id);
            notifyRemoved(id);
        }

        // 限制缓存大小
        while (notificationCache.size() > config.maxCachedNotifications) {
            // 移除最旧的通知
            String oldestId = null;
            long oldestTime = Long.MAX_VALUE;
            for (CrossDeviceNotification n : notificationCache.values()) {
                if (n.getCreatedAt() < oldestTime) {
                    oldestTime = n.getCreatedAt();
                    oldestId = n.getNotificationId();
                }
            }
            if (oldestId != null) {
                notificationCache.remove(oldestId);
            }
        }
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        activeNotifications.clear();
        notificationCache.clear();
        listeners.clear();

        Log.i(TAG, "NotificationClient cleaned up");
    }
}