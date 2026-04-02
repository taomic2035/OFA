package com.ofa.agent.memory;

import androidx.room.Dao;
import androidx.room.Delete;
import androidx.room.Insert;
import androidx.room.OnConflictStrategy;
import androidx.room.Query;
import androidx.room.Update;

import java.util.List;

/**
 * Memory DAO - Room数据库访问对象
 */
@Dao
public interface MemoryDao {

    // ===== 插入和更新 =====

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    long insert(MemoryEntity memory);

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    void insertAll(List<MemoryEntity> memories);

    @Update
    void update(MemoryEntity memory);

    // ===== 查询 =====

    @Query("SELECT * FROM memories WHERE `key` = :key ORDER BY score DESC, count DESC")
    List<MemoryEntity> getByKey(String key);

    @Query("SELECT * FROM memories WHERE `key` = :key ORDER BY score DESC, count DESC LIMIT :limit")
    List<MemoryEntity> getByKeyLimit(String key, int limit);

    @Query("SELECT * FROM memories WHERE category = :category ORDER BY timestamp DESC")
    List<MemoryEntity> getByCategory(String category);

    @Query("SELECT * FROM memories WHERE `key` = :key AND value = :value LIMIT 1")
    MemoryEntity getByKeyValue(String key, String value);

    @Query("SELECT * FROM memories WHERE `key` LIKE :prefix || '.%' ORDER BY score DESC")
    List<MemoryEntity> getByKeyPrefix(String prefix);

    // ===== 推荐查询 =====

    @Query("SELECT * FROM memories WHERE `key` = :key ORDER BY score DESC, count DESC, timestamp DESC LIMIT 1")
    MemoryEntity getTopRecommendation(String key);

    @Query("SELECT * FROM memories WHERE `key` = :key ORDER BY timestamp DESC LIMIT 1")
    MemoryEntity getLastUsed(String key);

    @Query("SELECT * FROM memories WHERE `key` = :key ORDER BY count DESC LIMIT 1")
    MemoryEntity getMostUsed(String key);

    // ===== 搜索 =====

    @Query("SELECT * FROM memories WHERE value LIKE '%' || :query || '%' ORDER BY score DESC")
    List<MemoryEntity> searchByValue(String query);

    @Query("SELECT * FROM memories WHERE `key` = :key AND value LIKE :prefix || '%' ORDER BY score DESC LIMIT :limit")
    List<MemoryEntity> autocomplete(String key, String prefix, int limit);

    // ===== 统计 =====

    @Query("SELECT COUNT(*) FROM memories")
    int getCount();

    @Query("SELECT COUNT(DISTINCT `key`) FROM memories")
    int getKeyCount();

    @Query("SELECT COUNT(*) FROM memories WHERE category = :category")
    int getCountByCategory(String category);

    @Query("SELECT SUM(count) FROM memories WHERE `key` = :key")
    int getTotalUsageCount(String key);

    // ===== 清理 =====

    @Delete
    void delete(MemoryEntity memory);

    @Query("DELETE FROM memories WHERE `key` = :key AND value = :value")
    void deleteByKeyValue(String key, String value);

    @Query("DELETE FROM memories WHERE `key` = :key")
    void deleteByKey(String key);

    @Query("DELETE FROM memories WHERE timestamp < :timestamp")
    int deleteOlderThan(long timestamp);

    @Query("DELETE FROM memories WHERE category = :category")
    void deleteByCategory(String category);

    @Query("DELETE FROM memories")
    void deleteAll();

    // ===== 批量操作 =====

    @Query("SELECT * FROM memories ORDER BY timestamp DESC")
    List<MemoryEntity> getAll();

    @Query("SELECT DISTINCT category FROM memories")
    List<String> getAllCategories();

    @Query("SELECT DISTINCT `key` FROM memories")
    List<String> getAllKeys();
}