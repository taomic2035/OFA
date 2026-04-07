package com.ofa.agent.messaging;

import android.content.Context;
import android.content.SharedPreferences;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 消息总线客户端 (v3.0.0)
 *
 * Center 是永远在线的灵魂载体，消息总线提供设备间通信能力。
 *
 * 主要功能：
 * - 发送消息到其他设备
 * - 接收来自 Center 或其他设备的消息
 * - 离线消息管理
 * - 消息确认机制
 */
public class MessageBus {

    private static final String TAG = "MessageBus";
    private static final String PREFS_NAME = "ofa_message_bus";
    private static final String KEY_OFFLINE_QUEUE = "offline_queue";
    private static final String KEY_PENDING_ACKS = "pending_acks";

    // 配置
    private static final int MAX_OFFLINE_MESSAGES = 100;
    private static final long DEFAULT_TTL = 24 * 60 * 60 * 1000L; // 24小时
    private static final long ACK_TIMEOUT = 30 * 1000L; // 30秒
    private static final int MAX_RETRIES = 3;

    private final Context context;
    private final SharedPreferences prefs;
    private final ExecutorService executor;
    private final Handler mainHandler;

    // 当前设备 ID
    private String agentId;
    private String identityId;

    // Center 连接
    private String centerAddress;
    private boolean connected = false;

    // 离线消息队列
    private final List<Message> offlineQueue;

    // 待确认消息
    private final Map<String, Long> pendingAcks;

    // 消息监听器
    private final List<MessageListener> listeners;

    // 连接状态监听器
    private final List<ConnectionListener> connectionListeners;

    // 消息处理器
    private MessageHandler messageHandler;

    public MessageBus(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
        this.executor = Executors.newSingleThreadExecutor();
        this.mainHandler = new Handler(Looper.getMainLooper());
        this.offlineQueue = new CopyOnWriteArrayList<>();
        this.pendingAcks = new ConcurrentHashMap<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.connectionListeners = new CopyOnWriteArrayList<>();

        // 加载存储的离线消息
        loadOfflineQueue();
    }

    /**
     * 初始化消息总线
     */
    public void initialize(@NonNull String agentId, @NonNull String identityId, @Nullable String centerAddress) {
        this.agentId = agentId;
        this.identityId = identityId;
        this.centerAddress = centerAddress;

        Log.i(TAG, "MessageBus initialized: agent=" + agentId + ", identity=" + identityId);
    }

    /**
     * 连接到 Center
     */
    public void connect() {
        if (centerAddress == null || centerAddress.isEmpty()) {
            Log.w(TAG, "No center address configured");
            return;
        }

        executor.execute(() -> {
            try {
                // 建立 WebSocket/HTTP 连接
                // TODO: 实现实际的连接逻辑

                connected = true;

                // 上报在线状态
                notifyCenterOnline();

                // 获取离线消息
                fetchOfflineMessages();

                // 发送本地离线队列中的消息
                flushOfflineQueue();

                // 通知连接成功
                mainHandler.post(() -> {
                    for (ConnectionListener l : connectionListeners) {
                        l.onConnected();
                    }
                });

                Log.i(TAG, "Connected to center: " + centerAddress);

            } catch (Exception e) {
                Log.e(TAG, "Failed to connect to center", e);
                connected = false;

                mainHandler.post(() -> {
                    for (ConnectionListener l : connectionListeners) {
                        l.onConnectionFailed(e.getMessage());
                    }
                });
            }
        });
    }

    /**
     * 断开连接
     */
    public void disconnect() {
        executor.execute(() -> {
            connected = false;

            // 通知 Center 离线
            notifyCenterOffline();

            mainHandler.post(() -> {
                for (ConnectionListener l : connectionListeners) {
                    l.onDisconnected();
                }
            });

            Log.i(TAG, "Disconnected from center");
        });
    }

    /**
     * 发送消息
     */
    public void send(@NonNull Message message) {
        message.fromAgent = agentId;
        message.identityId = identityId;
        message.createdAt = System.currentTimeMillis();

        if (message.id == null || message.id.isEmpty()) {
            message.id = generateMessageId();
        }

        executor.execute(() -> {
            if (connected) {
                sendMessageToCenter(message);
            } else {
                // 离线，加入队列
                addToOfflineQueue(message);
            }
        });
    }

    /**
     * 发送消息到指定设备
     */
    public void send(@NonNull String toAgent, @NonNull String type, @NonNull Map<String, Object> payload) {
        Message message = new Message();
        message.toAgent = toAgent;
        message.type = type;
        message.payload = payload;
        send(message);
    }

    /**
     * 广播消息到身份的所有设备
     */
    public void broadcast(@NonNull String type, @NonNull Map<String, Object> payload) {
        Message message = new Message();
        message.type = type;
        message.payload = payload;
        message.toAgent = "*"; // 广播标识
        send(message);
    }

    /**
     * 确认消息
     */
    public void ack(@NonNull String messageId) {
        executor.execute(() -> {
            Message ack = Message.createAck(agentId, "center", messageId);
            sendMessageToCenter(ack);

            // 从待确认列表移除
            pendingAcks.remove(messageId);
        });
    }

    /**
     * 接收消息（由 Center 推送调用）
     */
    public void onMessageReceived(@NonNull Message message) {
        Log.d(TAG, "Received message: " + message.id + " from " + message.fromAgent);

        // 更新状态
        message.status = Message.STATUS_DELIVERED;
        message.deliveredAt = System.currentTimeMillis();

        // 自动发送确认
        if (!Message.TYPE_ACK.equals(message.type)) {
            ack(message.id);
        }

        // 分发给监听器
        mainHandler.post(() -> {
            for (MessageListener l : listeners) {
                try {
                    l.onMessage(message);
                } catch (Exception e) {
                    Log.e(TAG, "Error in message listener", e);
                }
            }
        });

        // 交给消息处理器
        if (messageHandler != null) {
            messageHandler.handleMessage(message);
        }
    }

    /**
     * 添加消息监听器
     */
    public void addListener(@NonNull MessageListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除消息监听器
     */
    public void removeListener(@NonNull MessageListener listener) {
        listeners.remove(listener);
    }

    /**
     * 设置消息处理器
     */
    public void setMessageHandler(@Nullable MessageHandler handler) {
        this.messageHandler = handler;
    }

    /**
     * 添加连接状态监听器
     */
    public void addConnectionListener(@NonNull ConnectionListener listener) {
        connectionListeners.add(listener);
    }

    /**
     * 移除连接状态监听器
     */
    public void removeConnectionListener(@NonNull ConnectionListener listener) {
        connectionListeners.remove(listener);
    }

    /**
     * 检查是否已连接
     */
    public boolean isConnected() {
        return connected;
    }

    /**
     * 获取离线消息数量
     */
    public int getOfflineQueueSize() {
        return offlineQueue.size();
    }

    /**
     * 获取待确认消息数量
     */
    public int getPendingAckCount() {
        return pendingAcks.size();
    }

    // === 私有方法 ===

    private void loadOfflineQueue() {
        String json = prefs.getString(KEY_OFFLINE_QUEUE, "[]");
        try {
            JSONArray array = new JSONArray(json);
            for (int i = 0; i < array.length(); i++) {
                offlineQueue.add(Message.fromJson(array.getJSONObject(i)));
            }
            Log.d(TAG, "Loaded " + offlineQueue.size() + " offline messages");
        } catch (JSONException e) {
            Log.e(TAG, "Failed to load offline queue", e);
        }
    }

    private void saveOfflineQueue() {
        try {
            JSONArray array = new JSONArray();
            for (Message msg : offlineQueue) {
                array.put(msg.toJson());
            }
            prefs.edit().putString(KEY_OFFLINE_QUEUE, array.toString()).apply();
        } catch (JSONException e) {
            Log.e(TAG, "Failed to save offline queue", e);
        }
    }

    private void addToOfflineQueue(Message message) {
        // 检查队列大小限制
        while (offlineQueue.size() >= MAX_OFFLINE_MESSAGES) {
            // 移除最旧的低优先级消息
            removeOldestLowPriorityMessage();
        }

        offlineQueue.add(message);
        saveOfflineQueue();

        Log.d(TAG, "Message added to offline queue: " + message.id);
    }

    private void removeOldestLowPriorityMessage() {
        int oldestIdx = -1;
        long oldestTime = Long.MAX_VALUE;

        for (int i = 0; i < offlineQueue.size(); i++) {
            Message msg = offlineQueue.get(i);
            if (msg.priority <= Message.PRIORITY_LOW && msg.createdAt < oldestTime) {
                oldestIdx = i;
                oldestTime = msg.createdAt;
            }
        }

        if (oldestIdx >= 0) {
            offlineQueue.remove(oldestIdx);
        }
    }

    private void flushOfflineQueue() {
        if (offlineQueue.isEmpty()) {
            return;
        }

        Log.i(TAG, "Flushing " + offlineQueue.size() + " offline messages");

        // 按优先级排序
        List<Message> sorted = new ArrayList<>(offlineQueue);
        Collections.sort(sorted, (a, b) -> Integer.compare(b.priority, a.priority));

        for (Message msg : sorted) {
            if (!msg.isExpired()) {
                sendMessageToCenter(msg);
            }
        }

        // 清空队列
        offlineQueue.clear();
        saveOfflineQueue();
    }

    private void sendMessageToCenter(Message message) {
        if (!connected) {
            addToOfflineQueue(message);
            return;
        }

        try {
            // TODO: 实现实际的 HTTP/WebSocket 发送
            // POST /api/v1/messages 或 WebSocket send

            message.status = Message.STATUS_SENT;

            // 添加到待确认列表
            pendingAcks.put(message.id, System.currentTimeMillis());

            // 模拟发送成功
            Log.d(TAG, "Message sent: " + message.id + " to " + message.toAgent);

        } catch (Exception e) {
            Log.e(TAG, "Failed to send message", e);
            message.status = Message.STATUS_FAILED;
            message.retryCount++;

            if (message.shouldRetry()) {
                addToOfflineQueue(message);
            }
        }
    }

    private void fetchOfflineMessages() {
        try {
            // TODO: 从 Center 获取离线消息
            // GET /api/v1/messages/offline

            Log.d(TAG, "Fetching offline messages from center");

        } catch (Exception e) {
            Log.e(TAG, "Failed to fetch offline messages", e);
        }
    }

    private void notifyCenterOnline() {
        try {
            // TODO: 通知 Center 设备上线
            // POST /api/v1/agents/{agentId}/online

            Log.d(TAG, "Notified center of online status");

        } catch (Exception e) {
            Log.e(TAG, "Failed to notify center online", e);
        }
    }

    private void notifyCenterOffline() {
        try {
            // TODO: 通知 Center 设备离线
            // POST /api/v1/agents/{agentId}/offline

            Log.d(TAG, "Notified center of offline status");

        } catch (Exception e) {
            Log.e(TAG, "Failed to notify center offline", e);
        }
    }

    private String generateMessageId() {
        return "msg_" + System.currentTimeMillis() + "_" + (int) (Math.random() * 10000);
    }

    // === 接口定义 ===

    /**
     * 消息监听器
     */
    public interface MessageListener {
        void onMessage(Message message);
    }

    /**
     * 连接状态监听器
     */
    public interface ConnectionListener {
        void onConnected();
        void onDisconnected();
        void onConnectionFailed(String error);
    }

    /**
     * 消息处理器
     */
    public interface MessageHandler {
        void handleMessage(Message message);
    }
}