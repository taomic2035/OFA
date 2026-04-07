package com.ofa.agent.trust;

/**
 * 设备信任级别 (v2.8.0)
 *
 * Center 是永远在线的灵魂载体，设备信任级别定义了设备对身份的访问权限。
 */
public enum TrustLevel {
    NONE(0, "none"),
    LOW(1, "low"),        // 新设备，只读权限
    MEDIUM(2, "medium"),  // 验证过的设备，读+同步权限
    HIGH(3, "high"),      // 高信任设备，读+写+同步权限
    PRIMARY(4, "primary"); // 主设备，完全权限

    private final int value;
    private final String name;

    TrustLevel(int value, String name) {
        this.value = value;
        this.name = name;
    }

    public int getValue() {
        return value;
    }

    public String getName() {
        return name;
    }

    public static TrustLevel fromValue(int value) {
        for (TrustLevel level : values()) {
            if (level.value == value) {
                return level;
            }
        }
        return NONE;
    }

    public static TrustLevel fromName(String name) {
        for (TrustLevel level : values()) {
            if (level.name.equalsIgnoreCase(name)) {
                return level;
            }
        }
        return NONE;
    }

    /**
     * 检查是否有指定权限
     */
    public boolean hasPermission(String permission) {
        switch (permission) {
            case "read":
                return this != NONE;
            case "sync":
                return value >= MEDIUM.value;
            case "write":
                return value >= HIGH.value;
            case "admin":
            case "transfer":
                return this == PRIMARY;
            default:
                return false;
        }
    }
}