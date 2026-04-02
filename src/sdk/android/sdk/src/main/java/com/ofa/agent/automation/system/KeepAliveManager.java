package com.ofa.agent.automation.system;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.Service;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.os.Build;
import android.os.IBinder;
import android.os.PowerManager;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.app.NotificationCompat;

import com.ofa.agent.automation.AutomationResult;

import java.util.ArrayList;
import java.util.List;

/**
 * Keep Alive Manager - manages background service keep-alive strategies.
 * Provides multiple techniques to prevent the service from being killed.
 */
public class KeepAliveManager {

    private static final String TAG = "KeepAliveManager";

    private final Context context;
    private final SystemPermissionManager permissionManager;

    // Keep-alive strategies
    private boolean foregroundServiceEnabled = false;
    private boolean wakeLockEnabled = false;
    private boolean batteryOptimizationWhitelisted = false;
    private final List<KeepAliveStrategy> strategies = new ArrayList<>();

    // Wake lock reference
    private PowerManager.WakeLock wakeLock;

    // Notification constants
    private static final String CHANNEL_ID = "ofa_keepalive_channel";
    private static final int NOTIFICATION_ID = 10001;

    /**
     * Keep-alive strategy
     */
    public static class KeepAliveStrategy {
        public final String name;
        public final boolean available;
        public final boolean enabled;
        public final String description;

        public KeepAliveStrategy(String name, boolean available, boolean enabled, String description) {
            this.name = name;
            this.available = available;
            this.enabled = enabled;
            this.description = description;
        }
    }

    /**
     * Create keep-alive manager
     */
    public KeepAliveManager(@NonNull Context context,
                            @NonNull SystemPermissionManager permissionManager) {
        this.context = context;
        this.permissionManager = permissionManager;
        initializeStrategies();
    }

    /**
     * Initialize available strategies
     */
    private void initializeStrategies() {
        Log.i(TAG, "Initializing keep-alive strategies...");

        // 1. Foreground Service
        strategies.add(new KeepAliveStrategy(
            "ForegroundService",
            Build.VERSION.SDK_INT >= Build.VERSION_CODES.O,
            foregroundServiceEnabled,
            "Show persistent notification to keep service running"));

        // 2. Wake Lock
        strategies.add(new KeepAliveStrategy(
            "WakeLock",
            true,
            wakeLockEnabled,
            "Hold CPU wake lock to prevent sleep"));

        // 3. Battery Optimization Whitelist
        strategies.add(new KeepAliveStrategy(
            "BatteryOptimization",
            Build.VERSION.SDK_INT >= Build.VERSION_CODES.M,
            batteryOptimizationWhitelisted,
            "Request exemption from battery optimization"));

        // 4. System App (ROM)
        strategies.add(new KeepAliveStrategy(
            "SystemApp",
            permissionManager.isSystemApp(),
            permissionManager.isSystemApp(),
            "Running as system app provides inherent protection"));

        // 5. Root Keep Alive
        strategies.add(new KeepAliveStrategy(
            "RootKeepAlive",
            permissionManager.checkRootAccess(),
            false,
            "Use root to modify system kill policies"));

        Log.i(TAG, "Available strategies: " + getAvailableStrategies().size());
    }

    /**
     * Enable all available keep-alive strategies
     */
    @NonNull
    public AutomationResult enableAll() {
        Log.i(TAG, "Enabling all keep-alive strategies...");

        List<String> results = new ArrayList<>();

        // Enable foreground service
        if (enableForegroundService()) {
            results.add("ForegroundService: enabled");
        }

        // Enable wake lock
        if (enableWakeLock()) {
            results.add("WakeLock: enabled");
        }

        // Request battery optimization exemption
        if (requestBatteryOptimizationWhitelist()) {
            results.add("BatteryOptimization: requested");
        }

        // Apply root policies if available
        if (permissionManager.checkRootAccess()) {
            if (applyRootKeepAlivePolicies()) {
                results.add("RootKeepAlive: applied");
            }
        }

        if (results.isEmpty()) {
            return new AutomationResult("keepAlive", "No strategies available");
        }

        Log.i(TAG, "Keep-alive enabled: " + results.toString());
        return new AutomationResult("keepAlive", 0);
    }

    /**
     * Disable all keep-alive strategies
     */
    @NonNull
    public AutomationResult disableAll() {
        Log.i(TAG, "Disabling all keep-alive strategies...");

        disableWakeLock();
        foregroundServiceEnabled = false;

        return new AutomationResult("keepAlive", 0);
    }

    /**
     * Enable foreground service
     */
    public boolean enableForegroundService() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) {
            return true; // Not needed on older versions
        }

        foregroundServiceEnabled = true;
        Log.i(TAG, "Foreground service strategy enabled");
        return true;
    }

    /**
     * Create notification for foreground service
     */
    @NonNull
    public Notification createForegroundNotification(@NonNull String title, @NonNull String message) {
        createNotificationChannel();

        NotificationCompat.Builder builder = new NotificationCompat.Builder(context, CHANNEL_ID)
            .setContentTitle(title)
            .setContentText(message)
            .setSmallIcon(android.R.drawable.ic_menu_manage)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setOngoing(true)
            .setShowWhen(false);

        return builder.build();
    }

    /**
     * Create notification channel
     */
    private void createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID,
                "OFA Keep Alive",
                NotificationManager.IMPORTANCE_LOW);
            channel.setDescription("Keeps OFA service running in background");
            channel.setShowBadge(false);

            NotificationManager manager = context.getSystemService(NotificationManager.class);
            if (manager != null) {
                manager.createNotificationChannel(channel);
            }
        }
    }

    /**
     * Enable wake lock
     */
    public boolean enableWakeLock() {
        if (wakeLockEnabled) {
            return true;
        }

        try {
            PowerManager powerManager = (PowerManager) context.getSystemService(Context.POWER_SERVICE);
            if (powerManager != null) {
                wakeLock = powerManager.newWakeLock(
                    PowerManager.PARTIAL_WAKE_LOCK,
                    "OFA::KeepAlive");

                wakeLock.acquire(10 * 60 * 1000L); // 10 minutes max
                wakeLockEnabled = true;
                Log.i(TAG, "Wake lock acquired");
                return true;
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to acquire wake lock: " + e.getMessage());
        }

        return false;
    }

    /**
     * Disable wake lock
     */
    public void disableWakeLock() {
        if (wakeLock != null && wakeLockEnabled) {
            try {
                wakeLock.release();
                wakeLockEnabled = false;
                Log.i(TAG, "Wake lock released");
            } catch (Exception e) {
                Log.w(TAG, "Failed to release wake lock: " + e.getMessage());
            }
        }
    }

    /**
     * Request battery optimization whitelist
     */
    public boolean requestBatteryOptimizationWhitelist() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.M) {
            return true;
        }

        try {
            PowerManager powerManager = (PowerManager) context.getSystemService(Context.POWER_SERVICE);
            if (powerManager != null) {
                if (powerManager.isIgnoringBatteryOptimizations(context.getPackageName())) {
                    batteryOptimizationWhitelisted = true;
                    Log.i(TAG, "Already whitelisted from battery optimization");
                    return true;
                }

                // Request whitelist
                Intent intent = new Intent(android.provider.Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS);
                intent.setData(android.net.Uri.parse("package:" + context.getPackageName()));
                intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
                context.startActivity(intent);

                Log.i(TAG, "Battery optimization whitelist requested");
                return true;
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to request battery optimization whitelist: " + e.getMessage());
        }

        return false;
    }

    /**
     * Apply root keep-alive policies
     */
    public boolean applyRootKeepAlivePolicies() {
        if (!permissionManager.checkRootAccess()) {
            return false;
        }

        try {
            List<String> commands = new ArrayList<>();

            // Add to doze whitelist
            commands.add("dumpsys deviceidle whitelist +" + context.getPackageName());

            // Set as foreground service priority
            commands.add("am set-standby-bucket " + context.getPackageName() + " active");

            // Reduce OOM adjustment
            commands.add("echo -17 > /proc/$(pidof " + context.getPackageName() + ")/oom_adj");

            for (String command : commands) {
                Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
                int exitCode = process.waitFor();
                Log.d(TAG, "Root command: " + command + " -> " + exitCode);
            }

            Log.i(TAG, "Root keep-alive policies applied");
            return true;
        } catch (Exception e) {
            Log.e(TAG, "Failed to apply root policies: " + e.getMessage());
            return false;
        }
    }

    /**
     * Check if battery optimization is whitelisted
     */
    public boolean isBatteryOptimizationWhitelisted() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            PowerManager powerManager = (PowerManager) context.getSystemService(Context.POWER_SERVICE);
            if (powerManager != null) {
                return powerManager.isIgnoringBatteryOptimizations(context.getPackageName());
            }
        }
        return true; // Pre-M devices don't have battery optimization
    }

    /**
     * Get all strategies
     */
    @NonNull
    public List<KeepAliveStrategy> getAllStrategies() {
        return new ArrayList<>(strategies);
    }

    /**
     * Get available strategies
     */
    @NonNull
    public List<KeepAliveStrategy> getAvailableStrategies() {
        List<KeepAliveStrategy> available = new ArrayList<>();
        for (KeepAliveStrategy strategy : strategies) {
            if (strategy.available) {
                available.add(strategy);
            }
        }
        return available;
    }

    /**
     * Get enabled strategies
     */
    @NonNull
    public List<KeepAliveStrategy> getEnabledStrategies() {
        List<KeepAliveStrategy> enabled = new ArrayList<>();
        for (KeepAliveStrategy strategy : strategies) {
            if (strategy.enabled) {
                enabled.add(strategy);
            }
        }
        return enabled;
    }

    /**
     * Get keep-alive status report
     */
    @NonNull
    public String getStatusReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("Keep-Alive Status:\n");

        for (KeepAliveStrategy strategy : strategies) {
            sb.append("  ").append(strategy.name).append(": ");
            sb.append(strategy.available ? "Available" : "Not Available");
            sb.append(", ").append(strategy.enabled ? "Enabled" : "Disabled");
            sb.append("\n");
        }

        sb.append("\nCurrent State:\n");
        sb.append("  Foreground Service: ").append(foregroundServiceEnabled).append("\n");
        sb.append("  Wake Lock: ").append(wakeLockEnabled).append("\n");
        sb.append("  Battery Whitelisted: ").append(isBatteryOptimizationWhitelisted()).append("\n");

        return sb.toString();
    }

    /**
     * Start service as foreground
     */
    public void startForegroundService(@NonNull Service service, @NonNull String title, @NonNull String message) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && foregroundServiceEnabled) {
            Notification notification = createForegroundNotification(title, message);
            service.startForeground(NOTIFICATION_ID, notification);
            Log.i(TAG, "Service started as foreground");
        }
    }

    /**
     * Refresh wake lock (extend duration)
     */
    public void refreshWakeLock() {
        if (wakeLockEnabled && wakeLock != null) {
            wakeLock.acquire(10 * 60 * 1000L); // Extend by 10 minutes
            Log.d(TAG, "Wake lock refreshed");
        }
    }
}