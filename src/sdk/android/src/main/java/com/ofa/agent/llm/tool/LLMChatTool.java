package com.ofa.agent.llm.tool;

import android.content.Context;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.llm.LLMProvider;
import com.ofa.agent.llm.LLMResponse;
import com.ofa.agent.llm.Message;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * LLM 聊天工具
 * 将 LLM 能力暴露给 MCP
 */
public class LLMChatTool implements ToolExecutor {

    private static final String TOOL_ID = "llm.chat";

    private final LLMProvider provider;

    public LLMChatTool(@NonNull LLMProvider provider) {
        this.provider = provider;
    }

    @NonNull
    @Override
    public String getToolId() {
        return TOOL_ID;
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Chat with an AI language model. Returns AI-generated response.";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        try {
            String message = (String) args.get("message");
            if (message == null || message.isEmpty()) {
                return new ToolResult(TOOL_ID, "Missing 'message' parameter");
            }

            // 可选的系统提示
            String systemPrompt = (String) args.get("system");

            List<Message> messages = new ArrayList<>();
            if (systemPrompt != null && !systemPrompt.isEmpty()) {
                messages.add(Message.system(systemPrompt));
            }
            messages.add(Message.user(message));

            // 同步调用
            LLMResponse response = provider.chat(messages).join();

            if (!response.isSuccess()) {
                return new ToolResult(TOOL_ID, response.getError());
            }

            JSONObject output = new JSONObject();
            output.put("content", response.getContent());
            output.put("model", response.getModel());
            output.put("tokens", response.getTotalTokens());
            output.put("latency_ms", response.getLatencyMs());

            return new ToolResult(TOOL_ID, output, (int) response.getLatencyMs());

        } catch (Exception e) {
            return new ToolResult(TOOL_ID, "LLM chat failed: " + e.getMessage());
        }
    }

    @Override
    public boolean isAvailable() {
        return provider.isAvailable();
    }

    @Override
    public boolean supportsOffline() {
        return provider.supportsOffline();
    }
}