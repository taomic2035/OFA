package com.ofa.agent.distributed;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;
import com.ofa.agent.state.DeviceState;
import com.ofa.agent.state.StateSyncService;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 场景感知路由器 (v3.2.0)
 *
 * 根据当前场景智能路由消息到最合适的设备。
 * 支持自定义路由规则和设备能力匹配。
 */
public class SceneAwareRouter {

    private static final String TAG = "SceneAwareRouter";

    // 路由动作
    public static final String ACTION_ROUTE = "route";
    public static final String ACTION_BROADCAST = "broadcast";
    public static final String ACTION_PRIORITIZE = "prioritize";
    public static final String ACTION_CAPABILITY = "capability";
    public static final String ACTION_FILTER = "filter";
    public static final String ACTION_DELAY = "delay";

    // 默认规则优先级
    private static final int PRIORITY_RUNNING = 100;
    private static final int PRIORITY_DRIVING = 100;
    private static final int PRIORITY_MEETING = 90;
    private static final int PRIORITY_SLEEPING = 95;
    private static final int PRIORITY_HEALTH = 200;
    private static final int PRIORITY_DEFAULT = 0;

    private final ExecutorService executor;
    private final List<RoutingRule> rules;
    private final List<RoutingListener> listeners;
    private final List<RoutingRecord> routingHistory;

    // 依赖服务
    private StateSyncService stateSyncService;
    private MessageBus messageBus;
    private EnhancedSceneDetector sceneDetector;

    // 配置
    private RouterConfig config;

    // 统计
    private int totalRoutings = 0;
    private final Map<String, Integer> actionStats = new HashMap<>();
    private final Map<String, Integer> sceneStats = new HashMap<>();

    /**
     * 路由规则
     */
    public static class RoutingRule {
        public String ruleId;
        public String name;
        public int priority;
        public List<String> scenes;
        public List<String> messageTypes;
        public List<String> deviceTypes;
        public String action;
        public String targetDevice;
        public String targetRole;  // primary, secondary
        public boolean enabled;

        // 条件
        public List<RuleCondition> conditions;

        public RoutingRule() {
            this.scenes = new ArrayList<>();
            this.messageTypes = new ArrayList<>();
            this.deviceTypes = new ArrayList<>();
            this.conditions = new ArrayList<>();
            this.enabled = true;
        }
    }

    /**
     * 规则条件
     */
    public static class RuleCondition {
        public String field;
        public String operator; // eq, ne, gt, lt, gte, lte, in, contains
        public Object value;

        public RuleCondition(String field, String operator, Object value) {
            this.field = field;
            this.operator = operator;
            this.value = value;
        }
    }

    /**
     * 路由上下文
     */
    public static class RoutingContext {
        public String identityId;
        public String fromAgent;
        public String messageType;
        public String scene;
        public int priority;
        public Map<String, Object> payload;
        public List<DeviceState> deviceStates;

        public RoutingContext() {
            this.payload = new HashMap<>();
            this.deviceStates = new ArrayList<>();
        }
    }

    /**
     * 路由结果
     */
    public static class RoutingResult {
        public List<String> targetAgents = new ArrayList<>();
        public String action;
        public String reason;
        public RoutingRule matchedRule;
        public long delayMs;

        public boolean hasTargets() {
            return targetAgents != null && !targetAgents.isEmpty();
        }
    }

    /**
     * 路由记录
     */
    public static class RoutingRecord {
        public long timestamp;
        public String identityId;
        public String fromAgent;
        public List<String> toAgents;
        public String messageType;
        public String scene;
        public String action;
        public String ruleId;

        public RoutingRecord() {
            this.timestamp = System.currentTimeMillis();
            this.toAgents = new ArrayList<>();
        }
    }

    /**
     * 路由监听器
     */
    public interface RoutingListener {
        void onRoutingDecision(RoutingContext context, RoutingResult result);
        void onRuleMatched(RoutingRule rule, RoutingContext context);
    }

    /**
     * 路由器配置
     */
    public static class RouterConfig {
        public boolean smartRouting = true;
        public int lowBatteryThreshold = 20;
        public int maxHistorySize = 100;
        public String defaultAction = ACTION_PRIORITIZE;
    }

    public SceneAwareRouter() {
        this.executor = Executors.newSingleThreadExecutor();
        this.rules = new ArrayList<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.routingHistory = new ArrayList<>();
        this.config = new RouterConfig();

        initDefaultRules();
    }

    /**
     * 设置状态同步服务
     */
    public void setStateSyncService(@Nullable StateSyncService service) {
        this.stateSyncService = service;
    }

    /**
     * 设置消息总线
     */
    public void setMessageBus(@Nullable MessageBus bus) {
        this.messageBus = bus;
    }

    /**
     * 设置场景检测器
     */
    public void setSceneDetector(@Nullable EnhancedSceneDetector detector) {
        this.sceneDetector = detector;
    }

    /**
     * 设置配置
     */
    public void setConfig(@NonNull RouterConfig config) {
        this.config = config;
    }

    // === 规则管理 ===

    /**
     * 添加路由规则
     */
    public void addRule(@NonNull RoutingRule rule) {
        executor.execute(() -> {
            if (rule.ruleId == null || rule.ruleId.isEmpty()) {
                rule.ruleId = "rule_" + System.currentTimeMillis();
            }
            rules.add(rule);
            sortRules();
            Log.d(TAG, "Rule added: " + rule.ruleId);
        });
    }

    /**
     * 移除路由规则
     */
    public void removeRule(@NonNull String ruleId) {
        executor.execute(() -> {
            rules.removeIf(r -> r.ruleId.equals(ruleId));
            Log.d(TAG, "Rule removed: " + ruleId);
        });
    }

    /**
     * 获取所有规则
     */
    @NonNull
    public List<RoutingRule> getRules() {
        return new ArrayList<>(rules);
    }

    // === 路由决策 ===

    /**
     * 路由消息
     */
    @NonNull
    public RoutingResult route(@NonNull RoutingContext context) {
        RoutingResult result = new RoutingResult();

        executor.execute(() -> {
            // 获取设备状态
            if (stateSyncService != null) {
                context.deviceStates = stateSyncService.getAllDeviceStates();
            }

            // 获取当前场景
            if (sceneDetector != null && context.scene == null) {
                context.scene = sceneDetector.getCurrentScene().getSceneType();
            }

            // 匹配规则
            RoutingRule matchedRule = null;
            for (RoutingRule rule : rules) {
                if (!rule.enabled) {
                    continue;
                }
                if (matchRule(rule, context)) {
                    matchedRule = rule;
                    break;
                }
            }

            // 执行路由
            if (matchedRule != null) {
                result.matchedRule = matchedRule;
                result.action = matchedRule.action;
                executeAction(matchedRule, context, result);

                // 通知监听器
                notifyRuleMatched(matchedRule, context);
            } else {
                // 默认路由
                result.action = config.defaultAction;
                defaultRoute(context, result);
            }

            // 记录
            recordRouting(context, result, matchedRule);

            // 通知监听器
            notifyRoutingDecision(context, result);
        });

        return result;
    }

    /**
     * 快速路由（用于已知场景）
     */
    @NonNull
    public RoutingResult quickRoute(@NonNull String messageType, @NonNull String scene,
                                     int priority, @NonNull String fromAgent) {
        RoutingContext context = new RoutingContext();
        context.messageType = messageType;
        context.scene = scene;
        context.priority = priority;
        context.fromAgent = fromAgent;

        return route(context);
    }

    // === 规则匹配 ===

    private boolean matchRule(@NonNull RoutingRule rule, @NonNull RoutingContext context) {
        // 匹配场景
        if (!rule.scenes.isEmpty() && !rule.scenes.contains(context.scene)) {
            return false;
        }

        // 匹配消息类型
        if (!rule.messageTypes.isEmpty() && !rule.messageTypes.contains(context.messageType)) {
            return false;
        }

        // 匹配条件
        for (RuleCondition cond : rule.conditions) {
            if (!matchCondition(cond, context)) {
                return false;
            }
        }

        return true;
    }

    private boolean matchCondition(@NonNull RuleCondition cond, @NonNull RoutingContext context) {
        Object value = null;

        switch (cond.field) {
            case "priority":
                value = context.priority;
                break;
            case "scene":
                value = context.scene;
                break;
            case "message_type":
                value = context.messageType;
                break;
            default:
                if (context.payload != null) {
                    value = context.payload.get(cond.field);
                }
        }

        if (value == null) {
            return false;
        }

        // 执行比较
        switch (cond.operator) {
            case "eq":
                return value.equals(cond.value);
            case "ne":
                return !value.equals(cond.value);
            case "gt":
                if (value instanceof Number && cond.value instanceof Number) {
                    return ((Number) value).doubleValue() > ((Number) cond.value).doubleValue();
                }
                break;
            case "lt":
                if (value instanceof Number && cond.value instanceof Number) {
                    return ((Number) value).doubleValue() < ((Number) cond.value).doubleValue();
                }
                break;
            case "gte":
                if (value instanceof Number && cond.value instanceof Number) {
                    return ((Number) value).doubleValue() >= ((Number) cond.value).doubleValue();
                }
                break;
            case "lte":
                if (value instanceof Number && cond.value instanceof Number) {
                    return ((Number) value).doubleValue() <= ((Number) cond.value).doubleValue();
                }
                break;
        }

        return false;
    }

    // === 路由执行 ===

    private void executeAction(@NonNull RoutingRule rule, @NonNull RoutingContext context,
                               @NonNull RoutingResult result) {
        switch (rule.action) {
            case ACTION_ROUTE:
                routeToSpecific(rule, context, result);
                break;
            case ACTION_BROADCAST:
                broadcast(context, result);
                break;
            case ACTION_PRIORITIZE:
                prioritize(context, result);
                break;
            case ACTION_CAPABILITY:
                capabilityMatch(rule, context, result);
                break;
            case ACTION_FILTER:
                result.targetAgents.clear();
                result.reason = "Filtered by rule";
                break;
            case ACTION_DELAY:
                prioritize(context, result);
                result.delayMs = 5 * 60 * 1000; // 5分钟延迟
                result.reason = "Delayed delivery";
                break;
        }
    }

    private void routeToSpecific(@NonNull RoutingRule rule, @NonNull RoutingContext context,
                                 @NonNull RoutingResult result) {
        if (rule.targetDevice != null) {
            result.targetAgents.add(rule.targetDevice);
            result.reason = "Specified by rule";
        } else if ("primary".equals(rule.targetRole)) {
            // 路由到主设备（最高优先级设备）
            DeviceState primary = findPrimaryDevice(context.deviceStates);
            if (primary != null) {
                result.targetAgents.add(primary.getAgentId());
                result.reason = "Primary device";
            }
        }

        if (!result.hasTargets()) {
            prioritize(context, result);
        }
    }

    private void broadcast(@NonNull RoutingContext context, @NonNull RoutingResult result) {
        for (DeviceState state : context.deviceStates) {
            if (state.isOnline() && !state.getAgentId().equals(context.fromAgent)) {
                result.targetAgents.add(state.getAgentId());
            }
        }
        result.reason = "Broadcast to all devices";
    }

    private void prioritize(@NonNull RoutingContext context, @NonNull RoutingResult result) {
        List<DeviceState> sorted = new ArrayList<>(context.deviceStates);

        // 排除发送者
        sorted.removeIf(s -> s.getAgentId().equals(context.fromAgent));

        // 按优先级排序
        sorted.sort((a, b) -> Integer.compare(b.getPriority(), a.getPriority()));

        // 智能路由优化
        if (config.smartRouting) {
            sorted = optimizeByScene(sorted, context.scene);
        }

        for (DeviceState state : sorted) {
            if (state.isOnline()) {
                result.targetAgents.add(state.getAgentId());
            }
        }

        result.reason = "Prioritized by device priority";
    }

    private void capabilityMatch(@NonNull RoutingRule rule, @NonNull RoutingContext context,
                                 @NonNull RoutingResult result) {
        for (DeviceState state : context.deviceStates) {
            if (!state.isOnline()) continue;
            if (state.getAgentId().equals(context.fromAgent)) continue;

            // 匹配设备类型
            if (!rule.deviceTypes.isEmpty() && !rule.deviceTypes.contains(state.getDeviceType())) {
                continue;
            }

            result.targetAgents.add(state.getAgentId());
        }

        result.reason = "Matched by capability";
    }

    private void defaultRoute(@NonNull RoutingContext context, @NonNull RoutingResult result) {
        prioritize(context, result);
    }

    // === 智能优化 ===

    private List<DeviceState> optimizeByScene(@NonNull List<DeviceState> devices, @NonNull String scene) {
        List<DeviceState> optimized = new ArrayList<>();

        for (DeviceState state : devices) {
            // 低电量设备降权
            if (state.getBatteryLevel() < config.lowBatteryThreshold && !state.isCharging()) {
                continue; // 跳过低电量设备
            }

            // 场景感知
            switch (scene) {
                case EnhancedSceneDetector.SCENE_RUNNING:
                case EnhancedSceneDetector.SCENE_WALKING:
                    // 运动场景优先手表
                    if ("watch".equals(state.getDeviceType())) {
                        optimized.add(0, state); // 插入到前面
                    } else if (!"mobile".equals(state.getDeviceType())) {
                        optimized.add(state);
                    }
                    break;

                case EnhancedSceneDetector.SCENE_DRIVING:
                    // 驾驶场景优先车载设备或手表
                    if ("car".equals(state.getDeviceType()) || "watch".equals(state.getDeviceType())) {
                        optimized.add(0, state);
                    } else {
                        optimized.add(state);
                    }
                    break;

                case EnhancedSceneDetector.SCENE_MEETING:
                case EnhancedSceneDetector.SCENE_WORKING:
                    // 工作/会议场景优先电脑或平板
                    if ("desktop".equals(state.getDeviceType()) || "tablet".equals(state.getDeviceType())) {
                        optimized.add(0, state);
                    } else {
                        optimized.add(state);
                    }
                    break;

                case EnhancedSceneDetector.SCENE_SLEEPING:
                    // 睡眠场景仅手表
                    if ("watch".equals(state.getDeviceType())) {
                        optimized.add(state);
                    }
                    break;

                default:
                    optimized.add(state);
            }
        }

        return optimized;
    }

    private DeviceState findPrimaryDevice(@NonNull List<DeviceState> states) {
        DeviceState primary = null;
        for (DeviceState state : states) {
            if (!state.isOnline()) continue;
            if (primary == null || state.getPriority() > primary.getPriority()) {
                primary = state;
            }
        }
        return primary;
    }

    // === 初始化 ===

    private void initDefaultRules() {
        // 跑步/走路场景 - 路由到手表
        RoutingRule runningRule = new RoutingRule();
        runningRule.ruleId = "rule-running-watch";
        runningRule.name = "Running Scene - Route to Watch";
        runningRule.priority = PRIORITY_RUNNING;
        runningRule.scenes.add(EnhancedSceneDetector.SCENE_RUNNING);
        runningRule.scenes.add(EnhancedSceneDetector.SCENE_WALKING);
        runningRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_NOTIFICATION);
        runningRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_HEALTH);
        runningRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_SOCIAL);
        runningRule.deviceTypes.add("watch");
        runningRule.action = ACTION_ROUTE;
        rules.add(runningRule);

        // 驾驶场景 - 延迟投递
        RoutingRule drivingRule = new RoutingRule();
        drivingRule.ruleId = "rule-driving-delay";
        drivingRule.name = "Driving Scene - Delay Delivery";
        drivingRule.priority = PRIORITY_DRIVING;
        drivingRule.scenes.add(EnhancedSceneDetector.SCENE_DRIVING);
        drivingRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_NOTIFICATION);
        drivingRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_SOCIAL);
        drivingRule.conditions.add(new RuleCondition("priority", "lt", 3));
        drivingRule.action = ACTION_DELAY;
        rules.add(drivingRule);

        // 会议场景 - 过滤社交消息
        RoutingRule meetingRule = new RoutingRule();
        meetingRule.ruleId = "rule-meeting-filter";
        meetingRule.name = "Meeting Scene - Filter Social";
        meetingRule.priority = PRIORITY_MEETING;
        meetingRule.scenes.add(EnhancedSceneDetector.SCENE_MEETING);
        meetingRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_SOCIAL);
        meetingRule.conditions.add(new RuleCondition("priority", "lt", 2));
        meetingRule.action = ACTION_FILTER;
        rules.add(meetingRule);

        // 健康告警 - 广播到所有设备
        RoutingRule healthRule = new RoutingRule();
        healthRule.ruleId = "rule-health-broadcast";
        healthRule.name = "Health Alert - Broadcast";
        healthRule.priority = PRIORITY_HEALTH;
        healthRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_HEALTH);
        healthRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_ALERT);
        healthRule.action = ACTION_BROADCAST;
        rules.add(healthRule);

        // 睡眠场景 - 静音模式
        RoutingRule sleepingRule = new RoutingRule();
        sleepingRule.ruleId = "rule-sleeping-filter";
        sleepingRule.name = "Sleeping Scene - Quiet Mode";
        sleepingRule.priority = PRIORITY_SLEEPING;
        sleepingRule.scenes.add(EnhancedSceneDetector.SCENE_SLEEPING);
        sleepingRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_NOTIFICATION);
        sleepingRule.messageTypes.add(EnhancedSceneDetector.MSG_TYPE_SOCIAL);
        sleepingRule.conditions.add(new RuleCondition("priority", "lt", 3));
        sleepingRule.action = ACTION_FILTER;
        rules.add(sleepingRule);

        sortRules();
    }

    private void sortRules() {
        rules.sort((a, b) -> Integer.compare(b.priority, a.priority));
    }

    // === 记录与通知 ===

    private void recordRouting(@NonNull RoutingContext context, @NonNull RoutingResult result,
                               @Nullable RoutingRule matchedRule) {
        totalRoutings++;

        // 统计
        actionStats.merge(result.action, 1, Integer::sum);
        sceneStats.merge(context.scene, 1, Integer::sum);

        // 记录历史
        RoutingRecord record = new RoutingRecord();
        record.identityId = context.identityId;
        record.fromAgent = context.fromAgent;
        record.toAgents = result.targetAgents;
        record.messageType = context.messageType;
        record.scene = context.scene;
        record.action = result.action;
        record.ruleId = matchedRule != null ? matchedRule.ruleId : null;

        routingHistory.add(record);

        // 限制历史大小
        while (routingHistory.size() > config.maxHistorySize) {
            routingHistory.remove(0);
        }
    }

    private void notifyRoutingDecision(@NonNull RoutingContext context, @NonNull RoutingResult result) {
        for (RoutingListener l : listeners) {
            l.onRoutingDecision(context, result);
        }
    }

    private void notifyRuleMatched(@NonNull RoutingRule rule, @NonNull RoutingContext context) {
        for (RoutingListener l : listeners) {
            l.onRuleMatched(rule, context);
        }
    }

    // === 监听器管理 ===

    public void addListener(@NonNull RoutingListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull RoutingListener listener) {
        listeners.remove(listener);
    }

    // === 统计 ===

    @NonNull
    public RouterStats getStats() {
        RouterStats stats = new RouterStats();
        stats.totalRoutings = totalRoutings;
        stats.byAction = new HashMap<>(actionStats);
        stats.byScene = new HashMap<>(sceneStats);
        stats.totalRules = rules.size();
        stats.enabledRules = 0;
        for (RoutingRule rule : rules) {
            if (rule.enabled) {
                stats.enabledRules++;
            }
        }
        return stats;
    }

    /**
     * 路由统计
     */
    public static class RouterStats {
        public int totalRoutings;
        public Map<String, Integer> byAction;
        public Map<String, Integer> byScene;
        public int totalRules;
        public int enabledRules;

        @NonNull
        @Override
        public String toString() {
            return "RouterStats{" +
                    "total=" + totalRoutings +
                    ", rules=" + enabledRules + "/" + totalRules +
                    '}';
        }
    }

    /**
     * 获取路由历史
     */
    @NonNull
    public List<RoutingRecord> getRoutingHistory(int limit) {
        if (limit <= 0 || limit >= routingHistory.size()) {
            return new ArrayList<>(routingHistory);
        }
        return new ArrayList<>(routingHistory.subList(
                routingHistory.size() - limit, routingHistory.size()));
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        executor.shutdown();
        listeners.clear();
        routingHistory.clear();
    }
}