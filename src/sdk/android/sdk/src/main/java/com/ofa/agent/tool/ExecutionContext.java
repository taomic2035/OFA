package com.ofa.agent.tool;

import android.content.Context;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.BatteryManager;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.constraint.ConstraintResult;
import com.ofa.agent.offline.OfflineLevel;

/**
 * Execution Context - context information for tool execution.
 * Provides offline state, device status, and constraint information.
 */
public class ExecutionContext {

    private final Context context;
    private final boolean isOffline;
    private final OfflineLevel offlineLevel;
    private final NetworkInfo networkInfo;
    private final BatteryInfo batteryInfo;
    private final ConstraintResult constraints;
    private final long timeoutMs;
    private final String executionId;
    private final Callback callback;
    private final boolean permissionGranted;
    private final String userId;

    /**
     * Create execution context
     */
    public ExecutionContext(@NonNull Context context, boolean isOffline,
                            @NonNull OfflineLevel offlineLevel,
                            @Nullable NetworkInfo networkInfo,
                            @Nullable BatteryInfo batteryInfo,
                            @Nullable ConstraintResult constraints,
                            long timeoutMs, @Nullable String executionId,
                            @Nullable Callback callback, boolean permissionGranted,
                            @Nullable String userId) {
        this.context = context.getApplicationContext();
        this.isOffline = isOffline;
        this.offlineLevel = offlineLevel;
        this.networkInfo = networkInfo;
        this.batteryInfo = batteryInfo;
        this.constraints = constraints;
        this.timeoutMs = timeoutMs;
        this.executionId = executionId;
        this.callback = callback;
        this.permissionGranted = permissionGranted;
        this.userId = userId;
    }

    /**
     * Simple builder-based creation
     */
    public static Builder builder(@NonNull Context context) {
        return new Builder(context);
    }

    @NonNull
    public Context getContext() {
        return context;
    }

    public boolean isOffline() {
        return isOffline;
    }

    @NonNull
    public OfflineLevel getOfflineLevel() {
        return offlineLevel;
    }

    @Nullable
    public NetworkInfo getNetworkInfo() {
        return networkInfo;
    }

    @Nullable
    public BatteryInfo getBatteryInfo() {
        return batteryInfo;
    }

    @Nullable
    public ConstraintResult getConstraints() {
        return constraints;
    }

    public long getTimeoutMs() {
        return timeoutMs;
    }

    @Nullable
    public String getExecutionId() {
        return executionId;
    }

    @Nullable
    public Callback getCallback() {
        return callback;
    }

    public boolean isPermissionGranted() {
        return permissionGranted;
    }

    @Nullable
    public String getUserId() {
        return userId;
    }

    /**
     * Check if network is available
     */
    public boolean hasNetwork() {
        if (networkInfo == null) return false;
        return networkInfo.isConnected();
    }

    /**
     * Check if battery level is sufficient
     */
    public boolean hasSufficientBattery(int minLevel) {
        if (batteryInfo == null) return true;
        return batteryInfo.level >= minLevel;
    }

    /**
     * Report progress during execution
     */
    public void reportProgress(int progress, @Nullable String message) {
        if (callback != null) {
            callback.onProgress(executionId, progress, message);
        }
    }

    /**
     * Request cancellation check
     */
    public boolean shouldCancel() {
        if (callback != null) {
            return callback.shouldCancel(executionId);
        }
        return false;
    }

    /**
     * Builder class
     */
    public static class Builder {
        private final Context context;
        private boolean isOffline = false;
        private OfflineLevel offlineLevel = OfflineLevel.L4;
        private NetworkInfo networkInfo = null;
        private BatteryInfo batteryInfo = null;
        private ConstraintResult constraints = null;
        private long timeoutMs = 30000;
        private String executionId = null;
        private Callback callback = null;
        private boolean permissionGranted = true;
        private String userId = null;

        public Builder(@NonNull Context context) {
            this.context = context.getApplicationContext();
        }

        public Builder offline(boolean isOffline) {
            this.isOffline = isOffline;
            return this;
        }

        public Builder offlineLevel(@NonNull OfflineLevel level) {
            this.offlineLevel = level;
            return this;
        }

        public Builder networkInfo(@Nullable NetworkInfo info) {
            this.networkInfo = info;
            return this;
        }

        public Builder batteryInfo(@Nullable BatteryInfo info) {
            this.batteryInfo = info;
            return this;
        }

        public Builder constraints(@Nullable ConstraintResult result) {
            this.constraints = result;
            return this;
        }

        public Builder timeoutMs(long timeoutMs) {
            this.timeoutMs = timeoutMs;
            return this;
        }

        public Builder executionId(@Nullable String id) {
            this.executionId = id;
            return this;
        }

        public Builder callback(@Nullable Callback callback) {
            this.callback = callback;
            return this;
        }

        public Builder permissionGranted(boolean granted) {
            this.permissionGranted = granted;
            return this;
        }

        public Builder userId(@Nullable String userId) {
            this.userId = userId;
            return this;
        }

        /**
         * Auto-populate device info
         */
        public Builder autoPopulateDeviceInfo() {
            // Network info
            ConnectivityManager cm = (ConnectivityManager) context
                    .getSystemService(Context.CONNECTIVITY_SERVICE);
            if (cm != null) {
                this.networkInfo = cm.getActiveNetworkInfo();
                this.isOffline = (networkInfo == null || !networkInfo.isConnected());
            }

            // Battery info
            BatteryManager bm = (BatteryManager) context
                    .getSystemService(Context.BATTERY_SERVICE);
            if (bm != null) {
                int level = bm.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY);
                boolean charging = bm.isCharging();
                this.batteryInfo = new BatteryInfo(level, charging);
            }

            return this;
        }

        @NonNull
        public ExecutionContext build() {
            return new ExecutionContext(context, isOffline, offlineLevel, networkInfo,
                    batteryInfo, constraints, timeoutMs, executionId, callback,
                    permissionGranted, userId);
        }
    }

    /**
     * Execution callback interface
     */
    public interface Callback {
        void onProgress(@Nullable String executionId, int progress, @Nullable String message);
        boolean shouldCancel(@Nullable String executionId);
    }

    /**
     * Battery information
     */
    public static class BatteryInfo {
        public final int level;
        public final boolean charging;

        public BatteryInfo(int level, boolean charging) {
            this.level = level;
            this.charging = charging;
        }

        public boolean isLow() {
            return level < 20;
        }

        public boolean isCritical() {
            return level < 10;
        }
    }
}