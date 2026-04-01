package com.ofa.agent.p2p;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.net.InetAddress;

/**
 * 设备信息
 */
public class PeerInfo {
    public final String id;
    public final String name;
    public final InetAddress address;
    public final int port;
    public volatile boolean online;
    public volatile long lastSeen;
    public int latencyMs;

    public PeerInfo(@NonNull String id, @NonNull String name, @NonNull InetAddress address, int port) {
        this.id = id;
        this.name = name;
        this.address = address;
        this.port = port;
        this.online = true;
        this.lastSeen = System.currentTimeMillis();
        this.latencyMs = 0;
    }

    @NonNull
    @Override
    public String toString() {
        return "PeerInfo{" +
                "id='" + id + '\'' +
                ", name='" + name + '\'' +
                ", address=" + address +
                ", port=" + port +
                ", online=" + online +
                '}';
    }
}