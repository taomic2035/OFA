package com.ofa.agent.automation;

import androidx.annotation.NonNull;

/**
 * Automation capability level enumeration.
 */
public enum AutomationCapability {

    /**
     * No automation available
     */
    NONE(0, "No automation capability"),

    /**
     * Basic accessibility service capability
     * - Can perform clicks, swipes, input
     * - Can find elements by text, ID, description
     * - Limited to foreground app
     */
    BASIC(1, "Basic accessibility service"),

    /**
     * Enhanced accessibility capability
     * - Gestures with more precision
     * - Scroll-to-find
     * - Page stability detection
     * - Screenshot capture (if permitted)
     */
    ENHANCED(2, "Enhanced accessibility service"),

    /**
     * Full accessibility capability
     * - All accessibility features
     * - Multi-touch gestures
     * - Action recording and replay
     * - OCR integration (if available)
     */
    FULL_ACCESSIBILITY(3, "Full accessibility service"),

    /**
     * System-level capability (requires root/system permissions)
     * - Silent installation
     * - System settings modification
     * - Background service management
     * - All accessibility features plus system APIs
     */
    SYSTEM_LEVEL(4, "System-level automation");

    private final int level;
    private final String description;

    AutomationCapability(int level, @NonNull String description) {
        this.level = level;
        this.description = description;
    }

    public int getLevel() {
        return level;
    }

    @NonNull
    public String getDescription() {
        return description;
    }

    /**
     * Check if this capability supports at least the given level
     */
    public boolean supports(@NonNull AutomationCapability required) {
        return this.level >= required.level;
    }

    /**
     * Check if this capability supports basic operations
     */
    public boolean supportsBasicOperations() {
        return this.level >= BASIC.level;
    }

    /**
     * Check if this capability supports scroll operations
     */
    public boolean supportsScrollOperations() {
        return this.level >= ENHANCED.level;
    }

    /**
     * Check if this capability supports system-level operations
     */
    public boolean supportsSystemOperations() {
        return this.level >= SYSTEM_LEVEL.level;
    }
}