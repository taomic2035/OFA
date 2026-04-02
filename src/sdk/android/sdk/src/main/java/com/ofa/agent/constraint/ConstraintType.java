package com.ofa.agent.constraint;

/**
 * 约束类型
 */
public enum ConstraintType {
    NONE(0),
    PRIVACY(1),
    FINANCIAL(2),
    SECURITY(4),
    AUTH_REQUIRED(8),
    LOCATION(16),
    PERSONAL(32),
    DEVICE(64);

    private final int value;

    ConstraintType(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }
}