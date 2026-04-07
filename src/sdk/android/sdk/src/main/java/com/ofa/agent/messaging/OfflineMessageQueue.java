package com.ofa.agent.messaging;

import android.content.Context;
import android.content.SharedPreferences;
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

/**
 * 离线消息队列 (v3.0.0)
 *
 * 管理离线时待发送的消息队列。
 * 支持消息持久化、优先级排序、过期清理。
 */
public class OfflineMessageQueue {

    private static final String TAG = "OfflineMessageQueue";
    private static final String PREFS_NAME = "ofa_offline_queue";
    private static final String KEY_QUEUE = "message_queue";

    private static final int MAX_QUEUE_SIZE = 100;
    private static final long DEFAULT_TTL = 24 * 60 * 60 * 1000L; // 24小时

    private final Context context;
    private final SharedPreferences prefs;
    private final List<Message> queue;

    public OfflineMessageQueue(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
        this.queue = new ArrayList<>();

        loadQueue();
    }

    /**
     * 添加消息到队列
     */
    public void enqueue(@NonNull Message message) {
        // 设置默认过期时间
        if (message.expiresAt == null) {
            message.setTTL(DEFAULT_TTL);
        }

        // 检查是否过期
        if (message.isExpired()) {
            Log.w(TAG, "Message already expired, not adding to queue");
            return;
        }

        // 检查队列大小
        if (queue.size() >= MAX_QUEUE_SIZE) {
            // 移除最旧的低优先级消息
            removeOldestLowPriority();
        }

        queue.add(message);

        // 按优先级排序
        Collections.sort(queue, (a, b) -> Integer.compare(b.priority, a.priority));

        saveQueue();

        Log.d(TAG, "Message enqueued: " + message.id + ", queue size=" + queue.size());
    }

    /**
     * 批量添加消息
     */
    public void enqueueAll(@NonNull List<Message> messages) {
        for (Message msg : messages) {
            if (queue.size() >= MAX_QUEUE_SIZE) {
                break;
            }
            if (!msg.isExpired()) {
                queue.add(msg);
            }
        }

        // 排序
        Collections.sort(queue, (a, b) -> Integer.compare(b.priority, a.priority));

        saveQueue();
    }

    /**
     * 取出下一条消息
     */
    @Nullable
    public Message dequeue() {
        if (queue.isEmpty()) {
            return null;
        }

        Message msg = queue.remove(0);
        saveQueue();

        Log.d(TAG, "Message dequeued: " + msg.id + ", remaining=" + queue.size());

        return msg;
    }

    /**
     * 查看队首消息（不移除）
     */
    @Nullable
    public Message peek() {
        if (queue.isEmpty()) {
            return null;
        }
        return queue.get(0);
    }

    /**
     * 获取所有有效消息
     */
    @NonNull
    public List<Message> drainAll() {
        // 过滤过期消息
        List<Message> valid = new ArrayList<>();
        List<Message> expired = new ArrayList<>();

        long now = System.currentTimeMillis();
        for (Message msg : queue) {
            if (msg.isExpired()) {
                expired.add(msg);
            } else {
                valid.add(msg);
            }
        }

        // 移除过期消息
        if (!expired.isEmpty()) {
            queue.removeAll(expired);
            Log.d(TAG, "Removed " + expired.size() + " expired messages");
        }

        // 清空队列
        queue.clear();
        saveQueue();

        return valid;
    }

    /**
     * 获取队列大小
     */
    public int size() {
        return queue.size();
    }

    /**
     * 检查队列是否为空
     */
    public boolean isEmpty() {
        return queue.isEmpty();
    }

    /**
     * 清空队列
     */
    public void clear() {
        queue.clear();
        saveQueue();
        Log.d(TAG, "Queue cleared");
    }

    /**
     * 按优先级获取消息
     */
    @NonNull
    public List<Message> getByPriority(int minPriority) {
        List<Message> result = new ArrayList<>();
        for (Message msg : queue) {
            if (msg.priority >= minPriority && !msg.isExpired()) {
                result.add(msg);
            }
        }
        return result;
    }

    /**
     * 按类型获取消息
     */
    @NonNull
    public List<Message> getByType(@NonNull String type) {
        List<Message> result = new ArrayList<>();
        for (Message msg : queue) {
            if (type.equals(msg.type) && !msg.isExpired()) {
                result.add(msg);
            }
        }
        return result;
    }

    /**
     * 移除指定消息
     */
    public boolean remove(@NonNull String messageId) {
        for (int i = 0; i < queue.size(); i++) {
            if (messageId.equals(queue.get(i).id)) {
                queue.remove(i);
                saveQueue();
                return true;
            }
        }
        return false;
    }

    /**
     * 清理过期消息
     */
    public int cleanupExpired() {
        int removed = 0;
        List<Message> expired = new ArrayList<>();

        for (Message msg : queue) {
            if (msg.isExpired()) {
                expired.add(msg);
                removed++;
            }
        }

        if (!expired.isEmpty()) {
            queue.removeAll(expired);
            saveQueue();
        }

        Log.d(TAG, "Cleaned up " + removed + " expired messages");
        return removed;
    }

    /**
     * 获取队列统计信息
     */
    @NonNull
    public QueueStats getStats() {
        QueueStats stats = new QueueStats();
        stats.totalCount = queue.size();

        for (Message msg : queue) {
            if (msg.isExpired()) {
                stats.expiredCount++;
            } else {
                switch (msg.priority) {
                    case Message.PRIORITY_URGENT:
                        stats.urgentCount++;
                        break;
                    case Message.PRIORITY_HIGH:
                        stats.highCount++;
                        break;
                    case Message.PRIORITY_NORMAL:
                        stats.normalCount++;
                        break;
                    case Message.PRIORITY_LOW:
                        stats.lowCount++;
                        break;
                }
            }
        }

        return stats;
    }

    // === 私有方法 ===

    private void loadQueue() {
        String json = prefs.getString(KEY_QUEUE, "[]");
        try {
            JSONArray array = new JSONArray(json);
            for (int i = 0; i < array.length() && i < MAX_QUEUE_SIZE; i++) {
                queue.add(Message.fromJson(array.getJSONObject(i)));
            }
            Log.d(TAG, "Loaded " + queue.size() + " messages from storage");
        } catch (JSONException e) {
            Log.e(TAG, "Failed to load queue", e);
        }
    }

    private void saveQueue() {
        try {
            JSONArray array = new JSONArray();
            for (Message msg : queue) {
                array.put(msg.toJson());
            }
            prefs.edit().putString(KEY_QUEUE, array.toString()).apply();
        } catch (JSONException e) {
            Log.e(TAG, "Failed to save queue", e);
        }
    }

    private void removeOldestLowPriority() {
        int oldestIdx = -1;
        long oldestTime = Long.MAX_VALUE;

        for (int i = 0; i < queue.size(); i++) {
            Message msg = queue.get(i);
            if (msg.priority <= Message.PRIORITY_LOW && msg.createdAt < oldestTime) {
                oldestIdx = i;
                oldestTime = msg.createdAt;
            }
        }

        if (oldestIdx >= 0) {
            queue.remove(oldestIdx);
            Log.d(TAG, "Removed oldest low priority message");
        }
    }

    /**
     * 队列统计信息
     */
    public static class QueueStats {
        public int totalCount;
        public int urgentCount;
        public int highCount;
        public int normalCount;
        public int lowCount;
        public int expiredCount;

        @NonNull
        @Override
        public String toString() {
            return String.format("QueueStats{total=%d, urgent=%d, high=%d, normal=%d, low=%d, expired=%d}",
                    totalCount, urgentCount, highCount, normalCount, lowCount, expiredCount);
        }
    }
}