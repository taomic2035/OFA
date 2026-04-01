package com.ofa.agent.llm.orchestrator;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.llm.LLMConfig;
import com.ofa.agent.llm.LLMProvider;
import com.ofa.agent.llm.LLMResponse;
import com.ofa.agent.llm.LLMStats;
import com.ofa.agent.llm.Message;
import com.ofa.agent.llm.StreamCallback;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;

/**
 * LLM 编排器
 * 管理多个 LLM 提供者，实现自动故障转移和负载均衡
 */
public class LLMOrchestrator implements LLMProvider {

    private static final String TAG = "LLMOrchestrator";

    private final List<LLMProvider> providers;
    private LLMProvider primaryProvider;
    private LLMProvider fallbackProvider;

    private FailoverStrategy strategy;
    private boolean autoFailover;
    private long failoverThreshold = 30000; // 30秒无响应触发切换

    private final List<ProviderState> providerStates = new ArrayList<>();

    public LLMOrchestrator() {
        this.providers = new ArrayList<>();
        this.strategy = FailoverStrategy.PRIORITY;
        this.autoFailover = true;
    }

    /**
     * 添加提供者
     */
    public void addProvider(@NonNull LLMProvider provider, boolean isPrimary) {
        providers.add(provider);
        providerStates.add(new ProviderState(provider));

        if (isPrimary || primaryProvider == null) {
            primaryProvider = provider;
        }
    }

    /**
     * 设置主提供者
     */
    public void setPrimaryProvider(@NonNull LLMProvider provider) {
        this.primaryProvider = provider;
        if (!providers.contains(provider)) {
            providers.add(provider);
            providerStates.add(new ProviderState(provider));
        }
    }

    /**
     * 设置备用提供者
     */
    public void setFallbackProvider(@NonNull LLMProvider provider) {
        this.fallbackProvider = provider;
        if (!providers.contains(provider)) {
            providers.add(provider);
            providerStates.add(new ProviderState(provider));
        }
    }

    /**
     * 设置故障转移策略
     */
    public void setFailoverStrategy(@NonNull FailoverStrategy strategy) {
        this.strategy = strategy;
    }

    /**
     * 启用/禁用自动故障转移
     */
    public void setAutoFailover(boolean enable) {
        this.autoFailover = enable;
    }

    @Override
    @NonNull
    public String getId() {
        return "orchestrator";
    }

    @Override
    @NonNull
    public String getName() {
        return "LLM Orchestrator";
    }

    @Override
    @NonNull
    public ProviderType getType() {
        return ProviderType.CLOUD; // 默认返回云端类型
    }

    @Override
    public boolean isAvailable() {
        return getActiveProvider() != null;
    }

    @Override
    public boolean supportsOffline() {
        // 如果有本地提供者，则支持离线
        for (LLMProvider provider : providers) {
            if (provider.supportsOffline()) {
                return true;
            }
        }
        return false;
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chat(@NonNull String message) {
        return executeWithFailover(provider -> provider.chat(message));
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chat(@NonNull List<Message> messages) {
        return executeWithFailover(provider -> provider.chat(messages));
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chatWithTools(
            @NonNull List<Message> messages,
            @Nullable List<ToolDefinition> tools) {
        return executeWithFailover(provider -> provider.chatWithTools(messages, tools));
    }

    @Override
    public void streamChat(@NonNull List<Message> messages, @NonNull StreamCallback callback) {
        LLMProvider provider = getActiveProvider();
        if (provider == null) {
            callback.onError("No available LLM provider");
            return;
        }

        // 包装回调以支持故障转移
        StreamCallback wrappedCallback = new StreamCallback() {
            @Override
            public void onToken(@NonNull String token) {
                callback.onToken(token);
            }

            @Override
            public void onComplete(@NonNull LLMResponse response) {
                recordSuccess(provider);
                callback.onComplete(response);
            }

            @Override
            public void onToolCall(@NonNull LLMResponse.ToolCall toolCall) {
                callback.onToolCall(toolCall);
            }

            @Override
            public void onError(@NonNull String error) {
                recordFailure(provider);
                if (autoFailover) {
                    LLMProvider fallback = selectFallback(provider);
                    if (fallback != null) {
                        Log.i(TAG, "Failing over to " + fallback.getName());
                        fallback.streamChat(messages, callback);
                        return;
                    }
                }
                callback.onError(error);
            }
        };

        provider.streamChat(messages, wrappedCallback);
    }

    @Override
    @NonNull
    public CompletableFuture<float[]> embed(@NonNull String text) {
        return executeWithFailover(provider -> provider.embed(text));
    }

    @Override
    public void configure(@NonNull LLMConfig config) {
        for (LLMProvider provider : providers) {
            provider.configure(config);
        }
    }

    @Override
    @NonNull
    public LLMStats getStats() {
        // 返回聚合统计
        long totalRequests = 0;
        long successRequests = 0;
        long failedRequests = 0;
        long totalInputTokens = 0;
        long totalOutputTokens = 0;
        long totalLatencyMs = 0;
        long lastRequestTime = 0;

        for (LLMProvider provider : providers) {
            LLMStats stats = provider.getStats();
            totalRequests += stats.getTotalRequests();
            successRequests += stats.getSuccessRequests();
            failedRequests += stats.getFailedRequests();
            totalInputTokens += stats.getTotalInputTokens();
            totalOutputTokens += stats.getTotalOutputTokens();
            totalLatencyMs += stats.getTotalLatencyMs();
            lastRequestTime = Math.max(lastRequestTime, stats.getLastRequestTime());
        }

        return new LLMStats(totalRequests, successRequests, failedRequests,
                totalInputTokens, totalOutputTokens, totalLatencyMs, lastRequestTime);
    }

    @Override
    public void shutdown() {
        for (LLMProvider provider : providers) {
            provider.shutdown();
        }
        providers.clear();
        providerStates.clear();
    }

    // ===== Private Methods =====

    /**
     * 获取当前活跃的提供者
     */
    private LLMProvider getActiveProvider() {
        if (primaryProvider != null && primaryProvider.isAvailable()) {
            return primaryProvider;
        }

        if (fallbackProvider != null && fallbackProvider.isAvailable()) {
            return fallbackProvider;
        }

        // 查找任何可用的提供者
        for (LLMProvider provider : providers) {
            if (provider.isAvailable()) {
                return provider;
            }
        }

        return null;
    }

    /**
     * 选择备用提供者
     */
    private LLMProvider selectFallback(LLMProvider failed) {
        // 如果有指定的备用提供者
        if (fallbackProvider != null && fallbackProvider != failed && fallbackProvider.isAvailable()) {
            return fallbackProvider;
        }

        // 根据策略选择
        for (LLMProvider provider : providers) {
            if (provider != failed && provider.isAvailable()) {
                return provider;
            }
        }

        return null;
    }

    /**
     * 带故障转移的执行
     */
    private CompletableFuture<LLMResponse> executeWithFailover(ProviderOperation operation) {
        LLMProvider provider = getActiveProvider();
        if (provider == null) {
            return CompletableFuture.completedFuture(new LLMResponse("No available LLM provider"));
        }

        CompletableFuture<LLMResponse> future = operation.execute(provider);

        if (autoFailover) {
            return future.thenApply(response -> {
                if (response.isSuccess()) {
                    recordSuccess(provider);
                } else {
                    recordFailure(provider);
                }
                return response;
            }).exceptionally(throwable -> {
                recordFailure(provider);
                LLMProvider fallback = selectFallback(provider);
                if (fallback != null) {
                    Log.i(TAG, "Failing over to " + fallback.getName());
                    try {
                        return operation.execute(fallback).join();
                    } catch (Exception e) {
                        return new LLMResponse("All providers failed: " + e.getMessage());
                    }
                }
                return new LLMResponse(throwable.getMessage());
            });
        }

        return future;
    }

    private void recordSuccess(LLMProvider provider) {
        for (ProviderState state : providerStates) {
            if (state.provider == provider) {
                state.recordSuccess();
                break;
            }
        }
    }

    private void recordFailure(LLMProvider provider) {
        for (ProviderState state : providerStates) {
            if (state.provider == provider) {
                state.recordFailure();
                break;
            }
        }
    }

    /**
     * 提供者操作接口
     */
    @FunctionalInterface
    private interface ProviderOperation {
        CompletableFuture<LLMResponse> execute(LLMProvider provider);
    }

    /**
     * 提供者状态
     */
    private static class ProviderState {
        final LLMProvider provider;
        int consecutiveFailures;
        long lastFailureTime;
        long lastSuccessTime;
        boolean healthy;

        ProviderState(LLMProvider provider) {
            this.provider = provider;
            this.healthy = true;
        }

        void recordSuccess() {
            consecutiveFailures = 0;
            lastSuccessTime = System.currentTimeMillis();
            healthy = true;
        }

        void recordFailure() {
            consecutiveFailures++;
            lastFailureTime = System.currentTimeMillis();
            healthy = consecutiveFailures < 3;
        }
    }

    /**
     * 故障转移策略
     */
    public enum FailoverStrategy {
        PRIORITY,    // 按优先级
        ROUND_ROBIN, // 轮询
        LEAST_LATENCY // 最低延迟
    }
}