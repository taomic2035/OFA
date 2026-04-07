package com.ofa.agent.core;

import android.content.Context;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationOrchestrator;
import com.ofa.agent.intent.IntentEngine;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.social.SocialOrchestrator;
import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillRegistry;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.function.Consumer;

/**
 * Agent Mode Manager - manages different running modes and task routing.
 *
 * v2.1.0 运行模式（更新）：
 * 1. STANDALONE - 完全独立运行，不连接 Center（默认）
 * 2. SYNC - 定期与 Center 同步数据，本地优先
 *
 * 废弃的模式（向后兼容）：
 * - CONNECTED - 现在等同于 SYNC
 * - HYBRID - 现在等同于 SYNC
 *
 * 核心变化：
 * - Center 不再调度任务，Agent 完全自主
 * - Center 仅作为数据中心，提供同步服务
 * - Agent 主动拉取/推送数据
 */
public class AgentModeManager {

    private static final String TAG = "AgentModeManager";

    private final Context context;
    private final AgentProfile profile;

    // Current mode
    private AgentProfile.RunMode currentMode;
    private AgentProfile.AgentStatus currentStatus;

    // Components (can be null depending on mode)
    private final LocalExecutionEngine localEngine;
    private CenterConnection centerConnection;
    private PeerNetwork peerNetwork;

    // Status
    private boolean isInitialized = false;
    private boolean networkAvailable = false;

    // Listeners
    private final List<ModeChangeListener> modeChangeListeners = new ArrayList<>();
    private final List<StatusChangeListener> statusChangeListeners = new ArrayList<>();

    // Network callback
    private ConnectivityManager.NetworkCallback networkCallback;

    /**
     * Mode change listener
     */
    public interface ModeChangeListener {
        void onModeChanged(AgentProfile.RunMode oldMode, AgentProfile.RunMode newMode);
    }

    /**
     * Status change listener
     */
    public interface StatusChangeListener {
        void onStatusChanged(AgentProfile.AgentStatus oldStatus, AgentProfile.AgentStatus newStatus);
    }

    /**
     * Task result callback
     */
    public interface TaskCallback {
        void onSuccess(Object result);
        void onFailure(String error);
    }

    /**
     * Create mode manager
     */
    public AgentModeManager(@NonNull Context context,
                            @NonNull AgentProfile profile,
                            @Nullable UserMemoryManager memoryManager,
                            @Nullable AutomationOrchestrator automationOrchestrator,
                            @Nullable SocialOrchestrator socialOrchestrator) {
        this.context = context;
        this.profile = profile;
        this.currentMode = profile.getPreferredRunMode();
        this.currentStatus = AgentProfile.AgentStatus.OFFLINE;

        // Create local execution engine
        this.localEngine = new LocalExecutionEngine(context, memoryManager,
            automationOrchestrator, socialOrchestrator);
    }

    /**
     * Initialize the mode manager
     */
    public void initialize() {
        Log.i(TAG, "Initializing with mode: " + currentMode);

        // Initialize local engine first (always needed)
        localEngine.initialize();

        // Setup network monitoring
        setupNetworkMonitoring();

        // Initialize based on mode
        switch (currentMode) {
            case STANDALONE:
                initializeStandaloneMode();
                break;
            case SYNC:
                initializeSyncMode();
                break;
            case CONNECTED: // deprecated, same as SYNC
            case HYBRID:    // deprecated, same as SYNC
                Log.w(TAG, "Mode " + currentMode + " is deprecated, using SYNC mode");
                currentMode = AgentProfile.RunMode.SYNC;
                initializeSyncMode();
                break;
        }

        isInitialized = true;
        setStatus(AgentProfile.AgentStatus.ONLINE);

        Log.i(TAG, "Initialized successfully in " + currentMode + " mode");
    }

    /**
     * Initialize standalone mode - everything local
     */
    private void initializeStandaloneMode() {
        Log.i(TAG, "Initializing STANDALONE mode");
        // No Center connection needed
        // All components work locally
        centerConnection = null;
        peerNetwork = null;
    }

    /**
     * Initialize sync mode - local-first with Center data sync (v2.1.0)
     */
    private void initializeSyncMode() {
        Log.i(TAG, "Initializing SYNC mode - Center as data center");
        // Connect to Center for data sync (not task dispatch)
        centerConnection = new CenterConnection(context, profile);
        centerConnection.setConnectionListener(new CenterConnection.ConnectionListener() {
            @Override
            public void onConnected() {
                Log.i(TAG, "Center connected, enabling data sync");
                // 启动数据同步
                startDataSync();
            }

            @Override
            public void onDisconnected() {
                Log.i(TAG, "Center disconnected, continuing locally");
                stopDataSync();
            }

            @Override
            public void onError(String error) {
                Log.w(TAG, "Center connection error: " + error);
            }
        });
        centerConnection.connect();

        // Enable peer network for local collaboration
        if (profile.isAllowPeerCommunication()) {
            peerNetwork = new PeerNetwork(context, profile);
            peerNetwork.start();
        }
    }

    /**
     * Start data sync with Center
     */
    private void startDataSync() {
        // 数据同步由 IdentityManager 和 UserMemoryManager 负责
        // 这里只做状态通知
        Log.i(TAG, "Data sync started");
    }

    /**
     * Stop data sync with Center
     */
    private void stopDataSync() {
        Log.i(TAG, "Data sync stopped");
    }

    /**
     * Setup network monitoring
     */
    private void setupNetworkMonitoring() {
        ConnectivityManager cm = (ConnectivityManager) context
            .getSystemService(Context.CONNECTIVITY_SERVICE);

        networkCallback = new ConnectivityManager.NetworkCallback() {
            @Override
            public void onAvailable(@NonNull Network network) {
                networkAvailable = true;
                Log.d(TAG, "Network available");
                handleNetworkChange(true);
            }

            @Override
            public void onLost(@NonNull Network network) {
                networkAvailable = false;
                Log.d(TAG, "Network lost");
                handleNetworkChange(false);
            }
        };

        if (cm != null) {
            cm.registerDefaultNetworkCallback(networkCallback);

            // Check initial state
            Network activeNetwork = cm.getActiveNetwork();
            networkAvailable = activeNetwork != null;
        }
    }

    /**
     * Handle network availability change
     */
    private void handleNetworkChange(boolean available) {
        if (currentMode == AgentProfile.RunMode.HYBRID) {
            if (available && centerConnection != null && !centerConnection.isConnected()) {
                // Try to reconnect to Center
                centerConnection.connect();
            }
        }
    }

    /**
     * Switch running mode
     */
    public void switchMode(@NonNull AgentProfile.RunMode newMode) {
        if (newMode == currentMode) {
            Log.d(TAG, "Already in " + newMode + " mode");
            return;
        }

        Log.i(TAG, "Switching mode: " + currentMode + " → " + newMode);

        AgentProfile.RunMode oldMode = currentMode;

        // Cleanup old mode
        cleanupCurrentMode();

        // Switch mode
        currentMode = newMode;

        // Initialize new mode
        switch (newMode) {
            case STANDALONE:
                initializeStandaloneMode();
                break;
            case SYNC:
                initializeSyncMode();
                break;
            case CONNECTED: // deprecated
            case HYBRID:    // deprecated
                Log.w(TAG, "Mode " + newMode + " is deprecated, using SYNC mode");
                currentMode = AgentProfile.RunMode.SYNC;
                initializeSyncMode();
                break;
        }

        // Update profile
        profile.setPreferredRunMode(newMode);

        // Notify listeners
        for (ModeChangeListener listener : modeChangeListeners) {
            listener.onModeChanged(oldMode, newMode);
        }
    }

    /**
     * Cleanup current mode resources
     */
    private void cleanupCurrentMode() {
        if (centerConnection != null) {
            centerConnection.disconnect();
            centerConnection = null;
        }
        if (peerNetwork != null) {
            peerNetwork.stop();
            peerNetwork = null;
        }
    }

    /**
     * Execute task - always executes locally (v2.1.0)
     *
     * Center 不再调度任务，所有任务由 Agent 本地执行。
     * Center 仅用于数据同步。
     */
    @NonNull
    public CompletableFuture<TaskResult> executeTask(@NonNull TaskRequest request) {
        Log.i(TAG, "Executing task locally: " + request.taskId + " in mode: " + currentMode);

        setStatus(AgentProfile.AgentStatus.BUSY);

        CompletableFuture<TaskResult> future = executeLocally(request);

        // Reset status after completion
        future.whenComplete((result, error) -> setStatus(AgentProfile.AgentStatus.IDLE));

        return future;
    }

    /**
     * Execute locally
     */
    @NonNull
    private CompletableFuture<TaskResult> executeLocally(@NonNull TaskRequest request) {
        return localEngine.execute(request);
    }

    /**
     * Sync data with Center (v2.1.0)
     *
     * 同步身份、记忆、偏好到 Center
     */
    public void syncDataWithCenter() {
        if (centerConnection == null || !centerConnection.isConnected()) {
            Log.w(TAG, "Cannot sync data: Center not connected");
            return;
        }

        Log.i(TAG, "Syncing data with Center...");
        centerConnection.sync();
    }

    /**
     * Sync with Center - upload local state, download updates
     */
    public void syncWithCenter() {
        if (centerConnection == null || !centerConnection.isConnected()) {
            Log.w(TAG, "Cannot sync: Center not connected");
            return;
        }

        Log.i(TAG, "Syncing with Center...");
        centerConnection.sync();
    }

    /**
     * Get peer agents
     */
    @NonNull
    public List<AgentProfile> getPeerAgents() {
        if (peerNetwork == null) {
            return new ArrayList<>();
        }
        return peerNetwork.getDiscoveredAgents();
    }

    /**
     * Send message to peer agent
     */
    public boolean sendToPeer(@NonNull String peerId, @NonNull String message) {
        if (peerNetwork == null || !profile.isAllowPeerCommunication()) {
            Log.w(TAG, "Peer communication not available");
            return false;
        }
        return peerNetwork.send(peerId, message);
    }

    /**
     * Request task from peer
     */
    @Nullable
    public TaskResult requestFromPeer(@NonNull String peerId, @NonNull TaskRequest request) {
        if (peerNetwork == null) {
            return null;
        }
        return peerNetwork.requestTask(peerId, request);
    }

    // ===== Status Management =====

    /**
     * Set agent status
     */
    public void setStatus(@NonNull AgentProfile.AgentStatus status) {
        if (status == currentStatus) return;

        AgentProfile.AgentStatus oldStatus = currentStatus;
        currentStatus = status;
        profile.setStatus(status);

        Log.i(TAG, "Status changed: " + oldStatus + " → " + status);

        for (StatusChangeListener listener : statusChangeListeners) {
            listener.onStatusChanged(oldStatus, status);
        }

        // Report to Center if connected
        if (centerConnection != null && centerConnection.isConnected()) {
            centerConnection.reportStatus(status);
        }
    }

    // ===== Listeners =====

    public void addModeChangeListener(@NonNull ModeChangeListener listener) {
        modeChangeListeners.add(listener);
    }

    public void removeModeChangeListener(@NonNull ModeChangeListener listener) {
        modeChangeListeners.remove(listener);
    }

    public void addStatusChangeListener(@NonNull StatusChangeListener listener) {
        statusChangeListeners.add(listener);
    }

    public void removeStatusChangeListener(@NonNull StatusChangeListener listener) {
        statusChangeListeners.remove(listener);
    }

    // ===== Getters =====

    @NonNull
    public AgentProfile.RunMode getCurrentMode() {
        return currentMode;
    }

    @NonNull
    public AgentProfile.AgentStatus getCurrentStatus() {
        return currentStatus;
    }

    @NonNull
    public AgentProfile getProfile() {
        return profile;
    }

    public boolean isNetworkAvailable() {
        return networkAvailable;
    }

    public boolean isCenterConnected() {
        return centerConnection != null && centerConnection.isConnected();
    }

    public boolean canReachPeers() {
        return peerNetwork != null && peerNetwork.isActive();
    }

    /**
     * Get peer network instance
     */
    @Nullable
    public PeerNetwork getPeerNetwork() {
        return peerNetwork;
    }

    /**
     * Get status report
     */
    @NonNull
    public String getStatusReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== Agent Mode Manager ===\n");
        sb.append("Mode: ").append(currentMode).append("\n");
        sb.append("Status: ").append(currentStatus).append("\n");
        sb.append("Network: ").append(networkAvailable ? "Available" : "Unavailable").append("\n");
        sb.append("Center: ").append(isCenterConnected() ? "Connected" : "Disconnected").append("\n");
        sb.append("Peers: ").append(getPeerAgents().size()).append(" discovered\n");
        sb.append("\nProfile:\n").append(profile.toString());
        return sb.toString();
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        Log.i(TAG, "Shutting down...");

        cleanupCurrentMode();

        if (networkCallback != null) {
            ConnectivityManager cm = (ConnectivityManager) context
                .getSystemService(Context.CONNECTIVITY_SERVICE);
            if (cm != null) {
                cm.unregisterNetworkCallback(networkCallback);
            }
        }

        localEngine.shutdown();
        isInitialized = false;
        setStatus(AgentProfile.AgentStatus.OFFLINE);

        Log.i(TAG, "Shutdown complete");
    }
}