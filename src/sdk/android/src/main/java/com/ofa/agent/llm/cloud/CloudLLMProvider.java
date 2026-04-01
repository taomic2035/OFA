package com.ofa.agent.llm.cloud;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.llm.LLMConfig;
import com.ofa.agent.llm.LLMProvider;
import com.ofa.agent.llm.LLMResponse;
import com.ofa.agent.llm.LLMStats;
import com.ofa.agent.llm.Message;
import com.ofa.agent.llm.StreamCallback;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 云端 LLM 提供者 (OpenAI 兼容)
 * 支持 OpenAI, Azure OpenAI, 以及其他兼容端点
 */
public class CloudLLMProvider implements LLMProvider {

    private static final String TAG = "CloudLLMProvider";

    private final String id;
    private LLMConfig config;
    private LLMStats stats;
    private final ExecutorService executor;
    private volatile boolean available;

    public CloudLLMProvider(@NonNull LLMConfig config) {
        this.id = "cloud-" + System.currentTimeMillis();
        this.config = config;
        this.stats = new LLMStats();
        this.executor = Executors.newCachedThreadPool();
        this.available = true;
    }

    public CloudLLMProvider(@NonNull String endpoint, @NonNull String apiKey) {
        this(new LLMConfig.Builder()
                .providerType(ProviderType.CLOUD)
                .endpoint(endpoint)
                .apiKey(apiKey)
                .build());
    }

    @Override
    @NonNull
    public String getId() {
        return id;
    }

    @Override
    @NonNull
    public String getName() {
        return "Cloud LLM (" + (config.getModel() != null ? config.getModel() : "unknown") + ")";
    }

    @Override
    @NonNull
    public ProviderType getType() {
        return ProviderType.CLOUD;
    }

    @Override
    public boolean isAvailable() {
        return available && config.getApiKey() != null && config.getEndpoint() != null;
    }

    @Override
    public boolean supportsOffline() {
        return false;
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chat(@NonNull String message) {
        List<Message> messages = new ArrayList<>();
        messages.add(Message.user(message));
        return chat(messages);
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chat(@NonNull List<Message> messages) {
        return chatWithTools(messages, null);
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chatWithTools(
            @NonNull List<Message> messages,
            @Nullable List<ToolDefinition> tools) {
        return CompletableFuture.supplyAsync(() -> {
            long startTime = System.currentTimeMillis();
            try {
                JSONObject requestBody = buildRequestBody(messages, tools, false);

                HttpURLConnection conn = createConnection();
                conn.setRequestMethod("POST");
                conn.setRequestProperty("Content-Type", "application/json");
                conn.setRequestProperty("Authorization", "Bearer " + config.getApiKey());
                conn.setDoOutput(true);
                conn.setConnectTimeout(config.getTimeout());
                conn.setReadTimeout(config.getTimeout());

                conn.getOutputStream().write(requestBody.toString().getBytes(StandardCharsets.UTF_8));

                int responseCode = conn.getResponseCode();
                if (responseCode != 200) {
                    String error = readError(conn);
                    return updateStats(new LLMResponse("HTTP " + responseCode + ": " + error), 0, 0, startTime);
                }

                String response = readResponse(conn);
                LLMResponse llmResponse = parseResponse(response, startTime);
                return updateStats(llmResponse, llmResponse.getInputTokens(), llmResponse.getOutputTokens(), startTime);

            } catch (Exception e) {
                Log.e(TAG, "Chat failed", e);
                return updateStats(new LLMResponse(e.getMessage()), 0, 0, startTime);
            }
        }, executor);
    }

    @Override
    public void streamChat(@NonNull List<Message> messages, @NonNull StreamCallback callback) {
        executor.execute(() -> {
            long startTime = System.currentTimeMillis();
            try {
                JSONObject requestBody = buildRequestBody(messages, null, true);

                HttpURLConnection conn = createConnection();
                conn.setRequestMethod("POST");
                conn.setRequestProperty("Content-Type", "application/json");
                conn.setRequestProperty("Authorization", "Bearer " + config.getApiKey());
                conn.setRequestProperty("Accept", "text/event-stream");
                conn.setDoOutput(true);
                conn.setConnectTimeout(config.getTimeout());
                conn.setReadTimeout(config.getTimeout());

                conn.getOutputStream().write(requestBody.toString().getBytes(StandardCharsets.UTF_8));

                StringBuilder fullContent = new StringBuilder();
                int inputTokens = 0;
                int outputTokens = 0;

                try (BufferedReader reader = new BufferedReader(
                        new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        if (line.startsWith("data: ")) {
                            String data = line.substring(6);
                            if ("[DONE]".equals(data)) break;

                            JSONObject chunk = new JSONObject(data);
                            JSONArray choices = chunk.optJSONArray("choices");
                            if (choices != null && choices.length() > 0) {
                                JSONObject delta = choices.getJSONObject(0).optJSONObject("delta");
                                if (delta != null) {
                                    String content = delta.optString("content", "");
                                    if (!content.isEmpty()) {
                                        fullContent.append(content);
                                        callback.onToken(content);
                                    }
                                }
                            }

                            JSONObject usage = chunk.optJSONObject("usage");
                            if (usage != null) {
                                inputTokens = usage.optInt("prompt_tokens", 0);
                                outputTokens = usage.optInt("completion_tokens", 0);
                            }
                        }
                    }
                }

                LLMResponse response = new LLMResponse(fullContent.toString(), inputTokens, outputTokens,
                        System.currentTimeMillis() - startTime, config.getModel());
                callback.onComplete(updateStats(response, inputTokens, outputTokens, startTime));

            } catch (Exception e) {
                Log.e(TAG, "Stream chat failed", e);
                callback.onError(e.getMessage());
                updateStats(new LLMResponse(e.getMessage()), 0, 0, startTime);
            }
        });
    }

    @Override
    @NonNull
    public CompletableFuture<float[]> embed(@NonNull String text) {
        return CompletableFuture.supplyAsync(() -> {
            // TODO: 实现 embedding 接口
            return new float[0];
        }, executor);
    }

    @Override
    public void configure(@NonNull LLMConfig config) {
        this.config = config;
    }

    @Override
    @NonNull
    public LLMStats getStats() {
        return stats;
    }

    @Override
    public void shutdown() {
        available = false;
        executor.shutdown();
    }

    // ===== Private Methods =====

    private HttpURLConnection createConnection() throws Exception {
        String endpoint = config.getEndpoint();
        if (!endpoint.endsWith("/")) {
            endpoint += "/";
        }
        endpoint += "chat/completions";
        URL url = new URL(endpoint);
        return (HttpURLConnection) url.openConnection();
    }

    private JSONObject buildRequestBody(List<Message> messages, List<ToolDefinition> tools, boolean stream) throws Exception {
        JSONObject body = new JSONObject();

        body.put("model", config.getModel());
        body.put("stream", stream);
        body.put("max_tokens", config.getMaxTokens());
        body.put("temperature", config.getTemperature());
        body.put("top_p", config.getTopP());

        // Messages
        JSONArray messagesArray = new JSONArray();
        for (Message msg : messages) {
            JSONObject msgObj = new JSONObject();
            msgObj.put("role", msg.getRole().name().toLowerCase());
            if (msg.getContent() != null) {
                msgObj.put("content", msg.getContent());
            }
            if (msg.getName() != null) {
                msgObj.put("name", msg.getName());
            }
            messagesArray.put(msgObj);
        }
        body.put("messages", messagesArray);

        // Tools
        if (tools != null && !tools.isEmpty()) {
            JSONArray toolsArray = new JSONArray();
            for (ToolDefinition tool : tools) {
                JSONObject toolObj = new JSONObject();
                toolObj.put("type", "function");
                JSONObject function = new JSONObject();
                function.put("name", tool.getName());
                function.put("description", tool.getDescription());
                if (tool.getParameters() != null) {
                    function.put("parameters", new JSONObject(tool.getParameters()));
                }
                toolObj.put("function", function);
                toolsArray.put(toolObj);
            }
            body.put("tools", toolsArray);
        }

        return body;
    }

    private LLMResponse parseResponse(String response, long startTime) throws Exception {
        JSONObject json = new JSONObject(response);

        JSONArray choices = json.getJSONArray("choices");
        JSONObject choice = choices.getJSONObject(0);

        JSONObject usage = json.optJSONObject("usage");
        int inputTokens = usage != null ? usage.optInt("prompt_tokens", 0) : 0;
        int outputTokens = usage != null ? usage.optInt("completion_tokens", 0) : 0;

        String finishReason = choice.optString("finish_reason", "stop");

        // 检查是否有工具调用
        JSONObject message = choice.getJSONObject("message");
        JSONArray toolCalls = message.optJSONArray("tool_calls");

        if (toolCalls != null && toolCalls.length() > 0) {
            JSONObject toolCall = toolCalls.getJSONObject(0);
            String id = toolCall.getString("id");
            JSONObject function = toolCall.getJSONObject("function");
            String name = function.getString("name");
            JSONObject arguments = new JSONObject(function.getString("arguments"));

            LLMResponse.ToolCall tc = new LLMResponse.ToolCall(id, name, arguments);
            return new LLMResponse(tc, inputTokens, outputTokens,
                    System.currentTimeMillis() - startTime, config.getModel());
        }

        String content = message.optString("content", "");
        return new LLMResponse(content, inputTokens, outputTokens,
                System.currentTimeMillis() - startTime, config.getModel());
    }

    private String readResponse(HttpURLConnection conn) throws Exception {
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8))) {
            StringBuilder sb = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                sb.append(line);
            }
            return sb.toString();
        }
    }

    private String readError(HttpURLConnection conn) throws Exception {
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(conn.getErrorStream(), StandardCharsets.UTF_8))) {
            StringBuilder sb = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                sb.append(line);
            }
            return sb.toString();
        }
    }

    private LLMResponse updateStats(LLMResponse response, int inputTokens, int outputTokens, long startTime) {
        long latency = System.currentTimeMillis() - startTime;
        stats = new LLMStats.Builder()
                .from(stats)
                .addRequest(response.isSuccess(), inputTokens, outputTokens, latency)
                .build();
        return response;
    }
}