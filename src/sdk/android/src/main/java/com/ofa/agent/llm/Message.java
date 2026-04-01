package com.ofa.agent.llm;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * 聊天消息
 */
public class Message {

    private final Role role;
    private final String content;
    private final String name;
    private final LLMResponse.ToolCall toolCall;

    public Message(@NonNull Role role, @NonNull String content) {
        this.role = role;
        this.content = content;
        this.name = null;
        this.toolCall = null;
    }

    public Message(@NonNull Role role, @NonNull String content, @Nullable String name) {
        this.role = role;
        this.content = content;
        this.name = name;
        this.toolCall = null;
    }

    public Message(@NonNull Role role, @Nullable LLMResponse.ToolCall toolCall) {
        this.role = role;
        this.content = null;
        this.name = null;
        this.toolCall = toolCall;
    }

    @NonNull
    public Role getRole() { return role; }

    @Nullable
    public String getContent() { return content; }

    @Nullable
    public String getName() { return name; }

    @Nullable
    public LLMResponse.ToolCall getToolCall() { return toolCall; }

    public boolean isToolResult() {
        return role == Role.TOOL;
    }

    /**
     * 消息角色
     */
    public enum Role {
        SYSTEM,    // 系统消息
        USER,      // 用户消息
        ASSISTANT, // 助手消息
        TOOL       // 工具结果
    }

    // 便捷工厂方法

    public static Message system(@NonNull String content) {
        return new Message(Role.SYSTEM, content);
    }

    public static Message user(@NonNull String content) {
        return new Message(Role.USER, content);
    }

    public static Message assistant(@NonNull String content) {
        return new Message(Role.ASSISTANT, content);
    }

    public static Message toolResult(@NonNull String name, @NonNull String content) {
        return new Message(Role.TOOL, content, name);
    }
}