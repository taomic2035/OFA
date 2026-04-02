package com.ofa.agent.core;

import android.content.Context;
import android.net.nsd.NsdManager;
import android.net.nsd.NsdServiceInfo;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Peer Network - enables direct communication between agents.
 *
 * Features:
 * - Local network discovery (NSD/mDNS)
 * - P2P messaging
 * - Task delegation to peers
 * - Capability-based peer selection
 */
public class PeerNetwork {

    private static final String TAG = "PeerNetwork";
    private static final String SERVICE_TYPE = "_ofa_agent._tcp.";

    private final Context context;
    private final AgentProfile profile;

    // NSD
    private NsdManager nsdManager;
    private NsdManager.RegistrationListener registrationListener;
    private NsdManager.DiscoveryListener discoveryListener;

    // Server
    private ServerSocket serverSocket;
    private int localPort;
    private volatile boolean running = false;

    // Peers
    private final Map<String, PeerInfo> discoveredPeers = new ConcurrentHashMap<>();
    private final Map<String, PeerConnection> peerConnections = new ConcurrentHashMap<>();

    // Thread pool
    private ExecutorService executor;

    // Listeners
    private PeerListener peerListener;

    /**
     * Peer information
     */
    public static class PeerInfo {
        public final String agentId;
        public final String name;
        public final String host;
        public final int port;
        public final AgentProfile.AgentType type;
        public final List<String> capabilities;

        public PeerInfo(String agentId, String name, String host, int port,
                        AgentProfile.AgentType type, List<String> capabilities) {
            this.agentId = agentId;
            this.name = name;
            this.host = host;
            this.port = port;
            this.type = type;
            this.capabilities = capabilities;
        }
    }

    /**
     * Peer connection
     */
    private static class PeerConnection {
        final Socket socket;
        final PrintWriter writer;
        final BufferedReader reader;

        PeerConnection(Socket socket) throws IOException {
            this.socket = socket;
            this.writer = new PrintWriter(socket.getOutputStream(), true);
            this.reader = new BufferedReader(new InputStreamReader(socket.getInputStream()));
        }

        void close() {
            try {
                socket.close();
            } catch (IOException e) {
                // Ignore
            }
        }
    }

    /**
     * Peer listener interface
     */
    public interface PeerListener {
        void onPeerDiscovered(PeerInfo peer);
        void onPeerLost(String agentId);
        void onMessageReceived(String peerId, String message);
    }

    /**
     * Create peer network
     */
    public PeerNetwork(@NonNull Context context, @NonNull AgentProfile profile) {
        this.context = context;
        this.profile = profile;
    }

    /**
     * Start peer network
     */
    public void start() {
        if (running) return;

        Log.i(TAG, "Starting peer network...");

        executor = Executors.newCachedThreadPool();

        // Start server
        startServer();

        // Register service
        registerService();

        // Start discovery
        startDiscovery();

        running = true;
        Log.i(TAG, "Peer network started on port " + localPort);
    }

    /**
     * Stop peer network
     */
    public void stop() {
        if (!running) return;

        Log.i(TAG, "Stopping peer network...");

        running = false;

        // Unregister service
        unregisterService();

        // Stop discovery
        stopDiscovery();

        // Close server
        if (serverSocket != null) {
            try {
                serverSocket.close();
            } catch (IOException e) {
                // Ignore
            }
        }

        // Close all peer connections
        for (PeerConnection conn : peerConnections.values()) {
            conn.close();
        }
        peerConnections.clear();

        // Shutdown executor
        if (executor != null) {
            executor.shutdown();
        }

        discoveredPeers.clear();

        Log.i(TAG, "Peer network stopped");
    }

    /**
     * Check if active
     */
    public boolean isActive() {
        return running;
    }

    /**
     * Get discovered agents
     */
    @NonNull
    public List<AgentProfile> getDiscoveredAgents() {
        List<AgentProfile> agents = new ArrayList<>();
        for (PeerInfo peer : discoveredPeers.values()) {
            AgentProfile.Builder builder = new AgentProfile.Builder()
                .agentId(peer.agentId)
                .name(peer.name)
                .type(peer.type);
            agents.add(builder.build());
        }
        return agents;
    }

    /**
     * Send message to peer
     */
    public boolean send(@NonNull String peerId, @NonNull String message) {
        PeerInfo peer = discoveredPeers.get(peerId);
        if (peer == null) {
            Log.w(TAG, "Peer not found: " + peerId);
            return false;
        }

        try {
            PeerConnection conn = getOrCreateConnection(peer);
            if (conn != null) {
                conn.writer.println(message);
                return true;
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to send to peer: " + e.getMessage());
        }

        return false;
    }

    /**
     * Request task from peer
     */
    @Nullable
    public TaskResult requestTask(@NonNull String peerId, @NonNull TaskRequest request) {
        PeerInfo peer = discoveredPeers.get(peerId);
        if (peer == null) {
            return TaskResult.failure(request.taskId, "Peer not found");
        }

        try {
            PeerConnection conn = getOrCreateConnection(peer);
            if (conn == null) {
                return TaskResult.failure(request.taskId, "Cannot connect to peer");
            }

            // Send request
            JSONObject requestJson = new JSONObject();
            requestJson.put("type", "task_request");
            requestJson.put("taskId", request.taskId);
            requestJson.put("taskType", request.type);
            requestJson.put("params", new JSONObject(request.params));

            conn.writer.println(requestJson.toString());

            // Wait for response (simplified)
            String response = conn.reader.readLine();
            if (response != null) {
                JSONObject responseJson = new JSONObject(response);
                boolean success = responseJson.optBoolean("success", false);
                if (success) {
                    Map<String, Object> data = new HashMap<>();
                    data.put("result", responseJson.optString("result"));
                    return TaskResult.success(request.taskId, data, 0, "peer:" + peerId);
                } else {
                    return TaskResult.failure(request.taskId, responseJson.optString("error"));
                }
            }

        } catch (Exception e) {
            Log.e(TAG, "Task request failed: " + e.getMessage());
            return TaskResult.failure(request.taskId, e.getMessage());
        }

        return TaskResult.failure(request.taskId, "No response from peer");
    }

    /**
     * Find peer with capability
     */
    @Nullable
    public PeerInfo findPeerWithCapability(@NonNull String capabilityId) {
        for (PeerInfo peer : discoveredPeers.values()) {
            if (peer.capabilities.contains(capabilityId)) {
                return peer;
            }
        }
        return null;
    }

    // ===== Private Methods =====

    private void startServer() {
        executor.execute(() -> {
            try {
                serverSocket = new ServerSocket(0); // Auto-assign port
                localPort = serverSocket.getLocalPort();

                Log.i(TAG, "Server listening on port " + localPort);

                while (running) {
                    try {
                        Socket clientSocket = serverSocket.accept();
                        handleClient(clientSocket);
                    } catch (IOException e) {
                        if (running) {
                            Log.w(TAG, "Accept error: " + e.getMessage());
                        }
                    }
                }
            } catch (IOException e) {
                Log.e(TAG, "Server start failed: " + e.getMessage());
            }
        });
    }

    private void handleClient(Socket socket) {
        executor.execute(() -> {
            try {
                BufferedReader reader = new BufferedReader(
                    new InputStreamReader(socket.getInputStream()));
                PrintWriter writer = new PrintWriter(socket.getOutputStream(), true);

                String line;
                while ((line = reader.readLine()) != null) {
                    handleMessage(socket, line, writer);
                }

            } catch (IOException e) {
                Log.d(TAG, "Client disconnected: " + e.getMessage());
            } finally {
                try {
                    socket.close();
                } catch (IOException e) {
                    // Ignore
                }
            }
        });
    }

    private void handleMessage(Socket socket, String message, PrintWriter writer) {
        Log.d(TAG, "Received: " + message.substring(0, Math.min(100, message.length())));

        try {
            JSONObject json = new JSONObject(message);
            String type = json.optString("type");

            switch (type) {
                case "task_request":
                    handleTaskRequest(json, writer);
                    break;

                case "ping":
                    writer.println("{\"type\":\"pong\"}");
                    break;

                default:
                    if (peerListener != null) {
                        String peerId = json.optString("agentId", "unknown");
                        peerListener.onMessageReceived(peerId, message);
                    }
            }
        } catch (Exception e) {
            Log.w(TAG, "Message parse error: " + e.getMessage());
        }
    }

    private void handleTaskRequest(JSONObject request, PrintWriter writer) {
        try {
            String taskId = request.getString("taskId");
            String taskType = request.getString("taskType");

            // This would delegate to LocalExecutionEngine
            // For now, return a simple response
            JSONObject response = new JSONObject();
            response.put("taskId", taskId);
            response.put("success", false);
            response.put("error", "Task execution not implemented");

            writer.println(response.toString());
        } catch (Exception e) {
            Log.e(TAG, "Task request handling failed: " + e.getMessage());
        }
    }

    @Nullable
    private PeerConnection getOrCreateConnection(@NonNull PeerInfo peer) {
        PeerConnection conn = peerConnections.get(peer.agentId);
        if (conn != null && !conn.socket.isClosed()) {
            return conn;
        }

        try {
            Socket socket = new Socket(peer.host, peer.port);
            conn = new PeerConnection(socket);
            peerConnections.put(peer.agentId, conn);
            return conn;
        } catch (IOException e) {
            Log.e(TAG, "Connection failed: " + e.getMessage());
            return null;
        }
    }

    private void registerService() {
        nsdManager = (NsdManager) context.getSystemService(Context.NSD_SERVICE);
        if (nsdManager == null) {
            Log.w(TAG, "NSD not available");
            return;
        }

        NsdServiceInfo serviceInfo = new NsdServiceInfo();
        serviceInfo.setServiceName("OFA_" + profile.getName());
        serviceInfo.setServiceType(SERVICE_TYPE);
        serviceInfo.setPort(localPort);

        registrationListener = new NsdManager.RegistrationListener() {
            @Override
            public void onRegistrationFailed(NsdServiceInfo serviceInfo, int errorCode) {
                Log.e(TAG, "Registration failed: " + errorCode);
            }

            @Override
            public void onUnregistrationFailed(NsdServiceInfo serviceInfo, int errorCode) {
                Log.w(TAG, "Unregistration failed: " + errorCode);
            }

            @Override
            public void onServiceRegistered(NsdServiceInfo serviceInfo) {
                Log.i(TAG, "Service registered: " + serviceInfo.getServiceName());
            }

            @Override
            public void onServiceUnregistered(NsdServiceInfo serviceInfo) {
                Log.i(TAG, "Service unregistered");
            }
        };

        nsdManager.registerService(serviceInfo, NsdManager.PROTOCOL_DNS_SD, registrationListener);
    }

    private void unregisterService() {
        if (nsdManager != null && registrationListener != null) {
            nsdManager.unregisterService(registrationListener);
        }
    }

    private void startDiscovery() {
        if (nsdManager == null) return;

        discoveryListener = new NsdManager.DiscoveryListener() {
            @Override
            public void onStartDiscoveryFailed(String serviceType, int errorCode) {
                Log.e(TAG, "Discovery start failed: " + errorCode);
            }

            @Override
            public void onStopDiscoveryFailed(String serviceType, int errorCode) {
                Log.w(TAG, "Discovery stop failed: " + errorCode);
            }

            @Override
            public void onDiscoveryStarted(String serviceType) {
                Log.i(TAG, "Discovery started");
            }

            @Override
            public void onDiscoveryStopped(String serviceType) {
                Log.i(TAG, "Discovery stopped");
            }

            @Override
            public void onServiceFound(NsdServiceInfo serviceInfo) {
                Log.d(TAG, "Service found: " + serviceInfo.getServiceName());
                resolveService(serviceInfo);
            }

            @Override
            public void onServiceLost(NsdServiceInfo serviceInfo) {
                Log.d(TAG, "Service lost: " + serviceInfo.getServiceName());
                String name = serviceInfo.getServiceName().replace("OFA_", "");
                discoveredPeers.remove(name);
                if (peerListener != null) {
                    peerListener.onPeerLost(name);
                }
            }
        };

        nsdManager.discoverServices(SERVICE_TYPE, NsdManager.PROTOCOL_DNS_SD, discoveryListener);
    }

    private void resolveService(NsdServiceInfo serviceInfo) {
        nsdManager.resolveService(serviceInfo, new NsdManager.ResolveListener() {
            @Override
            public void onResolveFailed(NsdServiceInfo serviceInfo, int errorCode) {
                Log.w(TAG, "Resolve failed: " + errorCode);
            }

            @Override
            public void onServiceResolved(NsdServiceInfo resolvedInfo) {
                String name = resolvedInfo.getServiceName().replace("OFA_", "");

                // Skip self
                if (name.equals(profile.getName())) {
                    return;
                }

                PeerInfo peer = new PeerInfo(
                    name,
                    name,
                    resolvedInfo.getHost().getHostAddress(),
                    resolvedInfo.getPort(),
                    AgentProfile.AgentType.MOBILE,
                    new ArrayList<>()
                );

                discoveredPeers.put(name, peer);
                Log.i(TAG, "Peer resolved: " + peer.agentId + " @ " + peer.host + ":" + peer.port);

                if (peerListener != null) {
                    peerListener.onPeerDiscovered(peer);
                }
            }
        });
    }

    private void stopDiscovery() {
        if (nsdManager != null && discoveryListener != null) {
            nsdManager.stopServiceDiscovery(discoveryListener);
        }
    }

    // ===== Listeners =====

    public void setPeerListener(@Nullable PeerListener listener) {
        this.peerListener = listener;
    }
}