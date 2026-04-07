package com.ofa.agent.security;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;

import org.json.JSONException;
import org.json.JSONObject;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

import javax.crypto.Cipher;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;

/**
 * 安全客户端 (v3.7.0)
 *
 * 负责：
 * - 端到端加密/解密
 * - 密钥管理
 * - 安全会话管理
 * - 安全通道建立
 */
public class SecurityClient {

    private static final String TAG = "SecurityClient";

    // 加密算法
    public static final String ALGO_AES_GCM = "AES-256-GCM";
    public static final String ALGO_AES_CBC = "AES-256-CBC";

    // 密钥类型
    public static final String KEY_TYPE_MASTER = "master";
    public static final String KEY_TYPE_SESSION = "session";
    public static final String KEY_TYPE_DEVICE = "device";
    public static final String KEY_TYPE_IDENTITY = "identity";

    // 安全级别
    public static final String SECURITY_NONE = "none";
    public static final String SECURITY_BASIC = "basic";
    public static final String SECURITY_E2E = "e2e";
    public static final String SECURITY_E2E_AUTH = "e2e_auth";

    // 消息类型
    private static final String MSG_KEY_REQUEST = "key_request";
    private static final String MSG_KEY_RESPONSE = "key_response";
    private static final String MSG_SESSION_CREATE = "session_create";
    private static final String MSG_CHANNEL_ESTABLISH = "channel_establish";
    private static final String MSG_ENCRYPTED_DATA = "encrypted_data";

    // 本地缓存
    private final Map<String, SecurityKey> keyCache;
    private final Map<String, SecuritySession> sessionCache;
    private final Map<String, SecureChannel> channelCache;
    private final List<SecurityListener> listeners;

    private MessageBus messageBus;
    private String agentId;
    private String identityId;

    // 配置
    private SecurityConfig config;

    // 主密钥
    private SecurityKey masterKey;

    /**
     * 安全监听器
     */
    public interface SecurityListener {
        void onKeyCreated(@NonNull SecurityKey key);
        void onKeyRotated(@NonNull String oldKeyId, @NonNull String newKeyId);
        void onSessionCreated(@NonNull SecuritySession session);
        void onSessionExpired(@NonNull String sessionId);
        void onChannelEstablished(@NonNull SecureChannel channel);
        void onChannelClosed(@NonNull String channelId);
        void onEncryptSuccess(@NonNull String dataId, @NonNull String keyId);
        void onDecryptSuccess(@NonNull String dataId, @NonNull String keyId);
        void onSecurityError(@NonNull String errorType, @NonNull String details);
    }

    /**
     * 配置
     */
    public static class SecurityConfig {
        public String defaultAlgorithm = ALGO_AES_GCM;
        public long sessionKeyTtl = 24 * 60 * 60 * 1000; // 24小时
        public long channelKeyTtl = 60 * 60 * 1000; // 1小时
        public String defaultSecurityLevel = SECURITY_E2E;
        public boolean enableKeyRotation = true;
        public int maxSessionsPerAgent = 5;
    }

    public SecurityClient() {
        this.keyCache = new ConcurrentHashMap<>();
        this.sessionCache = new ConcurrentHashMap<>();
        this.channelCache = new ConcurrentHashMap<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.config = new SecurityConfig();

        // 生成本地主密钥
        this.masterKey = generateLocalKey(KEY_TYPE_MASTER, null);
    }

    /**
     * 初始化
     */
    public void initialize(@NonNull String agentId, @NonNull String identityId,
                           @Nullable MessageBus messageBus) {
        this.agentId = agentId;
        this.identityId = identityId;
        this.messageBus = messageBus;

        if (messageBus != null) {
            messageBus.addListener(this::handleMessage);
        }

        Log.i(TAG, "SecurityClient initialized for " + agentId);
    }

    // === 密钥管理 ===

    /**
     * 生成本地密钥
     */
    @NonNull
    private SecurityKey generateLocalKey(@NonNull String keyType, @Nullable String identityId) {
        String keyId = generateKeyId();
        byte[] keyData = new byte[32]; // AES-256

        SecureRandom random = new SecureRandom();
        random.nextBytes(keyData);

        SecurityKey key = new SecurityKey();
        key.setKeyId(keyId);
        key.setKeyType(keyType);
        key.setAlgorithm(config.defaultAlgorithm);
        key.setKeyData(keyData);
        key.setIdentityId(identityId);
        key.setCreatedAt(System.currentTimeMillis());
        key.setActive(true);

        keyCache.put(keyId, key);

        notifyKeyCreated(key);

        return key;
    }

    /**
     * 创建会话密钥
     */
    @NonNull
    public SecurityKey createSessionKey() {
        SecurityKey key = generateLocalKey(KEY_TYPE_SESSION, identityId);
        key.setAgentId(agentId);
        key.setExpiresAt(System.currentTimeMillis() + config.sessionKeyTtl);

        // 通知 Center
        if (messageBus != null) {
            sendKeyToCenter(key);
        }

        return key;
    }

    /**
     * 创建设备密钥
     */
    @NonNull
    public SecurityKey createDeviceKey() {
        SecurityKey key = generateLocalKey(KEY_TYPE_DEVICE, identityId);
        key.setAgentId(agentId);
        return key;
    }

    /**
     * 获取密钥
     */
    @Nullable
    public SecurityKey getKey(@NonNull String keyId) {
        return keyCache.get(keyId);
    }

    /**
     * 获取活动密钥
     */
    @Nullable
    public SecurityKey getActiveKey(@NonNull String keyType) {
        for (SecurityKey key : keyCache.values()) {
            if (key.isActive() && key.getKeyType().equals(keyType)) {
                if (key.getExpiresAt() == 0 || key.getExpiresAt() > System.currentTimeMillis()) {
                    return key;
                }
            }
        }
        return null;
    }

    /**
     * 轮换密钥
     */
    @Nullable
    public SecurityKey rotateKey(@NonNull String oldKeyId) {
        SecurityKey oldKey = keyCache.get(oldKeyId);
        if (oldKey == null) {
            return null;
        }

        // 生成新密钥
        SecurityKey newKey = generateLocalKey(oldKey.getKeyType(), oldKey.getIdentityId());
        newKey.setAgentId(oldKey.getAgentId());

        // 标记旧密钥为非活动
        oldKey.setActive(false);

        notifyKeyRotated(oldKeyId, newKey.getKeyId());

        return newKey;
    }

    // === 加密/解密 ===

    /**
     * 加密数据
     */
    @Nullable
    public EncryptedData encrypt(@NonNull byte[] plaintext, @NonNull String keyId) {
        SecurityKey key = keyCache.get(keyId);
        if (key == null || !key.isActive()) {
            Log.e(TAG, "Key not found or inactive: " + keyId);
            return null;
        }

        try {
            if (ALGO_AES_GCM.equals(key.getAlgorithm())) {
                return encryptAESGCM(plaintext, key);
            } else if (ALGO_AES_CBC.equals(key.getAlgorithm())) {
                return encryptAESCBC(plaintext, key);
            } else {
                return encryptAESGCM(plaintext, key);
            }
        } catch (Exception e) {
            Log.e(TAG, "Encryption failed", e);
            notifySecurityError("encryption_failed", e.getMessage());
            return null;
        }
    }

    /**
     * AES-GCM 加密
     */
    private EncryptedData encryptAESGCM(byte[] plaintext, SecurityKey key) throws Exception {
        Cipher cipher = Cipher.getInstance("AES/GCM/NoPadding");
        SecretKeySpec keySpec = new SecretKeySpec(key.getKeyData(), "AES");

        // 生成 IV
        byte[] iv = new byte[12];
        SecureRandom random = new SecureRandom();
        random.nextBytes(iv);

        GCMParameterSpec gcmSpec = new GCMParameterSpec(128, iv);
        cipher.init(Cipher.ENCRYPT_MODE, keySpec, gcmSpec);

        byte[] ciphertext = cipher.doFinal(plaintext);

        EncryptedData data = new EncryptedData();
        data.setDataId(generateDataId());
        data.setAlgorithm(ALGO_AES_GCM);
        data.setKeyId(key.getKeyId());
        data.setIv(android.util.Base64.encodeToString(iv, android.util.Base64.NO_WRAP));
        data.setCiphertext(android.util.Base64.encodeToString(ciphertext, android.util.Base64.NO_WRAP));
        data.setTimestamp(System.currentTimeMillis());
        data.setChecksum(calculateChecksum(plaintext));

        notifyEncryptSuccess(data.getDataId(), key.getKeyId());

        return data;
    }

    /**
     * AES-CBC 加密
     */
    private EncryptedData encryptAESCBC(byte[] plaintext, SecurityKey key) throws Exception {
        Cipher cipher = Cipher.getInstance("AES/CBC/PKCS7Padding");
        SecretKeySpec keySpec = new SecretKeySpec(key.getKeyData(), "AES");

        // 生成 IV
        byte[] iv = new byte[16];
        SecureRandom random = new SecureRandom();
        random.nextBytes(iv);

        IvParameterSpec ivSpec = new IvParameterSpec(iv);
        cipher.init(Cipher.ENCRYPT_MODE, keySpec, ivSpec);

        byte[] ciphertext = cipher.doFinal(plaintext);

        EncryptedData data = new EncryptedData();
        data.setDataId(generateDataId());
        data.setAlgorithm(ALGO_AES_CBC);
        data.setKeyId(key.getKeyId());
        data.setIv(android.util.Base64.encodeToString(iv, android.util.Base64.NO_WRAP));
        data.setCiphertext(android.util.Base64.encodeToString(ciphertext, android.util.Base64.NO_WRAP));
        data.setTimestamp(System.currentTimeMillis());
        data.setChecksum(calculateChecksum(plaintext));

        notifyEncryptSuccess(data.getDataId(), key.getKeyId());

        return data;
    }

    /**
     * 解密数据
     */
    @Nullable
    public byte[] decrypt(@NonNull EncryptedData data) {
        SecurityKey key = keyCache.get(data.getKeyId());
        if (key == null || !key.isActive()) {
            Log.e(TAG, "Key not found or inactive: " + data.getKeyId());
            return null;
        }

        try {
            if (ALGO_AES_GCM.equals(data.getAlgorithm())) {
                return decryptAESGCM(data, key);
            } else if (ALGO_AES_CBC.equals(data.getAlgorithm())) {
                return decryptAESCBC(data, key);
            } else {
                return decryptAESGCM(data, key);
            }
        } catch (Exception e) {
            Log.e(TAG, "Decryption failed", e);
            notifySecurityError("decryption_failed", e.getMessage());
            return null;
        }
    }

    /**
     * AES-GCM 解密
     */
    private byte[] decryptAESGCM(EncryptedData data, SecurityKey key) throws Exception {
        Cipher cipher = Cipher.getInstance("AES/GCM/NoPadding");
        SecretKeySpec keySpec = new SecretKeySpec(key.getKeyData(), "AES");

        byte[] iv = android.util.Base64.decode(data.getIv(), android.util.Base64.NO_WRAP);
        byte[] ciphertext = android.util.Base64.decode(data.getCiphertext(), android.util.Base64.NO_WRAP);

        GCMParameterSpec gcmSpec = new GCMParameterSpec(128, iv);
        cipher.init(Cipher.DECRYPT_MODE, keySpec, gcmSpec);

        byte[] plaintext = cipher.doFinal(ciphertext);

        notifyDecryptSuccess(data.getDataId(), key.getKeyId());

        return plaintext;
    }

    /**
     * AES-CBC 解密
     */
    private byte[] decryptAESCBC(EncryptedData data, SecurityKey key) throws Exception {
        Cipher cipher = Cipher.getInstance("AES/CBC/PKCS7Padding");
        SecretKeySpec keySpec = new SecretKeySpec(key.getKeyData(), "AES");

        byte[] iv = android.util.Base64.decode(data.getIv(), android.util.Base64.NO_WRAP);
        byte[] ciphertext = android.util.Base64.decode(data.getCiphertext(), android.util.Base64.NO_WRAP);

        IvParameterSpec ivSpec = new IvParameterSpec(iv);
        cipher.init(Cipher.DECRYPT_MODE, keySpec, ivSpec);

        byte[] plaintext = cipher.doFinal(ciphertext);

        notifyDecryptSuccess(data.getDataId(), key.getKeyId());

        return plaintext;
    }

    /**
     * 加密字符串
     */
    @Nullable
    public EncryptedData encryptString(@NonNull String plaintext, @NonNull String keyId) {
        return encrypt(plaintext.getBytes(StandardCharsets.UTF_8), keyId);
    }

    /**
     * 解密为字符串
     */
    @Nullable
    public String decryptToString(@NonNull EncryptedData data) {
        byte[] plaintext = decrypt(data);
        if (plaintext == null) {
            return null;
        }
        return new String(plaintext, StandardCharsets.UTF_8);
    }

    // === 会话管理 ===

    /**
     * 创建安全会话
     */
    @Nullable
    public SecuritySession createSession(@NonNull String securityLevel) {
        if (sessionCache.size() >= config.maxSessionsPerAgent) {
            // 清理过期会话
            cleanupExpiredSessions();
        }

        SecurityKey sessionKey = createSessionKey();

        SecuritySession session = new SecuritySession();
        session.setSessionId(generateSessionId());
        session.setIdentityId(identityId);
        session.setAgentId(agentId);
        session.setSessionKeyId(sessionKey.getKeyId());
        session.setSecurityLevel(securityLevel);
        session.setCreatedAt(System.currentTimeMillis());
        session.setExpiresAt(System.currentTimeMillis() + config.sessionKeyTtl);
        session.setLastActiveAt(System.currentTimeMillis());
        session.setActive(true);

        sessionCache.put(session.getSessionId(), session);

        // 通知 Center
        if (messageBus != null) {
            sendSessionToCenter(session);
        }

        notifySessionCreated(session);

        return session;
    }

    /**
     * 获取会话
     */
    @Nullable
    public SecuritySession getSession(@NonNull String sessionId) {
        return sessionCache.get(sessionId);
    }

    /**
     * 验证会话
     */
    public boolean validateSession(@NonNull String sessionId) {
        SecuritySession session = sessionCache.get(sessionId);
        if (session == null || !session.isActive()) {
            return false;
        }
        return session.getExpiresAt() > System.currentTimeMillis();
    }

    /**
     * 刷新会话
     */
    public boolean refreshSession(@NonNull String sessionId) {
        SecuritySession session = sessionCache.get(sessionId);
        if (session == null) {
            return false;
        }

        session.setLastActiveAt(System.currentTimeMillis());
        session.setExpiresAt(System.currentTimeMillis() + config.sessionKeyTtl);
        return true;
    }

    /**
     * 关闭会话
     */
    public void closeSession(@NonNull String sessionId) {
        SecuritySession session = sessionCache.get(sessionId);
        if (session != null) {
            session.setActive(false);
            sessionCache.remove(sessionId);
        }
    }

    private void cleanupExpiredSessions() {
        long now = System.currentTimeMillis();
        List<String> expired = new ArrayList<>();
        for (Map.Entry<String, SecuritySession> entry : sessionCache.entrySet()) {
            if (entry.getValue().getExpiresAt() < now) {
                expired.add(entry.getKey());
            }
        }
        for (String sessionId : expired) {
            closeSession(sessionId);
        }
    }

    // === 安全通道 ===

    /**
     * 建立安全通道
     */
    @Nullable
    public SecureChannel establishChannel(@NonNull String targetAgent) {
        SecurityKey channelKey = generateLocalKey(KEY_TYPE_SESSION, identityId);

        SecureChannel channel = new SecureChannel();
        channel.setChannelId(generateChannelId());
        channel.setIdentityId(identityId);
        channel.setSourceAgent(agentId);
        channel.setTargetAgent(targetAgent);
        channel.setChannelKeyId(channelKey.getKeyId());
        channel.setSecurityLevel(config.defaultSecurityLevel);
        channel.setCreatedAt(System.currentTimeMillis());
        channel.setExpiresAt(System.currentTimeMillis() + config.channelKeyTtl);
        channel.setActive(true);

        channelCache.put(channel.getChannelId(), channel);

        // 通知 Center
        if (messageBus != null) {
            sendChannelToCenter(channel);
        }

        notifyChannelEstablished(channel);

        return channel;
    }

    /**
     * 获取通道
     */
    @Nullable
    public SecureChannel getChannel(@NonNull String channelId) {
        return channelCache.get(channelId);
    }

    /**
     * 关闭通道
     */
    public void closeChannel(@NonNull String channelId) {
        SecureChannel channel = channelCache.get(channelId);
        if (channel != null) {
            channel.setActive(false);
            channelCache.remove(channelId);
            notifyChannelClosed(channelId);
        }
    }

    /**
     * 通过通道加密
     */
    @Nullable
    public EncryptedData encryptForChannel(@NonNull String channelId, @NonNull byte[] data) {
        SecureChannel channel = channelCache.get(channelId);
        if (channel == null || !channel.isActive()) {
            Log.e(TAG, "Channel not found or inactive: " + channelId);
            return null;
        }

        if (channel.getExpiresAt() < System.currentTimeMillis()) {
            Log.e(TAG, "Channel expired: " + channelId);
            return null;
        }

        EncryptedData encrypted = encrypt(data, channel.getChannelKeyId());
        if (encrypted != null) {
            channel.incrementMessageCount();
        }

        return encrypted;
    }

    // === 校验和 ===

    /**
     * 计算校验和
     */
    @NonNull
    public String calculateChecksum(@NonNull byte[] data) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(data);
            return android.util.Base64.encodeToString(hash, android.util.Base64.NO_WRAP);
        } catch (Exception e) {
            return "";
        }
    }

    /**
     * 验证校验和
     */
    public boolean verifyChecksum(@NonNull byte[] data, @NonNull String expectedChecksum) {
        return calculateChecksum(data).equals(expectedChecksum);
    }

    // === 消息处理 ===

    private void handleMessage(@NonNull Message message) {
        if (message.payload == null) {
            return;
        }

        Object typeObj = message.payload.get("type");
        if (!"security_event".equals(typeObj)) {
            return;
        }

        try {
            String event = (String) message.payload.get("event");
            if (event == null) {
                return;
            }

            switch (event) {
                case "key_response":
                    handleKeyResponse(message.payload);
                    break;

                case "session_created":
                    handleSessionCreated(message.payload);
                    break;

                case "channel_established":
                    handleChannelEstablished(message.payload);
                    break;
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to handle security event", e);
        }
    }

    private void handleKeyResponse(Map<String, Object> payload) {
        Object keyObj = payload.get("key");
        if (keyObj instanceof JSONObject) {
            try {
                SecurityKey key = SecurityKey.fromJson((JSONObject) keyObj);
                keyCache.put(key.getKeyId(), key);
            } catch (JSONException e) {
                Log.e(TAG, "Failed to parse key", e);
            }
        }
    }

    private void handleSessionCreated(Map<String, Object> payload) {
        Object sessionObj = payload.get("session");
        if (sessionObj instanceof JSONObject) {
            try {
                SecuritySession session = SecuritySession.fromJson((JSONObject) sessionObj);
                sessionCache.put(session.getSessionId(), session);
            } catch (JSONException e) {
                Log.e(TAG, "Failed to parse session", e);
            }
        }
    }

    private void handleChannelEstablished(Map<String, Object> payload) {
        Object channelObj = payload.get("channel");
        if (channelObj instanceof JSONObject) {
            try {
                SecureChannel channel = SecureChannel.fromJson((JSONObject) channelObj);
                channelCache.put(channel.getChannelId(), channel);
            } catch (JSONException e) {
                Log.e(TAG, "Failed to parse channel", e);
            }
        }
    }

    // === 发送到 Center ===

    private void sendKeyToCenter(SecurityKey key) {
        if (messageBus == null) {
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_KEY_REQUEST);
        payload.put("key", key.toJson());

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_HIGH;
        msg.payload = payload;

        messageBus.send(msg);
    }

    private void sendSessionToCenter(SecuritySession session) {
        if (messageBus == null) {
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_SESSION_CREATE);
        payload.put("session", session.toJson());

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_HIGH;
        msg.payload = payload;

        messageBus.send(msg);
    }

    private void sendChannelToCenter(SecureChannel channel) {
        if (messageBus == null) {
            return;
        }

        Map<String, Object> payload = new HashMap<>();
        payload.put("action", MSG_CHANNEL_ESTABLISH);
        payload.put("channel", channel.toJson());

        Message msg = new Message();
        msg.id = generateMessageId();
        msg.fromAgent = agentId;
        msg.toAgent = "center";
        msg.identityId = identityId;
        msg.type = Message.TYPE_DATA;
        msg.priority = Message.PRIORITY_HIGH;
        msg.payload = payload;

        messageBus.send(msg);
    }

    // === 监听器管理 ===

    public void addListener(@NonNull SecurityListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull SecurityListener listener) {
        listeners.remove(listener);
    }

    private void notifyKeyCreated(SecurityKey key) {
        for (SecurityListener l : listeners) {
            l.onKeyCreated(key);
        }
    }

    private void notifyKeyRotated(String oldKeyId, String newKeyId) {
        for (SecurityListener l : listeners) {
            l.onKeyRotated(oldKeyId, newKeyId);
        }
    }

    private void notifySessionCreated(SecuritySession session) {
        for (SecurityListener l : listeners) {
            l.onSessionCreated(session);
        }
    }

    private void notifySessionExpired(String sessionId) {
        for (SecurityListener l : listeners) {
            l.onSessionExpired(sessionId);
        }
    }

    private void notifyChannelEstablished(SecureChannel channel) {
        for (SecurityListener l : listeners) {
            l.onChannelEstablished(channel);
        }
    }

    private void notifyChannelClosed(String channelId) {
        for (SecurityListener l : listeners) {
            l.onChannelClosed(channelId);
        }
    }

    private void notifyEncryptSuccess(String dataId, String keyId) {
        for (SecurityListener l : listeners) {
            l.onEncryptSuccess(dataId, keyId);
        }
    }

    private void notifyDecryptSuccess(String dataId, String keyId) {
        for (SecurityListener l : listeners) {
            l.onDecryptSuccess(dataId, keyId);
        }
    }

    private void notifySecurityError(String errorType, String details) {
        for (SecurityListener l : listeners) {
            l.onSecurityError(errorType, details);
        }
    }

    // === 统计信息 ===

    @NonNull
    public SecurityStats getStats() {
        SecurityStats stats = new SecurityStats();
        stats.totalKeys = keyCache.size();
        stats.activeSessions = sessionCache.size();
        stats.activeChannels = channelCache.size();

        for (SecurityKey key : keyCache.values()) {
            if (key.isActive()) {
                stats.activeKeys++;
            }
        }

        return stats;
    }

    // === 辅助方法 ===

    private String generateKeyId() {
        return "key-" + System.currentTimeMillis() + "-" + randomHex(4);
    }

    private String generateSessionId() {
        return "session-" + System.currentTimeMillis() + "-" + randomHex(4);
    }

    private String generateChannelId() {
        return "channel-" + System.currentTimeMillis() + "-" + randomHex(4);
    }

    private String generateDataId() {
        return "data-" + System.currentTimeMillis() + "-" + randomHex(4);
    }

    private String generateMessageId() {
        return "sec_msg_" + System.currentTimeMillis() + "_" + agentId;
    }

    private String randomHex(int length) {
        SecureRandom random = new SecureRandom();
        byte[] bytes = new byte[length];
        random.nextBytes(bytes);
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        keyCache.clear();
        sessionCache.clear();
        channelCache.clear();
        listeners.clear();

        Log.i(TAG, "SecurityClient cleaned up");
    }

    /**
     * 统计信息
     */
    public static class SecurityStats {
        public int totalKeys;
        public int activeKeys;
        public int activeSessions;
        public int activeChannels;

        @NonNull
        @Override
        public String toString() {
            return "SecurityStats{" +
                    "keys=" + totalKeys +
                    ", activeKeys=" + activeKeys +
                    ", sessions=" + activeSessions +
                    ", channels=" + activeChannels +
                    '}';
        }
    }
}