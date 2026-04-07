package com.ofa.agent.core;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationOrchestrator;
import com.ofa.agent.behavior.BehaviorCollector;
import com.ofa.agent.distributed.DistributedOrchestrator;
import com.ofa.agent.identity.IdentityManager;
import com.ofa.agent.identity.PersonalIdentity;
import com.ofa.agent.identity.DecisionContext;
import com.ofa.agent.intent.IntentEngine;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillRegistry;
import com.ofa.agent.social.SocialOrchestrator;
import com.ofa.agent.tool.ToolRegistry;

import org.json.JSONObject;

import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * OFA Android Agent - Unified entry point for Android agent.
 *
 * Supports three running modes:
 * 1. STANDALONE - Complete local execution, no network dependency
 * 2. CONNECTED - Connected to Center, receives remote tasks
 * 3. HYBRID - Local-first with cloud enhancement (recommended)
 *
 * Task sources:
 * - Local triggers (Intent, UI, Scheduled)
 * - Center assignments
 * - Peer requests
 *
 * Capabilities:
 * - Intent understanding
 * - Skill orchestration
 * - UI automation
 * - Social notifications
 * - Memory management
 * - Peer communication
 */
public class OFAAndroidAgent {

    private static final String TAG = "OFAAndroidAgent";

    // Singleton instance
    private static OFAAndroidAgent instance;

    // Core components
    private final Context context;
    private final AgentProfile profile;
    private final AgentModeManager modeManager;
    private final LocalExecutionEngine localEngine;

    // Subsystems
    private UserMemoryManager memoryManager;
    private IdentityManager identityManager;  // v2.0.0: Identity Manager
    private BehaviorCollector behaviorCollector;  // v2.4.0: Behavior Collector
    private AutomationOrchestrator automationOrchestrator;
    private SocialOrchestrator socialOrchestrator;
    private ToolRegistry toolRegistry;
    private DistributedOrchestrator distributedOrchestrator;

    // State
    private boolean initialized = false;
    private boolean enableDistributed = true;

    /**
     * Builder for OFAAndroidAgent
     */
    public static class Builder {
        private final Context context;
        private AgentProfile.RunMode runMode = AgentProfile.RunMode.SYNC; // v2.1.0: 默认 SYNC 模式
        private String centerAddress = "localhost";
        private int centerPort = 9090;
        private boolean enableAutomation = true;
        private boolean enableSocial = true;
        private boolean enablePeerNetwork = true;
        private boolean enableDistributed = true;  // Enable distributed agent features

        public Builder(@NonNull Context context) {
            this.context = context.getApplicationContext();
        }

        public Builder runMode(@NonNull AgentProfile.RunMode mode) {
            this.runMode = mode;
            return this;
        }

        public Builder center(@NonNull String address, int port) {
            this.centerAddress = address;
            this.centerPort = port;
            return this;
        }

        public Builder enableAutomation(boolean enable) {
            this.enableAutomation = enable;
            return this;
        }

        public Builder enableSocial(boolean enable) {
            this.enableSocial = enable;
            return this;
        }

        public Builder enablePeerNetwork(boolean enable) {
            this.enablePeerNetwork = enable;
            return this;
        }

        public Builder enableDistributed(boolean enable) {
            this.enableDistributed = enable;
            return this;
        }

        public OFAAndroidAgent build() {
            if (instance != null) {
                instance.shutdown();
            }
            instance = new OFAAndroidAgent(this);
            return instance;
        }
    }

    /**
     * Get singleton instance
     */
    @Nullable
    public static OFAAndroidAgent getInstance() {
        return instance;
    }

    private OFAAndroidAgent(Builder builder) {
        this.context = builder.context;
        this.enableDistributed = builder.enableDistributed;

        // Create profile
        this.profile = new AgentProfile.Builder()
            .type(AgentProfile.AgentType.MOBILE)
            .preferredRunMode(builder.runMode)
            .allowRemoteControl(true)
            .allowPeerCommunication(builder.enablePeerNetwork)
            .addCapability(AgentProfile.Capability.UI_AUTOMATION)
            .addCapability(AgentProfile.Capability.INTENT_UNDERSTANDING)
            .addCapability(AgentProfile.Capability.MEMORY_SYSTEM)
            .addCapability(AgentProfile.Capability.SKILL_ORCHESTRATION)
            .build();

        // Initialize memory
        this.memoryManager = new UserMemoryManager(context);

        // v2.0.0: Initialize identity manager
        this.identityManager = new IdentityManager(context);
        this.identityManager.initialize();

        // v2.4.0: Initialize behavior collector
        this.behaviorCollector = new BehaviorCollector(context, identityManager);

        // Initialize automation if enabled
        if (builder.enableAutomation) {
            this.automationOrchestrator = new AutomationOrchestrator(context);
        }

        // Initialize social if enabled
        if (builder.enableSocial) {
            this.socialOrchestrator = new SocialOrchestrator(
                context, null, memoryManager);
        }

        // Create local execution engine
        this.localEngine = new LocalExecutionEngine(
            context, memoryManager, automationOrchestrator, socialOrchestrator);

        // Create mode manager
        this.modeManager = new AgentModeManager(
            context, profile, memoryManager, automationOrchestrator, socialOrchestrator);

        // Initialize distributed orchestrator if enabled
        if (builder.enableDistributed && builder.enablePeerNetwork) {
            PeerNetwork peerNetwork = modeManager.getPeerNetwork();
            if (peerNetwork != null) {
                this.distributedOrchestrator = new DistributedOrchestrator(
                    context, profile, this, peerNetwork);
            }
        }

        Log.i(TAG, "OFA Android Agent created with mode: " + builder.runMode +
              ", distributed: " + builder.enableDistributed);
    }

    /**
     * Initialize the agent
     */
    public void initialize() {
        if (initialized) {
            Log.w(TAG, "Already initialized");
            return;
        }

        Log.i(TAG, "Initializing OFA Android Agent...");

        // Initialize automation
        if (automationOrchestrator != null) {
            automationOrchestrator.initialize();
        }

        // Initialize mode manager
        modeManager.initialize();

        // v2.4.0: Enable behavior collection
        if (behaviorCollector != null) {
            behaviorCollector.enable();
        }

        // Initialize distributed orchestrator
        if (distributedOrchestrator != null) {
            distributedOrchestrator.initialize();
            Log.i(TAG, "Distributed orchestrator initialized");
        }

        initialized = true;
        Log.i(TAG, "OFA Android Agent initialized");
    }

    // ===== Running Mode =====

    /**
     * Get current running mode
     */
    @NonNull
    public AgentProfile.RunMode getRunMode() {
        return modeManager.getCurrentMode();
    }

    /**
     * Switch running mode
     */
    public void switchMode(@NonNull AgentProfile.RunMode mode) {
        modeManager.switchMode(mode);
    }

    /**
     * Check if Center connected
     */
    public boolean isCenterConnected() {
        return modeManager.isCenterConnected();
    }

    /**
     * Check if network available
     */
    public boolean isNetworkAvailable() {
        return modeManager.isNetworkAvailable();
    }

    // ===== Task Execution =====

    /**
     * Execute natural language input
     */
    @NonNull
    public CompletableFuture<TaskResult> execute(@NonNull String input) {
        return modeManager.executeTask(TaskRequest.naturalLanguage(input));
    }

    /**
     * Execute skill
     */
    @NonNull
    public CompletableFuture<TaskResult> executeSkill(@NonNull String skillId,
                                                       @NonNull Map<String, String> inputs) {
        return modeManager.executeTask(TaskRequest.skill(skillId, inputs));
    }

    /**
     * Execute automation
     */
    @NonNull
    public CompletableFuture<TaskResult> executeAutomation(@NonNull String operation,
                                                            @NonNull Map<String, String> params) {
        return modeManager.executeTask(TaskRequest.automation(operation, params));
    }

    /**
     * Send social notification
     */
    @NonNull
    public CompletableFuture<TaskResult> sendNotification(@NonNull String message,
                                                          @Nullable String recipient,
                                                          @Nullable String phone) {
        return modeManager.executeTask(TaskRequest.social(message, recipient, phone));
    }

    /**
     * Execute raw task request
     */
    @NonNull
    public CompletableFuture<TaskResult> executeTask(@NonNull TaskRequest request) {
        return modeManager.executeTask(request);
    }

    // ===== Memory =====

    /**
     * Get memory manager
     */
    @Nullable
    public UserMemoryManager getMemoryManager() {
        return memoryManager;
    }

    /**
     * Remember value
     */
    public void remember(@NonNull String key, @NonNull String value) {
        if (memoryManager != null) {
            memoryManager.set(key, value);
        }
    }

    /**
     * Recall value
     */
    @Nullable
    public String recall(@NonNull String key) {
        if (memoryManager != null) {
            return memoryManager.get(key);
        }
        return null;
    }

    // ===== Identity (v2.0.0) =====

    /**
     * Get identity manager
     */
    @Nullable
    public IdentityManager getIdentityManager() {
        return identityManager;
    }

    /**
     * Get current identity
     */
    @Nullable
    public PersonalIdentity getIdentity() {
        if (identityManager != null) {
            return identityManager.getIdentity();
        }
        return null;
    }

    /**
     * Get decision context for AI
     */
    @Nullable
    public DecisionContext getDecisionContext() {
        if (identityManager != null) {
            return identityManager.getDecisionContext();
        }
        return null;
    }

    /**
     * Generate AI prompt context
     */
    @NonNull
    public String generatePromptContext() {
        if (identityManager != null) {
            return identityManager.generatePromptContext();
        }
        return "";
    }

    /**
     * Sync identity with Center
     */
    public void syncIdentity() {
        if (identityManager != null && modeManager.isCenterConnected()) {
            identityManager.syncToCenter();
        }
    }

    // ===== Behavior Collection (v2.4.0) =====

    /**
     * Get behavior collector
     */
    @Nullable
    public BehaviorCollector getBehaviorCollector() {
        return behaviorCollector;
    }

    /**
     * Observe decision behavior
     */
    public void observeDecision(@NonNull String decisionType, @NonNull Map<String, Object> details) {
        if (behaviorCollector != null) {
            behaviorCollector.observeDecision(decisionType, details);
        }
    }

    /**
     * Observe interaction behavior
     */
    public void observeInteraction(@NonNull String interactionType, @NonNull Map<String, Object> details) {
        if (behaviorCollector != null) {
            behaviorCollector.observeInteraction(interactionType, details);
        }
    }

    /**
     * Observe preference behavior
     */
    public void observePreference(@NonNull String preferenceType, @NonNull Map<String, Object> details) {
        if (behaviorCollector != null) {
            behaviorCollector.observePreference(preferenceType, details);
        }
    }

    /**
     * Observe activity behavior
     */
    public void observeActivity(@NonNull String activityType, @NonNull Map<String, Object> details) {
        if (behaviorCollector != null) {
            behaviorCollector.observeActivity(activityType, details);
        }
    }

    /**
     * Record purchase decision (convenience method)
     */
    public void recordPurchase(@NonNull String item, double price, boolean isImpulse) {
        if (behaviorCollector != null) {
            behaviorCollector.recordPurchase(item, price, isImpulse);
        }
    }

    /**
     * Record social interaction (convenience method)
     */
    public void recordSocialInteraction(@NonNull String type, int participantCount, boolean usedEmoji) {
        if (behaviorCollector != null) {
            behaviorCollector.recordSocialInteraction(type, participantCount, usedEmoji);
        }
    }

    // ===== Peer Communication =====

    /**
     * Get discovered peers
     */
    @NonNull
    public List<AgentProfile> getPeers() {
        return modeManager.getPeerAgents();
    }

    /**
     * Send message to peer
     */
    public boolean sendToPeer(@NonNull String peerId, @NonNull String message) {
        return modeManager.sendToPeer(peerId, message);
    }

    /**
     * Request task from peer
     */
    @Nullable
    public TaskResult requestFromPeer(@NonNull String peerId, @NonNull TaskRequest request) {
        return modeManager.requestFromPeer(peerId, request);
    }

    // ===== Profile =====

    /**
     * Get agent profile
     */
    @NonNull
    public AgentProfile getProfile() {
        return profile;
    }

    /**
     * Get agent ID
     */
    @NonNull
    public String getAgentId() {
        return profile.getAgentId();
    }

    /**
     * Get agent status
     */
    @NonNull
    public AgentProfile.AgentStatus getStatus() {
        return modeManager.getCurrentStatus();
    }

    // ===== Listeners =====

    /**
     * Add mode change listener
     */
    public void addModeChangeListener(@NonNull AgentModeManager.ModeChangeListener listener) {
        modeManager.addModeChangeListener(listener);
    }

    /**
     * Add status change listener
     */
    public void addStatusChangeListener(@NonNull AgentModeManager.StatusChangeListener listener) {
        modeManager.addStatusChangeListener(listener);
    }

    // ===== Subsystems =====

    /**
     * Get automation orchestrator
     */
    @Nullable
    public AutomationOrchestrator getAutomationOrchestrator() {
        return automationOrchestrator;
    }

    /**
     * Get social orchestrator
     */
    @Nullable
    public SocialOrchestrator getSocialOrchestrator() {
        return socialOrchestrator;
    }

    /**
     * Get distributed orchestrator
     */
    @Nullable
    public DistributedOrchestrator getDistributedOrchestrator() {
        return distributedOrchestrator;
    }

    /**
     * Get peer network
     */
    @Nullable
    public PeerNetwork getPeerNetwork() {
        return modeManager.getPeerNetwork();
    }

    // ===== Status =====

    /**
     * Get comprehensive status report
     */
    @NonNull
    public String getStatusReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== OFA Android Agent Status ===\n\n");
        sb.append("Agent ID: ").append(profile.getAgentId()).append("\n");
        sb.append("Name: ").append(profile.getName()).append("\n");
        sb.append("Type: ").append(profile.getType()).append("\n");
        sb.append("Status: ").append(getStatus()).append("\n\n");

        sb.append(modeManager.getStatusReport()).append("\n");

        sb.append("Capabilities:\n");
        for (AgentProfile.Capability cap : profile.getCapabilities()) {
            sb.append("  - ").append(cap.name).append(" (").append(cap.id).append(")\n");
        }

        // v2.0.0: Add identity info
        if (identityManager != null && identityManager.hasIdentity()) {
            sb.append("\nIdentity:\n");
            sb.append("  ID: ").append(identityManager.getIdentityId()).append("\n");
            PersonalIdentity identity = identityManager.getIdentity();
            if (identity != null) {
                sb.append("  Name: ").append(identity.getName()).append("\n");
                sb.append("  Version: ").append(identity.getVersion()).append("\n");
            }
        }

        return sb.toString();
    }

    /**
     * Shutdown the agent
     */
    public void shutdown() {
        if (!initialized) return;

        Log.i(TAG, "Shutting down OFA Android Agent...");

        // Shutdown distributed orchestrator first
        if (distributedOrchestrator != null) {
            distributedOrchestrator.shutdown();
        }

        modeManager.shutdown();

        if (automationOrchestrator != null) {
            automationOrchestrator.shutdown();
        }

        if (socialOrchestrator != null) {
            socialOrchestrator.shutdown();
        }

        if (memoryManager != null) {
            memoryManager.close();
        }

        // v2.0.0: Shutdown identity manager
        if (identityManager != null) {
            identityManager.shutdown();
        }

        // v2.4.0: Shutdown behavior collector
        if (behaviorCollector != null) {
            behaviorCollector.shutdown();
        }

        initialized = false;
        instance = null;

        Log.i(TAG, "OFA Android Agent shutdown complete");
    }

    /**
     * Check if initialized
     */
    public boolean isInitialized() {
        return initialized;
    }
}