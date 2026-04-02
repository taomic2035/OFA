package com.ofa.agent;

import android.content.Context;
import android.os.BatteryManager;
import android.os.Build;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.work.BackoffPolicy;
import androidx.work.Constraints;
import androidx.work.ExistingPeriodicWorkPolicy;
import androidx.work.NetworkType;
import androidx.work.PeriodicWorkRequest;
import androidx.work.WorkManager;

import com.ofa.agent.ai.AIAgentInterface;
import com.ofa.agent.ai.ToolCallingAdapter;
import com.ofa.agent.constraint.ConstraintChecker;
import com.ofa.agent.grpc.AgentGrpc;
import com.ofa.agent.grpc.AgentOuterClass;
import com.ofa.agent.llm.LLMConfig;
import com.ofa.agent.llm.LLMProvider;
import com.ofa.agent.llm.cloud.CloudLLMProvider;
import com.ofa.agent.llm.local.LocalLLMProvider;
import com.ofa.agent.llm.orchestrator.LLMOrchestrator;
import com.ofa.agent.llm.tool.LLMChatTool;
import com.ofa.agent.mcp.MCPServer;
import com.ofa.agent.mcp.MCPServerImpl;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.offline.OfflineLevel;
import com.ofa.agent.offline.OfflineManager;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.tool.BuiltInTools;
import com.ofa.agent.tool.PermissionManager;
import com.ofa.agent.tool.ToolRegistry;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;

/**
 * OFA Agent for Android devices.
 * Connects to OFA Center and executes skills locally.
 */
public class OFAAgent {
    private static final String TAG = "OFAAgent";

    // Agent configuration
    private final String agentId;
    private final String agentName;
    private final AgentType type;
    private final String centerAddress;
    private final int centerPort;

    // gRPC
    private ManagedChannel channel;
    private AgentGrpc.AgentStub asyncStub;
    private StreamObserver<AgentOuterClass.AgentMessage> messageStream;

    // Context
    private final Context context;

    // Skills
    private final Map<String, SkillExecutor> skills = new ConcurrentHashMap<>();

    // MCP & Tools
    private ToolRegistry toolRegistry;
    private MCPServer mcpServer;
    private ConstraintChecker constraintChecker;
    private PermissionManager permissionManager;
    private OfflineManager offlineManager;
    private ToolCallingAdapter aiAdapter;

    // LLM
    private LLMProvider llmProvider;
    private LLMOrchestrator llmOrchestrator;

    // State
    private volatile boolean connected = false;
    private final Handler handler = new Handler(Looper.getMainLooper());
    private Runnable heartbeatRunnable;

    // Listeners
    private ConnectionListener connectionListener;
    private TaskListener taskListener;

    /**
     * Agent type enum
     */
    public enum AgentType {
        FULL(1),
        MOBILE(2),
        LITE(3),
        IOT(4),
        EDGE(5);

        private final int value;

        AgentType(int value) {
            this.value = value;
        }

        public int getValue() {
            return value;
        }
    }

    /**
     * Builder pattern for OFAAgent
     */
    public static class Builder {
        private Context context;
        private String agentId;
        private String agentName;
        private AgentType type = AgentType.MOBILE;
        private String centerAddress = "localhost";
        private int centerPort = 9090;
        private OfflineLevel offlineLevel = OfflineLevel.L4;
        private boolean enableTools = true;

        // LLM 配置
        private LLMProvider llmProvider;
        private LLMProvider fallbackLLMProvider;
        private LLMConfig cloudLLMConfig;
        private LLMConfig localLLMConfig;
        private boolean autoLLMFailover = true;

        public Builder(Context context) {
            this.context = context.getApplicationContext();
        }

        public Builder agentId(String agentId) {
            this.agentId = agentId;
            return this;
        }

        public Builder agentName(String agentName) {
            this.agentName = agentName;
            return this;
        }

        public Builder type(AgentType type) {
            this.type = type;
            return this;
        }

        public Builder centerAddress(String address) {
            this.centerAddress = address;
            return this;
        }

        public Builder centerPort(int port) {
            this.centerPort = port;
            return this;
        }

        public Builder offlineLevel(OfflineLevel level) {
            this.offlineLevel = level;
            return this;
        }

        public Builder enableTools(boolean enable) {
            this.enableTools = enable;
            return this;
        }

        // ===== LLM 配置方法 =====

        /**
         * 设置云端 LLM 提供者
         */
        public Builder llmProvider(LLMProvider provider) {
            this.llmProvider = provider;
            return this;
        }

        /**
         * 设置备用 LLM 提供者 (通常用于离线)
         */
        public Builder fallbackLLMProvider(LLMProvider provider) {
            this.fallbackLLMProvider = provider;
            return this;
        }

        /**
         * 配置云端 LLM (便捷方法)
         */
        public Builder cloudLLM(String endpoint, String apiKey, String model) {
            this.cloudLLMConfig = new LLMConfig.Builder()
                    .providerType(LLMProvider.ProviderType.CLOUD)
                    .endpoint(endpoint)
                    .apiKey(apiKey)
                    .model(model)
                    .build();
            return this;
        }

        /**
         * 配置云端 LLM (完整配置)
         */
        public Builder cloudLLM(LLMConfig config) {
            this.cloudLLMConfig = config;
            return this;
        }

        /**
         * 配置本地 LLM (便捷方法)
         */
        public Builder localLLM(String modelPath) {
            this.localLLMConfig = new LLMConfig.Builder()
                    .providerType(LLMProvider.ProviderType.LOCAL)
                    .modelPath(modelPath)
                    .build();
            return this;
        }

        /**
         * 配置本地 LLM (完整配置)
         */
        public Builder localLLM(LLMConfig config) {
            this.localLLMConfig = config;
            return this;
        }

        /**
         * 启用/禁用自动 LLM 故障转移
         */
        public Builder autoLLMFailover(boolean enable) {
            this.autoLLMFailover = enable;
            return this;
        }

        public OFAAgent build() {
            if (agentId == null) {
                agentId = UUID.randomUUID().toString();
            }
            if (agentName == null) {
                agentName = Build.MODEL;
            }
            return new OFAAgent(this);
        }
    }

    private OFAAgent(Builder builder) {
        this.context = builder.context;
        this.agentId = builder.agentId;
        this.agentName = builder.agentName;
        this.type = builder.type;
        this.centerAddress = builder.centerAddress;
        this.centerPort = builder.centerPort;

        // Initialize heartbeat runnable
        heartbeatRunnable = () -> {
            if (connected) {
                sendHeartbeat();
                handler.postDelayed(heartbeatRunnable, 30000); // 30s interval
            }
        };

        // Register built-in skills
        registerBuiltInSkills();

        // Initialize LLM
        initializeLLM(builder);

        // Initialize MCP & Tools if enabled
        if (builder.enableTools) {
            initializeTools(builder.offlineLevel);
        }
    }

    /**
     * Initialize LLM providers
     */
    private void initializeLLM(Builder builder) {
        // 如果直接提供了提供者
        if (builder.llmProvider != null) {
            this.llmProvider = builder.llmProvider;
            if (builder.fallbackLLMProvider != null) {
                llmOrchestrator = new LLMOrchestrator();
                llmOrchestrator.setPrimaryProvider(builder.llmProvider);
                llmOrchestrator.setFallbackProvider(builder.fallbackLLMProvider);
                llmOrchestrator.setAutoFailover(builder.autoLLMFailover);
                this.llmProvider = llmOrchestrator;
            }
            return;
        }

        // 从配置创建
        boolean hasCloud = builder.cloudLLMConfig != null;
        boolean hasLocal = builder.localLLMConfig != null;

        if (hasCloud || hasLocal) {
            llmOrchestrator = new LLMOrchestrator();
            llmOrchestrator.setAutoFailover(builder.autoLLMFailover);

            if (hasCloud) {
                CloudLLMProvider cloudProvider = new CloudLLMProvider(builder.cloudLLMConfig);
                llmOrchestrator.setPrimaryProvider(cloudProvider);
            }

            if (hasLocal) {
                LocalLLMProvider localProvider = new LocalLLMProvider(context, builder.localLLMConfig);
                localProvider.initialize();
                llmOrchestrator.setFallbackProvider(localProvider);
            }

            this.llmProvider = llmOrchestrator;
            Log.i(TAG, "LLM initialized with orchestrator");
        }
    }

    /**
     * Initialize MCP Server and Tools
     */
    private void initializeTools(@NonNull OfflineLevel offlineLevel) {
        // Create tool registry
        toolRegistry = new ToolRegistry(context);

        // Create permission manager
        permissionManager = new PermissionManager(context);

        // Create constraint checker
        constraintChecker = new ConstraintChecker();

        // Create offline manager
        offlineManager = new OfflineManager(context, offlineLevel);
        offlineManager.start();

        // Register built-in tools
        BuiltInTools.registerAll(context, toolRegistry);

        // Register LLM tool if available
        if (llmProvider != null) {
            toolRegistry.register(
                    new com.ofa.agent.mcp.ToolDefinition(
                            "llm.chat",
                            "Chat with AI language model",
                            "{}"
                    ),
                    new LLMChatTool(llmProvider)
            );
            Log.i(TAG, "LLM tool registered");
        }

        // Create MCP Server
        mcpServer = new MCPServerImpl(context, toolRegistry, permissionManager, constraintChecker, offlineManager);
        ((MCPServerImpl) mcpServer).start();

        // Create AI adapter
        aiAdapter = new ToolCallingAdapter(mcpServer);

        Log.i(TAG, "MCP Server initialized with " + toolRegistry.getCount() + " tools");
    }

    /**
     * Register a skill executor
     */
    public void registerSkill(@NonNull String skillId, @NonNull SkillExecutor executor) {
        skills.put(skillId, executor);
        Log.i(TAG, "Registered skill: " + skillId);
    }

    /**
     * Unregister a skill
     */
    public void unregisterSkill(@NonNull String skillId) {
        skills.remove(skillId);
        Log.i(TAG, "Unregistered skill: " + skillId);
    }

    /**
     * Get registered skills
     */
    public List<String> getRegisteredSkills() {
        return new ArrayList<>(skills.keySet());
    }

    /**
     * Connect to OFA Center
     */
    public void connect() {
        new Thread(() -> {
            try {
                // Create gRPC channel
                channel = ManagedChannelBuilder
                        .forAddress(centerAddress, centerPort)
                        .usePlaintext()
                        .build();

                asyncStub = AgentGrpc.newStub(channel);

                // Start bidirectional stream
                messageStream = asyncStub.connect(new StreamObserver<AgentOuterClass.CenterMessage>() {
                    @Override
                    public void onNext(AgentOuterClass.CenterMessage value) {
                        handleCenterMessage(value);
                    }

                    @Override
                    public void onError(Throwable t) {
                        Log.e(TAG, "Stream error", t);
                        connected = false;
                        notifyConnectionError(t.getMessage());
                        scheduleReconnect();
                    }

                    @Override
                    public void onCompleted() {
                        Log.i(TAG, "Stream completed");
                        connected = false;
                        notifyDisconnected();
                    }
                });

                // Send registration
                sendRegistration();

                connected = true;
                notifyConnected();

                // Start heartbeat
                handler.post(heartbeatRunnable);

                Log.i(TAG, "Connected to Center: " + centerAddress + ":" + centerPort);

            } catch (Exception e) {
                Log.e(TAG, "Connection failed", e);
                notifyConnectionError(e.getMessage());
                scheduleReconnect();
            }
        }).start();
    }

    /**
     * Disconnect from Center
     */
    public void disconnect() {
        connected = false;
        handler.removeCallbacks(heartbeatRunnable);

        if (messageStream != null) {
            messageStream.onCompleted();
        }

        if (channel != null) {
            channel.shutdown();
        }

        notifyDisconnected();
        Log.i(TAG, "Disconnected from Center");
    }

    /**
     * Check if connected
     */
    public boolean isConnected() {
        return connected;
    }

    /**
     * Get agent ID
     */
    public String getAgentId() {
        return agentId;
    }

    // ===== MCP & Tool Methods =====

    /**
     * Get MCP Server instance
     */
    @Nullable
    public MCPServer getMCPServer() {
        return mcpServer;
    }

    /**
     * Get Tool Registry
     */
    @Nullable
    public ToolRegistry getToolRegistry() {
        return toolRegistry;
    }

    /**
     * Get AI Agent Interface for tool calling
     */
    @Nullable
    public AIAgentInterface getAIAgentInterface() {
        return aiAdapter;
    }

    /**
     * Get LLM Provider
     */
    @Nullable
    public LLMProvider getLLMProvider() {
        return llmProvider;
    }

    /**
     * Check if LLM is available
     */
    public boolean hasLLM() {
        return llmProvider != null && llmProvider.isAvailable();
    }

    /**
     * Chat with LLM directly
     */
    @Nullable
    public LLMProvider getLLM() {
        return llmProvider;
    }

    /**
     * Call a tool by name
     */
    @NonNull
    public ToolResult callTool(@NonNull String toolName, @NonNull Map<String, Object> args) {
        if (mcpServer == null) {
            return new ToolResult(toolName, "MCP Server not initialized");
        }
        return mcpServer.callTool(toolName, args);
    }

    /**
     * Get list of available tools
     */
    @NonNull
    public List<ToolDefinition> getAvailableTools() {
        if (mcpServer == null) {
            return new ArrayList<>();
        }
        return mcpServer.listTools();
    }

    /**
     * Get Offline Manager
     */
    @Nullable
    public OfflineManager getOfflineManager() {
        return offlineManager;
    }

    /**
     * Set offline mode
     */
    public void setOfflineMode(boolean offline) {
        if (offlineManager != null) {
            offlineManager.setOfflineMode(offline);
        }
        if (constraintChecker != null) {
            constraintChecker.setOfflineMode(offline);
        }
    }

    /**
     * Check if in offline mode
     */
    public boolean isOfflineMode() {
        if (offlineManager != null) {
            return offlineManager.isOfflineMode();
        }
        return false;
    }

    /**
     * Shutdown the agent
     */
    public void shutdown() {
        disconnect();

        if (mcpServer != null) {
            mcpServer.shutdown();
        }

        if (offlineManager != null) {
            offlineManager.stop();
        }

        if (aiAdapter != null) {
            aiAdapter.shutdown();
        }

        if (llmProvider != null) {
            llmProvider.shutdown();
        }

        Log.i(TAG, "Agent shutdown complete");
    }

    // ===== Private Methods =====

    private void registerBuiltInSkills() {
        // Built-in skills can be added here
        // e.g., registerSkill("echo", new EchoSkill());
    }

    private void sendRegistration() {
        AgentOuterClass.RegisterRequest register = AgentOuterClass.RegisterRequest.newBuilder()
                .setAgentId(agentId)
                .setName(agentName)
                .setTypeValue(type.getValue())
                .setDeviceInfo(getDeviceInfo())
                .addAllCapabilities(getCapabilities())
                .build();

        AgentOuterClass.AgentMessage message = AgentOuterClass.AgentMessage.newBuilder()
                .setMsgId(UUID.randomUUID().toString())
                .setRegister(register)
                .build();

        messageStream.onNext(message);
    }

    private void sendHeartbeat() {
        AgentOuterClass.HeartbeatRequest heartbeat = AgentOuterClass.HeartbeatRequest.newBuilder()
                .setAgentId(agentId)
                .setStatusValue(1) // ONLINE
                .setResources(getResourceUsage())
                .build();

        AgentOuterClass.AgentMessage message = AgentOuterClass.AgentMessage.newBuilder()
                .setMsgId(UUID.randomUUID().toString())
                .setHeartbeat(heartbeat)
                .build();

        messageStream.onNext(message);
    }

    private void handleCenterMessage(AgentOuterClass.CenterMessage message) {
        switch (message.getPayloadCase()) {
            case TASK:
                handleTaskAssignment(message.getTask());
                break;
            case CONFIG:
                handleConfigUpdate(message.getConfig());
                break;
            case BROADCAST:
                handleBroadcast(message.getBroadcast());
                break;
            default:
                Log.d(TAG, "Unknown message type: " + message.getPayloadCase());
        }
    }

    private void handleTaskAssignment(AgentOuterClass.TaskAssignment task) {
        Log.i(TAG, "Received task: " + task.getTaskId());

        new Thread(() -> {
            String skillId = task.getSkillId();
            SkillExecutor executor = skills.get(skillId);

            AgentOuterClass.TaskResult.Builder resultBuilder = AgentOuterClass.TaskResult.newBuilder()
                    .setTaskId(task.getTaskId());

            try {
                if (executor == null) {
                    throw new Exception("Skill not found: " + skillId);
                }

                byte[] output = executor.execute(task.getInput().toByteArray());
                resultBuilder
                        .setStatusValue(3) // COMPLETED
                        .setOutput(com.google.protobuf.ByteString.copyFrom(output));

            } catch (Exception e) {
                Log.e(TAG, "Task execution failed", e);
                resultBuilder
                        .setStatusValue(4) // FAILED
                        .setError(e.getMessage());
            }

            AgentOuterClass.AgentMessage result = AgentOuterClass.AgentMessage.newBuilder()
                    .setMsgId(UUID.randomUUID().toString())
                    .setTaskResult(resultBuilder.build())
                    .build();

            messageStream.onNext(result);
        }).start();
    }

    private void handleConfigUpdate(AgentOuterClass.ConfigUpdate config) {
        Log.i(TAG, "Received config update");
    }

    private void handleBroadcast(AgentOuterClass.BroadcastMessage broadcast) {
        Log.i(TAG, "Received broadcast: " + broadcast.getAction());
    }

    private AgentOuterClass.DeviceInfo getDeviceInfo() {
        return AgentOuterClass.DeviceInfo.newBuilder()
                .setOs("android")
                .setOsVersion(Build.VERSION.RELEASE)
                .setModel(Build.MODEL)
                .setManufacturer(Build.MANUFACTURER)
                .setArch(Build.SUPPORTED_ABIS[0])
                .build();
    }

    private AgentOuterClass.ResourceUsage getResourceUsage() {
        BatteryManager bm = (BatteryManager) context.getSystemService(Context.BATTERY_SERVICE);
        int batteryLevel = bm != null ? bm.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY) : 0;

        Runtime runtime = Runtime.getRuntime();
        long usedMem = runtime.totalMemory() - runtime.freeMemory();
        double memoryUsage = (double) usedMem / runtime.maxMemory() * 100;

        return AgentOuterClass.ResourceUsage.newBuilder()
                .setCpuUsage(0) // Would need native code
                .setMemoryUsage((int) memoryUsage)
                .setBatteryLevel(batteryLevel)
                .build();
    }

    private List<AgentOuterClass.Capability> getCapabilities() {
        List<AgentOuterClass.Capability> list = new ArrayList<>();
        for (String skillId : skills.keySet()) {
            list.add(AgentOuterClass.Capability.newBuilder()
                    .setId(skillId)
                    .setName(skillId)
                    .build());
        }
        return list;
    }

    private void scheduleReconnect() {
        handler.postDelayed(() -> {
            if (!connected) {
                connect();
            }
        }, 5000); // 5s retry
    }

    // ===== Listeners =====

    public void setConnectionListener(ConnectionListener listener) {
        this.connectionListener = listener;
    }

    public void setTaskListener(TaskListener listener) {
        this.taskListener = listener;
    }

    private void notifyConnected() {
        if (connectionListener != null) {
            handler.post(() -> connectionListener.onConnected());
        }
    }

    private void notifyDisconnected() {
        if (connectionListener != null) {
            handler.post(() -> connectionListener.onDisconnected());
        }
    }

    private void notifyConnectionError(String error) {
        if (connectionListener != null) {
            handler.post(() -> connectionListener.onError(error));
        }
    }

    // ===== Interfaces =====

    public interface ConnectionListener {
        void onConnected();
        void onDisconnected();
        void onError(String message);
    }

    public interface TaskListener {
        void onTaskReceived(String taskId, String skillId);
        void onTaskCompleted(String taskId);
        void onTaskFailed(String taskId, String error);
    }
}