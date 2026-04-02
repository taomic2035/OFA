package com.ofa.agent.skill;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.skill.builtin.OfflineSkills;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.FileWriter;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 技能注册表
 * 管理所有注册的技能
 */
public class SkillRegistry {

    private static final String TAG = "SkillRegistry";
    private static SkillRegistry instance;

    private final Map<String, SkillDefinition> skills;
    private final Map<String, SkillExecutor> executors;
    private final Map<String, List<SkillDefinition>> categoryIndex;
    private final Context context;
    private File skillsDir;

    private SkillRegistry(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.skills = new ConcurrentHashMap<>();
        this.executors = new ConcurrentHashMap<>();
        this.categoryIndex = new ConcurrentHashMap<>();

        // 初始化技能存储目录
        skillsDir = new File(context.getFilesDir(), "skills");
        if (!skillsDir.exists()) {
            skillsDir.mkdirs();
        }

        // 注册内置技能
        OfflineSkills.registerAll(this);

        // 加载用户定义的技能
        loadUserSkills();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static synchronized SkillRegistry getInstance(@NonNull Context context) {
        if (instance == null) {
            instance = new SkillRegistry(context);
        }
        return instance;
    }

    /**
     * 获取实例（需要先初始化）
     */
    @NonNull
    public static SkillRegistry getInstance() {
        if (instance == null) {
            throw new IllegalStateException("SkillRegistry not initialized. Call getInstance(context) first.");
        }
        return instance;
    }

    // ===== 注册方法 =====

    /**
     * 注册技能
     */
    public void register(@NonNull SkillDefinition skill) {
        skills.put(skill.getId(), skill);
        categoryIndex.computeIfAbsent(skill.getCategory(), k -> new ArrayList<>()).add(skill);
        Log.i(TAG, "Registered skill: " + skill.getId());
    }

    /**
     * 注册技能（带执行器）
     */
    public void register(@NonNull SkillDefinition skill, @NonNull SkillExecutor executor) {
        skills.put(skill.getId(), skill);
        executors.put(skill.getId(), executor);
        categoryIndex.computeIfAbsent(skill.getCategory(), k -> new ArrayList<>()).add(skill);
        Log.i(TAG, "Registered skill with executor: " + skill.getId());
    }

    /**
     * 注册简单技能执行器
     */
    public void register(@NonNull SkillExecutor executor) {
        executors.put(executor.getSkillId(), executor);
        Log.i(TAG, "Registered simple executor: " + executor.getSkillId());
    }

    /**
     * 注销技能
     */
    public void unregister(@NonNull String skillId) {
        SkillDefinition removed = skills.remove(skillId);
        if (removed != null) {
            List<SkillDefinition> categorySkills = categoryIndex.get(removed.getCategory());
            if (categorySkills != null) {
                categorySkills.remove(removed);
            }
        }
        executors.remove(skillId);
        Log.i(TAG, "Unregistered skill: " + skillId);
    }

    // ===== 查询方法 =====

    /**
     * 获取技能定义
     */
    @Nullable
    public SkillDefinition getSkill(@NonNull String skillId) {
        return skills.get(skillId);
    }

    /**
     * 获取执行器
     */
    @Nullable
    public SkillExecutor getExecutor(@NonNull String skillId) {
        return executors.get(skillId);
    }

    /**
     * 检查技能是否注册
     */
    public boolean hasSkill(@NonNull String skillId) {
        return skills.containsKey(skillId) || executors.containsKey(skillId);
    }

    /**
     * 获取所有技能
     */
    @NonNull
    public List<SkillDefinition> getAllSkills() {
        return new ArrayList<>(skills.values());
    }

    /**
     * 获取某分类下的技能
     */
    @NonNull
    public List<SkillDefinition> getSkillsByCategory(@NonNull String category) {
        List<SkillDefinition> result = categoryIndex.get(category);
        return result != null ? new ArrayList<>(result) : Collections.emptyList();
    }

    /**
     * 搜索技能
     */
    @NonNull
    public List<SkillDefinition> searchSkills(@NonNull String query) {
        List<SkillDefinition> results = new ArrayList<>();
        String lowerQuery = query.toLowerCase();

        for (SkillDefinition skill : skills.values()) {
            if (skill.getName().toLowerCase().contains(lowerQuery) ||
                    skill.getDescription().toLowerCase().contains(lowerQuery) ||
                    skill.getTags().stream().anyMatch(t -> t.toLowerCase().contains(lowerQuery))) {
                results.add(skill);
            }
        }

        // 按名称排序
        Collections.sort(results, Comparator.comparing(SkillDefinition::getName));
        return results;
    }

    /**
     * 根据触发器匹配技能
     */
    @Nullable
    public SkillDefinition matchTrigger(@NonNull String input) {
        String lowerInput = input.toLowerCase();

        for (SkillDefinition skill : skills.values()) {
            for (SkillDefinition.Trigger trigger : skill.getTriggers()) {
                if ("intent".equals(trigger.type) || "voice".equals(trigger.type)) {
                    if (trigger.pattern != null && lowerInput.contains(trigger.pattern.toLowerCase())) {
                        return skill;
                    }
                }
            }
        }
        return null;
    }

    /**
     * 获取技能数量
     */
    public int getSkillCount() {
        return skills.size();
    }

    // ===== 用户技能管理 =====

    /**
     * 保存用户定义的技能
     */
    public boolean saveSkill(@NonNull SkillDefinition skill) {
        try {
            File skillFile = new File(skillsDir, skill.getId() + ".json");
            JSONObject json = skillToJson(skill);
            FileWriter writer = new FileWriter(skillFile);
            writer.write(json.toString(2));
            writer.close();

            // 注册到内存
            register(skill);
            Log.i(TAG, "Saved skill: " + skill.getId());
            return true;
        } catch (Exception e) {
            Log.e(TAG, "Failed to save skill: " + skill.getId(), e);
            return false;
        }
    }

    /**
     * 删除用户定义的技能
     */
    public boolean deleteSkill(@NonNull String skillId) {
        try {
            File skillFile = new File(skillsDir, skillId + ".json");
            if (skillFile.exists()) {
                skillFile.delete();
            }
            unregister(skillId);
            Log.i(TAG, "Deleted skill: " + skillId);
            return true;
        } catch (Exception e) {
            Log.e(TAG, "Failed to delete skill: " + skillId, e);
            return false;
        }
    }

    /**
     * 加载用户定义的技能
     */
    private void loadUserSkills() {
        if (skillsDir == null || !skillsDir.exists()) return;

        File[] files = skillsDir.listFiles((dir, name) -> name.endsWith(".json"));
        if (files == null) return;

        for (File file : files) {
            try {
                BufferedReader reader = new BufferedReader(new FileReader(file));
                StringBuilder sb = new StringBuilder();
                String line;
                while ((line = reader.readLine()) != null) {
                    sb.append(line);
                }
                reader.close();

                JSONObject json = new JSONObject(sb.toString());
                SkillDefinition skill = jsonToSkill(json);
                if (skill != null) {
                    register(skill);
                }
            } catch (Exception e) {
                Log.e(TAG, "Failed to load skill from: " + file.getName(), e);
            }
        }
    }

    // ===== JSON 序列化 =====

    @NonNull
    private JSONObject skillToJson(@NonNull SkillDefinition skill) throws Exception {
        JSONObject json = new JSONObject();
        json.put("id", skill.getId());
        json.put("name", skill.getName());
        json.put("description", skill.getDescription());
        json.put("category", skill.getCategory());
        json.put("version", skill.getVersion());

        // 标签
        JSONArray tagsArray = new JSONArray();
        for (String tag : skill.getTags()) {
            tagsArray.put(tag);
        }
        json.put("tags", tagsArray);

        // 步骤
        JSONArray stepsArray = new JSONArray();
        for (SkillStep step : skill.getSteps()) {
            stepsArray.put(stepToJson(step));
        }
        json.put("steps", stepsArray);

        // 触发器
        JSONArray triggersArray = new JSONArray();
        for (SkillDefinition.Trigger trigger : skill.getTriggers()) {
            JSONObject triggerJson = new JSONObject();
            triggerJson.put("type", trigger.type);
            triggerJson.put("pattern", trigger.pattern);
            triggersArray.put(triggerJson);
        }
        json.put("triggers", triggersArray);

        return json;
    }

    @NonNull
    private JSONObject stepToJson(@NonNull SkillStep step) throws Exception {
        JSONObject json = new JSONObject();
        json.put("id", step.getId());
        json.put("name", step.getName());
        json.put("type", step.getType().name());
        json.put("action", step.getAction());
        json.put("timeout", step.getTimeout());
        json.put("retryCount", step.getRetryCount());
        json.put("retryDelay", step.getRetryDelay());
        json.put("optional", step.isOptional());
        json.put("nextStep", step.getNextStep());
        json.put("description", step.getDescription());

        // 参数
        JSONObject paramsJson = new JSONObject(step.getParams());
        json.put("params", paramsJson);

        return json;
    }

    @Nullable
    private SkillDefinition jsonToSkill(@NonNull JSONObject json) {
        try {
            SkillDefinition.Builder builder = new SkillDefinition.Builder()
                    .id(json.getString("id"))
                    .name(json.getString("name"))
                    .description(json.optString("description", ""))
                    .category(json.optString("category", "custom"))
                    .version(json.optString("version", "1.0.0"));

            // 标签
            JSONArray tagsArray = json.optJSONArray("tags");
            if (tagsArray != null) {
                for (int i = 0; i < tagsArray.length(); i++) {
                    builder.tag(tagsArray.getString(i));
                }
            }

            // 步骤
            JSONArray stepsArray = json.optJSONArray("steps");
            if (stepsArray != null) {
                for (int i = 0; i < stepsArray.length(); i++) {
                    JSONObject stepJson = stepsArray.getJSONObject(i);
                    SkillStep step = jsonToStep(stepJson);
                    if (step != null) {
                        builder.step(step);
                    }
                }
            }

            // 触发器
            JSONArray triggersArray = json.optJSONArray("triggers");
            if (triggersArray != null) {
                for (int i = 0; i < triggersArray.length(); i++) {
                    JSONObject triggerJson = triggersArray.getJSONObject(i);
                    builder.trigger(triggerJson.getString("type"), triggerJson.optString("pattern"));
                }
            }

            return builder.build();
        } catch (Exception e) {
            Log.e(TAG, "Failed to parse skill JSON", e);
            return null;
        }
    }

    @Nullable
    private SkillStep jsonToStep(@NonNull JSONObject json) {
        try {
            SkillStep.Builder builder = new SkillStep.Builder()
                    .id(json.getString("id"))
                    .name(json.getString("name"))
                    .type(SkillStep.StepType.valueOf(json.getString("type")))
                    .action(json.optString("action", null))
                    .timeout(json.optInt("timeout", 30000))
                    .retry(json.optInt("retryCount", 0), json.optInt("retryDelay", 1000))
                    .optional(json.optBoolean("optional", false))
                    .nextStep(json.optString("nextStep", null))
                    .description(json.optString("description", null));

            // 参数
            JSONObject paramsJson = json.optJSONObject("params");
            if (paramsJson != null) {
                Map<String, Object> params = new ConcurrentHashMap<>();
                for (java.util.Iterator<String> it = paramsJson.keys(); it.hasNext(); ) {
                    String key = it.next();
                    params.put(key, paramsJson.get(key));
                }
                builder.params(params);
            }

            return builder.build();
        } catch (Exception e) {
            Log.e(TAG, "Failed to parse step JSON", e);
            return null;
        }
    }

    /**
     * 清空所有技能
     */
    public void clear() {
        skills.clear();
        executors.clear();
        categoryIndex.clear();
    }
}