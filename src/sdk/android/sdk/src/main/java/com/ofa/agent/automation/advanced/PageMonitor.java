package com.ofa.agent.automation.advanced;

import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationListener;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Page Monitor - monitors page changes and stability.
 * Detects when pages load, change, or stabilize.
 */
public class PageMonitor implements AutomationListener {

    private static final String TAG = "PageMonitor";

    private final AutomationEngine engine;
    private final Handler handler;
    private final List<PageChangeListener> listeners = new ArrayList<>();

    // Monitoring state
    private final AtomicBoolean monitoring = new AtomicBoolean(false);
    private final AtomicLong lastChangeTime = new AtomicLong(0);
    private volatile String lastPageSource;
    private volatile String lastPackageName;
    private volatile String lastActivityName;

    // Configuration
    private long checkInterval = 500; // ms
    private long stableThreshold = 1000; // ms without change to be considered stable
    private int maxHistorySize = 100;

    // Page history for debugging/analysis
    private final List<PageSnapshot> pageHistory = new ArrayList<>();

    /**
     * Page snapshot for history
     */
    public static class PageSnapshot {
        public final long timestamp;
        public final String packageName;
        public final String activityName;
        public final int nodeCount;
        public final String sourceHash;

        public PageSnapshot(long timestamp, String packageName, String activityName,
                           int nodeCount, String sourceHash) {
            this.timestamp = timestamp;
            this.packageName = packageName;
            this.activityName = activityName;
            this.nodeCount = nodeCount;
            this.sourceHash = sourceHash;
        }
    }

    /**
     * Page change listener interface
     */
    public interface PageChangeListener {
        void onPageChanged(String packageName, String activityName);
        void onPageStable();
        void onPackageChanged(String oldPackage, String newPackage);
    }

    public PageMonitor(@NonNull AutomationEngine engine) {
        this.engine = engine;
        this.handler = new Handler(Looper.getMainLooper());
    }

    /**
     * Start monitoring page changes
     */
    public void startMonitoring() {
        if (monitoring.compareAndSet(false, true)) {
            Log.i(TAG, "Starting page monitoring");
            lastChangeTime.set(System.currentTimeMillis());
            lastPageSource = engine.getPageSource();
            lastPackageName = engine.getForegroundPackage();

            // Start periodic check
            handler.postDelayed(this::checkPageChange, checkInterval);
        }
    }

    /**
     * Stop monitoring page changes
     */
    public void stopMonitoring() {
        if (monitoring.compareAndSet(true, false)) {
            Log.i(TAG, "Stopping page monitoring");
            handler.removeCallbacksAndMessages(null);
        }
    }

    /**
     * Check if monitoring is active
     */
    public boolean isMonitoring() {
        return monitoring.get();
    }

    /**
     * Add page change listener
     */
    public void addListener(@NonNull PageChangeListener listener) {
        synchronized (listeners) {
            if (!listeners.contains(listener)) {
                listeners.add(listener);
            }
        }
    }

    /**
     * Remove page change listener
     */
    public void removeListener(@NonNull PageChangeListener listener) {
        synchronized (listeners) {
            listeners.remove(listener);
        }
    }

    /**
     * Wait for page to stabilize
     * @param timeoutMs Maximum wait time
     * @return true if page stabilized, false if timeout
     */
    public boolean waitForStable(long timeoutMs) {
        Log.i(TAG, "Waiting for page stable, timeout=" + timeoutMs + "ms");

        long startTime = System.currentTimeMillis();
        String previousSource = engine.getPageSource();
        long lastChangeTime = startTime;

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            sleep(200);

            String currentSource = engine.getPageSource();
            if (!currentSource.equals(previousSource)) {
                previousSource = currentSource;
                lastChangeTime = System.currentTimeMillis();
            } else {
                // Check if stable threshold reached
                if (System.currentTimeMillis() - lastChangeTime >= stableThreshold) {
                    Log.i(TAG, "Page stable after " + (System.currentTimeMillis() - startTime) + "ms");
                    return true;
                }
            }
        }

        Log.w(TAG, "Page did not stabilize within timeout");
        return false;
    }

    /**
     * Wait for page to change
     * @param timeoutMs Maximum wait time
     * @return true if page changed, false if timeout
     */
    public boolean waitForChange(long timeoutMs) {
        Log.i(TAG, "Waiting for page change, timeout=" + timeoutMs + "ms");

        long startTime = System.currentTimeMillis();
        String initialSource = engine.getPageSource();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            sleep(200);

            String currentSource = engine.getPageSource();
            if (!currentSource.equals(initialSource)) {
                Log.i(TAG, "Page changed after " + (System.currentTimeMillis() - startTime) + "ms");
                return true;
            }
        }

        Log.w(TAG, "Page did not change within timeout");
        return false;
    }

    /**
     * Wait for specific package to be in foreground
     */
    public boolean waitForPackage(@NonNull String packageName, long timeoutMs) {
        Log.i(TAG, "Waiting for package: " + packageName);

        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            sleep(200);

            String currentPackage = engine.getForegroundPackage();
            if (packageName.equals(currentPackage)) {
                Log.i(TAG, "Package appeared after " + (System.currentTimeMillis() - startTime) + "ms");
                return true;
            }
        }

        Log.w(TAG, "Package not found within timeout");
        return false;
    }

    /**
     * Get page history
     */
    @NonNull
    public List<PageSnapshot> getPageHistory() {
        synchronized (pageHistory) {
            return new ArrayList<>(pageHistory);
        }
    }

    /**
     * Clear page history
     */
    public void clearHistory() {
        synchronized (pageHistory) {
            pageHistory.clear();
        }
    }

    /**
     * Get last page snapshot
     */
    @Nullable
    public PageSnapshot getLastSnapshot() {
        synchronized (pageHistory) {
            if (pageHistory.isEmpty()) {
                return null;
            }
            return pageHistory.get(pageHistory.size() - 1);
        }
    }

    /**
     * Set monitoring interval
     */
    public void setCheckInterval(long intervalMs) {
        this.checkInterval = intervalMs;
    }

    /**
     * Set stability threshold
     */
    public void setStableThreshold(long thresholdMs) {
        this.stableThreshold = thresholdMs;
    }

    // ===== Internal Methods =====

    private void checkPageChange() {
        if (!monitoring.get()) return;

        try {
            String currentSource = engine.getPageSource();
            String currentPackage = engine.getForegroundPackage();
            long currentTime = System.currentTimeMillis();

            // Check for package change
            if (lastPackageName != null && !lastPackageName.equals(currentPackage)) {
                Log.d(TAG, "Package changed: " + lastPackageName + " -> " + currentPackage);
                notifyPackageChanged(lastPackageName, currentPackage);
            }

            // Check for page source change
            if (lastPageSource != null && !lastPageSource.equals(currentSource)) {
                Log.d(TAG, "Page source changed");
                lastChangeTime.set(currentTime);
                notifyPageChanged(currentPackage, null);

                // Add to history
                addToHistory(currentPackage, null, currentSource);
            } else {
                // Check for stability
                if (currentTime - lastChangeTime.get() >= stableThreshold) {
                    notifyPageStable();
                }
            }

            lastPageSource = currentSource;
            lastPackageName = currentPackage;

        } catch (Exception e) {
            Log.e(TAG, "Error checking page change", e);
        }

        // Schedule next check
        if (monitoring.get()) {
            handler.postDelayed(this::checkPageChange, checkInterval);
        }
    }

    private void addToHistory(String packageName, String activityName, String source) {
        synchronized (pageHistory) {
            int nodeCount = countNodes(source);
            String hash = Integer.toHexString(source.hashCode());

            PageSnapshot snapshot = new PageSnapshot(
                System.currentTimeMillis(),
                packageName,
                activityName,
                nodeCount,
                hash
            );

            pageHistory.add(snapshot);

            // Trim history if needed
            while (pageHistory.size() > maxHistorySize) {
                pageHistory.remove(0);
            }
        }
    }

    private int countNodes(String source) {
        if (source == null || source.isEmpty()) return 0;
        // Simple count of opening tags
        int count = 0;
        for (int i = 0; i < source.length(); i++) {
            if (source.charAt(i) == '<' &&
                i + 1 < source.length() &&
                source.charAt(i + 1) != '/' &&
                source.charAt(i + 1) != '?') {
                count++;
            }
        }
        return count;
    }

    private void notifyPageChanged(String packageName, String activityName) {
        synchronized (listeners) {
            for (PageChangeListener listener : listeners) {
                try {
                    listener.onPageChanged(packageName, activityName);
                } catch (Exception e) {
                    Log.e(TAG, "Error notifying listener", e);
                }
            }
        }
    }

    private void notifyPageStable() {
        synchronized (listeners) {
            for (PageChangeListener listener : listeners) {
                try {
                    listener.onPageStable();
                } catch (Exception e) {
                    Log.e(TAG, "Error notifying listener", e);
                }
            }
        }
    }

    private void notifyPackageChanged(String oldPackage, String newPackage) {
        synchronized (listeners) {
            for (PageChangeListener listener : listeners) {
                try {
                    listener.onPackageChanged(oldPackage, newPackage);
                } catch (Exception e) {
                    Log.e(TAG, "Error notifying listener", e);
                }
            }
        }
    }

    private void sleep(long ms) {
        try {
            Thread.sleep(ms);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    // ===== AutomationListener Implementation =====

    @Override
    public void onPageChange(@Nullable String packageName, @Nullable String activityName) {
        lastPackageName = packageName;
        lastActivityName = activityName;
        lastChangeTime.set(System.currentTimeMillis());
    }
}