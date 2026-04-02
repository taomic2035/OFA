package com.ofa.agent.skill;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 技能步骤定义
 * 定义技能中的单个步骤
 */
public class SkillStep {

    private final String id;
    private final String name;
    private final StepType type;
    private final String action;           // 工具名/意图名/条件表达式
    private final Map<String, Object> params;  // 参数（支持变量引用 ${var}）
    private final List<Branch> branches;   // 条件分支
    private final String nextStep;         // 下一步骤ID
    private final int timeout;             // 超时时间（毫秒）
    private final int retryCount;          // 重试次数
    private final int retryDelay;          // 重试间隔（毫秒）
    private final String onErrorStep;      // 错误时跳转的步骤
    private final boolean optional;        // 是否可选（失败不影响整体）
    private final String description;      // 步骤描述

    /**
     * 步骤类型
     */
    public enum StepType {
        TOOL,           // 调用工具
        INTENT,         // 执行意图
        DELAY,          // 延迟等待
        WAIT_FOR,       // 等待条件
        CONDITION,      // 条件判断
        LOOP,           // 循环
        ASSIGN,         // 变量赋值
        INPUT,          // 请求用户输入
        CONFIRM,        // 请求用户确认
        NOTIFY,         // 发送通知
        PARALLEL,       // 并行执行
        SUB_SKILL       // 调用子技能
    }

    /**
     * 条件分支
     */
    public static class Branch {
        public final String condition;  // 条件表达式
        public final String targetStep; // 目标步骤ID

        public Branch(@NonNull String condition, @NonNull String targetStep) {
            this.condition = condition;
            this.targetStep = targetStep;
        }
    }

    private SkillStep(Builder builder) {
        this.id = builder.id;
        this.name = builder.name;
        this.type = builder.type;
        this.action = builder.action;
        this.params = new HashMap<>(builder.params);
        this.branches = new ArrayList<>(builder.branches);
        this.nextStep = builder.nextStep;
        this.timeout = builder.timeout;
        this.retryCount = builder.retryCount;
        this.retryDelay = builder.retryDelay;
        this.onErrorStep = builder.onErrorStep;
        this.optional = builder.optional;
        this.description = builder.description;
    }

    @NonNull
    public String getId() { return id; }

    @NonNull
    public String getName() { return name; }

    @NonNull
    public StepType getType() { return type; }

    @Nullable
    public String getAction() { return action; }

    @NonNull
    public Map<String, Object> getParams() { return new HashMap<>(params); }

    @NonNull
    public List<Branch> getBranches() { return new ArrayList<>(branches); }

    @Nullable
    public String getNextStep() { return nextStep; }

    public int getTimeout() { return timeout; }

    public int getRetryCount() { return retryCount; }

    public int getRetryDelay() { return retryDelay; }

    @Nullable
    public String getOnErrorStep() { return onErrorStep; }

    public boolean isOptional() { return optional; }

    @Nullable
    public String getDescription() { return description; }

    /**
     * Builder
     */
    public static class Builder {
        private String id;
        private String name;
        private StepType type = StepType.TOOL;
        private String action;
        private Map<String, Object> params = new HashMap<>();
        private List<Branch> branches = new ArrayList<>();
        private String nextStep;
        private int timeout = 30000;
        private int retryCount = 0;
        private int retryDelay = 1000;
        private String onErrorStep;
        private boolean optional = false;
        private String description;

        public Builder id(@NonNull String id) {
            this.id = id;
            return this;
        }

        public Builder name(@NonNull String name) {
            this.name = name;
            return this;
        }

        public Builder type(@NonNull StepType type) {
            this.type = type;
            return this;
        }

        public Builder action(@NonNull String action) {
            this.action = action;
            return this;
        }

        public Builder param(@NonNull String key, @Nullable Object value) {
            this.params.put(key, value);
            return this;
        }

        public Builder params(@NonNull Map<String, Object> params) {
            this.params.putAll(params);
            return this;
        }

        public Builder branch(@NonNull String condition, @NonNull String targetStep) {
            this.branches.add(new Branch(condition, targetStep));
            return this;
        }

        public Builder nextStep(@Nullable String nextStep) {
            this.nextStep = nextStep;
            return this;
        }

        public Builder timeout(int timeoutMs) {
            this.timeout = timeoutMs;
            return this;
        }

        public Builder retry(int count, int delayMs) {
            this.retryCount = count;
            this.retryDelay = delayMs;
            return this;
        }

        public Builder onError(@Nullable String errorStep) {
            this.onErrorStep = errorStep;
            return this;
        }

        public Builder optional(boolean optional) {
            this.optional = optional;
            return this;
        }

        public Builder description(@Nullable String description) {
            this.description = description;
            return this;
        }

        public SkillStep build() {
            if (id == null) {
                id = "step_" + System.currentTimeMillis();
            }
            if (name == null) {
                name = id;
            }
            return new SkillStep(this);
        }
    }
}