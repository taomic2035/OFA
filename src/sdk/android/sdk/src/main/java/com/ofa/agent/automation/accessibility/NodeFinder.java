package com.ofa.agent.automation.accessibility;

import android.graphics.Rect;
import android.util.Log;
import android.view.accessibility.AccessibilityNodeInfo;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.BySelector;

import java.util.ArrayList;
import java.util.List;

/**
 * Node finder - locates UI elements using AccessibilityNodeInfo.
 */
public class NodeFinder {

    private static final String TAG = "NodeFinder";

    private final AccessibilityEngine engine;

    public NodeFinder(@NonNull AccessibilityEngine engine) {
        this.engine = engine;
    }

    /**
     * Find single element matching selector
     */
    @Nullable
    public AutomationNode findElement(@NonNull BySelector selector) {
        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            Log.w(TAG, "Service not available for findElement");
            return null;
        }

        AccessibilityNodeInfo root = service.getRootNode();
        if (root == null) {
            Log.w(TAG, "Root node not available");
            return null;
        }

        try {
            AccessibilityNodeInfo foundNode = findNodeRecursive(root, selector, false);
            if (foundNode != null) {
                return convertToAutomationNode(foundNode);
            }
        } finally {
            // Note: AccessibilityNodeInfo should not be recycled manually in modern Android
            // as they are managed by the system
        }

        return null;
    }

    /**
     * Find all elements matching selector
     */
    @NonNull
    public List<AutomationNode> findElements(@NonNull BySelector selector) {
        List<AutomationNode> result = new ArrayList<>();

        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            Log.w(TAG, "Service not available for findElements");
            return result;
        }

        AccessibilityNodeInfo root = service.getRootNode();
        if (root == null) {
            Log.w(TAG, "Root node not available");
            return result;
        }

        try {
            List<AccessibilityNodeInfo> foundNodes = findNodesRecursive(root, selector);
            for (AccessibilityNodeInfo node : foundNodes) {
                result.add(convertToAutomationNode(node));
            }
        } finally {
            // Nodes managed by system
        }

        return result;
    }

    /**
     * Get page source as XML-like string
     */
    @NonNull
    public String getPageSource() {
        StringBuilder sb = new StringBuilder();

        OFAAccessibilityService service = engine.getService();
        if (service == null) {
            return "<error>Service not available</error>";
        }

        AccessibilityNodeInfo root = service.getRootNode();
        if (root == null) {
            return "<error>Root node not available</error>";
        }

        appendNodeXml(root, sb, 0);
        return sb.toString();
    }

    // ===== Recursive Search Methods =====

    /**
     * Find single node recursively
     */
    @Nullable
    private AccessibilityNodeInfo findNodeRecursive(@NonNull AccessibilityNodeInfo node,
                                                     @NonNull BySelector selector,
                                                     boolean matchIndex) {
        // Check if this node matches
        if (matchesSelector(node, selector)) {
            if (!matchIndex || selector.getIndex() < 0) {
                return node;
            }
        }

        // Search children
        int childCount = node.getChildCount();
        for (int i = 0; i < childCount; i++) {
            AccessibilityNodeInfo child = node.getChild(i);
            if (child != null) {
                AccessibilityNodeInfo found = findNodeRecursive(child, selector, matchIndex);
                if (found != null) {
                    return found;
                }
            }
        }

        return null;
    }

    /**
     * Find all matching nodes recursively
     */
    @NonNull
    private List<AccessibilityNodeInfo> findNodesRecursive(@NonNull AccessibilityNodeInfo node,
                                                            @NonNull BySelector selector) {
        List<AccessibilityNodeInfo> result = new ArrayList<>();

        if (matchesSelector(node, selector)) {
            result.add(node);
        }

        int childCount = node.getChildCount();
        for (int i = 0; i < childCount; i++) {
            AccessibilityNodeInfo child = node.getChild(i);
            if (child != null) {
                result.addAll(findNodesRecursive(child, selector));
            }
        }

        return result;
    }

    /**
     * Check if node matches selector criteria
     */
    private boolean matchesSelector(@NonNull AccessibilityNodeInfo node, @NonNull BySelector selector) {
        // Text matching
        String nodeText = node.getText() != null ? node.getText().toString() : null;

        if (selector.getText() != null) {
            if (nodeText == null || !nodeText.equals(selector.getText())) return false;
        }
        if (selector.getTextContains() != null) {
            if (nodeText == null || !nodeText.contains(selector.getTextContains())) return false;
        }
        if (selector.getTextStartsWith() != null) {
            if (nodeText == null || !nodeText.startsWith(selector.getTextStartsWith())) return false;
        }
        if (selector.getTextEndsWith() != null) {
            if (nodeText == null || !nodeText.endsWith(selector.getTextEndsWith())) return false;
        }

        // Resource ID matching
        String nodeId = node.getViewIdResourceName();
        if (selector.getResourceId() != null) {
            if (nodeId == null || !nodeId.equals(selector.getResourceId())) {
                // Also check if it ends with the ID (for ID without package prefix)
                if (nodeId == null || !nodeId.endsWith("/" + selector.getResourceId())) return false;
            }
        }

        // Class name matching
        String nodeClassName = node.getClassName() != null ? node.getClassName().toString() : null;
        if (selector.getClassName() != null) {
            if (nodeClassName == null || !nodeClassName.equals(selector.getClassName())) return false;
        }

        // Content description matching
        String nodeDesc = node.getContentDescription() != null ?
                node.getContentDescription().toString() : null;

        if (selector.getContentDescription() != null) {
            if (nodeDesc == null || !nodeDesc.equals(selector.getContentDescription())) return false;
        }
        if (selector.getDescContains() != null) {
            if (nodeDesc == null || !nodeDesc.contains(selector.getDescContains())) return false;
        }

        // State matching
        if (selector.isClickable() && !node.isClickable()) return false;
        if (!selector.isEnabled() && node.isEnabled()) return false;
        if (selector.isFocusable() && !node.isFocusable()) return false;
        if (selector.isScrollable() && !node.isScrollable()) return false;

        return true;
    }

    // ===== Conversion Methods =====

    /**
     * Convert AccessibilityNodeInfo to AutomationNode
     */
    @NonNull
    private AutomationNode convertToAutomationNode(@NonNull AccessibilityNodeInfo node) {
        String className = node.getClassName() != null ? node.getClassName().toString() : null;
        String text = node.getText() != null ? node.getText().toString() : null;
        String contentDesc = node.getContentDescription() != null ?
                node.getContentDescription().toString() : null;
        String resourceId = node.getViewIdResourceName();
        String packageName = node.getPackageName() != null ?
                node.getPackageName().toString() : null;

        Rect bounds = new Rect();
        node.getBoundsInScreen(bounds);

        return new AutomationNode(
                className,
                text,
                contentDesc,
                resourceId,
                -1, // index not tracked for single node
                node.isClickable(),
                node.isEnabled(),
                node.isFocusable(),
                node.isScrollable(),
                node.isCheckable(),
                node.isChecked(),
                node.isVisibleToUser(),
                bounds.left,
                bounds.top,
                bounds.right,
                bounds.bottom,
                packageName
        );
    }

    /**
     * Append node as XML to string builder
     */
    private void appendNodeXml(@NonNull AccessibilityNodeInfo node, @NonNull StringBuilder sb, int depth) {
        String indent = "";
        for (int i = 0; i < depth; i++) indent += "  ";

        String className = node.getClassName() != null ?
                node.getClassName().toString().replace("$", ".") : "View";
        String text = node.getText() != null ? node.getText().toString() : "";
        String resourceId = node.getViewIdResourceName() != null ?
                node.getViewIdResourceName() : "";

        // Escape text for XML
        text = escapeXml(text);

        sb.append(indent).append("<").append(className);

        if (!resourceId.isEmpty()) {
            sb.append(" id=\"").append(escapeXml(resourceId)).append("\"");
        }
        if (!text.isEmpty()) {
            sb.append(" text=\"").append(text).append("\"");
        }

        String contentDesc = node.getContentDescription() != null ?
                node.getContentDescription().toString() : "";
        if (!contentDesc.isEmpty()) {
            sb.append(" desc=\"").append(escapeXml(contentDesc)).append("\"");
        }

        sb.append(" clickable=\"").append(node.isClickable()).append("\"");
        sb.append(" enabled=\"").append(node.isEnabled()).append("\"");
        sb.append(" scrollable=\"").append(node.isScrollable()).append("\"");

        Rect bounds = new Rect();
        node.getBoundsInScreen(bounds);
        sb.append(" bounds=\"[").append(bounds.left).append(",").append(bounds.top)
                .append("][").append(bounds.right).append(",").append(bounds.bottom).append("]\"");

        int childCount = node.getChildCount();
        if (childCount > 0) {
            sb.append(">\n");
            for (int i = 0; i < childCount; i++) {
                AccessibilityNodeInfo child = node.getChild(i);
                if (child != null) {
                    appendNodeXml(child, sb, depth + 1);
                }
            }
            sb.append(indent).append("</").append(className).append(">\n");
        } else {
            sb.append("/>\n");
        }
    }

    /**
     * Escape special characters for XML
     */
    @NonNull
    private String escapeXml(@NonNull String text) {
        return text
                .replace("&", "&amp;")
                .replace("<", "&lt;")
                .replace(">", "&gt;")
                .replace("\"", "&quot;")
                .replace("'", "&apos;");
    }
}