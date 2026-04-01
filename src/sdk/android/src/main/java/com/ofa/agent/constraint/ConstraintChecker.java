package com.ofa.agent.constraint;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.regex.Pattern;

/**
 * 约束类型
 */
public enum ConstraintType {
    NONE(0),
    PRIVACY(1),
    FINANCIAL(2),
    SECURITY(4),
    AUTH_REQUIRED(8),
    LOCATION(16),
    PERSONAL(32),
    DEVICE(64);

    private final int value;

    ConstraintType(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }
}

/**
 * 约束检查结果
 */
public class ConstraintResult {
    public final boolean allowed;
    public final ConstraintType violated;
    public final String reason;
    public final boolean requiresAuth;
    public final List<String> suggestions;

    public ConstraintResult(boolean allowed, ConstraintType violated, String reason, boolean requiresAuth) {
        this.allowed = allowed;
        this.violated = violated;
        this.reason = reason;
        this.requiresAuth = requiresAuth;
        this.suggestions = new ArrayList<>();
    }

    public void addSuggestion(String suggestion) {
        suggestions.add(suggestion);
    }
}

/**
 * 约束规则
 */
public class ConstraintRule {
    public final String name;
    public final ConstraintType type;
    public final Pattern actionPattern;
    public final Pattern dataPattern;
    public final boolean offlineRestricted;
    public final boolean requiresAuth;
    public final String message;

    public ConstraintRule(String name, ConstraintType type, String actionPattern, String dataPattern,
                          boolean offlineRestricted, boolean requiresAuth, String message) {
        this.name = name;
        this.type = type;
        this.actionPattern = actionPattern != null ? Pattern.compile(actionPattern, Pattern.CASE_INSENSITIVE) : null;
        this.dataPattern = dataPattern != null ? Pattern.compile(dataPattern, Pattern.CASE_INSENSITIVE) : null;
        this.offlineRestricted = offlineRestricted;
        this.requiresAuth = requiresAuth;
        this.message = message;
    }
}

/**
 * 约束检查器
 */
public class ConstraintChecker {
    private static final String TAG = "ConstraintChecker";

    private final List<ConstraintRule> rules = new CopyOnWriteArrayList<>();
    private final Set<String> offlineRestrictedActions = new HashSet<>();
    private final Set<String> sensitiveFields = new HashSet<>();
    private boolean offlineMode = false;

    public ConstraintChecker() {
        loadDefaultRules();
        loadSensitiveFields();
    }

    private void loadDefaultRules() {
        // 财务操作
        addRule(new ConstraintRule(
            "financial_operations",
            ConstraintType.FINANCIAL,
            "(payment|transfer|withdraw|pay)",
            null,
            true,
            true,
            "Financial operations require online mode and authorization"
        ));

        // 隐私数据
        addRule(new ConstraintRule(
            "privacy_data",
            ConstraintType.PRIVACY,
            null,
            "(idcard|id_card|身份证|passport|护照)",
            false,
            false,
            "Data contains sensitive personal information"
        ));

        // 位置信息
        addRule(new ConstraintRule(
            "location_data",
            ConstraintType.LOCATION,
            null,
            "(location|gps|latitude|longitude|经纬度)",
            false,
            true,
            "Location data sharing requires authorization"
        ));

        // 安全操作
        addRule(new ConstraintRule(
            "security_operations",
            ConstraintType.SECURITY,
            "(delete|password|auth|login|logout)",
            null,
            true,
            true,
            "Security operations require online mode and authorization"
        ));
    }

    private void loadSensitiveFields() {
        String[] fields = {
            "idcard", "id_card", "身份证", "passport", "护照",
            "phone", "mobile", "电话", "手机",
            "email", "邮箱",
            "address", "地址",
            "bank_account", "银行卡",
            "password", "密码",
            "token", "令牌",
            "secret", "密钥",
            "location", "gps", "位置"
        };
        for (String field : fields) {
            sensitiveFields.add(field.toLowerCase());
        }
    }

    public void addRule(@NonNull ConstraintRule rule) {
        rules.add(rule);
        if (rule.offlineRestricted && rule.actionPattern != null) {
            // 提取动作关键词
            String pattern = rule.actionPattern.pattern();
            pattern = pattern.replace("(", "").replace(")", "");
            for (String action : pattern.split("\\|")) {
                offlineRestrictedActions.add(action.trim().toLowerCase());
            }
        }
    }

    public void removeRule(@NonNull String name) {
        rules.removeIf(rule -> rule.name.equals(name));
    }

    public void setOfflineMode(boolean offline) {
        this.offlineMode = offline;
    }

    @NonNull
    public ConstraintResult check(@NonNull String action, @Nullable String dataJson) {
        ConstraintResult result = new ConstraintResult(true, ConstraintType.NONE, null, false);

        // 1. 检查离线受限操作
        if (offlineMode) {
            String actionLower = action.toLowerCase();
            for (String restricted : offlineRestrictedActions) {
                if (actionLower.contains(restricted)) {
                    result = new ConstraintResult(false, ConstraintType.FINANCIAL,
                        "Action '" + action + "' requires online mode", false);
                    result.addSuggestion("Connect to network or use alternative offline action");
                    return result;
                }
            }
        }

        // 2. 应用规则
        for (ConstraintRule rule : rules) {
            ConstraintResult ruleResult = applyRule(rule, action, dataJson);
            if (!ruleResult.allowed) {
                return ruleResult;
            }
        }

        // 3. 检查敏感数据
        if (dataJson != null && !dataJson.isEmpty()) {
            ConstraintResult dataResult = checkSensitiveData(dataJson);
            if (!dataResult.allowed) {
                return dataResult;
            }
        }

        return result;
    }

    private ConstraintResult applyRule(ConstraintRule rule, String action, String dataJson) {
        ConstraintResult result = new ConstraintResult(true, ConstraintType.NONE, null, false);

        // 检查操作模式
        if (rule.actionPattern != null && !rule.actionPattern.matcher(action).find()) {
            return result;
        }

        // 检查数据模式
        if (rule.dataPattern != null && dataJson != null) {
            if (!rule.dataPattern.matcher(dataJson).find()) {
                return result;
            }
        }

        // 离线限制
        if (rule.offlineRestricted && offlineMode) {
            return new ConstraintResult(false, rule.type, rule.message, rule.requiresAuth);
        }

        // 授权要求
        if (rule.requiresAuth) {
            // TODO: 检查用户授权状态
            return new ConstraintResult(false, ConstraintType.AUTH_REQUIRED, rule.message, true);
        }

        return result;
    }

    private ConstraintResult checkSensitiveData(String dataJson) {
        String dataLower = dataJson.toLowerCase();

        for (String field : sensitiveFields) {
            if (dataLower.contains(field)) {
                ConstraintType type = ConstraintType.PRIVACY;
                String reason = "Data contains sensitive information";

                if (field.contains("bank") || field.contains("card")) {
                    type = ConstraintType.FINANCIAL;
                    reason = "Data contains financial information";
                } else if (field.contains("location") || field.contains("gps")) {
                    type = ConstraintType.LOCATION;
                    reason = "Data contains location information";
                } else if (field.contains("password") || field.contains("token") || field.contains("secret")) {
                    type = ConstraintType.SECURITY;
                    reason = "Data contains security credentials";
                }

                return new ConstraintResult(false, type, reason, false);
            }
        }

        return new ConstraintResult(true, ConstraintType.NONE, null, false);
    }

    public void addSensitiveField(@NonNull String field) {
        sensitiveFields.add(field.toLowerCase());
    }

    public void removeSensitiveField(@NonNull String field) {
        sensitiveFields.remove(field.toLowerCase());
    }

    @NonNull
    public Set<String> getOfflineRestrictedActions() {
        return new HashSet<>(offlineRestrictedActions);
    }

    @NonNull
    public List<ConstraintRule> getRules() {
        return new ArrayList<>(rules);
    }
}