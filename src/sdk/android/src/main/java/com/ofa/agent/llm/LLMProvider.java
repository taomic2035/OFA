package com.ofa.agent.llm;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * LLM 提供者接口
 * 支持云端和本地 LLM 的统一抽象
 */
public interface LLMProvider {

    /**
     * 获取提供者 ID
     */
    @NonNull
    String getId();

    /**
     * 获取提供者名称
     */
    @NonNull
    String getName();

    /**
     * 获取提供者类型
     */
    @NonNull
    ProviderType getType();

    /**
     * 检查是否可用
     */
    boolean isAvailable();

    /**
     * 是否支持离线
     */
    boolean supportsOffline();

    /**
     * 聊天 - 单轮
     */
    @NonNull
    CompletableFuture<LLMResponse> chat(@NonNull String message);

    /**
     * 聊天 - 多轮
     */
    @NonNull
    CompletableFuture<LLMResponse> chat(@NonNull List<Message> messages);

    /**
     * 聊天 - 带工具调用
     */
    @NonNull
    CompletableFuture<LLMResponse> chatWithTools(
            @NonNull List<Message> messages,
            @Nullable List<ToolDefinition> tools
    );

    /**
     * 流式聊天
     */
    void streamChat(
            @NonNull List<Message> messages,
            @NonNull StreamCallback callback
    );

    /**
     * 文本嵌入
     */
    @NonNull
    CompletableFuture<float[]> embed(@NonNull String text);

    /**
     * 配置提供者
     */
    void configure(@NonNull LLMConfig config);

    /**
     * 获取统计信息
     */
    @NonNull
    LLMStats getStats();

    /**
     * 关闭提供者
     */
    void shutdown();

    /**
     * 提供者类型
     */
    enum ProviderType {
        CLOUD,  // 云端 LLM (OpenAI, Claude, etc.)
        LOCAL   // 本地 LLM (TensorFlow Lite)
    }

    /**
     * 工具定义 (用于 Function Calling)
     */
    class ToolDefinition {
        private final String name;
        private final String description;
        private final Map<String, Object> parameters;

        public ToolDefinition(@NonNull String name, @NonNull String description,
                              @Nullable Map<String, Object> parameters) {
            this.name = name;
            this.description = description;
            this.parameters = parameters;
        }

        @NonNull
        public String getName() { return name; }
        @NonNull
        public String getDescription() { return description; }
        @Nullable
        public Map<String, Object> getParameters() { return parameters; }
    }
}