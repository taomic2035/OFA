package com.ofa.agent.offline;

/**
 * 离线能力等级
 */
public enum OfflineLevel {
    NONE(0),    // 不支持离线
    L1(1),      // 完全离线 (本地执行)
    L2(2),      // 局域网协作
    L3(3),      // 弱网同步
    L4(4);      // 在线模式

    private final int value;

    OfflineLevel(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }

    public static OfflineLevel fromValue(int value) {
        for (OfflineLevel level : values()) {
            if (level.value == value) {
                return level;
            }
        }
        return NONE;
    }
}