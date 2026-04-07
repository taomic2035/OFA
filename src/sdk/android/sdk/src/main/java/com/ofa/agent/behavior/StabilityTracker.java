package com.ofa.agent.behavior;

import android.content.Context;
import android.content.SharedPreferences;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 性格稳定性追踪器 (v2.9.0)
 *
 * 追踪性格稳定性和 MBTI 类型收敛趋势。
 * Center 是永远在线的灵魂载体，性格随时间趋于稳定。
 *
 * 主要功能：
 * - 性格稳定性追踪
 * - MBTI 类型收敛检测
 * - 行为时间衰减
 * - 重大事件检测
 */
public class StabilityTracker {

    private static final String TAG = "StabilityTracker";
    private static final String PREFS_NAME = "ofa_stability";
    private static final String KEY_SNAPSHOTS = "snapshots";
    private static final String KEY_EVENTS = "major_events";
    private static final String KEY_STATE = "state";

    // 配置
    private static final float STABILITY_THRESHOLD = 0.8f;
    private static final int CONVERGENCE_WINDOW = 20;
    private static final int MBTI_STABILITY_WINDOW = 50;
    private static final long DECAY_HALF_LIFE_MS = 30L * 24 * 60 * 60 * 1000; // 30天
    private static final float MAJOR_EVENT_THRESHOLD = 0.3f;
    private static final int MAX_SNAPSHOTS = 90;

    private final Context context;
    private final SharedPreferences prefs;
    private final ExecutorService executor;

    // 性格快照历史
    private final List<PersonalitySnapshot> snapshots;

    // 重大事件
    private final List<MajorEvent> majorEvents;

    // 当前状态
    private PersonalityState currentState;

    // 监听器
    private StabilityListener listener;

    public StabilityTracker(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
        this.executor = Executors.newSingleThreadExecutor();
        this.snapshots = new ArrayList<>();
        this.majorEvents = new ArrayList<>();

        // 加载存储的数据
        loadData();
    }

    /**
     * 记录性格快照
     */
    public void recordSnapshot(@NonNull PersonalityData personality, @NonNull String trigger) {
        executor.execute(() -> {
            PersonalitySnapshot snapshot = new PersonalitySnapshot();
            snapshot.timestamp = System.currentTimeMillis();
            snapshot.openness = personality.openness;
            snapshot.conscientiousness = personality.conscientiousness;
            snapshot.extraversion = personality.extraversion;
            snapshot.agreeableness = personality.agreeableness;
            snapshot.neuroticism = personality.neuroticism;
            snapshot.mbtiType = personality.mbtiType;
            snapshot.mbtiConfidence = personality.mbtiConfidence;
            snapshot.stabilityScore = personality.stabilityScore;
            snapshot.observedCount = personality.observedCount;
            snapshot.trigger = trigger;

            synchronized (snapshots) {
                snapshots.add(snapshot);

                // 限制快照数量
                while (snapshots.size() > MAX_SNAPSHOTS) {
                    snapshots.remove(0);
                }
            }

            // 更新状态
            updateState(personality);

            // 保存数据
            saveData();

            // 通知监听器
            notifySnapshotRecorded(snapshot);
        });
    }

    /**
     * 检查性格稳定性
     */
    @NonNull
    public StabilityReport checkStability() {
        StabilityReport report = new StabilityReport();
        report.checkedAt = System.currentTimeMillis();

        synchronized (snapshots) {
            if (snapshots.size() < CONVERGENCE_WINDOW) {
                report.isStable = false;
                report.reason = "insufficient_data";
                report.requiredObservations = CONVERGENCE_WINDOW - snapshots.size();
                return report;
            }

            // 计算方差
            float variance = calculateVariance();
            report.variance = variance;

            // 获取最新的稳定分数
            PersonalitySnapshot latest = snapshots.get(snapshots.size() - 1);
            report.stabilityScore = latest.stabilityScore;

            // 判断稳定性
            if (variance < 0.05f && latest.stabilityScore >= STABILITY_THRESHOLD) {
                report.isStable = true;
                report.reason = "converged";
            } else if (latest.stabilityScore >= STABILITY_THRESHOLD) {
                report.isStable = true;
                report.reason = "high_stability";
            } else {
                report.isStable = false;
                report.reason = "still_evolving";
            }

            // MBTI 稳定性
            report.mbtiStable = checkMBTIStability();
            report.stableMbtiType = currentState != null ? currentState.stableMbtiType : null;
        }

        return report;
    }

    /**
     * 获取趋势分析
     */
    @Nullable
    public TrendAnalysis getTrendAnalysis() {
        synchronized (snapshots) {
            if (snapshots.size() < 5) {
                return null;
            }

            TrendAnalysis analysis = new TrendAnalysis();
            analysis.analyzedAt = System.currentTimeMillis();

            // 计算各维度趋势
            analysis.opennessTrend = calculateTrend(s -> s.openness);
            analysis.conscientiousnessTrend = calculateTrend(s -> s.conscientiousness);
            analysis.extraversionTrend = calculateTrend(s -> s.extraversion);
            analysis.agreeablenessTrend = calculateTrend(s -> s.agreeableness);
            analysis.neuroticismTrend = calculateTrend(s -> s.neuroticism);

            // MBTI 变化历史
            analysis.mbtiHistory = buildMBTIHistory();

            // 预测 MBTI
            if (!snapshots.isEmpty()) {
                PersonalitySnapshot latest = snapshots.get(snapshots.size() - 1);
                analysis.predictedMbti = latest.mbtiType;
                analysis.predictionConfidence = latest.mbtiConfidence;
            }

            return analysis;
        }
    }

    /**
     * 记录重大事件
     */
    public void recordMajorEvent(@NonNull String type, @NonNull String description,
                                  @NonNull Map<String, Float> impact) {
        MajorEvent event = new MajorEvent();
        event.id = "evt_" + System.currentTimeMillis();
        event.type = type;
        event.description = description;
        event.observedAt = System.currentTimeMillis();
        event.impact = new HashMap<>(impact);

        synchronized (majorEvents) {
            majorEvents.add(event);
        }

        saveData();
        Log.i(TAG, "Major event recorded: " + type);
    }

    /**
     * 应用时间衰减到观察
     */
    public float applyTimeDecay(float value, long observationTime) {
        long now = System.currentTimeMillis();
        float ageHours = (now - observationTime) / (1000f * 60 * 60);

        // 指数衰减
        double decayFactor = Math.pow(0.5, ageHours / (DECAY_HALF_LIFE_MS / (1000f * 60 * 60)));

        return (float) (value * decayFactor);
    }

    /**
     * 获取当前状态
     */
    @Nullable
    public PersonalityState getCurrentState() {
        return currentState;
    }

    /**
     * 设置监听器
     */
    public void setListener(StabilityListener listener) {
        this.listener = listener;
    }

    // === 私有方法 ===

    private void loadData() {
        // 加载快照
        String snapshotsJson = prefs.getString(KEY_SNAPSHOTS, "[]");
        try {
            JSONArray array = new JSONArray(snapshotsJson);
            synchronized (snapshots) {
                snapshots.clear();
                for (int i = 0; i < array.length(); i++) {
                    snapshots.add(PersonalitySnapshot.fromJson(array.getJSONObject(i)));
                }
            }
        } catch (JSONException e) {
            Log.e(TAG, "Failed to load snapshots", e);
        }

        // 加载重大事件
        String eventsJson = prefs.getString(KEY_EVENTS, "[]");
        try {
            JSONArray array = new JSONArray(eventsJson);
            synchronized (majorEvents) {
                majorEvents.clear();
                for (int i = 0; i < array.length(); i++) {
                    majorEvents.add(MajorEvent.fromJson(array.getJSONObject(i)));
                }
            }
        } catch (JSONException e) {
            Log.e(TAG, "Failed to load events", e);
        }

        // 加载状态
        String stateJson = prefs.getString(KEY_STATE, null);
        if (stateJson != null) {
            try {
                currentState = PersonalityState.fromJson(new JSONObject(stateJson));
            } catch (JSONException e) {
                Log.e(TAG, "Failed to load state", e);
            }
        }

        if (currentState == null) {
            currentState = new PersonalityState();
        }
    }

    private void saveData() {
        executor.execute(() -> {
            try {
                // 保存快照
                JSONArray snapshotsArray = new JSONArray();
                synchronized (snapshots) {
                    for (PersonalitySnapshot s : snapshots) {
                        snapshotsArray.put(s.toJson());
                    }
                }
                prefs.edit().putString(KEY_SNAPSHOTS, snapshotsArray.toString()).apply();

                // 保存重大事件
                JSONArray eventsArray = new JSONArray();
                synchronized (majorEvents) {
                    for (MajorEvent e : majorEvents) {
                        eventsArray.put(e.toJson());
                    }
                }
                prefs.edit().putString(KEY_EVENTS, eventsArray.toString()).apply();

                // 保存状态
                if (currentState != null) {
                    prefs.edit().putString(KEY_STATE, currentState.toJson().toString()).apply();
                }
            } catch (JSONException e) {
                Log.e(TAG, "Failed to save data", e);
            }
        });
    }

    private void updateState(PersonalityData personality) {
        if (currentState == null) {
            currentState = new PersonalityState();
        }

        currentState.lastUpdated = System.currentTimeMillis();
        currentState.observationCount++;
        currentState.stabilityScore = personality.stabilityScore;

        // 检查稳定性
        if (personality.stabilityScore >= STABILITY_THRESHOLD) {
            currentState.isStable = true;
            if (currentState.stabilizedAt == 0) {
                currentState.stabilizedAt = System.currentTimeMillis();
            }
        }

        // MBTI 稳定性
        if (checkMBTIStability()) {
            currentState.mbtiStable = true;
            currentState.stableMbtiType = personality.mbtiType;
        }
    }

    private float calculateVariance() {
        if (snapshots.size() < 2) {
            return 1.0f;
        }

        // 计算各维度方差
        float opennessVar = calculateTraitVariance(s -> s.openness);
        float conscientiousnessVar = calculateTraitVariance(s -> s.conscientiousness);
        float extraversionVar = calculateTraitVariance(s -> s.extraversion);
        float agreeablenessVar = calculateTraitVariance(s -> s.agreeableness);
        float neuroticismVar = calculateTraitVariance(s -> s.neuroticism);

        return (opennessVar + conscientiousnessVar + extraversionVar + agreeablenessVar + neuroticismVar) / 5;
    }

    private float calculateTraitVariance(TraitExtractor extractor) {
        float sum = 0, sumSq = 0;
        int n = snapshots.size();

        for (PersonalitySnapshot s : snapshots) {
            float v = extractor.extract(s);
            sum += v;
            sumSq += v * v;
        }

        float mean = sum / n;
        return sumSq / n - mean * mean;
    }

    private boolean checkMBTIStability() {
        if (snapshots.size() < MBTI_STABILITY_WINDOW) {
            return false;
        }

        // 获取最近的 MBTI 类型
        String currentType = snapshots.get(snapshots.size() - 1).mbtiType;
        if (currentType == null || currentType.isEmpty()) {
            return false;
        }

        // 检查最近 N 个快照是否一致
        int start = Math.max(0, snapshots.size() - MBTI_STABILITY_WINDOW);
        for (int i = start; i < snapshots.size(); i++) {
            if (!currentType.equals(snapshots.get(i).mbtiType)) {
                return false;
            }
        }

        return true;
    }

    private float calculateTrend(TraitExtractor extractor) {
        int n = snapshots.size();
        if (n < 2) {
            return 0;
        }

        float sumX = 0, sumY = 0, sumXY = 0, sumX2 = 0;

        for (int i = 0; i < n; i++) {
            float x = i;
            float y = extractor.extract(snapshots.get(i));
            sumX += x;
            sumY += y;
            sumXY += x * y;
            sumX2 += x * x;
        }

        // 线性回归斜率
        return (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
    }

    private List<MbtiTransition> buildMBTIHistory() {
        List<MbtiTransition> history = new ArrayList<>();
        String prevType = null;

        for (PersonalitySnapshot s : snapshots) {
            if (prevType != null && !prevType.equals(s.mbtiType)) {
                MbtiTransition transition = new MbtiTransition();
                transition.fromType = prevType;
                transition.toType = s.mbtiType;
                transition.timestamp = s.timestamp;
                transition.confidence = s.mbtiConfidence;
                history.add(transition);
            }
            prevType = s.mbtiType;
        }

        return history;
    }

    private void notifySnapshotRecorded(PersonalitySnapshot snapshot) {
        if (listener != null) {
            listener.onSnapshotRecorded(snapshot);
        }
    }

    // === 内部类 ===

    public static class PersonalitySnapshot {
        public long timestamp;
        public float openness;
        public float conscientiousness;
        public float extraversion;
        public float agreeableness;
        public float neuroticism;
        public String mbtiType;
        public float mbtiConfidence;
        public float stabilityScore;
        public int observedCount;
        public String trigger;

        public JSONObject toJson() throws JSONException {
            JSONObject json = new JSONObject();
            json.put("timestamp", timestamp);
            json.put("openness", openness);
            json.put("conscientiousness", conscientiousness);
            json.put("extraversion", extraversion);
            json.put("agreeableness", agreeableness);
            json.put("neuroticism", neuroticism);
            json.put("mbti_type", mbtiType);
            json.put("mbti_confidence", mbtiConfidence);
            json.put("stability_score", stabilityScore);
            json.put("observed_count", observedCount);
            json.put("trigger", trigger);
            return json;
        }

        public static PersonalitySnapshot fromJson(JSONObject json) throws JSONException {
            PersonalitySnapshot s = new PersonalitySnapshot();
            s.timestamp = json.getLong("timestamp");
            s.openness = (float) json.optDouble("openness", 0.5);
            s.conscientiousness = (float) json.optDouble("conscientiousness", 0.5);
            s.extraversion = (float) json.optDouble("extraversion", 0.5);
            s.agreeableness = (float) json.optDouble("agreeableness", 0.5);
            s.neuroticism = (float) json.optDouble("neuroticism", 0.5);
            s.mbtiType = json.optString("mbti_type", "");
            s.mbtiConfidence = (float) json.optDouble("mbti_confidence", 0);
            s.stabilityScore = (float) json.optDouble("stability_score", 0);
            s.observedCount = json.optInt("observed_count", 0);
            s.trigger = json.optString("trigger", "");
            return s;
        }
    }

    public static class PersonalityState {
        public boolean isStable;
        public float stabilityScore;
        public long stabilizedAt;
        public boolean mbtiStable;
        public String stableMbtiType;
        public boolean mbtiLocked;
        public float convergenceRate;
        public float variance;
        public long lastUpdated;
        public int observationCount;

        public JSONObject toJson() throws JSONException {
            JSONObject json = new JSONObject();
            json.put("is_stable", isStable);
            json.put("stability_score", stabilityScore);
            json.put("stabilized_at", stabilizedAt);
            json.put("mbti_stable", mbtiStable);
            json.put("stable_mbti_type", stableMbtiType);
            json.put("mbti_locked", mbtiLocked);
            json.put("convergence_rate", convergenceRate);
            json.put("variance", variance);
            json.put("last_updated", lastUpdated);
            json.put("observation_count", observationCount);
            return json;
        }

        public static PersonalityState fromJson(JSONObject json) throws JSONException {
            PersonalityState s = new PersonalityState();
            s.isStable = json.optBoolean("is_stable", false);
            s.stabilityScore = (float) json.optDouble("stability_score", 0);
            s.stabilizedAt = json.optLong("stabilized_at", 0);
            s.mbtiStable = json.optBoolean("mbti_stable", false);
            s.stableMbtiType = json.optString("stable_mbti_type", null);
            s.mbtiLocked = json.optBoolean("mbti_locked", false);
            s.convergenceRate = (float) json.optDouble("convergence_rate", 0);
            s.variance = (float) json.optDouble("variance", 0);
            s.lastUpdated = json.optLong("last_updated", 0);
            s.observationCount = json.optInt("observation_count", 0);
            return s;
        }
    }

    public static class StabilityReport {
        public boolean isStable;
        public String reason;
        public float stabilityScore;
        public float variance;
        public boolean mbtiStable;
        public String stableMbtiType;
        public int requiredObservations;
        public long checkedAt;
    }

    public static class TrendAnalysis {
        public float opennessTrend;
        public float conscientiousnessTrend;
        public float extraversionTrend;
        public float agreeablenessTrend;
        public float neuroticismTrend;
        public List<MbtiTransition> mbtiHistory;
        public String predictedMbti;
        public float predictionConfidence;
        public long analyzedAt;
    }

    public static class MbtiTransition {
        public String fromType;
        public String toType;
        public long timestamp;
        public float confidence;
    }

    public static class MajorEvent {
        public String id;
        public String type;
        public String description;
        public long observedAt;
        public Map<String, Float> impact;

        public JSONObject toJson() throws JSONException {
            JSONObject json = new JSONObject();
            json.put("id", id);
            json.put("type", type);
            json.put("description", description);
            json.put("observed_at", observedAt);
            if (impact != null) {
                JSONObject impactJson = new JSONObject();
                for (Map.Entry<String, Float> e : impact.entrySet()) {
                    impactJson.put(e.getKey(), e.getValue());
                }
                json.put("impact", impactJson);
            }
            return json;
        }

        public static MajorEvent fromJson(JSONObject json) throws JSONException {
            MajorEvent e = new MajorEvent();
            e.id = json.getString("id");
            e.type = json.getString("type");
            e.description = json.getString("description");
            e.observedAt = json.getLong("observed_at");
            e.impact = new HashMap<>();
            JSONObject impactJson = json.optJSONObject("impact");
            if (impactJson != null) {
                for (java.util.Iterator<String> it = impactJson.keys(); it.hasNext(); ) {
                    String key = it.next();
                    e.impact.put(key, (float) impactJson.optDouble(key, 0));
                }
            }
            return e;
        }
    }

    public static class PersonalityData {
        public float openness;
        public float conscientiousness;
        public float extraversion;
        public float agreeableness;
        public float neuroticism;
        public String mbtiType;
        public float mbtiConfidence;
        public float stabilityScore;
        public int observedCount;
    }

    private interface TraitExtractor {
        float extract(PersonalitySnapshot s);
    }

    public interface StabilityListener {
        void onSnapshotRecorded(PersonalitySnapshot snapshot);
    }
}