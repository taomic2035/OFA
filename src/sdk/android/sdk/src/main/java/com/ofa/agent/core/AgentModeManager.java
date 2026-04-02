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
 * Three running modes:
 * 1. STANDALONE - Complete local execution, no Center connection
 * 2. CONNECTED - Connected to Center, receives remote tasks
 * 3. HYBRID - Local-first with cloud enhancement
 *
 * Task sources in HYBRID mode:
 * - Local triggers (Intent, UI, Scheduled)
 * - Center assignments
 * - Peer requests
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
            case CONNECTED:
                initializeConnectedMode();
                break;
            case HYBRID:
                initializeHybridMode();
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
     * Initialize connected mode - connected to Center
     */
    private void initializeConnectedMode() {
        Log.i(TAG, "Initializing CONNECTED mode");
        // Connect to Center
        centerConnection = new CenterConnection(context, profile);
        centerConnection.connect();

        // Optionally enable peer network
        if (profile.isAllowPeerCommunication()) {
            peerNetwork = new PeerNetwork(context, profile);
            peerNetwork.start();
        }
    }

    /**
     * Initialize hybrid mode - local-first with cloud enhancement
     */
    private void initializeHybridMode() {
        Log.i(TAG, "Initializing HYBRID mode");
        // Try to connect to Center, but don't fail if unavailable
        centerConnection = new CenterConnection(context, profile);
        centerConnection.setConnectionListener(new CenterConnection.ConnectionListener() {
            @Override
            public void onConnected() {
                Log.i(TAG, "Center connected in hybrid mode");
                syncWithCenter();
            }

            @Override
            public void onDisconnected() {
                Log.i(TAG, "Center disconnected, continuing locally");
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
            case CONNECTED:
                initializeConnectedMode();
                break;
            case HYBRID:
                initializeHybridMode();
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
     * Execute task - routes to appropriate handler based on mode
     */
    @NonNull
    public CompletableFuture<TaskResult> executeTask(@NonNull TaskRequest request) {
        Log.i(TAG, "Executing task: " + request.taskId + " in mode: " + currentMode);

        setStatus(AgentProfile.AgentStatus.BUSY);

        CompletableFuture<TaskResult> future = new CompletableFuture<>();

        try {
            switch (currentMode) {
                case STANDALONE:
                    // Always execute locally
                    future = executeLocally(request);
                    break;

                case CONNECTED:
                    // Prefer Center if available
                    if (centerConnection != null && centerConnection.isConnected()) {
                        future = executeViaCenter(request);
                    } else {
                        future.completeExceptionally(new Exception("Center not connected"));
                    }
                    break;

                case HYBRID:
                    // Decide based on task requirements and availability
                    future = executeHybrid(request);
                    break;
            }
        } catch (Exception e) {
            future.completeExceptionally(e);
        }

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
     * Execute via Center
     */
    @NonNull
    private CompletableFuture<TaskResult> executeViaCenter(@NonNull TaskRequest request) {
        return centerConnection.executeTask(request);
    }

    /**
     * Execute in hybrid mode - intelligent routing
     */
    @NonNull
    private CompletableFuture<TaskResult> executeHybrid(@NonNull TaskRequest request) {
        // Decision logic:
        // 1. Can execute offline? → Local if no network
        // 2. Needs cloud LLM? → Center if available
        // 3. Local skills? → Local first
        // 4. Complex coordination? → Center

        boolean canBeOffline = localEngine.canExecute(request);
        boolean needsCloud = request.requiresCloudCapability();
        boolean centerAvailable = centerConnection != null && centerConnection.isConnected();

        if (!networkAvailable && canBeOffline) {
            Log.d(TAG, "Hybrid: executing locally (no network)");
            return executeLocally(request);
        }

        if (needsCloud && centerAvailable) {
            Log.d(TAG, "Hybrid: executing via Center (needs cloud)");
            return executeViaCenter(request);
        }

        if (canBeOffline) {
            Log.d(TAG, "Hybrid: executing locally (local capability)");
            return executeLocally(request);
        }

        if (centerAvailable) {
            Log.d(TAG, "Hybrid: executing via Center (fallback)");
            return executeViaCenter(request);
        }

        // Cannot execute
        CompletableFuture<TaskResult> future = new CompletableFuture<>();
        future.completeExceptionally(new Exception("No execution path available"));
        return future;
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