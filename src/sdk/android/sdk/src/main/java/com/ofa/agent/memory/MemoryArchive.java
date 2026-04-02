package com.ofa.agent.memory;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.FileWriter;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Locale;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Memory Archive - L3文件归档层
 * 用于备份、导出、导入和归档冷数据
 */
public class MemoryArchive {

    private static final String TAG = "MemoryArchive";
    private static final String ARCHIVE_DIR = "memory_archive";
    private static final String EXPORT_FILE = "memory_export.json";
    private static final SimpleDateFormat DATE_FORMAT = new SimpleDateFormat("yyyy-MM-dd_HH-mm-ss", Locale.US);

    private final Context context;
    private final File archiveDir;
    private final ExecutorService executor;

    public MemoryArchive(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.archiveDir = new File(context.getFilesDir(), ARCHIVE_DIR);
        if (!archiveDir.exists()) {
            archiveDir.mkdirs();
        }
        this.executor = Executors.newSingleThreadExecutor();
    }

    // ===== 导出 =====

    /**
     * 导出记忆到文件
     */
    public void exportToFile(@NonNull List<MemoryEntry> entries, @Nullable ExportCallback callback) {
        executor.execute(() -> {
            try {
                File exportFile = new File(context.getExternalFilesDir(null), EXPORT_FILE);

                JSONObject json = new JSONObject();
                json.put("version", 1);
                json.put("exportTime", System.currentTimeMillis());
                json.put("count", entries.size());

                JSONArray array = new JSONArray();
                for (MemoryEntry entry : entries) {
                    array.put(entryToJson(entry));
                }
                json.put("memories", array);

                FileWriter writer = new FileWriter(exportFile);
                writer.write(json.toString(2));
                writer.close();

                Log.i(TAG, "Exported " + entries.size() + " memories to " + exportFile);

                if (callback != null) {
                    callback.onSuccess(exportFile);
                }
            } catch (Exception e) {
                Log.e(TAG, "Export failed", e);
                if (callback != null) {
                    callback.onError(e.getMessage());
                }
            }
        });
    }

    /**
     * 从文件导入记忆
     */
    public void importFromFile(@NonNull File file, @NonNull ImportCallback callback) {
        executor.execute(() -> {
            try {
                BufferedReader reader = new BufferedReader(new FileReader(file));
                StringBuilder sb = new StringBuilder();
                String line;
                while ((line = reader.readLine()) != null) {
                    sb.append(line);
                }
                reader.close();

                JSONObject json = new JSONObject(sb.toString());
                JSONArray array = json.getJSONArray("memories");

                List<MemoryEntry> entries = new ArrayList<>();
                for (int i = 0; i < array.length(); i++) {
                    MemoryEntry entry = jsonToEntry(array.getJSONObject(i));
                    if (entry != null) {
                        entries.add(entry);
                    }
                }

                Log.i(TAG, "Imported " + entries.size() + " memories from " + file);
                callback.onSuccess(entries);

            } catch (Exception e) {
                Log.e(TAG, "Import failed", e);
                callback.onError(e.getMessage());
            }
        });
    }

    // ===== 归档 =====

    /**
     * 归档旧记忆
     */
    public void archiveOldMemories(@NonNull List<MemoryEntity> oldMemories, @Nullable ArchiveCallback callback) {
        if (oldMemories.isEmpty()) {
            if (callback != null) callback.onSuccess(null);
            return;
        }

        executor.execute(() -> {
            try {
                String timestamp = DATE_FORMAT.format(new Date());
                String fileName = "archive_" + timestamp + ".json";
                File archiveFile = new File(archiveDir, fileName);

                JSONObject json = new JSONObject();
                json.put("archiveTime", System.currentTimeMillis());
                json.put("count", oldMemories.size());

                JSONArray array = new JSONArray();
                for (MemoryEntity entity : oldMemories) {
                    array.put(entityToJson(entity));
                }
                json.put("memories", array);

                FileWriter writer = new FileWriter(archiveFile);
                writer.write(json.toString(2));
                writer.close();

                Log.i(TAG, "Archived " + oldMemories.size() + " memories to " + archiveFile);

                if (callback != null) {
                    callback.onSuccess(archiveFile);
                }
            } catch (Exception e) {
                Log.e(TAG, "Archive failed", e);
                if (callback != null) {
                    callback.onError(e.getMessage());
                }
            }
        });
    }

    /**
     * 列出所有归档文件
     */
    @NonNull
    public List<ArchiveInfo> listArchives() {
        List<ArchiveInfo> archives = new ArrayList<>();
        File[] files = archiveDir.listFiles((dir, name) -> name.startsWith("archive_") && name.endsWith(".json"));

        if (files != null) {
            for (File file : files) {
                archives.add(new ArchiveInfo(file.getName(), file.length(), file.lastModified()));
            }
        }

        return archives;
    }

    /**
     * 恢复归档
     */
    public void restoreArchive(@NonNull String archiveName, @NonNull ImportCallback callback) {
        File archiveFile = new File(archiveDir, archiveName);
        if (!archiveFile.exists()) {
            callback.onError("Archive not found: " + archiveName);
            return;
        }
        importFromFile(archiveFile, callback);
    }

    /**
     * 删除归档
     */
    public boolean deleteArchive(@NonNull String archiveName) {
        File archiveFile = new File(archiveDir, archiveName);
        if (archiveFile.exists()) {
            return archiveFile.delete();
        }
        return false;
    }

    /**
     * 清理所有归档
     */
    public void clearArchives() {
        File[] files = archiveDir.listFiles();
        if (files != null) {
            for (File file : files) {
                file.delete();
            }
        }
        Log.i(TAG, "Cleared all archives");
    }

    // ===== JSON转换 =====

    @NonNull
    private JSONObject entryToJson(@NonNull MemoryEntry entry) throws Exception {
        JSONObject json = new JSONObject();
        json.put("key", entry.getKey());
        json.put("category", entry.getCategory());
        json.put("value", entry.getValue());
        json.put("timestamp", entry.getTimestamp());
        json.put("count", entry.getCount());
        json.put("score", entry.getScore());
        json.put("context", entry.getContext());

        JSONObject attrs = new JSONObject();
        for (java.util.Map.Entry<String, String> e : entry.getAttributes().entrySet()) {
            attrs.put(e.getKey(), e.getValue());
        }
        json.put("attributes", attrs);

        return json;
    }

    @NonNull
    private JSONObject entityToJson(@NonNull MemoryEntity entity) throws Exception {
        JSONObject json = new JSONObject();
        json.put("key", entity.key);
        json.put("category", entity.category);
        json.put("value", entity.value);
        json.put("attributes", entity.attributes);
        json.put("timestamp", entity.timestamp);
        json.put("count", entity.count);
        json.put("score", entity.score);
        json.put("context", entity.context);
        return json;
    }

    @Nullable
    private MemoryEntry jsonToEntry(@NonNull JSONObject json) {
        try {
            MemoryEntry.Builder builder = new MemoryEntry.Builder()
                    .key(json.getString("key"))
                    .category(json.optString("category", "general"))
                    .value(json.getString("value"))
                    .timestamp(json.optLong("timestamp", System.currentTimeMillis()))
                    .count(json.optInt("count", 1))
                    .score((float) json.optDouble("score", 1.0))
                    .context(json.optString("context", null));

            JSONObject attrs = json.optJSONObject("attributes");
            if (attrs != null) {
                for (java.util.Iterator<String> it = attrs.keys(); it.hasNext(); ) {
                    String key = it.next();
                    builder.attribute(key, attrs.optString(key));
                }
            }

            return builder.build();
        } catch (Exception e) {
            Log.e(TAG, "Failed to parse memory entry", e);
            return null;
        }
    }

    // ===== 回调接口 =====

    public interface ExportCallback {
        void onSuccess(File file);
        void onError(String error);
    }

    public interface ImportCallback {
        void onSuccess(List<MemoryEntry> entries);
        void onError(String error);
    }

    public interface ArchiveCallback {
        void onSuccess(File archiveFile);
        void onError(String error);
    }

    // ===== 数据类 =====

    public static class ArchiveInfo {
        public final String name;
        public final long size;
        public final long timestamp;

        public ArchiveInfo(String name, long size, long timestamp) {
            this.name = name;
            this.size = size;
            this.timestamp = timestamp;
        }
    }
}