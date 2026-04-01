package com.ofa.agent.llm;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * 流式响应回调
 */
public interface StreamCallback {

    /**
     * 收到 token
     * @param token 文本片段
     */
    void onToken(@NonNull String token);

    /**
     * 响应完成
     * @param response 完整响应
     */
    void onComplete(@NonNull LLMResponse response);

    /**
     * 工具调用
     * @param toolCall 工具调用信息
     */
    void onToolCall(@NonNull LLMResponse.ToolCall toolCall);

    /**
     * 发生错误
     * @param error 错误信息
     */
    void onError(@NonNull String error);

    /**
     * 空实现
     */
    StreamCallback EMPTY = new StreamCallback() {
        @Override public void onToken(@NonNull String token) {}
        @Override public void onComplete(@NonNull LLMResponse response) {}
        @Override public void onToolCall(@NonNull LLMResponse.ToolCall toolCall) {}
        @Override public void onError(@NonNull String error) {}
    };
}