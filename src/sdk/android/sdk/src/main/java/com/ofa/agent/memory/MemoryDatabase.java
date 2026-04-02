package com.ofa.agent.memory;

import android.content.Context;

import androidx.annotation.NonNull;
import androidx.room.Database;
import androidx.room.Room;
import androidx.room.RoomDatabase;
import androidx.room.migration.Migration;
import androidx.sqlite.db.SupportSQLiteDatabase;

/**
 * Memory Database - Room数据库
 */
@Database(entities = {MemoryEntity.class}, version = 1, exportSchema = false)
public abstract class MemoryDatabase extends RoomDatabase {

    private static volatile MemoryDatabase instance;
    private static final String DATABASE_NAME = "user_memory.db";

    public abstract MemoryDao memoryDao();

    /**
     * 获取单例实例
     */
    @NonNull
    public static synchronized MemoryDatabase getInstance(@NonNull Context context) {
        if (instance == null) {
            instance = buildDatabase(context.getApplicationContext());
        }
        return instance;
    }

    @NonNull
    private static MemoryDatabase buildDatabase(@NonNull Context context) {
        return Room.databaseBuilder(context, MemoryDatabase.class, DATABASE_NAME)
                .fallbackToDestructiveMigration() // 开发阶段允许破坏性迁移
                .addCallback(new Callback() {
                    @Override
                    public void onCreate(@NonNull SupportSQLiteDatabase db) {
                        super.onCreate(db);
                        // 数据库创建时的初始化
                    }
                })
                .build();
    }

    /**
     * 清除实例（用于测试）
     */
    public static void clearInstance() {
        if (instance != null) {
            instance.close();
            instance = null;
        }
    }
}