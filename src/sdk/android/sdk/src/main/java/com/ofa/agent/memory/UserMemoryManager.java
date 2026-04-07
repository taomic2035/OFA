package com.ofa.agent.memory;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 用户记忆管理器 - 三层架构
 * L1: MemoryCache (热数据缓存, LRU策略)
 * L2: Room Database (持久化存储)
 * L3: MemoryArchive (文件归档, 冷数据备份)
 *
 * v2.2.0 新增：跨设备同步支持
 */
public class UserMemoryManager {

    private static final String TAG = "UserMemory";
    private static final long DEFAULT_MAX_AGE = 30L * 24 * 60 * 60 * 1000; // 30天
    private static final int ARCHIVE_THRESHOLD_DAYS = 60; // 60天以上的数据归档

    private static volatile UserMemoryManager instance;

    // Context for sync service
    private final Context context;

    // 三层存储
    private final MemoryCache cache;           // L1: 内存缓存
    private final MemoryDao dao;               // L2: Room数据库
    private final MemoryArchive archive;       // L3: 文件归档

    // 辅助存储
    private final Map<String, MemoryEntry> lastActions;  // 最近行为快速访问

    // 异步执行器
    private final ExecutorService executor;

    // v2.2.0: 记忆同步服务
    private MemorySyncService syncService;

    private UserMemoryManager(@NonNull Context context) {
        this.context = context.getApplicationContext();

        // 初始化三层存储
        this.cache = new MemoryCache(100);  // L1: 缓存100条热数据
        this.dao = MemoryDatabase.getInstance(this.context).memoryDao();  // L2
        this.archive = new MemoryArchive(this.context);  // L3

        this.lastActions = new ConcurrentHashMap<>();
        this.executor = Executors.newSingleThreadExecutor();

        // 异步加载数据
        loadFromDatabase();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static synchronized UserMemoryManager getInstance(@NonNull Context context) {
        if (instance == null) {
            instance = new UserMemoryManager(context);
        }
        return instance;
    }

    // ===== 记忆存储 =====

    /**
     * 记录用户行为/偏好
     * 写入三层存储: L1缓存 + L2数据库
     */
    public void remember(@NonNull MemoryEntry entry) {
        String key = entry.getKey();
        String value = entry.getValue();

        // 更新最近行为
        lastActions.put(key, entry);

        // L1: 写入缓存
        cache.put(key, entry);

        // L2: 异步写入数据库
        executor.execute(() -> {
            MemoryEntity existing = dao.getByKeyValue(key, value);
            MemoryEntity entity;

            if (existing != null) {
                // 更新已存在条目: 增加计数、更新时间戳
                existing.count++;
                existing.score = Math.min(existing.score + 0.5f, 10.0f);
                existing.lastAccessed = System.currentTimeMillis();
                existing.context = entry.getContext();
                dao.update(existing);
                entity = existing;
            } else {
                // 新条目
                entity = MemoryEntity.fromEntry(entry);
                dao.insert(entity);
            }

            Log.d(TAG, "Remembered: " + key + " = " + value + " (count: " + entity.count + ")");
        });
    }

    /**
     * 批量记录
     */
    public void rememberAll(@NonNull List<MemoryEntry> entries) {
        for (MemoryEntry entry : entries) {
            remember(entry);
        }
    }

    /**
     * 快速记录偏好
     */
    public void rememberPreference(@NonNull String key, @NonNull String value,
                                    @Nullable String category, @Nullable Map<String, String> attributes) {
        MemoryEntry.Builder builder = new MemoryEntry.Builder()
                .key(key)
                .value(value)
                .category(category != null ? category : "preference");

        if (attributes != null) {
            builder.attributes(attributes);
        }

        remember(builder.build());
    }

    /**
     * 记录技能执行参数
     */
    public void rememberSkillParams(@NonNull String skillId, @NonNull Map<String, Object> params) {
        for (Map.Entry<String, Object> param : params.entrySet()) {
            String key = skillId + "." + param.getKey();
            String value = String.valueOf(param.getValue());

            rememberPreference(key, value, "skill_params", null);
        }
    }

    // ===== 记忆查询 =====

    /**
     * 获取推荐值（按分数排序的第一个）
     * 查询顺序: L1缓存 → L2数据库
     */
    @Nullable
    public String getRecommendedValue(@NonNull String key) {
        // L1: 先查缓存
        String cachedValue = cache.getRecommendedValue(key);
        if (cachedValue != null) {
            return cachedValue;
        }

        // L2: 查数据库
        MemoryEntity top = dao.getTopRecommendation(key);
        if (top != null) {
            // 回填L1缓存
            cache.put(key, top.toEntry());
            return top.value;
        }

        return null;
    }

    /**
     * 获取推荐列表（按分数排序）
     */
    @NonNull
    public List<String> getRecommendedValues(@NonNull String key, int limit) {
        // L1: 先查缓存
        List<String> cachedValues = cache.getRecommendedValues(key, limit);
        if (!cachedValues.isEmpty()) {
            return cachedValues;
        }

        // L2: 查数据库
        List<MemoryEntity> entities = dao.getByKeyLimit(key, limit);
        if (entities.isEmpty()) {
            return Collections.emptyList();
        }

        List<String> values = new ArrayList<>();
        for (MemoryEntity entity : entities) {
            values.add(entity.value);
            // 回填L1缓存
            cache.put(key, entity.toEntry());
        }

        return values;
    }

    /**
     * 获取所有匹配的记忆条目
     */
    @NonNull
    public List<MemoryEntry> getEntries(@NonNull String key) {
        // L1: 先查缓存
        List<MemoryEntry> cachedEntries = cache.getEntries(key);
        if (!cachedEntries.isEmpty()) {
            return cachedEntries;
        }

        // L2: 查数据库
        List<MemoryEntity> entities = dao.getByKey(key);
        List<MemoryEntry> entries = new ArrayList<>();

        for (MemoryEntity entity : entities) {
            MemoryEntry entry = entity.toEntry();
            entries.add(entry);
            // 回填L1缓存
            cache.put(key, entry);
        }

        return entries;
    }

    /**
     * 获取某分类下的所有记忆
     */
    @NonNull
    public List<MemoryEntry> getEntriesByCategory(@NonNull String category) {
        List<MemoryEntity> entities = dao.getByCategory(category);
        List<MemoryEntry> entries = new ArrayList<>();

        for (MemoryEntity entity : entities) {
            entries.add(entity.toEntry());
        }

        return entries;
    }

    /**
     * 获取最后一次行为
     */
    @Nullable
    public MemoryEntry getLastAction(@NonNull String key) {
        // 先查快速访问缓存
        MemoryEntry cached = lastActions.get(key);
        if (cached != null) {
            return cached;
        }

        // L1: 查缓存
        String lastValue = cache.getLastUsedValue(key);
        if (lastValue != null) {
            MemoryEntry entry = cache.getEntry(key, lastValue);
            if (entry != null) {
                lastActions.put(key, entry);
                return entry;
            }
        }

        // L2: 查数据库
        MemoryEntity entity = dao.getLastUsed(key);
        if (entity != null) {
            MemoryEntry entry = entity.toEntry();
            lastActions.put(key, entry);
            cache.put(key, entry);
            return entry;
        }

        return null;
    }

    /**
     * 获取最后一次值
     */
    @Nullable
    public String getLastValue(@NonNull String key) {
        MemoryEntry entry = getLastAction(key);
        return entry != null ? entry.getValue() : null;
    }

    /**
     * 检查是否有记忆
     */
    public boolean hasMemory(@NonNull String key) {
        // L1: 检查缓存
        if (cache.hasKey(key)) {
            return true;
        }

        // L2: 检查数据库
        return dao.getTotalUsageCount(key) > 0;
    }

    /**
     * 获取使用次数
     */
    public int getUsageCount(@NonNull String key, @NonNull String value) {
        // L1: 先查缓存
        MemoryEntry cached = cache.getEntry(key, value);
        if (cached != null) {
            return cached.getCount();
        }

        // L2: 查数据库
        MemoryEntity entity = dao.getByKeyValue(key, value);
        return entity != null ? entity.count : 0;
    }

    // ===== 智能推荐 =====

    /**
     * 根据部分输入推荐完整值
     */
    @NonNull
    public List<String> autocomplete(@NonNull String key, @NonNull String prefix, int limit) {
        // L2: 直接查数据库（autocomplete通常数据量较大）
        List<MemoryEntity> entities = dao.autocomplete(key, prefix, limit);
        List<String> results = new ArrayList<>();

        for (MemoryEntity entity : entities) {
            results.add(entity.value);
        }

        return results;
    }

    /**
     * 搜索记忆
     */
    @NonNull
    public List<MemoryEntry> search(@NonNull String query) {
        List<MemoryEntity> entities = dao.searchByValue(query);
        List<MemoryEntry> entries = new ArrayList<>();

        for (MemoryEntity entity : entities) {
            entries.add(entity.toEntry());
        }

        return entries;
    }

    /**
     * 获取智能默认值
     * 考虑：最近使用、使用频率、推荐分数
     */
    @Nullable
    public SmartDefault getSmartDefault(@NonNull String key) {
        // 获取推荐值
        String recommended = getRecommendedValue(key);

        // 获取最近使用的值
        MemoryEntity lastEntity = dao.getLastUsed(key);
        String lastUsed = lastEntity != null ? lastEntity.value : null;

        // 获取最常用的值
        MemoryEntity mostUsedEntity = dao.getMostUsed(key);
        String mostUsed = mostUsedEntity != null ? mostUsedEntity.value : null;

        if (recommended == null) {
            return null;
        }

        // 计算置信度
        float confidence = 0.0f;
        if (recommended.equals(lastUsed)) confidence += 0.3f;
        if (recommended.equals(mostUsed)) confidence += 0.4f;

        MemoryEntity topEntity = dao.getTopRecommendation(key);
        if (topEntity != null) {
            confidence += topEntity.calculateRecommendationScore() * 0.3f;
        }

        return new SmartDefault(recommended, lastUsed, mostUsed, confidence);
    }

    /**
     * 智能默认值
     */
    public static class SmartDefault {
        public final String recommendedValue;  // 推荐值
        public final String lastUsedValue;     // 最后使用的值
        public final String mostUsedValue;     // 最常用的值
        public final float confidence;         // 置信度

        public SmartDefault(String recommendedValue, String lastUsedValue,
                            String mostUsedValue, float confidence) {
            this.recommendedValue = recommendedValue;
            this.lastUsedValue = lastUsedValue;
            this.mostUsedValue = mostUsedValue;
            this.confidence = confidence;
        }
    }

    // ===== 记忆管理 =====

    /**
     * 忘记某个值
     */
    public void forget(@NonNull String key, @NonNull String value) {
        // L1: 从缓存删除
        cache.remove(key, value);
        lastActions.remove(key);

        // L2: 从数据库删除
        executor.execute(() -> {
            dao.deleteByKeyValue(key, value);
            Log.d(TAG, "Forgot: " + key + " = " + value);
        });
    }

    /**
     * 清除某个key的所有记忆
     */
    public void forgetAll(@NonNull String key) {
        // L1: 从缓存删除
        cache.removeAll(key);
        lastActions.remove(key);

        // L2: 从数据库删除
        executor.execute(() -> {
            dao.deleteByKey(key);
            Log.d(TAG, "Forgot all: " + key);
        });
    }

    /**
     * 清除过期记忆并归档
     */
    public void cleanupExpired() {
        executor.execute(() -> {
            long archiveThreshold = System.currentTimeMillis() - ARCHIVE_THRESHOLD_DAYS * 24 * 60 * 60 * 1000;
            long expireThreshold = System.currentTimeMillis() - DEFAULT_MAX_AGE;

            // 查找需要归档的旧数据
            List<MemoryEntity> oldMemories = new ArrayList<>();
            List<MemoryEntity> allMemories = dao.getAll();

            for (MemoryEntity entity : allMemories) {
                if (entity.timestamp < archiveThreshold) {
                    oldMemories.add(entity);
                }
            }

            // 归档到L3
            if (!oldMemories.isEmpty()) {
                archive.archiveOldMemories(oldMemories, new MemoryArchive.ArchiveCallback() {
                    @Override
                    public void onSuccess(File archiveFile) {
                        // 删除已归档的数据
                        for (MemoryEntity entity : oldMemories) {
                            dao.delete(entity);
                        }
                        Log.i(TAG, "Archived and deleted " + oldMemories.size() + " old memories");
                    }

                    @Override
                    public void onError(String error) {
                        Log.e(TAG, "Archive failed: " + error);
                    }
                });
            }

            // 删除过期数据
            int deleted = dao.deleteOlderThan(expireThreshold);
            Log.i(TAG, "Cleaned up " + deleted + " expired memories");
        });
    }

    /**
     * 清除所有记忆
     */
    public void clearAll() {
        // L1: 清空缓存
        cache.clear();
        lastActions.clear();

        // L2: 清空数据库
        executor.execute(() -> {
            dao.deleteAll();
            Log.i(TAG, "Cleared all memories");
        });
    }

    // ===== 导入导出 =====

    /**
     * 导出记忆
     */
    public void exportMemories(@NonNull MemoryArchive.ExportCallback callback) {
        executor.execute(() -> {
            List<MemoryEntity> entities = dao.getAll();
            List<MemoryEntry> entries = new ArrayList<>();

            for (MemoryEntity entity : entities) {
                entries.add(entity.toEntry());
            }

            archive.exportToFile(entries, callback);
        });
    }

    /**
     * 导入记忆
     */
    public void importMemories(@NonNull File file, @NonNull MemoryArchive.ImportCallback callback) {
        archive.importFromFile(file, new MemoryArchive.ImportCallback() {
            @Override
            public void onSuccess(List<MemoryEntry> entries) {
                // 写入数据库和缓存
                for (MemoryEntry entry : entries) {
                    remember(entry);
                }
                callback.onSuccess(entries);
            }

            @Override
            public void onError(String error) {
                callback.onError(error);
            }
        });
    }

    /**
     * 列出所有归档
     */
    @NonNull
    public List<MemoryArchive.ArchiveInfo> listArchives() {
        return archive.listArchives();
    }

    /**
     * 恢复归档
     */
    public void restoreArchive(@NonNull String archiveName, @NonNull MemoryArchive.ImportCallback callback) {
        archive.restoreArchive(archiveName, new MemoryArchive.ImportCallback() {
            @Override
            public void onSuccess(List<MemoryEntry> entries) {
                for (MemoryEntry entry : entries) {
                    remember(entry);
                }
                callback.onSuccess(entries);
            }

            @Override
            public void onError(String error) {
                callback.onError(error);
            }
        });
    }

    // ===== 统计 =====

    /**
     * 获取记忆统计
     */
    @NonNull
    public MemoryStats getStats() {
        int totalEntries = dao.getCount();
        int totalKeys = dao.getKeyCount();
        int totalCategories = dao.getAllCategories().size();

        MemoryCache.CacheStats cacheStats = cache.getStats();

        return new MemoryStats(totalKeys, totalEntries, totalCategories,
                cacheStats.cacheSize, cacheStats.recentKeyCount);
    }

    public static class MemoryStats {
        public final int totalKeys;
        public final int totalEntries;
        public final int totalCategories;
        public final int cacheSize;
        public final int hotKeyCount;

        public MemoryStats(int totalKeys, int totalEntries, int totalCategories,
                           int cacheSize, int hotKeyCount) {
            this.totalKeys = totalKeys;
            this.totalEntries = totalEntries;
            this.totalCategories = totalCategories;
            this.cacheSize = cacheSize;
            this.hotKeyCount = hotKeyCount;
        }
    }

    // ===== 内部方法 =====

    /**
     * 从数据库加载数据到缓存
     */
    private void loadFromDatabase() {
        executor.execute(() -> {
            try {
                // 加载热点数据到L1缓存
                List<String> keys = dao.getAllKeys();
                for (String key : keys) {
                    // 加载每个key的top推荐到缓存
                    MemoryEntity top = dao.getTopRecommendation(key);
                    if (top != null) {
                        cache.put(key, top.toEntry());
                    }

                    // 加载最近使用的到lastActions
                    MemoryEntity last = dao.getLastUsed(key);
                    if (last != null) {
                        lastActions.put(key, last.toEntry());
                    }
                }

                Log.i(TAG, "Loaded memories: " + dao.getCount() + " entries, " +
                        cache.size() + " cached");
            } catch (Exception e) {
                Log.e(TAG, "Failed to load memories", e);
            }
        });
    }

    /**
     * 清除实例（用于测试）
     */
    public static void clearInstance() {
        if (instance != null) {
            instance.executor.shutdown();
            MemoryDatabase.clearInstance();
            instance = null;
        }
    }

    // ===== v2.2.0: 跨设备同步 =====

    /**
     * 启用记忆同步
     */
    public void enableSync(@NonNull String identityId, @Nullable String centerAddress, int centerPort) {
        if (syncService != null) {
            syncService.disableSync();
        }

        syncService = new MemorySyncService(
            context,  // 使用 this.context 但这里没有 context 变量，需要修复
            this,
            identityId,
            centerAddress,
            centerPort
        );

        syncService.enableSync();
        Log.i(TAG, "Memory sync enabled with Center: " + centerAddress);
    }

    /**
     * 禁用记忆同步
     */
    public void disableSync() {
        if (syncService != null) {
            syncService.disableSync();
            syncService = null;
        }
        Log.i(TAG, "Memory sync disabled");
    }

    /**
     * 同步记忆到 Center
     */
    public void syncToCenter() {
        if (syncService != null) {
            syncService.sync();
        }
    }

    /**
     * 从 Center 全量恢复记忆
     */
    public void restoreFromCenter() {
        if (syncService != null) {
            syncService.fullSync();
        }
    }

    /**
     * 获取同步服务
     */
    @Nullable
    public MemorySyncService getSyncService() {
        return syncService;
    }

    /**
     * 设置同步状态监听器
     */
    public void setSyncStatusListener(@Nullable MemorySyncService.SyncStatusListener listener) {
        if (syncService != null) {
            syncService.setStatusListener(listener);
        }
    }

    /**
     * 获取所有记忆（用于同步）
     */
    @NonNull
    public List<MemoryEntry> getAllMemories() {
        List<MemoryEntity> entities = dao.getAll();
        List<MemoryEntry> entries = new ArrayList<>();
        for (MemoryEntity entity : entities) {
            entries.add(entity.toEntry());
        }
        return entries;
    }
}