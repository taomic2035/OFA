package com.ofa.agent.trust;

import android.content.Context;
import android.content.SharedPreferences;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 设备信任管理器 (v2.8.0)
 *
 * 管理设备的信任级别、认证和授权。
 * Center 是永远在线的灵魂载体，设备通过信任链获得访问权限。
 *
 * 主要功能：
 * - 设备注册与认证
 * - 信任级别管理
 * - 设备更换支持
 * - 本地凭证存储
 */
public class DeviceTrustManager {

    private static final String TAG = "DeviceTrustManager";
    private static final String PREFS_NAME = "ofa_device_trust";
    private static final String KEY_DEVICE_TOKEN = "device_token";
    private static final String KEY_REFRESH_TOKEN = "refresh_token";
    private static final String KEY_TOKEN_EXPIRY = "token_expiry";
    private static final String KEY_AUTHORIZATION = "authorization";

    private final Context context;
    private final SharedPreferences prefs;
    private final ExecutorService executor;
    private final String centerAddress;

    private String deviceToken;
    private String refreshToken;
    private long tokenExpiry;
    private DeviceAuthorization authorization;

    // 监听器
    private TrustChangeListener trustChangeListener;

    public DeviceTrustManager(@NonNull Context context, @Nullable String centerAddress) {
        this.context = context.getApplicationContext();
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
        this.executor = Executors.newSingleThreadExecutor();
        this.centerAddress = centerAddress;

        // 加载本地存储的凭证
        loadCredentials();
    }

    /**
     * 获取当前设备授权
     */
    @Nullable
    public DeviceAuthorization getAuthorization() {
        return authorization;
    }

    /**
     * 获取当前信任级别
     */
    @NonNull
    public TrustLevel getTrustLevel() {
        return authorization != null ? authorization.getTrustLevel() : TrustLevel.NONE;
    }

    /**
     * 检查是否为主设备
     */
    public boolean isPrimaryDevice() {
        return authorization != null && authorization.isPrimary();
    }

    /**
     * 检查是否有指定权限
     */
    public boolean hasPermission(String permission) {
        return authorization != null && authorization.hasPermission(permission);
    }

    /**
     * 检查凭证是否有效
     */
    public boolean hasValidCredentials() {
        if (deviceToken == null || deviceToken.isEmpty()) {
            return false;
        }
        if (tokenExpiry > 0 && System.currentTimeMillis() > tokenExpiry) {
            return false;
        }
        return authorization != null && !authorization.isExpired();
    }

    /**
     * 注册设备到 Center
     */
    public CompletableFuture<RegisterResult> registerDevice(@NonNull String agentId,
                                                            @NonNull String identityId,
                                                            @NonNull String deviceType,
                                                            @NonNull String deviceName) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                JSONObject request = new JSONObject();
                request.put("agent_id", agentId);
                request.put("identity_id", identityId);
                request.put("device_type", deviceType);
                request.put("device_name", deviceName);

                // 发送注册请求到 Center
                String responseJson = sendRegisterRequest(request);

                JSONObject response = new JSONObject(responseJson);
                if (response.optBoolean("success", false)) {
                    // 保存凭证
                    deviceToken = response.getString("device_token");
                    refreshToken = response.getString("refresh_token");
                    tokenExpiry = response.getLong("expires_at");

                    // 解析授权
                    authorization = new DeviceAuthorization();
                    authorization.setAgentId(agentId);
                    authorization.setIdentityId(identityId);
                    authorization.setTrustLevel(TrustLevel.fromName(response.optString("trust_level")));

                    // 保存到本地
                    saveCredentials();

                    notifyTrustChanged();

                    return new RegisterResult(true, authorization, null);
                } else {
                    return new RegisterResult(false, null, response.optString("error", "Unknown error"));
                }
            } catch (Exception e) {
                Log.e(TAG, "Failed to register device", e);
                return new RegisterResult(false, null, e.getMessage());
            }
        }, executor);
    }

    /**
     * 刷新设备令牌
     */
    public CompletableFuture<Boolean> refreshToken() {
        return CompletableFuture.supplyAsync(() -> {
            if (refreshToken == null) {
                return false;
            }

            try {
                JSONObject request = new JSONObject();
                request.put("refresh_token", refreshToken);

                String responseJson = sendRefreshRequest(request);
                JSONObject response = new JSONObject(responseJson);

                if (response.optBoolean("success", false)) {
                    deviceToken = response.getString("device_token");
                    refreshToken = response.getString("refresh_token");
                    tokenExpiry = response.getLong("expires_at");
                    saveCredentials();
                    return true;
                }
                return false;
            } catch (Exception e) {
                Log.e(TAG, "Failed to refresh token", e);
                return false;
            }
        }, executor);
    }

    /**
     * 请求提升信任级别
     */
    public CompletableFuture<Boolean> requestTrustUpgrade(@NonNull TrustLevel targetLevel) {
        return CompletableFuture.supplyAsync(() -> {
            if (authorization == null) {
                return false;
            }

            try {
                JSONObject request = new JSONObject();
                request.put("agent_id", authorization.getAgentId());
                request.put("target_level", targetLevel.getName());

                String responseJson = sendTrustUpgradeRequest(request);
                JSONObject response = new JSONObject(responseJson);

                if (response.optBoolean("success", false)) {
                    authorization.setTrustLevel(targetLevel);
                    saveCredentials();
                    notifyTrustChanged();
                    return true;
                }
                return false;
            } catch (Exception e) {
                Log.e(TAG, "Failed to upgrade trust level", e);
                return false;
            }
        }, executor);
    }

    /**
     * 准备设备迁移（设备丢失场景）
     */
    public CompletableFuture<TransferResult> prepareTransfer(@NonNull String toDeviceId,
                                                             @NonNull String reason) {
        return CompletableFuture.supplyAsync(() -> {
            if (authorization == null) {
                return new TransferResult(false, "Not authorized");
            }

            try {
                JSONObject request = new JSONObject();
                request.put("from_agent_id", authorization.getAgentId());
                request.put("to_agent_id", toDeviceId);
                request.put("identity_id", authorization.getIdentityId());
                request.put("reason", reason);

                String responseJson = sendTransferRequest(request);
                JSONObject response = new JSONObject(responseJson);

                return new TransferResult(
                        response.optBoolean("success", false),
                        response.optString("error")
                );
            } catch (Exception e) {
                Log.e(TAG, "Failed to prepare transfer", e);
                return new TransferResult(false, e.getMessage());
            }
        }, executor);
    }

    /**
     * 接收灵魂迁移（新设备接收）
     */
    public CompletableFuture<Boolean> receiveTransfer(@NonNull String transferToken,
                                                      @NonNull String fromDeviceId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                JSONObject request = new JSONObject();
                request.put("transfer_token", transferToken);
                request.put("from_agent_id", fromDeviceId);

                String responseJson = sendReceiveTransferRequest(request);
                JSONObject response = new JSONObject(responseJson);

                if (response.optBoolean("success", false)) {
                    // 更新授权
                    if (authorization != null) {
                        authorization.setTrustLevel(TrustLevel.PRIMARY);
                        saveCredentials();
                        notifyTrustChanged();
                    }
                    return true;
                }
                return false;
            } catch (Exception e) {
                Log.e(TAG, "Failed to receive transfer", e);
                return false;
            }
        }, executor);
    }

    /**
     * 注销设备
     */
    public void revoke() {
        deviceToken = null;
        refreshToken = null;
        tokenExpiry = 0;
        authorization = null;

        prefs.edit()
                .remove(KEY_DEVICE_TOKEN)
                .remove(KEY_REFRESH_TOKEN)
                .remove(KEY_TOKEN_EXPIRY)
                .remove(KEY_AUTHORIZATION)
                .apply();

        notifyTrustChanged();
    }

    /**
     * 设置信任变更监听器
     */
    public void setTrustChangeListener(TrustChangeListener listener) {
        this.trustChangeListener = listener;
    }

    // === 私有方法 ===

    private void loadCredentials() {
        deviceToken = prefs.getString(KEY_DEVICE_TOKEN, null);
        refreshToken = prefs.getString(KEY_REFRESH_TOKEN, null);
        tokenExpiry = prefs.getLong(KEY_TOKEN_EXPIRY, 0);

        String authJson = prefs.getString(KEY_AUTHORIZATION, null);
        if (authJson != null) {
            try {
                authorization = DeviceAuthorization.fromJson(new JSONObject(authJson));
            } catch (JSONException e) {
                Log.e(TAG, "Failed to parse authorization", e);
            }
        }
    }

    private void saveCredentials() {
        SharedPreferences.Editor editor = prefs.edit();
        editor.putString(KEY_DEVICE_TOKEN, deviceToken);
        editor.putString(KEY_REFRESH_TOKEN, refreshToken);
        editor.putLong(KEY_TOKEN_EXPIRY, tokenExpiry);

        if (authorization != null) {
            try {
                editor.putString(KEY_AUTHORIZATION, authorization.toJson().toString());
            } catch (JSONException e) {
                Log.e(TAG, "Failed to serialize authorization", e);
            }
        }
        editor.apply();
    }

    private void notifyTrustChanged() {
        if (trustChangeListener != null) {
            trustChangeListener.onTrustChanged(authorization);
        }
    }

    // HTTP 请求方法（简化实现）
    private String sendRegisterRequest(JSONObject request) throws Exception {
        // 实际实现需要使用 HTTP 客户端发送请求到 Center
        // POST /api/v1/trust/register
        if (centerAddress == null) {
            throw new Exception("Center address not configured");
        }

        // TODO: 实现 HTTP 请求
        // 模拟返回
        JSONObject response = new JSONObject();
        response.put("success", true);
        response.put("device_token", "dev_" + System.currentTimeMillis());
        response.put("refresh_token", "ref_" + System.currentTimeMillis());
        response.put("expires_at", System.currentTimeMillis() + 30L * 24 * 60 * 60 * 1000);
        response.put("trust_level", "low");

        return response.toString();
    }

    private String sendRefreshRequest(JSONObject request) throws Exception {
        // POST /api/v1/trust/refresh
        JSONObject response = new JSONObject();
        response.put("success", true);
        return response.toString();
    }

    private String sendTrustUpgradeRequest(JSONObject request) throws Exception {
        // POST /api/v1/trust/upgrade
        JSONObject response = new JSONObject();
        response.put("success", false);
        response.put("error", "Upgrade requires Center approval");
        return response.toString();
    }

    private String sendTransferRequest(JSONObject request) throws Exception {
        // POST /api/v1/trust/transfer
        JSONObject response = new JSONObject();
        response.put("success", true);
        return response.toString();
    }

    private String sendReceiveTransferRequest(JSONObject request) throws Exception {
        // POST /api/v1/trust/receive
        JSONObject response = new JSONObject();
        response.put("success", true);
        return response.toString();
    }

    // === 结果类 ===

    public static class RegisterResult {
        public final boolean success;
        public final DeviceAuthorization authorization;
        public final String error;

        public RegisterResult(boolean success, DeviceAuthorization authorization, String error) {
            this.success = success;
            this.authorization = authorization;
            this.error = error;
        }
    }

    public static class TransferResult {
        public final boolean success;
        public final String error;

        public TransferResult(boolean success, String error) {
            this.success = success;
            this.error = error;
        }
    }

    // === 监听器接口 ===

    public interface TrustChangeListener {
        void onTrustChanged(@Nullable DeviceAuthorization authorization);
    }
}