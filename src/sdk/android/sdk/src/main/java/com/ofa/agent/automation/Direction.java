package com.ofa.agent.automation;

import androidx.annotation.NonNull;

/**
 * Direction enumeration for swipe operations.
 */
public enum Direction {
    UP,
    DOWN,
    LEFT,
    RIGHT;

    /**
     * Parse from string
     */
    @NonNull
    public static Direction fromString(@NonNull String value) {
        switch (value.toUpperCase()) {
            case "UP":
                return UP;
            case "DOWN":
                return DOWN;
            case "LEFT":
                return LEFT;
            case "RIGHT":
                return RIGHT;
            default:
                return DOWN; // Default to DOWN
        }
    }
}