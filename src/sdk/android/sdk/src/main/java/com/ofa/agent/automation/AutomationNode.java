package com.ofa.agent.automation;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

/**
 * Represents a UI automation node ( AccessibilityNodeInfo wrapper).
 */
public class AutomationNode {

    private final String className;
    private final String text;
    private final String contentDescription;
    private final String resourceId;
    private final int index;
    private final boolean clickable;
    private final boolean enabled;
    private final boolean focusable;
    private final boolean scrollable;
    private final boolean checkable;
    private final boolean checked;
    private final boolean visibleToUser;
    private final int boundsLeft;
    private final int boundsTop;
    private final int boundsRight;
    private final int boundsBottom;

    // Package name of the node's app
    @Nullable
    private final String packageName;

    /**
     * Full constructor
     */
    public AutomationNode(@Nullable String className, @Nullable String text,
                          @Nullable String contentDescription, @Nullable String resourceId,
                          int index, boolean clickable, boolean enabled, boolean focusable,
                          boolean scrollable, boolean checkable, boolean checked,
                          boolean visibleToUser, int boundsLeft, int boundsTop,
                          int boundsRight, int boundsBottom, @Nullable String packageName) {
        this.className = className;
        this.text = text;
        this.contentDescription = contentDescription;
        this.resourceId = resourceId;
        this.index = index;
        this.clickable = clickable;
        this.enabled = enabled;
        this.focusable = focusable;
        this.scrollable = scrollable;
        this.checkable = checkable;
        this.checked = checked;
        this.visibleToUser = visibleToUser;
        this.boundsLeft = boundsLeft;
        this.boundsTop = boundsTop;
        this.boundsRight = boundsRight;
        this.boundsBottom = boundsBottom;
        this.packageName = packageName;
    }

    // ===== Getters =====

    @Nullable
    public String getClassName() {
        return className;
    }

    @Nullable
    public String getText() {
        return text;
    }

    @Nullable
    public String getContentDescription() {
        return contentDescription;
    }

    @Nullable
    public String getResourceId() {
        return resourceId;
    }

    public int getIndex() {
        return index;
    }

    public boolean isClickable() {
        return clickable;
    }

    public boolean isEnabled() {
        return enabled;
    }

    public boolean isFocusable() {
        return focusable;
    }

    public boolean isScrollable() {
        return scrollable;
    }

    public boolean isCheckable() {
        return checkable;
    }

    public boolean isChecked() {
        return checked;
    }

    public boolean isVisibleToUser() {
        return visibleToUser;
    }

    public int getBoundsLeft() {
        return boundsLeft;
    }

    public int getBoundsTop() {
        return boundsTop;
    }

    public int getBoundsRight() {
        return boundsRight;
    }

    public int getBoundsBottom() {
        return boundsBottom;
    }

    @Nullable
    public String getPackageName() {
        return packageName;
    }

    // ===== Utility Methods =====

    /**
     * Get center X coordinate
     */
    public int getCenterX() {
        return (boundsLeft + boundsRight) / 2;
    }

    /**
     * Get center Y coordinate
     */
    public int getCenterY() {
        return (boundsTop + boundsBottom) / 2;
    }

    /**
     * Get width of node
     */
    public int getWidth() {
        return boundsRight - boundsLeft;
    }

    /**
     * Get height of node
     */
    public int getHeight() {
        return boundsBottom - boundsTop;
    }

    /**
     * Check if node matches selector
     */
    public boolean matches(@NonNull BySelector selector) {
        // Text matching
        if (selector.getText() != null) {
            if (text == null || !text.equals(selector.getText())) return false;
        }
        if (selector.getTextContains() != null) {
            if (text == null || !text.contains(selector.getTextContains())) return false;
        }
        if (selector.getTextStartsWith() != null) {
            if (text == null || !text.startsWith(selector.getTextStartsWith())) return false;
        }
        if (selector.getTextEndsWith() != null) {
            if (text == null || !text.endsWith(selector.getTextEndsWith())) return false;
        }

        // Resource ID matching
        if (selector.getResourceId() != null) {
            if (resourceId == null || !resourceId.equals(selector.getResourceId())) return false;
        }

        // Class name matching
        if (selector.getClassName() != null) {
            if (className == null || !className.equals(selector.getClassName())) return false;
        }

        // Content description matching
        if (selector.getContentDescription() != null) {
            if (contentDescription == null ||
                !contentDescription.equals(selector.getContentDescription())) return false;
        }
        if (selector.getDescContains() != null) {
            if (contentDescription == null ||
                !contentDescription.contains(selector.getDescContains())) return false;
        }

        // State matching
        if (selector.isClickable() && !clickable) return false;
        if (!selector.isEnabled() && enabled) return false;
        if (selector.isFocusable() && !focusable) return false;
        if (selector.isScrollable() && !scrollable) return false;

        return true;
    }

    /**
     * Convert to JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            if (className != null) json.put("className", className);
            if (text != null) json.put("text", text);
            if (contentDescription != null) json.put("contentDescription", contentDescription);
            if (resourceId != null) json.put("resourceId", resourceId);
            json.put("index", index);
            json.put("clickable", clickable);
            json.put("enabled", enabled);
            json.put("focusable", focusable);
            json.put("scrollable", scrollable);
            json.put("checkable", checkable);
            json.put("checked", checked);
            json.put("visibleToUser", visibleToUser);

            JSONObject bounds = new JSONObject();
            bounds.put("left", boundsLeft);
            bounds.put("top", boundsTop);
            bounds.put("right", boundsRight);
            bounds.put("bottom", boundsBottom);
            json.put("bounds", bounds);

            if (packageName != null) json.put("packageName", packageName);
        } catch (Exception e) {
            // Should not fail
        }
        return json;
    }

    @NonNull
    @Override
    public String toString() {
        return "AutomationNode{" +
                (text != null ? "text='" + text + "'" : "") +
                (resourceId != null ? ", id='" + resourceId + "'" : "") +
                ", bounds=[" + boundsLeft + "," + boundsTop + "," + boundsRight + "," + boundsBottom + "]" +
                ", clickable=" + clickable +
                '}';
    }
}