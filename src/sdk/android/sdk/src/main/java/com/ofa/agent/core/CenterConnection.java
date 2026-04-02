package com.ofa.agent.core;

import android.content.Context;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.grpc.AgentGrpc;
import com.ofa.agent.grpc.AgentOuterClass;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;

/**
 * Center Connection - manages connection to OFA Center.
 *
 * Responsibilities:
 * - Registration and heartbeat
 * - Task assignment handling
 * - Status reporting
 * - Data synchronization
 */
public class CenterConnection {

    private static final String TAG = "CenterConnection";

    private final Context context;
    private final AgentProfile profile;
    private final String centerAddress;
    private final int centerPort;

    private ManagedChannel channel;
    private AgentGrpc.AgentStub asyncStub;
    private StreamObserver<AgentOuterClass.AgentMessage> messageStream;

    private volatile boolean connected = false;
    private volatile boolean registered = false;

    private final Handler handler = new Handler(Looper.getMainLooper());
    private Runnable heartbeatRunnable;

    private ConnectionListener connectionListener;
    private TaskAssignmentListener taskListener;

    // Pending tasks
    private final Map<String, CompletableFuture<TaskResult>> pendingTasks = new ConcurrentHashMap<>();

    /**
     * Connection listener
     */
    public interface ConnectionListener {
        void onConnected();
        void onDisconnected();
        void onError(String error);
    }

    /**
     * Task assignment listener
     */
    public interface TaskAssignmentListener {
        void onTaskAssigned(TaskRequest request);
    }

    /**
     * Create center connection
     */
    public CenterConnection(@NonNull Context context, @NonNull AgentProfile profile) {
        this(context, profile, "localhost", 9090);
    }

    public CenterConnection(@NonNull Context context, @NonNull AgentProfile profile,
                            @NonNull String address, int port) {
        this.context = context;
        this.profile = profile;
        this.centerAddress = address;
        this.centerPort = port;

        // Setup heartbeat
        heartbeatRunnable = () -> {
            if (connected) {
                sendHeartbeat();
                handler.postDelayed(heartbeatRunnable, 30000);
            }
        };
    }

    /**
     * Connect to Center
     */
    public void connect() {
        new Thread(() -> {
            try {
                Log.i(TAG, "Connecting to Center: " + centerAddress + ":" + centerPort);

                // Create gRPC channel
                channel = ManagedChannelBuilder
                    .forAddress(centerAddress, centerPort)
                    .usePlaintext()
                    .build();

                asyncStub = AgentGrpc.newStub(channel);

                // Start bidirectional stream
                messageStream = asyncStub.connect(new StreamObserver<AgentOuterClass.CenterMessage>() {
                    @Override
                    public void onNext(AgentOuterClass.CenterMessage message) {
                        handleCenterMessage(message);
                    }

                    @Override
                    public void onError(Throwable t) {
                        Log.e(TAG, "Stream error", t);
                        connected = false;
                        notifyDisconnected(t.getMessage());
                        scheduleReconnect();
                    }

                    @Override
                    public void onCompleted() {
                        Log.i(TAG, "Stream completed");
                        connected = false;
                        notifyDisconnected("Stream completed");
                    }
                });

                // Send registration
                sendRegistration();

                connected = true;
                notifyConnected();

                // Start heartbeat
                handler.post(heartbeatRunnable);

                Log.i(TAG, "Connected to Center");

            } catch (Exception e) {
                Log.e(TAG, "Connection failed", e);
                notifyError(e.getMessage());
                scheduleReconnect();
            }
        }).start();
    }

    /**
     * Disconnect from Center
     */
    public void disconnect() {
        Log.i(TAG, "Disconnecting from Center...");

        connected = false;
        handler.removeCallbacks(heartbeatRunnable);

        if (messageStream != null) {
            messageStream.onCompleted();
            messageStream = null;
        }

        if (channel != null) {
            channel.shutdown();
            channel = null;
        }

        notifyDisconnected("Disconnect requested");
    }

    /**
     * Check if connected
     */
    public boolean isConnected() {
        return connected;
    }

    /**
     * Check if registered
     */
    public boolean isRegistered() {
        return registered;
    }

    /**
     * Execute task via Center
     */
    @NonNull
    public CompletableFuture<TaskResult> executeTask(@NonNull TaskRequest request) {
        CompletableFuture<TaskResult> future = new CompletableFuture<>();

        if (!connected || messageStream == null) {
            future.completeExceptionally(new Exception("Not connected to Center"));
            return future;
        }

        // Store pending task
        pendingTasks.put(request.taskId, future);

        // Send task request
        try {
            AgentOuterClass.TaskAssignment task = AgentOuterClass.TaskAssignment.newBuilder()
                .setTaskId(request.taskId)
                .setSkillId(request.params.getOrDefault("skillId", "").toString())
                .build();

            AgentOuterClass.AgentMessage message = AgentOuterClass.AgentMessage.newBuilder()
                .setMsgId(UUID.randomUUID().toString())
                .setTask(task)
                .build();

            messageStream.onNext(message);
            Log.d(TAG, "Task sent to Center: " + request.taskId);

        } catch (Exception e) {
            pendingTasks.remove(request.taskId);
            future.completeExceptionally(e);
        }

        return future;
    }

    /**
     * Report status to Center
     */
    public void reportStatus(@NonNull AgentProfile.AgentStatus status) {
        if (!connected || messageStream == null) return;

        try {
            AgentOuterClass.HeartbeatRequest heartbeat = AgentOuterClass.HeartbeatRequest.newBuilder()
                .setAgentId(profile.getAgentId())
                .setStatusValue(status.getValue())
                .build();

            AgentOuterClass.AgentMessage message = AgentOuterClass.AgentMessage.newBuilder()
                .setMsgId(UUID.randomUUID().toString())
                .setHeartbeat(heartbeat)
                .build();

            messageStream.onNext(message);
        } catch (Exception e) {
            Log.w(TAG, "Failed to report status: " + e.getMessage());
        }
    }

    /**
     * Sync data with Center
     */
    public void sync() {
        if (!connected) return;

        Log.i(TAG, "Syncing with Center...");

        // Upload local capabilities
        // Download updates
        // This would involve actual data sync logic
    }

    // ===== Private Methods =====

    private void sendRegistration() {
        AgentOuterClass.DeviceInfo deviceInfo = AgentOuterClass.DeviceInfo.newBuilder()
            .setOs(profile.getDeviceInfo().os)
            .setOsVersion(profile.getDeviceInfo().osVersion)
            .setModel(profile.getDeviceInfo().model)
            .setManufacturer(profile.getDeviceInfo().manufacturer)
            .setArch(profile.getDeviceInfo().arch)
            .build();

        List<AgentOuterClass.Capability> capabilities = new ArrayList<>();
        for (AgentProfile.Capability cap : profile.getCapabilities()) {
            capabilities.add(AgentOuterClass.Capability.newBuilder()
                .setId(cap.id)
                .setName(cap.name)
                .build());
        }

        AgentOuterClass.RegisterRequest register = AgentOuterClass.RegisterRequest.newBuilder()
            .setAgentId(profile.getAgentId())
            .setName(profile.getName())
            .setTypeValue(profile.getType().getValue())
            .setDeviceInfo(deviceInfo)
            .addAllCapabilities(capabilities)
            .build();

        AgentOuterClass.AgentMessage message = AgentOuterClass.AgentMessage.newBuilder()
            .setMsgId(UUID.randomUUID().toString())
            .setRegister(register)
            .build();

        messageStream.onNext(message);
        registered = true;

        Log.i(TAG, "Registration sent");
    }

    private void sendHeartbeat() {
        if (!connected || messageStream == null) return;

        AgentOuterClass.HeartbeatRequest heartbeat = AgentOuterClass.HeartbeatRequest.newBuilder()
            .setAgentId(profile.getAgentId())
            .setStatusValue(profile.getStatus().getValue())
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

            case TASKRESULT:
                handleTaskResult(message.getTaskResult());
                break;

            default:
                Log.d(TAG, "Unknown message type: " + message.getPayloadCase());
        }
    }

    private void handleTaskAssignment(AgentOuterClass.TaskAssignment task) {
        Log.i(TAG, "Received task from Center: " + task.getTaskId());

        TaskRequest request = new TaskRequest.Builder()
            .taskId(task.getTaskId())
            .type(TaskRequest.TYPE_SKILL)
            .param("skillId", task.getSkillId())
            .source("center")
            .build();

        if (taskListener != null) {
            taskListener.onTaskAssigned(request);
        }
    }

    private void handleConfigUpdate(AgentOuterClass.ConfigUpdate config) {
        Log.i(TAG, "Received config update from Center");
    }

    private void handleBroadcast(AgentOuterClass.BroadcastMessage broadcast) {
        Log.i(TAG, "Received broadcast from Center: " + broadcast.getAction());
    }

    private void handleTaskResult(AgentOuterClass.TaskResult result) {
        Log.d(TAG, "Received task result: " + result.getTaskId());

        CompletableFuture<TaskResult> future = pendingTasks.remove(result.getTaskId());
        if (future != null) {
            if (result.getStatusValue() == 3) { // COMPLETED
                Map<String, Object> data = new HashMap<>();
                data.put("output", result.getOutput().toString());
                future.complete(TaskResult.success(result.getTaskId(), data));
            } else {
                future.completeExceptionally(new Exception(result.getError()));
            }
        }
    }

    private void scheduleReconnect() {
        handler.postDelayed(() -> {
            if (!connected) {
                connect();
            }
        }, 5000);
    }

    // ===== Listeners =====

    public void setConnectionListener(@Nullable ConnectionListener listener) {
        this.connectionListener = listener;
    }

    public void setTaskListener(@Nullable TaskAssignmentListener listener) {
        this.taskListener = listener;
    }

    private void notifyConnected() {
        if (connectionListener != null) {
            handler.post(() -> connectionListener.onConnected());
        }
    }

    private void notifyDisconnected(String reason) {
        if (connectionListener != null) {
            handler.post(() -> connectionListener.onDisconnected());
        }
    }

    private void notifyError(String error) {
        if (connectionListener != null) {
            handler.post(() -> connectionListener.onError(error));
        }
    }
}