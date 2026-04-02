package com.ofa.agent.offline;

import android.content.Context;
import android.database.Cursor;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

/**
 * 离线缓存 - 数据持久化
 */
public class OfflineCache {
    private static final String TAG = "OfflineCache";
    private static final String DB_NAME = "ofa_offline_cache.db";
    private static final int DB_VERSION = 1;

    private final CacheDbHelper dbHelper;
    private final ConcurrentHashMap<String, CacheEntry> memoryCache = new ConcurrentHashMap<>();
    private final AtomicLong currentSize = new AtomicLong(0);
    private final AtomicInteger hits = new AtomicInteger(0);
    private final AtomicInteger misses = new AtomicInteger(0);
    private final long maxSize;

    public OfflineCache(@NonNull Context context) {
        this(context, 10 * 1024 * 1024); // 默认 10MB
    }

    public OfflineCache(@NonNull Context context, long maxSize) {
        this.maxSize = maxSize;
        this.dbHelper = new CacheDbHelper(context);
        loadFromDb();
    }

    /**
     * 存储数据
     */
    public void put(@NonNull String key, @NonNull byte[] data, long expiryMs) {
        long timestamp = System.currentTimeMillis();
        long expiry = expiryMs > 0 ? timestamp + expiryMs : 0;

        // 检查容量
        if (currentSize.get() + data.length > maxSize) {
            evictIfNeeded(data.length);
        }

        // 内存缓存
        CacheEntry entry = new CacheEntry(data, timestamp, expiry, false);
        memoryCache.put(key, entry);
        currentSize.addAndGet(data.length);

        // 数据库持久化
        saveToDb(key, entry);

        Log.d(TAG, "Cached: " + key + " (" + data.length + " bytes)");
    }

    public void put(@NonNull String key, @NonNull byte[] data) {
        put(key, data, 0);
    }

    /**
     * 获取数据
     */
    @Nullable
    public byte[] get(@NonNull String key) {
        CacheEntry entry = memoryCache.get(key);

        if (entry != null) {
            // 检查过期
            if (entry.expiry > 0 && System.currentTimeMillis() > entry.expiry) {
                remove(key);
                misses.incrementAndGet();
                return null;
            }

            hits.incrementAndGet();
            return entry.data;
        }

        misses.incrementAndGet();
        return null;
    }

    /**
     * 删除数据
     */
    public void remove(@NonNull String key) {
        CacheEntry entry = memoryCache.remove(key);
        if (entry != null) {
            currentSize.addAndGet(-entry.data.length);
        }

        deleteFromDb(key);
    }

    /**
     * 清空缓存
     */
    public void clear() {
        memoryCache.clear();
        currentSize.set(0);
        dbHelper.getWritableDatabase().delete("cache", null, null);
        dbHelper.getWritableDatabase().delete("pending_sync", null, null);
    }

    /**
     * 获取待同步键列表
     */
    @NonNull
    public List<String> getPendingKeys() {
        List<String> keys = new ArrayList<>();
        SQLiteDatabase db = dbHelper.getReadableDatabase();
        Cursor cursor = db.query("pending_sync", new String[]{"key"}, null, null, null, null, null);

        while (cursor.moveToNext()) {
            keys.add(cursor.getString(0));
        }
        cursor.close();

        return keys;
    }

    /**
     * 标记已同步
     */
    public void markSynced(@NonNull String key) {
        CacheEntry entry = memoryCache.get(key);
        if (entry != null) {
            memoryCache.put(key, new CacheEntry(entry.data, entry.timestamp, entry.expiry, true));
        }

        SQLiteDatabase db = dbHelper.getWritableDatabase();
        db.delete("pending_sync", "key = ?", new String[]{key});
    }

    /**
     * 获取待同步数量
     */
    public int getPendingCount() {
        SQLiteDatabase db = dbHelper.getReadableDatabase();
        Cursor cursor = db.rawQuery("SELECT COUNT(*) FROM pending_sync", null);
        cursor.moveToFirst();
        int count = cursor.getInt(0);
        cursor.close();
        return count;
    }

    /**
     * 获取命中率
     */
    public double hitRate() {
        int total = hits.get() + misses.get();
        return total > 0 ? (double) hits.get() / total : 0.0;
    }

    /**
     * 获取当前大小
     */
    public long getCurrentSize() {
        return currentSize.get();
    }

    private void evictIfNeeded(long needed) {
        long now = System.currentTimeMillis();

        // 清理过期项
        for (Map.Entry<String, CacheEntry> entry : memoryCache.entrySet()) {
            if (entry.getValue().expiry > 0 && now > entry.getValue().expiry) {
                remove(entry.getKey());
            }
        }

        // 如果仍不够，清理最旧的已同步项
        while (currentSize.get() + needed > maxSize && !memoryCache.isEmpty()) {
            String oldestKey = null;
            long oldestTime = Long.MAX_VALUE;

            for (Map.Entry<String, CacheEntry> entry : memoryCache.entrySet()) {
                if (entry.getValue().synced && entry.getValue().timestamp < oldestTime) {
                    oldestTime = entry.getValue().timestamp;
                    oldestKey = entry.getKey();
                }
            }

            if (oldestKey != null) {
                remove(oldestKey);
            } else {
                break;
            }
        }
    }

    private void saveToDb(String key, CacheEntry entry) {
        SQLiteDatabase db = dbHelper.getWritableDatabase();
        db.execSQL(
            "INSERT OR REPLACE INTO cache (key, data, timestamp, expiry, synced) VALUES (?, ?, ?, ?, ?)",
            new Object[]{key, entry.data, entry.timestamp, entry.expiry, entry.synced ? 1 : 0}
        );
        db.execSQL(
            "INSERT OR IGNORE INTO pending_sync (key) VALUES (?)",
            new Object[]{key}
        );
    }

    private void deleteFromDb(String key) {
        SQLiteDatabase db = dbHelper.getWritableDatabase();
        db.delete("cache", "key = ?", new String[]{key});
        db.delete("pending_sync", "key = ?", new String[]{key});
    }

    private void loadFromDb() {
        SQLiteDatabase db = dbHelper.getReadableDatabase();
        Cursor cursor = db.query("cache", null, null, null, null, null, null);

        while (cursor.moveToNext()) {
            String key = cursor.getString(cursor.getColumnIndexOrThrow("key"));
            byte[] data = cursor.getBlob(cursor.getColumnIndexOrThrow("data"));
            long timestamp = cursor.getLong(cursor.getColumnIndexOrThrow("timestamp"));
            long expiry = cursor.getLong(cursor.getColumnIndexOrThrow("expiry"));
            boolean synced = cursor.getInt(cursor.getColumnIndexOrThrow("synced")) == 1;

            memoryCache.put(key, new CacheEntry(data, timestamp, expiry, synced));
            currentSize.addAndGet(data.length);
        }
        cursor.close();
    }

    private static class CacheEntry {
        final byte[] data;
        final long timestamp;
        final long expiry;
        final boolean synced;

        CacheEntry(byte[] data, long timestamp, long expiry, boolean synced) {
            this.data = data;
            this.timestamp = timestamp;
            this.expiry = expiry;
            this.synced = synced;
        }
    }

    private static class CacheDbHelper extends SQLiteOpenHelper {
        CacheDbHelper(Context context) {
            super(context, DB_NAME, null, DB_VERSION);
        }

        @Override
        public void onCreate(SQLiteDatabase db) {
            db.execSQL(
                "CREATE TABLE cache (" +
                "key TEXT PRIMARY KEY, " +
                "data BLOB, " +
                "timestamp INTEGER, " +
                "expiry INTEGER, " +
                "synced INTEGER)"
            );
            db.execSQL(
                "CREATE TABLE pending_sync (key TEXT PRIMARY KEY)"
            );
        }

        @Override
        public void onUpgrade(SQLiteDatabase db, int oldVersion, int newVersion) {
            db.execSQL("DROP TABLE IF EXISTS cache");
            db.execSQL("DROP TABLE IF EXISTS pending_sync");
            onCreate(db);
        }
    }
}