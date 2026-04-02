package com.ofa.agent.automation.accessibility;

import android.accessibilityservice.AccessibilityService;
import android.accessibilityservice.AccessibilityServiceInfo;
import android.content.Intent;
import android.util.Log;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityNodeInfo;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationListener;

/**
 * OFA Accessibility Service implementation.
 * Provides UI automation capabilities through Android's accessibility framework.
 *
 * Required configuration in AndroidManifest:
 * <service
 *     android:name=".automation.accessibility.OFAAccessibilityService"
 *     android:permission="android.permission.BIND_ACCESSIBILITY_SERVICE"
 *     android:exported="true">
 *     <intent-filter>
 *         <action android:name="android.accessibilityservice.AccessibilityService" />
 *     </intent-filter>
 *     <meta-data
 *         android:name="android.accessibilityservice"
 *         android:resource="@xml/accessibility_config" />
 * </service>
 *
 * And accessibility_config.xml in res/xml:
 * <accessibility-service
 *     xmlns:android="http://schemas.android.com/apk/res/android"
 *     android:description="@string/accessibility_service_description"
 *     android:accessibilityEventTypes="typeAllMask"
 *     android:accessibilityFeedbackType="feedbackGeneric"
 *     android:canPerformGestures="true"
 *     android:canRetrieveWindowContent="true"
 *     android:notificationTimeout="100"
 *     android:settingsActivity="" />
 */
public class OFAAccessibilityService extends AccessibilityService {

    private static final String TAG = "OFAAccessibilityService";

    // Singleton instance for access
    @Nullable
    private static OFAAccessibilityService instance;

    // Engine reference
    @Nullable
    private AccessibilityEngine engine;

    // Listener for events
    @Nullable
    private AutomationListener listener;

    // Flag to track service state
    private volatile boolean serviceEnabled = false;

    /**
     * Get the singleton instance
     */
    @Nullable
    public static OFAAccessibilityService getInstance() {
        return instance;
    }

    /**
     * Check if service is running
     */
    public static boolean isRunning() {
        return instance != null && instance.serviceEnabled;
    }

    /**
     * Set the automation engine
     */
    public void setEngine(@Nullable AccessibilityEngine engine) {
        this.engine = engine;
        if (engine != null) {
            engine.setService(this);
        }
    }

    /**
     * Set event listener
     */
    public void setListener(@Nullable AutomationListener listener) {
        this.listener = listener;
    }

    // ===== AccessibilityService Lifecycle =====

    @Override
    public void onCreate() {
        super.onCreate();
        Log.i(TAG, "Accessibility service created");
    }

    @Override
    public void onServiceConnected() {
        super.onServiceConnected();
        instance = this;
        serviceEnabled = true;

        Log.i(TAG, "Accessibility service connected");

        // Configure service capabilities
        AccessibilityServiceInfo info = getServiceInfo();
        if (info != null) {
            Log.i(TAG, "Service capabilities: " + info.getCapabilities());
        }

        // Notify engine
        if (engine != null) {
            engine.setService(this);
        }

        if (listener != null) {
            listener.onAccessibilityServiceStateChanged(true);
        }
    }

    @Override
    public void onAccessibilityEvent(@NonNull AccessibilityEvent event) {
        // Forward events to engine
        if (engine != null) {
            engine.updateForegroundPackage(event);
        }

        // Log significant events
        int eventType = event.getEventType();
        if (eventType == AccessibilityEvent.TYPE_WINDOW_STATE_CHANGED ||
            eventType == AccessibilityEvent.TYPE_WINDOW_CONTENT_CHANGED) {
            String packageName = event.getPackageName() != null ?
                    event.getPackageName().toString() : "unknown";
            String className = event.getClassName() != null ?
                    event.getClassName().toString() : "";

            Log.d(TAG, "Event: " + AccessibilityEvent.eventTypeToString(eventType) +
                    ", package: " + packageName + ", class: " + className);
        }
    }

    @Override
    public void onInterrupt() {
        Log.w(TAG, "Accessibility service interrupted");
        serviceEnabled = false;

        if (listener != null) {
            listener.onAccessibilityServiceStateChanged(false);
        }
    }

    @Override
    public boolean onUnbind(@Nullable Intent intent) {
        Log.i(TAG, "Accessibility service unbound");
        instance = null;
        serviceEnabled = false;

        // Notify engine
        if (engine != null) {
            engine.setService(null);
        }

        if (listener != null) {
            listener.onAccessibilityServiceStateChanged(false);
        }

        return super.onUnbind(intent);
    }

    @Override
    public void onDestroy() {
        Log.i(TAG, "Accessibility service destroyed");
        instance = null;
        serviceEnabled = false;
        super.onDestroy();
    }

    // ===== Helper Methods =====

    /**
     * Get root node in active window
     */
    @Nullable
    public AccessibilityNodeInfo getRootNode() {
        return getRootInActiveWindow();
    }

    /**
     * Check if service can perform gestures
     */
    public boolean canPerformGestures() {
        AccessibilityServiceInfo info = getServiceInfo();
        if (info != null) {
            return (info.getCapabilities() &
                    AccessibilityServiceInfo.CAPABILITY_CAN_PERFORM_GESTURES) != 0;
        }
        return false;
    }

    /**
     * Check if service can retrieve window content
     */
    public boolean canRetrieveWindowContent() {
        AccessibilityServiceInfo info = getServiceInfo();
        if (info != null) {
            return (info.getCapabilities() &
                    AccessibilityServiceInfo.CAPABILITY_CAN_RETRIEVE_WINDOW_CONTENT) != 0;
        }
        return false;
    }

    /**
     * Request service configuration update
     */
    public void updateServiceInfo(@NonNull AccessibilityServiceInfo newInfo) {
        setServiceInfo(newInfo);
        Log.i(TAG, "Service info updated");
    }

    /**
     * Get current service info
     */
    @Nullable
    public AccessibilityServiceInfo getCurrentServiceInfo() {
        return getServiceInfo();
    }
}