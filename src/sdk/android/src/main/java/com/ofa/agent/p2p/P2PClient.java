package com.ofa.agent.p2p;

import android.net.nsd.NsdManager;
import android.net.nsd.NsdServiceInfo;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.io.IOException;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * P2P 客户端 - 设备发现与通信
 */
public class P2PClient {
    private static final String TAG = "P2PClient";
    private static final String SERVICE_TYPE = "_ofa._tcp.";

    private final String agentId;
    private final int port;
    private final NsdManager nsdManager;
    private final Handler mainHandler = new Handler(Looper.getMainLooper());

    private ServerSocket serverSocket;
    private volatile boolean running = false;
    private NsdManager.RegistrationListener registrationListener;
    private NsdManager.DiscoveryListener discoveryListener;

    private final Map<String, PeerInfo> peers = new ConcurrentHashMap<>();
    private final List<MessageListener> messageListeners = new CopyOnWriteArrayList<>();
    private final List<PeerListener> peerListeners = new CopyOnWriteArrayList<>();

    public P2PClient(@NonNull android.content.Context context, @NonNull String agentId) {
        this.agentId = agentId;
        this.port = findFreePort();
        this.nsdManager = (NsdManager) context.getSystemService(android.content.Context.NSD_SERVICE);
    }

    /**
     * 启动 P2P 服务
     */
    public void start() {
        if (running) return;
        running = true;

        // 启动服务器
        new Thread(this::startServer).start();

        // 注册服务
        registerService();

        // 启动发现
        startDiscovery();

        Log.i(TAG, "P2P client started on port " + port);
    }

    /**
     * 停止 P2P 服务
     */
    public void stop() {
        running = false;

        if (registrationListener != null) {
            nsdManager.unregisterService(registrationListener);
        }

        if (discoveryListener != null) {
            nsdManager.stopServiceDiscovery(discoveryListener);
        }

        if (serverSocket != null) {
            try {
                serverSocket.close();
            } catch (IOException e) {
                Log.e(TAG, "Error closing server socket", e);
            }
        }

        Log.i(TAG, "P2P client stopped");
    }

    private int findFreePort() {
        try (ServerSocket socket = new ServerSocket(0)) {
            return socket.getLocalPort();
        } catch (IOException e) {
            return 9090;
        }
    }

    private void startServer() {
        try {
            serverSocket = new ServerSocket(port);

            while (running) {
                try {
                    Socket client = serverSocket.accept();
                    handleConnection(client);
                } catch (IOException e) {
                    if (running) {
                        Log.e(TAG, "Error accepting connection", e);
                    }
                }
            }
        } catch (IOException e) {
            Log.e(TAG, "Error starting server", e);
        }
    }

    private void handleConnection(Socket socket) {
        new Thread(() -> {
            try {
                java.io.DataInputStream in = new java.io.DataInputStream(socket.getInputStream());
                int length = in.readInt();
                byte[] data = new byte[length];
                in.readFully(data);

                String json = new String(data, "UTF-8");
                P2PMessage message = P2PMessage.fromJson(json);

                // 更新设备信息
                PeerInfo peer = peers.get(message.fromId);
                if (peer != null) {
                    peer.lastSeen = System.currentTimeMillis();
                    peer.online = true;
                }

                // 通知监听器
                for (MessageListener listener : messageListeners) {
                    mainHandler.post(() -> listener.onMessage(message));
                }

            } catch (Exception e) {
                Log.e(TAG, "Error handling connection", e);
            } finally {
                try {
                    socket.close();
                } catch (IOException e) {
                    Log.e(TAG, "Error closing socket", e);
                }
            }
        }).start();
    }

    private void registerService() {
        NsdServiceInfo serviceInfo = new NsdServiceInfo();
        serviceInfo.setServiceName("OFA-" + agentId.substring(0, 8));
        serviceInfo.setServiceType(SERVICE_TYPE);
        serviceInfo.setPort(port);

        registrationListener = new NsdManager.RegistrationListener() {
            @Override
            public void onRegistrationFailed(NsdServiceInfo serviceInfo, int errorCode) {
                Log.e(TAG, "Service registration failed: " + errorCode);
            }

            @Override
            public void onUnregistrationFailed(NsdServiceInfo serviceInfo, int errorCode) {
                Log.e(TAG, "Service unregistration failed: " + errorCode);
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

    private void startDiscovery() {
        discoveryListener = new NsdManager.DiscoveryListener() {
            @Override
            public void onStartDiscoveryFailed(String serviceType, int errorCode) {
                Log.e(TAG, "Discovery start failed: " + errorCode);
            }

            @Override
            public void onStopDiscoveryFailed(String serviceType, int errorCode) {
                Log.e(TAG, "Discovery stop failed: " + errorCode);
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
                nsdManager.resolveService(serviceInfo, new NsdManager.ResolveListener() {
                    @Override
                    public void onResolveFailed(NsdServiceInfo serviceInfo, int errorCode) {
                        Log.e(TAG, "Resolve failed: " + errorCode);
                    }

                    @Override
                    public void onServiceResolved(NsdServiceInfo serviceInfo) {
                        String peerId = extractPeerId(serviceInfo.getServiceName());
                        if (!peerId.equals(agentId)) {
                            PeerInfo peer = new PeerInfo(
                                peerId,
                                serviceInfo.getServiceName(),
                                serviceInfo.getHost(),
                                serviceInfo.getPort()
                            );
                            addPeer(peer);
                        }
                    }
                });
            }

            @Override
            public void onServiceLost(NsdServiceInfo serviceInfo) {
                String peerId = extractPeerId(serviceInfo.getServiceName());
                removePeer(peerId);
            }
        };

        nsdManager.discoverServices(SERVICE_TYPE, NsdManager.PROTOCOL_DNS_SD, discoveryListener);
    }

    private String extractPeerId(String serviceName) {
        if (serviceName.startsWith("OFA-")) {
            return serviceName.substring(4);
        }
        return serviceName;
    }

    /**
     * 发送消息
     */
    public boolean send(@NonNull String peerId, @NonNull byte[] data) {
        PeerInfo peer = peers.get(peerId);
        if (peer == null || !peer.online) {
            Log.w(TAG, "Peer not found or offline: " + peerId);
            return false;
        }

        try {
            Socket socket = new Socket(peer.address, peer.port);
            java.io.DataOutputStream out = new java.io.DataOutputStream(socket.getOutputStream());

            P2PMessage message = new P2PMessage(
                P2PMessage.TYPE_DATA,
                agentId,
                peerId,
                data
            );

            byte[] jsonBytes = message.toJson().getBytes("UTF-8");
            out.writeInt(jsonBytes.length);
            out.write(jsonBytes);
            out.flush();
            socket.close();

            return true;
        } catch (Exception e) {
            Log.e(TAG, "Send failed", e);
            peer.online = false;
            return false;
        }
    }

    /**
     * 广播消息
     */
    public void broadcast(@NonNull byte[] data) {
        for (PeerInfo peer : peers.values()) {
            if (peer.online) {
                send(peer.id, data);
            }
        }
    }

    /**
     * 添加设备
     */
    public void addPeer(@NonNull PeerInfo peer) {
        peers.put(peer.id, peer);

        for (PeerListener listener : peerListeners) {
            mainHandler.post(() -> listener.onPeerAdded(peer.id, peer.name));
        }

        Log.i(TAG, "Peer added: " + peer.id);
    }

    /**
     * 移除设备
     */
    public void removePeer(@NonNull String peerId) {
        PeerInfo peer = peers.remove(peerId);

        if (peer != null) {
            for (PeerListener listener : peerListeners) {
                mainHandler.post(() -> listener.onPeerRemoved(peerId, peer.name));
            }

            Log.i(TAG, "Peer removed: " + peerId);
        }
    }

    /**
     * 获取设备列表
     */
    @NonNull
    public List<PeerInfo> getPeers() {
        return new ArrayList<>(peers.values());
    }

    /**
     * 获取在线设备列表
     */
    @NonNull
    public List<PeerInfo> getOnlinePeers() {
        List<PeerInfo> online = new ArrayList<>();
        for (PeerInfo peer : peers.values()) {
            if (peer.online) {
                online.add(peer);
            }
        }
        return online;
    }

    /**
     * 获取端口
     */
    public int getPort() {
        return port;
    }

    /**
     * 添加消息监听器
     */
    public void addMessageListener(@NonNull MessageListener listener) {
        messageListeners.add(listener);
    }

    /**
     * 移除消息监听器
     */
    public void removeMessageListener(@NonNull MessageListener listener) {
        messageListeners.remove(listener);
    }

    /**
     * 添加设备监听器
     */
    public void addPeerListener(@NonNull PeerListener listener) {
        peerListeners.add(listener);
    }

    /**
     * 移除设备监听器
     */
    public void removePeerListener(@NonNull PeerListener listener) {
        peerListeners.remove(listener);
    }

    public interface MessageListener {
        void onMessage(P2PMessage message);
    }

    public interface PeerListener {
        void onPeerAdded(String peerId, String peerName);
        void onPeerRemoved(String peerId, String peerName);
    }
}