package com.ofa.agent.distributed;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * Device Role - defines the role of a device in the distributed agent network.
 *
 * A device can have multiple roles:
 * - SOURCE: Generates data (watch with sensors)
 * - DISPLAY: Shows information (phone, tablet, TV)
 * - EXECUTOR: Executes actions (phone, smart home hub)
 * - COORDINATOR: Coordinates multiple devices (center, phone)
 * - RELAY: Relays messages between devices (router, hub)
 */
public class DeviceRole {

    public static final int SOURCE = 1;      // 数据源：手表传感器、手机GPS等
    public static final int DISPLAY = 2;     // 显示设备：手机屏幕、手表、电视
    public static final int EXECUTOR = 3;    // 执行设备：手机操作、智能家居控制
    public static final int COORDINATOR = 4; // 协调器：Center、主手机
    public static final int RELAY = 5;       // 中继：消息转发

    private final int role;
    private final int priority;  // 同角色优先级，数值越高越优先
    private final boolean active; // 当前是否激活

    public DeviceRole(int role, int priority, boolean active) {
        this.role = role;
        this.priority = priority;
        this.active = active;
    }

    public int getRole() { return role; }
    public int getPriority() { return priority; }
    public boolean isActive() { return active; }

    /**
     * Role name for display
     */
    @NonNull
    public String getRoleName() {
        switch (role) {
            case SOURCE: return "数据源";
            case DISPLAY: return "显示设备";
            case EXECUTOR: return "执行设备";
            case COORDINATOR: return "协调器";
            case RELAY: return "中继";
            default: return "未知";
        }
    }

    /**
     * Check if this role matches a specific role type
     */
    public boolean isRole(int roleType) {
        return this.role == roleType;
    }

    /**
     * Common role presets
     */
    public static DeviceRole asSource(int priority) {
        return new DeviceRole(SOURCE, priority, true);
    }

    public static DeviceRole asDisplay(int priority) {
        return new DeviceRole(DISPLAY, priority, true);
    }

    public static DeviceRole asExecutor(int priority) {
        return new DeviceRole(EXECUTOR, priority, true);
    }

    public static DeviceRole asCoordinator(int priority) {
        return new DeviceRole(COORDINATOR, priority, true);
    }
}