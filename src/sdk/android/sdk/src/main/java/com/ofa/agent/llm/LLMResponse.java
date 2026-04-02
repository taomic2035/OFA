package com.ofa.agent.llm;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

/**
 * LLM 响应
 */
public class LLMResponse {

    private final boolean success;
    private final String content;
    private final String error;
    private final int inputTokens;
    private final int outputTokens;
    private final int totalTokens;
    private final long latencyMs;
    private final String model;
    private final ToolCall toolCall;
    private final FinishReason finishReason;

    public LLMResponse(@NonNull String content, int inputTokens, int outputTokens,
                       long latencyMs, @NonNull String model) {
        this.success = true;
        this.content = content;
        this.error = null;
        this.inputTokens = inputTokens;
        this.outputTokens = outputTokens;
        this.totalTokens = inputTokens + outputTokens;
        this.latencyMs = latencyMs;
        this.model = model;
        this.toolCall = null;
        this.finishReason = FinishReason.STOP;
    }

    public LLMResponse(@NonNull ToolCall toolCall, int inputTokens, int outputTokens,
                       long latencyMs, @NonNull String model) {
        this.success = true;
        this.content = null;
        this.error = null;
        this.inputTokens = inputTokens;
        this.outputTokens = outputTokens;
        this.totalTokens = inputTokens + outputTokens;
        this.latencyMs = latencyMs;
        this.model = model;
        this.toolCall = toolCall;
        this.finishReason = FinishReason.TOOL_CALL;
    }

    public LLMResponse(@NonNull String error) {
        this.success = false;
        this.content = null;
        this.error = error;
        this.inputTokens = 0;
        this.outputTokens = 0;
        this.totalTokens = 0;
        this.latencyMs = 0;
        this.model = null;
        this.toolCall = null;
        this.finishReason = FinishReason.ERROR;
    }

    public boolean isSuccess() { return success; }

    @Nullable
    public String getContent() { return content; }

    @Nullable
    public String getError() { return error; }

    public int getInputTokens() { return inputTokens; }

    public int getOutputTokens() { return outputTokens; }

    public int getTotalTokens() { return totalTokens; }

    public long getLatencyMs() { return latencyMs; }

    @Nullable
    public String getModel() { return model; }

    @Nullable
    public ToolCall getToolCall() { return toolCall; }

    @NonNull
    public FinishReason getFinishReason() { return finishReason; }

    public boolean hasToolCall() { return toolCall != null; }

    /**
     * 工具调用
     */
    public static class ToolCall {
        private final String id;
        private final String name;
        private final JSONObject arguments;

        public ToolCall(@NonNull String id, @NonNull String name, @NonNull JSONObject arguments) {
            this.id = id;
            this.name = name;
            this.arguments = arguments;
        }

        @NonNull
        public String getId() { return id; }

        @NonNull
        public String getName() { return name; }

        @NonNull
        public JSONObject getArguments() { return arguments; }
    }

    /**
     * 完成原因
     */
    public enum FinishReason {
        STOP,       // 正常结束
        TOOL_CALL,  // 工具调用
        LENGTH,     // 达到长度限制
        ERROR       // 错误
    }
}