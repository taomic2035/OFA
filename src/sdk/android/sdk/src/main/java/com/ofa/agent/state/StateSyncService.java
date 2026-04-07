package com.ofa.agent.state;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.NetworkInfo;
import android.os.BatteryManager;
import android.os.Build;
import android.os.Handler;
import android.os.Looper;
import android.os.PowerManager;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 状态同步服务 (v3.1.0)
 *
 * Center 是永远在线的灵魂载体，状态同步服务负责：
 * - 收集本地设备状态
 * - 上报状态到 Center
 * - 接收其他设备状态变更通知
 * - 维护本地设备状态缓存
 */
public class StateSyncService {

    private static final String TAG = "StateSyncService";

    // 配置
    private static final long DEFAULT_SYNC_INTERVAL = 10 * 1000L; // 10秒
    private static final long HEARTBEAT_INTERVAL = 30 * 1000L;    // 30秒
    private static final int MAX_CACHED_STATES = 50;

    private final Context context;
    private final ExecutorService executor;
    private final Handler mainHandler;

    // 设备信息
    private String agentId;
    private String identityId;
    private String deviceType;
    private String deviceName;

    // 当前状态
    private DeviceState currentState;

    // 其他设备状态缓存
    private final Map<String, DeviceState> deviceStates;

    // 状态变更历史
    private final List<StateChange> stateHistory;

    // 消息总线
    private MessageBus messageBus;

    // 状态监听器
    private final List<StateListener> listeners;

    // 配置
    private StateSyncConfig config;

    // 系统状态监听
    private BroadcastReceiver batteryReceiver;
    private BroadcastReceiver networkReceiver;
    private ConnectivityManager.NetworkCallback networkCallback;

    // 同步状态
    private boolean syncing = false;
    private long lastSyncTime = 0;
    private long stateVersion = 0;

    // 停止标志
    private boolean stopped = false;

    /**
     * 状态同步配置
     */
    public static class StateSyncConfig {
        public long syncInterval = DEFAULT_SYNC_INTERVAL;
        public boolean reportBattery = true;
        public boolean reportNetwork = true;
        public boolean reportScene = true;
        public boolean reportLocation = false;  // 位置敏感，默认不上报
        public boolean reportActiveApps = false;
        public int maxHistorySize = 100;
    }

    /**
     * 状态监听器
     */
    public interface StateListener {
        /**
         * 本地状态变更
         */
        void onLocalStateChanged(@NonNull StateChange change);

        /**
         * 其他设备状态变更
         */
        void onRemoteStateChanged(@NonNull StateChange change);

        /**
         * 设备上线
         */
        void onDeviceOnline(@NonNull DeviceState state);

        /**
         * 设备离线
         */
        void onDeviceOffline(@NonNull String agentId);
    }

    public StateSyncService(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.executor = Executors.newSingleThreadExecutor();
        this.mainHandler = new Handler(Looper.getMainLooper());
        this.deviceStates = new HashMap<>();
        this.stateHistory = new ArrayList<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.config = new StateSyncConfig();

        initSystemListeners();
    }

    /**
     * 初始化状态同步服务
     */
    public void initialize(@NonNull String agentId, @NonNull String identityId,
                           @NonNull String deviceType, @NonNull String deviceName) {
        this.agentId = agentId;
        this.identityId = identityId;
        this.deviceType = deviceType;
        this.deviceName = deviceName;

        // 创建初始状态
        currentState = DeviceState.create(agentId, identityId, deviceType, deviceName);
        currentState.setCapabilities(getDefaultCapabilities());

        // 收集初始系统状态
        collectSystemState();

        Log.i(TAG, "StateSyncService initialized for " + agentId);
    }

    /**
     * 设置消息总线
     */
    public void setMessageBus(@Nullable MessageBus messageBus) {
        this.messageBus = messageBus;

        if (messageBus != null) {
            // 注册消息监听器
            messageBus.addListener(new MessageBus.MessageListener() {
                @Override
                public void onMessage(Message message) {
                    handleStateMessage(message);
                }
            });
        }
    }

    /**
     * 设置配置
     */
    public void setConfig(@NonNull StateSyncConfig config) {
        this.config = config;
    }

    /**
     * 开始同步
     */
    public void startSync() {
        if (syncing) {
            return;
        }

        syncing = true;
        stopped = false;

        // 启动定时同步
        startPeriodicSync();

        // 立即上报一次完整状态
        reportFullState();

        Log.i(TAG, "State sync started");
    }

    /**
     * 停止同步
     */
    public void stopSync() {
        syncing = false;
        stopped = true;

        // 取消定时任务
        mainHandler.removeCallbacksAndMessages(null);

        Log.i(TAG, "State sync stopped");
    }

    // === 状态收集 ===

    /**
     * 收集系统状态
     */
    private void collectSystemState() {
        if (currentState == null) {
            return;
        }

        // 收集电池状态
        if (config.reportBattery) {
            collectBatteryState();
        }

        // 收集网络状态
        if (config.reportNetwork) {
            collectNetworkState();
        }

        // 收集场景状态（由外部设置）
        if (config.reportScene && currentState.getScene() == null) {
            currentState.setScene("idle");
        }

        currentState.setUpdatedAt(System.currentTimeMillis());
    }

    /**
     * 收集电池状态
     */
    private void collectBatteryState() {
        IntentFilter filter = new IntentFilter(Intent.ACTION_BATTERY_CHANGED);
        Intent batteryStatus = context.registerReceiver(null, filter);

        if (batteryStatus == null) {
            return;
        }

        BatteryManager bm = (BatteryManager) context.getSystemService(Context.BATTERY_SERVICE);

        int level = batteryStatus.getIntExtra(BatteryManager.EXTRA_LEVEL, -1);
        int scale = batteryStatus.getIntExtra(BatteryManager.EXTRA_SCALE, -1);
        int batteryPct = (level * 100) / scale;

        int status = batteryStatus.getIntExtra(BatteryManager.EXTRA_STATUS, -1);
        boolean isCharging = status == BatteryManager.BATTERY_STATUS_CHARGING ||
                status == BatteryManager.BATTERY_STATUS_FULL;

        PowerManager pm = (PowerManager) context.getSystemService(Context.POWER_SERVICE);
        boolean powerSaver = false;
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            powerSaver = pm != null && pm.isPowerSaveMode();
        }

        currentState.setBatteryLevel(batteryPct);
        currentState.setCharging(isCharging);
        currentState.setPowerSaver(powerSaver);
    }

    /**
     * 收集网络状态
     */
    private void collectNetworkState() {
        ConnectivityManager cm = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);

        if (cm == null) {
            currentState.setNetworkType("unknown");
            currentState.setNetworkStrength(0);
            return;
        }

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            Network network = cm.getActiveNetwork();
            if (network == null) {
                currentState.setNetworkType("offline");
                currentState.setNetworkStrength(0);
                return;
            }

            NetworkCapabilities caps = cm.getNetworkCapabilities(network);
            if (caps == null) {
                currentState.setNetworkType("unknown");
                currentState.setNetworkStrength(0);
                return;
            }

            if (caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI)) {
                currentState.setNetworkType("wifi");
                // WiFi 信号强度估算
                int strength = estimateWiFiStrength(caps);
                currentState.setNetworkStrength(strength);
            } else if (caps.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)) {
                currentState.setNetworkType("cellular");
                int strength = estimateCellularStrength(caps);
                currentState.setNetworkStrength(strength);
            } else {
                currentState.setNetworkType("other");
                currentState.setNetworkStrength(50);
            }
        } else {
            NetworkInfo info = cm.getActiveNetworkInfo();
            if (info == null || !info.isConnected()) {
                currentState.setNetworkType("offline");
                currentState.setNetworkStrength(0);
            } else {
                switch (info.getType()) {
                    case ConnectivityManager.TYPE_WIFI:
                        currentState.setNetworkType("wifi");
                        currentState.setNetworkStrength(100);
                        break;
                    case ConnectivityManager.TYPE_MOBILE:
                        currentState.setNetworkType("cellular");
                        currentState.setNetworkStrength(estimateLegacyCellularStrength(info));
                        break;
                    default:
                        currentState.setNetworkType("other");
                        currentState.setNetworkStrength(50);
                }
            }
        }
    }

    private int estimateWiFiStrength(NetworkCapabilities caps) {
        // 基于带宽估算信号强度
        int bandwidth = caps.getLinkDownstreamBandwidthKbps();
        if (bandwidth >= 50000) return 100; // 50+ Mbps
        if (bandwidth >= 20000) return 80;  // 20+ Mbps
        if (bandwidth >= 10000) return 60;  // 10+ Mbps
        if (bandwidth >= 5000) return 40;   // 5+ Mbps
        return 20;
    }

    private int estimateCellularStrength(NetworkCapabilities caps) {
        if (caps.hasCapability(NetworkCapabilities.NET_CAPABILITY_NOT_METERED)) {
            return 100; // 非计费网络通常是高质量
        }
        int bandwidth = caps.getLinkDownstreamBandwidthKbps();
        if (bandwidth >= 10000) return 80;
        if (bandwidth >= 5000) return 60;
        if (bandwidth >= 2000) return 40;
        return 20;
    }

    private int estimateLegacyCellularStrength(NetworkInfo info) {
        int type = info.getSubtype();
        if (type == TelephonyInfo.NETWORK_TYPE_LTE ||
                type == TelephonyInfo.NETWORK_TYPE_HSPAP ||
                type == TelephonyInfo.NETWORK_TYPE_EHRPD) {
            return 80;
        }
        if (type == TelephonyInfo.NETWORK_TYPE_HSDPA ||
                type == TelephonyInfo.NETWORK_TYPE_HSPA ||
                type == TelephonyInfo.NETWORK_TYPE_HSUPA) {
            return 60;
        }
        if (type == TelephonyInfo.NETWORK_TYPE_UMTS ||
                type == TelephonyInfo.NETWORK_TYPE_EVDO_0 ||
                type == TelephonyInfo.NETWORK_TYPE_EVDO_A) {
            return 40;
        }
        return 20;
    }

    // === 状态上报 ===

    /**
     * 上报完整状态
     */
    public void reportFullState() {
        if (messageBus == null || currentState == null) {
            return;
        }

        collectSystemState();
        currentState.setStateVersion(++stateVersion);

        executor.execute(() -> {
            try {
                Message msg = new Message();
                msg.id = generateStateMessageId();
                msg.fromAgent = agentId;
                msg.toAgent = "center";
                msg.identityId = identityId;
                msg.type = Message.TYPE_DATA;
                msg.priority = Message.PRIORITY_NORMAL;

                Map<String, Object> payload = new HashMap<>();
                payload.put("action", "state_update");
                payload.put("change_type", DeviceState.CHANGE_FULL);
                payload.put("state", currentState.toJson());
                msg.payload = payload;

                messageBus.send(msg);
                lastSyncTime = System.currentTimeMillis();

                Log.d(TAG, "Full state reported: " + currentState);

            } catch (JSONException e) {
                Log.e(TAG, "Failed to report state", e);
            }
        });
    }

    /**
     * 上报部分状态变更
     */
    public void reportStateChange(@NonNull String changeType) {
        if (messageBus == null || currentState == null) {
            return;
        }

        currentState.setStateVersion(++stateVersion);
        currentState.setUpdatedAt(System.currentTimeMillis());

        executor.execute(() -> {
            try {
                Message msg = new Message();
                msg.id = generateStateMessageId();
                msg.fromAgent = agentId;
                msg.toAgent = "center";
                msg.identityId = identityId;
                msg.type = Message.TYPE_DATA;
                msg.priority = getChangePriority(changeType);

                Map<String, Object> payload = new HashMap<>();
                payload.put("action", "state_update");
                payload.put("change_type", changeType);
                payload.put("state", currentState.toJson());
                msg.payload = payload;

                messageBus.send(msg);

                Log.d(TAG, "State change reported: " + changeType);

            } catch (JSONException e) {
                Log.e(TAG, "Failed to report state change", e);
            }
        });
    }

    private int getChangePriority(String changeType) {
        if (DeviceState.CHANGE_ONLINE.equals(changeType) ||
                DeviceState.CHANGE_OFFLINE.equals(changeType)) {
            return Message.PRIORITY_HIGH;
        }
        if (DeviceState.CHANGE_BATTERY.equals(changeType)) {
            return Message.PRIORITY_NORMAL;
        }
        return Message.PRIORITY_NORMAL;
    }

    /**
     * 发送心跳
     */
    public void sendHeartbeat() {
        if (messageBus == null) {
            return;
        }

        executor.execute(() -> {
            Message msg = new Message();
            msg.id = generateStateMessageId();
            msg.fromAgent = agentId;
            msg.toAgent = "center";
            msg.identityId = identityId;
            msg.type = Message.TYPE_HEARTBEAT;
            msg.priority = Message.PRIORITY_NORMAL;

            Map<String, Object> payload = new HashMap<>();
            payload.put("state_version", stateVersion);
            payload.put("last_seen", System.currentTimeMillis());
            msg.payload = payload;

            messageBus.send(msg);
        });
    }

    // === 定时同步 ===

    private void startPeriodicSync() {
        mainHandler.postDelayed(new Runnable() {
            @Override
            public void run() {
                if (stopped) {
                    return;
                }

                // 收集并上报状态
                collectSystemState();

                // 检查是否有变更
                String changeType = currentState.detectChangeType(getPreviousState());
                if (!DeviceState.CHANGE_FULL.equals(changeType) &&
                        System.currentTimeMillis() - lastSyncTime > config.syncInterval) {
                    reportStateChange(changeType);
                }

                // 发送心跳
                sendHeartbeat();

                // 继续定时
                mainHandler.postDelayed(this, HEARTBEAT_INTERVAL);
            }
        }, HEARTBEAT_INTERVAL);
    }

    private DeviceState getPreviousState() {
        // 从历史获取上一个状态
        if (stateHistory.isEmpty()) {
            return null;
        }
        StateChange last = stateHistory.get(stateHistory.size() - 1);
        return last.getNewState();
    }

    // === 消息处理 ===

    private void handleStateMessage(Message message) {
        if (message == null || message.payload == null) {
            return;
        }

        Object changeTypeObj = message.payload.get("change_type");
        String changeType = changeTypeObj != null ? changeTypeObj.toString() : null;

        // 处理状态变更通知
        if ("state_change".equals(message.payload.get("action")) ||
                Message.TYPE_NOTIFICATION.equals(message.type)) {

            Object stateObj = message.payload.get("new_state");
            if (stateObj instanceof JSONObject) {
                try {
                    DeviceState newState = DeviceState.fromJson((JSONObject) stateObj);
                    handleRemoteStateChange(newState, changeType);
                } catch (JSONException e) {
                    Log.e(TAG, "Failed to parse remote state", e);
                }
            } else if (stateObj instanceof Map) {
                // 处理 Map 格式
                try {
                    JSONObject json = new JSONObject((Map) stateObj);
                    DeviceState newState = DeviceState.fromJson(json);
                    handleRemoteStateChange(newState, changeType);
                } catch (JSONException e) {
                    Log.e(TAG, "Failed to parse remote state from map", e);
                }
            }
        }

        // 处理完整状态同步
        if ("full_state_sync".equals(message.payload.get("type"))) {
            Object statesObj = message.payload.get("states");
            if (statesObj instanceof JSONArray) {
                handleFullStateSync((JSONArray) statesObj);
            }
        }
    }

    private void handleRemoteStateChange(@NonNull DeviceState newState, @Nullable String changeType) {
        String agentId = newState.getAgentId();

        // 不处理自己的状态
        if (agentId.equals(this.agentId)) {
            return;
        }

        // 获取旧状态
        DeviceState oldState = deviceStates.get(agentId);

        // 更新缓存
        deviceStates.put(agentId, newState);

        // 清理缓存
        if (deviceStates.size() > MAX_CACHED_STATES) {
            cleanDeviceStateCache();
        }

        // 创建变更事件
        StateChange change = StateChange.create(
                agentId,
                newState.getIdentityId(),
                changeType != null ? changeType : newState.detectChangeType(oldState),
                oldState,
                newState
        );

        // 通知监听器
        notifyRemoteStateChange(change);

        Log.d(TAG, "Remote state change: " + agentId + " -> " + changeType);
    }

    private void handleFullStateSync(@NonNull JSONArray states) {
        try {
            for (int i = 0; i < states.length(); i++) {
                JSONObject stateJson = states.getJSONObject(i);
                DeviceState state = DeviceState.fromJson(stateJson);

                if (!state.getAgentId().equals(this.agentId)) {
                    deviceStates.put(state.getAgentId(), state);
                }
            }

            Log.d(TAG, "Full state sync received: " + states.length() + " devices");

        } catch (JSONException e) {
            Log.e(TAG, "Failed to parse full state sync", e);
        }
    }

    // === 本地状态更新 ===

    /**
     * 更新场景状态
     */
    public void updateScene(@NonNull String scene, @Nullable Map<String, Object> context) {
        if (currentState == null) {
            return;
        }

        String oldScene = currentState.getScene();

        currentState.setScene(scene);
        if (context != null) {
            currentState.setSceneContext(context);
        }
        currentState.setUpdatedAt(System.currentTimeMillis());

        if (!oldScene.equals(scene)) {
            reportStateChange(DeviceState.CHANGE_SCENE);
            notifyLocalStateChange(DeviceState.CHANGE_SCENE);
        }
    }

    /**
     * 更新电池状态（由系统监听器调用）
     */
    public void updateBatteryState(int level, boolean charging, boolean powerSaver) {
        if (currentState == null) {
            return;
        }

        int oldLevel = currentState.getBatteryLevel();
        boolean oldCharging = currentState.isCharging();

        currentState.setBatteryLevel(level);
        currentState.setCharging(charging);
        currentState.setPowerSaver(powerSaver);
        currentState.setUpdatedAt(System.currentTimeMillis());

        // 电池变化超过 10% 或充电状态变化时上报
        if (Math.abs(level - oldLevel) > 10 || charging != oldCharging) {
            reportStateChange(DeviceState.CHANGE_BATTERY);
            notifyLocalStateChange(DeviceState.CHANGE_BATTERY);
        }
    }

    /**
     * 更新网络状态（由系统监听器调用）
     */
    public void updateNetworkState(@NonNull String networkType, int strength, boolean roaming) {
        if (currentState == null) {
            return;
        }

        String oldType = currentState.getNetworkType();

        currentState.setNetworkType(networkType);
        currentState.setNetworkStrength(strength);
        currentState.setRoaming(roaming);
        currentState.setUpdatedAt(System.currentTimeMillis());

        if (!oldType.equals(networkType)) {
            reportStateChange(DeviceState.CHANGE_NETWORK);
            notifyLocalStateChange(DeviceState.CHANGE_NETWORK);
        }
    }

    /**
     * 更新位置状态
     */
    public void updateLocation(@Nullable DeviceLocation location) {
        if (currentState == null || !config.reportLocation) {
            return;
        }

        currentState.setLocation(location);
        currentState.setUpdatedAt(System.currentTimeMillis());

        reportStateChange(DeviceState.CHANGE_LOCATION);
        notifyLocalStateChange(DeviceState.CHANGE_LOCATION);
    }

    /**
     * 更新活跃应用列表
     */
    public void updateActiveApps(@NonNull List<String> activeApps) {
        if (currentState == null || !config.reportActiveApps) {
            return;
        }

        currentState.setActiveApps(activeApps);
        currentState.setUpdatedAt(System.currentTimeMillis());
    }

    /**
     * 设置在线状态
     */
    public void setOnline(boolean online) {
        if (currentState == null) {
            return;
        }

        boolean wasOnline = currentState.isOnline();

        currentState.setOnline(online);
        currentState.setLastSeen(System.currentTimeMillis());
        currentState.setUpdatedAt(System.currentTimeMillis());

        if (wasOnline != online) {
            reportStateChange(online ? DeviceState.CHANGE_ONLINE : DeviceState.CHANGE_OFFLINE);
            notifyLocalStateChange(online ? DeviceState.CHANGE_ONLINE : DeviceState.CHANGE_OFFLINE);
        }
    }

    // === 状态查询 ===

    /**
     * 获取当前设备状态
     */
    @Nullable
    public DeviceState getCurrentState() {
        return currentState;
    }

    /**
     * 获取指定设备状态
     */
    @Nullable
    public DeviceState getDeviceState(@NonNull String agentId) {
        return deviceStates.get(agentId);
    }

    /**
     * 获取所有设备状态
     */
    @NonNull
    public List<DeviceState> getAllDeviceStates() {
        return new ArrayList<>(deviceStates.values());
    }

    /**
     * 获取身份下所有设备状态
     */
    @NonNull
    public List<DeviceState> getIdentityDeviceStates(@NonNull String identityId) {
        List<DeviceState> result = new ArrayList<>();
        for (DeviceState state : deviceStates.values()) {
            if (identityId.equals(state.getIdentityId())) {
                result.add(state);
            }
        }
        return result;
    }

    /**
     * 获取在线设备
     */
    @NonNull
    public List<DeviceState> getOnlineDevices(@NonNull String identityId) {
        List<DeviceState> result = new ArrayList<>();
        for (DeviceState state : deviceStates.values()) {
            if (identityId.equals(state.getIdentityId()) && state.isOnline()) {
                result.add(state);
            }
        }
        return result;
    }

    /**
     * 获取指定场景的设备
     */
    @NonNull
    public List<DeviceState> getDevicesInScene(@NonNull String identityId, @NonNull String scene) {
        List<DeviceState> result = new ArrayList<>();
        for (DeviceState state : deviceStates.values()) {
            if (identityId.equals(state.getIdentityId()) && scene.equals(state.getScene())) {
                result.add(state);
            }
        }
        return result;
    }

    /**
     * 获取状态历史
     */
    @NonNull
    public List<StateChange> getStateHistory(int limit) {
        if (limit <= 0 || limit >= stateHistory.size()) {
            return new ArrayList<>(stateHistory);
        }
        return new ArrayList<>(stateHistory.subList(
                stateHistory.size() - limit, stateHistory.size()));
    }

    // === 监听器管理 ===

    /**
     * 添加状态监听器
     */
    public void addListener(@NonNull StateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull StateListener listener) {
        listeners.remove(listener);
    }

    private void notifyLocalStateChange(@NonNull String changeType) {
        StateChange change = StateChange.create(
                agentId, identityId, changeType, null, currentState);

        stateHistory.add(change);
        trimHistory();

        mainHandler.post(() -> {
            for (StateListener l : listeners) {
                l.onLocalStateChanged(change);
            }
        });
    }

    private void notifyRemoteStateChange(@NonNull StateChange change) {
        mainHandler.post(() -> {
            for (StateListener l : listeners) {
                l.onRemoteStateChanged(change);

                if (change.isOnlineChange()) {
                    l.onDeviceOnline(change.getNewState());
                } else if (change.isOfflineChange()) {
                    l.onDeviceOffline(change.getAgentId());
                }
            }
        });
    }

    // === 系统监听器 ===

    private void initSystemListeners() {
        // 电池状态监听
        batteryReceiver = new BroadcastReceiver() {
            @Override
            public void onReceive(Context context, Intent intent) {
                collectBatteryState();
                if (currentState != null) {
                    updateBatteryState(
                            currentState.getBatteryLevel(),
                            currentState.isCharging(),
                            currentState.isPowerSaver()
                    );
                }
            }
        };

        IntentFilter batteryFilter = new IntentFilter();
        batteryFilter.addAction(Intent.ACTION_BATTERY_CHANGED);
        batteryFilter.addAction(Intent.ACTION_BATTERY_LOW);
        batteryFilter.addAction(Intent.ACTION_BATTERY_OKAY);
        context.registerReceiver(batteryReceiver, batteryFilter);

        // 网络状态监听
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            ConnectivityManager cm = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
            if (cm != null) {
                networkCallback = new ConnectivityManager.NetworkCallback() {
                    @Override
                    public void onAvailable(Network network) {
                        collectNetworkState();
                        if (currentState != null) {
                            updateNetworkState(
                                    currentState.getNetworkType(),
                                    currentState.getNetworkStrength(),
                                    currentState.isRoaming()
                            );
                        }
                    }

                    @Override
                    public void onLost(Network network) {
                        if (currentState != null) {
                            updateNetworkState("offline", 0, false);
                        }
                    }

                    @Override
                    public void onCapabilitiesChanged(Network network, NetworkCapabilities caps) {
                        collectNetworkState();
                        if (currentState != null) {
                            updateNetworkState(
                                    currentState.getNetworkType(),
                                    currentState.getNetworkStrength(),
                                    currentState.isRoaming()
                            );
                        }
                    }
                };

                cm.registerDefaultNetworkCallback(networkCallback);
            }
        } else {
            networkReceiver = new BroadcastReceiver() {
                @Override
                public void onReceive(Context context, Intent intent) {
                    collectNetworkState();
                    if (currentState != null) {
                        updateNetworkState(
                                currentState.getNetworkType(),
                                currentState.getNetworkStrength(),
                                currentState.isRoaming()
                        );
                    }
                }
            };

            IntentFilter networkFilter = new IntentFilter(ConnectivityManager.CONNECTIVITY_ACTION);
            context.registerReceiver(networkReceiver, networkFilter);
        }
    }

    /**
     * 清理系统监听器
     */
    public void cleanup() {
        try {
            if (batteryReceiver != null) {
                context.unregisterReceiver(batteryReceiver);
            }
            if (networkReceiver != null) {
                context.unregisterReceiver(networkReceiver);
            }
            if (networkCallback != null) {
                ConnectivityManager cm = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
                if (cm != null) {
                    cm.unregisterNetworkCallback(networkCallback);
                }
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to cleanup receivers", e);
        }

        stopSync();
    }

    // === 辅助方法 ===

    private List<String> getDefaultCapabilities() {
        List<String> caps = new ArrayList<>();
        caps.add("ui_automation");
        caps.add("speech");
        caps.add("camera");
        caps.add("sensor");
        return caps;
    }

    private String generateStateMessageId() {
        return "state_" + System.currentTimeMillis() + "_" + agentId;
    }

    private void trimHistory() {
        while (stateHistory.size() > config.maxHistorySize) {
            stateHistory.remove(0);
        }
    }

    private void cleanDeviceStateCache() {
        // 移除最旧的离线设备状态
        long oldestTime = Long.MAX_VALUE;
        String oldestAgent = null;

        for (Map.Entry<String, DeviceState> entry : deviceStates.entrySet()) {
            DeviceState state = entry.getValue();
            if (!state.isOnline() && state.getUpdatedAt() < oldestTime) {
                oldestTime = state.getUpdatedAt();
                oldestAgent = entry.getKey();
            }
        }

        if (oldestAgent != null) {
            deviceStates.remove(oldestAgent);
        }
    }

    /**
     * 状态同步统计
     */
    @NonNull
    public StateSyncStats getStats() {
        StateSyncStats stats = new StateSyncStats();
        stats.currentDevice = agentId;
        stats.totalDevices = deviceStates.size() + 1;
        stats.onlineDevices = 0;

        for (DeviceState state : deviceStates.values()) {
            if (state.isOnline()) {
                stats.onlineDevices++;
            }
        }

        if (currentState != null && currentState.isOnline()) {
            stats.onlineDevices++;
        }

        stats.historySize = stateHistory.size();
        stats.stateVersion = stateVersion;
        stats.lastSyncTime = lastSyncTime;

        return stats;
    }

    /**
     * 状态同步统计信息
     */
    public static class StateSyncStats {
        public String currentDevice;
        public int totalDevices;
        public int onlineDevices;
        public int historySize;
        public long stateVersion;
        public long lastSyncTime;

        @NonNull
        @Override
        public String toString() {
            return "StateSyncStats{" +
                    "total=" + totalDevices +
                    ", online=" + onlineDevices +
                    ", history=" + historySize +
                    ", version=" + stateVersion +
                    '}';
        }
    }

    // TelephonyInfo 辅助类（用于网络类型判断）
    private static class TelephonyInfo {
        static final int NETWORK_TYPE_LTE = 13;
        static final int NETWORK_TYPE_HSPAP = 14;
        static final int NETWORK_TYPE_EHRPD = 15;
        static final int NETWORK_TYPE_HSDPA = 8;
        static final int NETWORK_TYPE_HSPA = 10;
        static final int NETWORK_TYPE_HSUPA = 9;
        static final int NETWORK_TYPE_UMTS = 3;
        static final int NETWORK_TYPE_EVDO_0 = 5;
        static final int NETWORK_TYPE_EVDO_A = 6;
    }
}