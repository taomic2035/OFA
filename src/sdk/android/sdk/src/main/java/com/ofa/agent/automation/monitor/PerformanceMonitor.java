package com.ofa.agent.automation.monitor;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Performance Monitor - tracks automation operation performance metrics.
 */
public class PerformanceMonitor {

    private static final String TAG = "PerformanceMonitor";

    private final Map<String, OperationStats> operationStats;
    private final List<OperationRecord> recentOperations;
    private final int maxRecentOperations;
    private final boolean enabled;

    private PerformanceListener listener;

    /**
     * Performance listener
     */
    public interface PerformanceListener {
        void onOperationComplete(@NonNull OperationRecord record);
        void onSlowOperation(@NonNull OperationRecord record, long threshold);
        void onStatsUpdated(@NonNull String operation, @NonNull OperationStats stats);
    }

    /**
     * Operation statistics
     */
    public static class OperationStats {
        public final String operation;
        public int count;
        public int successCount;
        public int failureCount;
        public long totalTimeMs;
        public long minTimeMs = Long.MAX_VALUE;
        public long maxTimeMs = 0;
        public long lastTimeMs;

        public OperationStats(String operation) {
            this.operation = operation;
        }

        public void record(long durationMs, boolean success) {
            count++;
            if (success) successCount++;
            else failureCount++;
            totalTimeMs += durationMs;
            minTimeMs = Math.min(minTimeMs, durationMs);
            maxTimeMs = Math.max(maxTimeMs, durationMs);
            lastTimeMs = durationMs;
        }

        public double getSuccessRate() {
            return count > 0 ? (double) successCount / count : 0;
        }

        public double getAverageTimeMs() {
            return count > 0 ? (double) totalTimeMs / count : 0;
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("%s: count=%d, success=%.1f%%, avg=%.1fms, min=%dms, max=%dms",
                operation, count, getSuccessRate() * 100, getAverageTimeMs(), minTimeMs, maxTimeMs);
        }
    }

    /**
     * Operation record
     */
    public static class OperationRecord {
        public final String operation;
        public final long startTime;
        public final long endTime;
        public final long durationMs;
        public final boolean success;
        public final String errorMessage;
        public final Map<String, String> metadata;

        public OperationRecord(String operation, long startTime, long endTime,
                               boolean success, String errorMessage,
                               Map<String, String> metadata) {
            this.operation = operation;
            this.startTime = startTime;
            this.endTime = endTime;
            this.durationMs = endTime - startTime;
            this.success = success;
            this.errorMessage = errorMessage;
            this.metadata = metadata;
        }
    }

    /**
     * Create performance monitor
     */
    public PerformanceMonitor() {
        this(100, true);
    }

    /**
     * Create performance monitor with settings
     */
    public PerformanceMonitor(int maxRecentOperations, boolean enabled) {
        this.operationStats = new ConcurrentHashMap<>();
        this.recentOperations = new ArrayList<>();
        this.maxRecentOperations = maxRecentOperations;
        this.enabled = enabled;
    }

    /**
     * Set listener
     */
    public void setListener(@Nullable PerformanceListener listener) {
        this.listener = listener;
    }

    /**
     * Start timing an operation
     */
    @NonNull
    public OperationTimer startOperation(@NonNull String operation) {
        return new OperationTimer(this, operation);
    }

    /**
     * Record an operation
     */
    public void recordOperation(@NonNull String operation, long durationMs, boolean success,
                                 @Nullable String errorMessage, @Nullable Map<String, String> metadata) {
        if (!enabled) return;

        // Update stats
        OperationStats stats = operationStats.computeIfAbsent(operation, OperationStats::new);
        stats.record(durationMs, success);

        // Add to recent operations
        OperationRecord record = new OperationRecord(
            operation,
            System.currentTimeMillis() - durationMs,
            System.currentTimeMillis(),
            success,
            errorMessage,
            metadata
        );

        synchronized (recentOperations) {
            recentOperations.add(record);
            while (recentOperations.size() > maxRecentOperations) {
                recentOperations.remove(0);
            }
        }

        // Notify listener
        if (listener != null) {
            listener.onOperationComplete(record);
            listener.onStatsUpdated(operation, stats);

            // Check for slow operation (threshold: 5 seconds)
            if (durationMs > 5000) {
                listener.onSlowOperation(record, 5000);
            }
        }

        Log.v(TAG, "Recorded: " + operation + " -> " + durationMs + "ms (" + (success ? "OK" : "FAIL") + ")");
    }

    /**
     * Get stats for an operation
     */
    @Nullable
    public OperationStats getStats(@NonNull String operation) {
        return operationStats.get(operation);
    }

    /**
     * Get all operation stats
     */
    @NonNull
    public Map<String, OperationStats> getAllStats() {
        return new HashMap<>(operationStats);
    }

    /**
     * Get recent operations
     */
    @NonNull
    public List<OperationRecord> getRecentOperations() {
        synchronized (recentOperations) {
            return new ArrayList<>(recentOperations);
        }
    }

    /**
     * Get recent operations by type
     */
    @NonNull
    public List<OperationRecord> getRecentOperations(@NonNull String operation) {
        List<OperationRecord> filtered = new ArrayList<>();
        synchronized (recentOperations) {
            for (OperationRecord record : recentOperations) {
                if (record.operation.equals(operation)) {
                    filtered.add(record);
                }
            }
        }
        return filtered;
    }

    /**
     * Get overall statistics
     */
    @NonNull
    public OverallStats getOverallStats() {
        int totalOperations = 0;
        int totalSuccess = 0;
        long totalTime = 0;

        for (OperationStats stats : operationStats.values()) {
            totalOperations += stats.count;
            totalSuccess += stats.successCount;
            totalTime += stats.totalTimeMs;
        }

        return new OverallStats(totalOperations, totalSuccess, totalOperations - totalSuccess, totalTime);
    }

    /**
     * Clear all stats
     */
    public void clear() {
        operationStats.clear();
        synchronized (recentOperations) {
            recentOperations.clear();
        }
        Log.i(TAG, "Performance stats cleared");
    }

    /**
     * Generate report
     */
    @NonNull
    public String generateReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== Performance Report ===\n\n");

        OverallStats overall = getOverallStats();
        sb.append("Overall:\n");
        sb.append("  Total operations: ").append(overall.totalCount).append("\n");
        sb.append("  Success rate: ").append(String.format("%.1f%%", overall.getSuccessRate() * 100)).append("\n");
        sb.append("  Total time: ").append(overall.totalTimeMs).append("ms\n\n");

        sb.append("Operations:\n");
        for (Map.Entry<String, OperationStats> entry : operationStats.entrySet()) {
            sb.append("  ").append(entry.getValue().toString()).append("\n");
        }

        return sb.toString();
    }

    /**
     * Overall statistics
     */
    public static class OverallStats {
        public final int totalCount;
        public final int successCount;
        public final int failureCount;
        public final long totalTimeMs;

        public OverallStats(int totalCount, int successCount, int failureCount, long totalTimeMs) {
            this.totalCount = totalCount;
            this.successCount = successCount;
            this.failureCount = failureCount;
            this.totalTimeMs = totalTimeMs;
        }

        public double getSuccessRate() {
            return totalCount > 0 ? (double) successCount / totalCount : 0;
        }

        public double getAverageTimeMs() {
            return totalCount > 0 ? (double) totalTimeMs / totalCount : 0;
        }
    }

    /**
     * Operation timer for easy timing
     */
    public static class OperationTimer {
        private final PerformanceMonitor monitor;
        private final String operation;
        private final long startTime;
        private boolean completed = false;

        private OperationTimer(PerformanceMonitor monitor, String operation) {
            this.monitor = monitor;
            this.operation = operation;
            this.startTime = System.currentTimeMillis();
        }

        /**
         * Complete with success
         */
        public void success() {
            complete(true, null, null);
        }

        /**
         * Complete with failure
         */
        public void failure(@Nullable String error) {
            complete(false, error, null);
        }

        /**
         * Complete with metadata
         */
        public void complete(boolean success, @Nullable String error, @Nullable Map<String, String> metadata) {
            if (completed) return;
            completed = true;

            long endTime = System.currentTimeMillis();
            monitor.recordOperation(operation, endTime - startTime, success, error, metadata);
        }
    }
}