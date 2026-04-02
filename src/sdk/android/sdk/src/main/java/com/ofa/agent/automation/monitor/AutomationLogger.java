package com.ofa.agent.automation.monitor;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.FileWriter;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Automation Logger - logs automation operations for debugging and auditing.
 */
public class AutomationLogger {

    private static final String TAG = "AutomationLogger";

    private final Context context;
    private final ExecutorService executor;
    private final File logDir;
    private final SimpleDateFormat dateFormat;
    private final int maxLogFiles;
    private final long maxFileSize;

    private final List<LogEntry> inMemoryLog;
    private final int maxInMemoryEntries;

    private LogCallback callback;
    private LogLevel minLevel = LogLevel.DEBUG;
    private boolean writeToFile = true;

    /**
     * Log levels
     */
    public enum LogLevel {
        DEBUG, INFO, WARN, ERROR
    }

    /**
     * Log entry
     */
    public static class LogEntry {
        public final long timestamp;
        public final LogLevel level;
        public final String operation;
        public final String message;
        public final Map<String, String> data;
        public final String error;

        public LogEntry(LogLevel level, String operation, String message,
                        Map<String, String> data, String error) {
            this.timestamp = System.currentTimeMillis();
            this.level = level;
            this.operation = operation;
            this.message = message;
            this.data = data;
            this.error = error;
        }

        @NonNull
        public String format() {
            StringBuilder sb = new StringBuilder();
            sb.append(new SimpleDateFormat("yyyy-MM-dd HH:mm:ss.SSS", Locale.getDefault())
                .format(new Date(timestamp)));
            sb.append(" [").append(level).append("] ");
            sb.append(operation).append(": ").append(message);

            if (data != null && !data.isEmpty()) {
                sb.append(" | ");
                for (Map.Entry<String, String> entry : data.entrySet()) {
                    sb.append(entry.getKey()).append("=").append(entry.getValue()).append(" ");
                }
            }

            if (error != null) {
                sb.append(" | ERROR: ").append(error);
            }

            return sb.toString();
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("timestamp", timestamp);
                json.put("level", level.name());
                json.put("operation", operation);
                json.put("message", message);
                if (data != null) {
                    json.put("data", new JSONObject(data));
                }
                if (error != null) {
                    json.put("error", error);
                }
            } catch (Exception e) {
                // Ignore
            }
            return json;
        }
    }

    /**
     * Log callback
     */
    public interface LogCallback {
        void onLogEntry(@NonNull LogEntry entry);
    }

    /**
     * Create automation logger
     */
    public AutomationLogger(@NonNull Context context) {
        this(context, 100, 10, 5 * 1024 * 1024); // 100 in-memory, 10 files, 5MB max
    }

    /**
     * Create automation logger with settings
     */
    public AutomationLogger(@NonNull Context context, int maxInMemoryEntries,
                            int maxLogFiles, long maxFileSize) {
        this.context = context;
        this.executor = Executors.newSingleThreadExecutor();
        this.logDir = new File(context.getFilesDir(), "automation_logs");
        this.dateFormat = new SimpleDateFormat("yyyy-MM-dd", Locale.getDefault());
        this.maxLogFiles = maxLogFiles;
        this.maxFileSize = maxFileSize;
        this.inMemoryLog = new ArrayList<>();
        this.maxInMemoryEntries = maxInMemoryEntries;

        // Ensure log directory exists
        if (!logDir.exists()) {
            logDir.mkdirs();
        }
    }

    /**
     * Set callback
     */
    public void setCallback(@Nullable LogCallback callback) {
        this.callback = callback;
    }

    /**
     * Set minimum log level
     */
    public void setMinLevel(@NonNull LogLevel level) {
        this.minLevel = level;
    }

    /**
     * Enable/disable file writing
     */
    public void setWriteToFile(boolean enabled) {
        this.writeToFile = enabled;
    }

    /**
     * Log debug message
     */
    public void debug(@NonNull String operation, @NonNull String message) {
        log(LogLevel.DEBUG, operation, message, null, null);
    }

    /**
     * Log info message
     */
    public void info(@NonNull String operation, @NonNull String message) {
        log(LogLevel.INFO, operation, message, null, null);
    }

    /**
     * Log warning message
     */
    public void warn(@NonNull String operation, @NonNull String message) {
        log(LogLevel.WARN, operation, message, null, null);
    }

    /**
     * Log error message
     */
    public void error(@NonNull String operation, @NonNull String message, @Nullable String error) {
        log(LogLevel.ERROR, operation, message, null, error);
    }

    /**
     * Log with data
     */
    public void log(@NonNull LogLevel level, @NonNull String operation, @NonNull String message,
                    @Nullable Map<String, String> data, @Nullable String error) {
        if (level.ordinal() < minLevel.ordinal()) {
            return;
        }

        LogEntry entry = new LogEntry(level, operation, message, data, error);

        // Add to in-memory log
        synchronized (inMemoryLog) {
            inMemoryLog.add(entry);
            while (inMemoryLog.size() > maxInMemoryEntries) {
                inMemoryLog.remove(0);
            }
        }

        // Log to Android
        String formatted = entry.format();
        switch (level) {
            case DEBUG:
                Log.d(TAG, formatted);
                break;
            case INFO:
                Log.i(TAG, formatted);
                break;
            case WARN:
                Log.w(TAG, formatted);
                break;
            case ERROR:
                Log.e(TAG, formatted);
                break;
        }

        // Notify callback
        if (callback != null) {
            callback.onLogEntry(entry);
        }

        // Write to file
        if (writeToFile) {
            writeToFileAsync(entry);
        }
    }

    /**
     * Write log entry to file asynchronously
     */
    private void writeToFileAsync(@NonNull LogEntry entry) {
        executor.execute(() -> {
            try {
                String dateStr = dateFormat.format(new Date());
                File logFile = new File(logDir, "automation_" + dateStr + ".log");

                // Check file size
                if (logFile.exists() && logFile.length() > maxFileSize) {
                    rotateLogFiles();
                    logFile = new File(logDir, "automation_" + dateStr + ".log");
                }

                // Append entry
                try (FileWriter writer = new FileWriter(logFile, true)) {
                    writer.write(entry.format());
                    writer.write("\n");
                }
            } catch (Exception e) {
                Log.e(TAG, "Failed to write log: " + e.getMessage());
            }
        });
    }

    /**
     * Rotate log files
     */
    private void rotateLogFiles() {
        File[] files = logDir.listFiles((dir, name) -> name.startsWith("automation_") && name.endsWith(".log"));
        if (files != null && files.length >= maxLogFiles) {
            // Sort by name (oldest first)
            java.util.Arrays.sort(files);
            // Delete oldest
            files[0].delete();
        }
    }

    /**
     * Get in-memory log entries
     */
    @NonNull
    public List<LogEntry> getLogEntries() {
        synchronized (inMemoryLog) {
            return new ArrayList<>(inMemoryLog);
        }
    }

    /**
     * Get log entries by operation
     */
    @NonNull
    public List<LogEntry> getLogEntries(@NonNull String operation) {
        List<LogEntry> filtered = new ArrayList<>();
        synchronized (inMemoryLog) {
            for (LogEntry entry : inMemoryLog) {
                if (entry.operation.equals(operation)) {
                    filtered.add(entry);
                }
            }
        }
        return filtered;
    }

    /**
     * Get log entries by level
     */
    @NonNull
    public List<LogEntry> getLogEntries(@NonNull LogLevel level) {
        List<LogEntry> filtered = new ArrayList<>();
        synchronized (inMemoryLog) {
            for (LogEntry entry : inMemoryLog) {
                if (entry.level == level) {
                    filtered.add(entry);
                }
            }
        }
        return filtered;
    }

    /**
     * Export logs to JSON
     */
    @NonNull
    public JSONArray exportLogs() {
        JSONArray array = new JSONArray();
        synchronized (inMemoryLog) {
            for (LogEntry entry : inMemoryLog) {
                array.put(entry.toJson());
            }
        }
        return array;
    }

    /**
     * Clear in-memory logs
     */
    public void clearInMemoryLogs() {
        synchronized (inMemoryLog) {
            inMemoryLog.clear();
        }
        Log.i(TAG, "In-memory logs cleared");
    }

    /**
     * Delete all log files
     */
    public void deleteLogFiles() {
        executor.execute(() -> {
            File[] files = logDir.listFiles();
            if (files != null) {
                for (File file : files) {
                    file.delete();
                }
            }
            Log.i(TAG, "Log files deleted");
        });
    }

    /**
     * Get log file size
     */
    public long getLogFileSize() {
        long total = 0;
        File[] files = logDir.listFiles((dir, name) -> name.endsWith(".log"));
        if (files != null) {
            for (File file : files) {
                total += file.length();
            }
        }
        return total;
    }

    /**
     * Shutdown logger
     */
    public void shutdown() {
        executor.shutdown();
        Log.i(TAG, "AutomationLogger shutdown");
    }
}