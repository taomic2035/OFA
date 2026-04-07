package com.ofa.agent.identity;

import android.content.Context;
import android.content.SharedPreferences;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.util.ArrayList;
import java.util.List;

/**
 * LocalIdentityStore - 本地身份存储
 *
 * 使用 SharedPreferences 和文件存储身份数据。
 * 支持离线存储和快速读取。
 */
public class LocalIdentityStore {

    private static final String TAG = "LocalIdentityStore";

    private static final String PREFS_NAME = "ofa_identity_prefs";
    private static final String KEY_CURRENT_ID = "current_identity_id";
    private static final String KEY_IDENTITY_JSON = "identity_json";
    private static final String KEY_VERSION = "identity_version";

    private final Context context;
    private final SharedPreferences prefs;
    private final File identityDir;

    private PersonalIdentity cachedIdentity;

    /**
     * 创建本地存储
     */
    public LocalIdentityStore(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
        this.identityDir = new File(context.getFilesDir(), "identity");

        // 确保目录存在
        if (!identityDir.exists()) {
            identityDir.mkdirs();
        }
    }

    // === 基本存储操作 ===

    /**
     * 保存身份
     */
    public boolean saveIdentity(@NonNull PersonalIdentity identity) {
        try {
            // 缓存
            cachedIdentity = identity;

            // 保存到 SharedPreferences
            prefs.edit()
                .putString(KEY_CURRENT_ID, identity.getId())
                .putString(KEY_IDENTITY_JSON, identity.toJson())
                .putLong(KEY_VERSION, identity.getVersion())
                .apply();

            // 同时保存到文件（备份）
            saveToFile(identity);

            Log.i(TAG, "Identity saved: " + identity.getId() + " v" + identity.getVersion());
            return true;

        } catch (Exception e) {
            Log.e(TAG, "Failed to save identity", e);
            return false;
        }
    }

    /**
     * 加载身份
     */
    @Nullable
    public PersonalIdentity loadIdentity() {
        // 优先返回缓存
        if (cachedIdentity != null) {
            return cachedIdentity;
        }

        try {
            // 从 SharedPreferences 加载
            String json = prefs.getString(KEY_IDENTITY_JSON, null);
            if (json != null) {
                cachedIdentity = PersonalIdentity.fromJson(json);
                if (cachedIdentity != null) {
                    Log.i(TAG, "Identity loaded from prefs: " + cachedIdentity.getId());
                    return cachedIdentity;
                }
            }

            // 从文件加载
            cachedIdentity = loadFromFile();
            if (cachedIdentity != null) {
                Log.i(TAG, "Identity loaded from file: " + cachedIdentity.getId());
                return cachedIdentity;
            }

        } catch (Exception e) {
            Log.e(TAG, "Failed to load identity", e);
        }

        return null;
    }

    /**
     * 加载指定身份
     */
    @Nullable
    public PersonalIdentity loadIdentity(@NonNull String identityId) {
        try {
            File file = new File(identityDir, identityId + ".json");
            if (file.exists()) {
                return loadFromFile(file);
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to load identity: " + identityId, e);
        }
        return null;
    }

    /**
     * 获取当前身份 ID
     */
    @Nullable
    public String getCurrentIdentityId() {
        return prefs.getString(KEY_CURRENT_ID, null);
    }

    /**
     * 设置当前身份 ID
     */
    public void setCurrentIdentityId(@NonNull String identityId) {
        prefs.edit()
            .putString(KEY_CURRENT_ID, identityId)
            .apply();
    }

    /**
     * 获取版本号
     */
    public long getVersion() {
        return prefs.getLong(KEY_VERSION, 0);
    }

    /**
     * 检查是否有本地身份
     */
    public boolean hasLocalIdentity() {
        return prefs.contains(KEY_IDENTITY_JSON) ||
               new File(identityDir, "current.json").exists();
    }

    /**
     * 删除身份
     */
    public boolean deleteIdentity(@NonNull String identityId) {
        try {
            // 清除缓存
            if (cachedIdentity != null && cachedIdentity.getId().equals(identityId)) {
                cachedIdentity = null;
            }

            // 清除 SharedPreferences
            if (prefs.getString(KEY_CURRENT_ID, "").equals(identityId)) {
                prefs.edit()
                    .remove(KEY_CURRENT_ID)
                    .remove(KEY_IDENTITY_JSON)
                    .remove(KEY_VERSION)
                    .apply();
            }

            // 删除文件
            File file = new File(identityDir, identityId + ".json");
            if (file.exists()) {
                file.delete();
            }

            Log.i(TAG, "Identity deleted: " + identityId);
            return true;

        } catch (Exception e) {
            Log.e(TAG, "Failed to delete identity", e);
            return false;
        }
    }

    // === 列出所有身份 ===

    /**
     * 列出所有本地身份 ID
     */
    @NonNull
    public List<String> listIdentityIds() {
        List<String> ids = new ArrayList<>();

        // 从文件目录获取
        File[] files = identityDir.listFiles((dir, name) -> name.endsWith(".json"));
        if (files != null) {
            for (File file : files) {
                String name = file.getName();
                ids.add(name.substring(0, name.length() - 5)); // 移除 .json
            }
        }

        return ids;
    }

    // === 文件操作 ===

    private void saveToFile(@NonNull PersonalIdentity identity) {
        try {
            // 保存当前身份
            File currentFile = new File(identityDir, "current.json");
            writeFile(currentFile, identity.toJson());

            // 保存备份（按 ID）
            File idFile = new File(identityDir, identity.getId() + ".json");
            writeFile(idFile, identity.toJson());

        } catch (Exception e) {
            Log.e(TAG, "Failed to save identity to file", e);
        }
    }

    @Nullable
    private PersonalIdentity loadFromFile() {
        File currentFile = new File(identityDir, "current.json");
        if (currentFile.exists()) {
            return loadFromFile(currentFile);
        }
        return null;
    }

    @Nullable
    private PersonalIdentity loadFromFile(@NonNull File file) {
        try {
            String json = readFile(file);
            if (json != null) {
                return PersonalIdentity.fromJson(json);
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to load identity from file: " + file.getName(), e);
        }
        return null;
    }

    private void writeFile(@NonNull File file, @NonNull String content) throws Exception {
        FileOutputStream fos = new FileOutputStream(file);
        fos.write(content.getBytes("UTF-8"));
        fos.close();
    }

    @Nullable
    private String readFile(@NonNull File file) throws Exception {
        if (!file.exists()) return null;

        FileInputStream fis = new FileInputStream(file);
        byte[] data = new byte[(int) file.length()];
        int read = fis.read(data);
        fis.close();

        if (read > 0) {
            return new String(data, 0, read, "UTF-8");
        }
        return null;
    }

    // === 清理 ===

    /**
     * 清除所有本地身份
     */
    public void clearAll() {
        prefs.edit().clear().apply();

        File[] files = identityDir.listFiles();
        if (files != null) {
            for (File file : files) {
                file.delete();
            }
        }

        cachedIdentity = null;
        Log.i(TAG, "All identities cleared");
    }
}