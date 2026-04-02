package com.ofa.agent.automation;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * Element selector for finding UI elements.
 * Supports multiple selection criteria.
 */
public class BySelector {

    // Selection criteria
    @Nullable
    private String text;
    @Nullable
    private String textContains;
    @Nullable
    private String textStartsWith;
    @Nullable
    private String textEndsWith;
    @Nullable
    private String resourceId;
    @Nullable
    private String className;
    @Nullable
    private String contentDescription;
    @Nullable
    private String descContains;
    private int index = -1;
    private boolean clickable = false;
    private boolean enabled = true;
    private boolean focusable = false;
    private boolean scrollable = false;

    /**
     * Empty constructor
     */
    public BySelector() {}

    // ===== Static Factory Methods =====

    /**
     * Select by exact text
     */
    @NonNull
    public static BySelector text(@NonNull String text) {
        BySelector selector = new BySelector();
        selector.text = text;
        return selector;
    }

    /**
     * Select by text containing substring
     */
    @NonNull
    public static BySelector textContains(@NonNull String text) {
        BySelector selector = new BySelector();
        selector.textContains = text;
        return selector;
    }

    /**
     * Select by text starting with prefix
     */
    @NonNull
    public static BySelector textStartsWith(@NonNull String prefix) {
        BySelector selector = new BySelector();
        selector.textStartsWith = prefix;
        return selector;
    }

    /**
     * Select by text ending with suffix
     */
    @NonNull
    public static BySelector textEndsWith(@NonNull String suffix) {
        BySelector selector = new BySelector();
        selector.textEndsWith = suffix;
        return selector;
    }

    /**
     * Select by resource ID
     */
    @NonNull
    public static BySelector id(@NonNull String resourceId) {
        BySelector selector = new BySelector();
        selector.resourceId = resourceId;
        return selector;
    }

    /**
     * Select by class name
     */
    @NonNull
    public static BySelector className(@NonNull String className) {
        BySelector selector = new BySelector();
        selector.className = className;
        return selector;
    }

    /**
     * Select by content description (exact)
     */
    @NonNull
    public static BySelector desc(@NonNull String description) {
        BySelector selector = new BySelector();
        selector.contentDescription = description;
        return selector;
    }

    /**
     * Select by content description containing substring
     */
    @NonNull
    public static BySelector descContains(@NonNull String description) {
        BySelector selector = new BySelector();
        selector.descContains = description;
        return selector;
    }

    /**
     * Select clickable elements
     */
    @NonNull
    public static BySelector clickable() {
        BySelector selector = new BySelector();
        selector.clickable = true;
        return selector;
    }

    /**
     * Select scrollable elements
     */
    @NonNull
    public static BySelector scrollable() {
        BySelector selector = new BySelector();
        selector.scrollable = true;
        return selector;
    }

    // ===== Modifier Methods =====

    /**
     * Set index (for multiple matching elements)
     */
    @NonNull
    public BySelector index(int index) {
        this.index = index;
        return this;
    }

    /**
     * Set clickable requirement
     */
    @NonNull
    public BySelector clickable(boolean clickable) {
        this.clickable = clickable;
        return this;
    }

    /**
     * Set enabled requirement
     */
    @NonNull
    public BySelector enabled(boolean enabled) {
        this.enabled = enabled;
        return this;
    }

    /**
     * Set focusable requirement
     */
    @NonNull
    public BySelector focusable(boolean focusable) {
        this.focusable = focusable;
        return this;
    }

    /**
     * Set scrollable requirement
     */
    @NonNull
    public BySelector scrollable(boolean scrollable) {
        this.scrollable = scrollable;
        return this;
    }

    /**
     * Combine with resource ID
     */
    @NonNull
    public BySelector andId(@NonNull String resourceId) {
        this.resourceId = resourceId;
        return this;
    }

    /**
     * Combine with class name
     */
    @NonNull
    public BySelector andClassName(@NonNull String className) {
        this.className = className;
        return this;
    }

    // ===== Getters =====

    @Nullable
    public String getText() {
        return text;
    }

    @Nullable
    public String getTextContains() {
        return textContains;
    }

    @Nullable
    public String getTextStartsWith() {
        return textStartsWith;
    }

    @Nullable
    public String getTextEndsWith() {
        return textEndsWith;
    }

    @Nullable
    public String getResourceId() {
        return resourceId;
    }

    @Nullable
    public String getClassName() {
        return className;
    }

    @Nullable
    public String getContentDescription() {
        return contentDescription;
    }

    @Nullable
    public String getDescContains() {
        return descContains;
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

    // Alternative selectors (OR condition)
    @Nullable
    private BySelector alternative;

    /**
     * Add alternative selector (OR condition)
     * Matches either this selector OR the alternative
     */
    @NonNull
    public BySelector or(@NonNull BySelector alternative) {
        this.alternative = alternative;
        return this;
    }

    @Nullable
    public BySelector getAlternative() {
        return alternative;
    }

    /**
     * Check if selector has any criteria
     */
    public boolean hasCriteria() {
        return text != null || textContains != null || textStartsWith != null ||
               textEndsWith != null || resourceId != null || className != null ||
               contentDescription != null || descContains != null || index >= 0 ||
               clickable || focusable || scrollable;
    }

    /**
     * Get description of selector criteria
     */
    @NonNull
    public String describe() {
        StringBuilder sb = new StringBuilder("BySelector[");
        if (text != null) sb.append("text=").append(text);
        if (textContains != null) sb.append("textContains=").append(textContains);
        if (textStartsWith != null) sb.append("textStartsWith=").append(textStartsWith);
        if (textEndsWith != null) sb.append("textEndsWith=").append(textEndsWith);
        if (resourceId != null) sb.append("id=").append(resourceId);
        if (className != null) sb.append("class=").append(className);
        if (contentDescription != null) sb.append("desc=").append(contentDescription);
        if (descContains != null) sb.append("descContains=").append(descContains);
        if (index >= 0) sb.append("index=").append(index);
        if (clickable) sb.append("clickable");
        if (!enabled) sb.append("!enabled");
        if (focusable) sb.append("focusable");
        if (scrollable) sb.append("scrollable");
        sb.append("]");
        return sb.toString();
    }

    @NonNull
    @Override
    public String toString() {
        return describe();
    }
}