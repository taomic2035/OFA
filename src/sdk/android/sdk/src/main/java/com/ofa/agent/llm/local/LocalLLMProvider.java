package com.ofa.agent.llm.local;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.llm.LLMConfig;
import com.ofa.agent.llm.LLMProvider;
import com.ofa.agent.llm.LLMResponse;
import com.ofa.agent.llm.LLMStats;
import com.ofa.agent.llm.Message;
import com.ofa.agent.llm.StreamCallback;

import org.json.JSONObject;

import java.io.File;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 本地 LLM 提供者 (TensorFlow Lite)
 * 支持在设备端运行量化模型
 */
public class LocalLLMProvider implements LLMProvider {

    private static final String TAG = "LocalLLMProvider";

    private final String id;
    private final Context context;
    private LLMConfig config;
    private LLMStats stats;
    private final ExecutorService executor;
    private volatile boolean available;
    private volatile boolean initialized;

    // TensorFlow Lite 引擎
    private TFLiteEngine engine;
    private Tokenizer tokenizer;

    public LocalLLMProvider(@NonNull Context context, @NonNull LLMConfig config) {
        this.id = "local-" + System.currentTimeMillis();
        this.context = context.getApplicationContext();
        this.config = config;
        this.stats = new LLMStats();
        this.executor = Executors.newSingleThreadExecutor();
        this.available = false;
        this.initialized = false;
    }

    public LocalLLMProvider(@NonNull Context context, @NonNull String modelPath) {
        this(context, new LLMConfig.Builder()
                .providerType(ProviderType.LOCAL)
                .modelPath(modelPath)
                .build());
    }

    /**
     * 初始化模型
     */
    public synchronized void initialize() {
        if (initialized) return;

        executor.execute(() -> {
            try {
                Log.i(TAG, "Initializing local LLM: " + config.getModelPath());

                // 初始化分词器
                if (config.getTokenizerPath() != null) {
                    tokenizer = new Tokenizer(config.getTokenizerPath());
                } else {
                    // 使用默认分词器
                    tokenizer = new Tokenizer();
                }

                // 初始化 TFLite 引擎
                engine = new TFLiteEngine(context, config);
                engine.loadModel(config.getModelPath());

                initialized = true;
                available = true;
                Log.i(TAG, "Local LLM initialized successfully");

            } catch (Exception e) {
                Log.e(TAG, "Failed to initialize local LLM", e);
                available = false;
                initialized = false;
            }
        });
    }

    @Override
    @NonNull
    public String getId() {
        return id;
    }

    @Override
    @NonNull
    public String getName() {
        String model = config.getModelPath();
        if (model != null) {
            int lastSlash = model.lastIndexOf('/');
            if (lastSlash >= 0) {
                model = model.substring(lastSlash + 1);
            }
            return "Local LLM (" + model + ")";
        }
        return "Local LLM";
    }

    @Override
    @NonNull
    public ProviderType getType() {
        return ProviderType.LOCAL;
    }

    @Override
    public boolean isAvailable() {
        return available && initialized;
    }

    @Override
    public boolean supportsOffline() {
        return true;
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
        return CompletableFuture.supplyAsync(() -> {
            if (!isAvailable()) {
                return new LLMResponse("Local LLM not initialized");
            }

            long startTime = System.currentTimeMillis();
            try {
                // 构建提示词
                String prompt = buildPrompt(messages);

                // 分词
                int[] inputTokens = tokenizer.encode(prompt);

                // 推理
                int[] outputTokens = engine.generate(inputTokens, config.getMaxTokens());

                // 解码
                String response = tokenizer.decode(outputTokens);

                // 计算延迟
                long latency = System.currentTimeMillis() - startTime;

                // 估算 token 数量
                int inputCount = inputTokens.length;
                int outputCount = outputTokens.length;

                return updateStats(new LLMResponse(response, inputCount, outputCount, latency, "local"),
                        inputCount, outputCount, startTime);

            } catch (Exception e) {
                Log.e(TAG, "Local inference failed", e);
                return updateStats(new LLMResponse(e.getMessage()), 0, 0, startTime);
            }
        }, executor);
    }

    @Override
    @NonNull
    public CompletableFuture<LLMResponse> chatWithTools(
            @NonNull List<Message> messages,
            @Nullable List<ToolDefinition> tools) {
        // 本地模型暂不支持 Function Calling
        // 返回普通聊天响应
        return chat(messages);
    }

    @Override
    public void streamChat(@NonNull List<Message> messages, @NonNull StreamCallback callback) {
        executor.execute(() -> {
            if (!isAvailable()) {
                callback.onError("Local LLM not initialized");
                return;
            }

            long startTime = System.currentTimeMillis();
            try {
                String prompt = buildPrompt(messages);
                int[] inputTokens = tokenizer.encode(prompt);

                // 流式生成
                engine.streamGenerate(inputTokens, config.getMaxTokens(), (token, finished) -> {
                    String text = tokenizer.decode(new int[]{token});
                    callback.onToken(text);
                });

                // 完成
                int[] outputTokens = engine.getLastOutput();
                String response = tokenizer.decode(outputTokens);
                long latency = System.currentTimeMillis() - startTime;

                LLMResponse llmResponse = new LLMResponse(response, inputTokens.length,
                        outputTokens.length, latency, "local");
                callback.onComplete(updateStats(llmResponse, inputTokens.length, outputTokens.length, startTime));

            } catch (Exception e) {
                Log.e(TAG, "Local stream inference failed", e);
                callback.onError(e.getMessage());
            }
        });
    }

    @Override
    @NonNull
    public CompletableFuture<float[]> embed(@NonNull String text) {
        return CompletableFuture.supplyAsync(() -> {
            if (!isAvailable() || engine == null) {
                return new float[0];
            }
            try {
                int[] tokens = tokenizer.encode(text);
                return engine.embed(tokens);
            } catch (Exception e) {
                Log.e(TAG, "Embedding failed", e);
                return new float[0];
            }
        }, executor);
    }

    @Override
    public void configure(@NonNull LLMConfig config) {
        this.config = config;
        if (engine != null) {
            engine.configure(config);
        }
    }

    @Override
    @NonNull
    public LLMStats getStats() {
        return stats;
    }

    @Override
    public void shutdown() {
        available = false;
        initialized = false;
        if (engine != null) {
            engine.close();
            engine = null;
        }
        executor.shutdown();
    }

    // ===== Private Methods =====

    private String buildPrompt(List<Message> messages) {
        StringBuilder sb = new StringBuilder();

        for (Message msg : messages) {
            switch (msg.getRole()) {
                case SYSTEM:
                    sb.append("<|system|>\n").append(msg.getContent()).append("\n");
                    break;
                case USER:
                    sb.append("<|user|>\n").append(msg.getContent()).append("\n");
                    break;
                case ASSISTANT:
                    sb.append("<|assistant|>\n").append(msg.getContent()).append("\n");
                    break;
                case TOOL:
                    sb.append("<|tool|>\n").append(msg.getContent()).append("\n");
                    break;
            }
        }

        sb.append("<|assistant|>\n");
        return sb.toString();
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