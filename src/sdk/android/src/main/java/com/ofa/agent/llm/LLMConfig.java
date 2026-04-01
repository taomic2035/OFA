package com.ofa.agent.llm;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * LLM 配置
 */
public class LLMConfig {

    private final LLMProvider.ProviderType providerType;
    private final String endpoint;
    private final String apiKey;
    private final String model;
    private final int maxTokens;
    private final float temperature;
    private final float topP;
    private final int timeout;
    private final int maxRetries;

    // 本地 LLM 特有配置
    private final String modelPath;
    private final String tokenizerPath;
    private final int threads;
    private final boolean gpuAccel;

    private LLMConfig(Builder builder) {
        this.providerType = builder.providerType;
        this.endpoint = builder.endpoint;
        this.apiKey = builder.apiKey;
        this.model = builder.model;
        this.maxTokens = builder.maxTokens;
        this.temperature = builder.temperature;
        this.topP = builder.topP;
        this.timeout = builder.timeout;
        this.maxRetries = builder.maxRetries;
        this.modelPath = builder.modelPath;
        this.tokenizerPath = builder.tokenizerPath;
        this.threads = builder.threads;
        this.gpuAccel = builder.gpuAccel;
    }

    @NonNull
    public LLMProvider.ProviderType getProviderType() { return providerType; }

    @Nullable
    public String getEndpoint() { return endpoint; }

    @Nullable
    public String getApiKey() { return apiKey; }

    @Nullable
    public String getModel() { return model; }

    public int getMaxTokens() { return maxTokens; }

    public float getTemperature() { return temperature; }

    public float getTopP() { return topP; }

    public int getTimeout() { return timeout; }

    public int getMaxRetries() { return maxRetries; }

    @Nullable
    public String getModelPath() { return modelPath; }

    @Nullable
    public String getTokenizerPath() { return tokenizerPath; }

    public int getThreads() { return threads; }

    public boolean isGpuAccel() { return gpuAccel; }

    public static class Builder {
        private LLMProvider.ProviderType providerType = LLMProvider.ProviderType.CLOUD;
        private String endpoint;
        private String apiKey;
        private String model = "gpt-3.5-turbo";
        private int maxTokens = 2048;
        private float temperature = 0.7f;
        private float topP = 1.0f;
        private int timeout = 30000;
        private int maxRetries = 3;

        private String modelPath;
        private String tokenizerPath;
        private int threads = 4;
        private boolean gpuAccel = false;

        public Builder providerType(@NonNull LLMProvider.ProviderType type) {
            this.providerType = type;
            return this;
        }

        public Builder endpoint(@NonNull String endpoint) {
            this.endpoint = endpoint;
            return this;
        }

        public Builder apiKey(@NonNull String apiKey) {
            this.apiKey = apiKey;
            return this;
        }

        public Builder model(@NonNull String model) {
            this.model = model;
            return this;
        }

        public Builder maxTokens(int maxTokens) {
            this.maxTokens = maxTokens;
            return this;
        }

        public Builder temperature(float temperature) {
            this.temperature = temperature;
            return this;
        }

        public Builder topP(float topP) {
            this.topP = topP;
            return this;
        }

        public Builder timeout(int timeout) {
            this.timeout = timeout;
            return this;
        }

        public Builder maxRetries(int maxRetries) {
            this.maxRetries = maxRetries;
            return this;
        }

        public Builder modelPath(@NonNull String path) {
            this.modelPath = path;
            return this;
        }

        public Builder tokenizerPath(@NonNull String path) {
            this.tokenizerPath = path;
            return this;
        }

        public Builder threads(int threads) {
            this.threads = threads;
            return this;
        }

        public Builder gpuAccel(boolean enable) {
            this.gpuAccel = enable;
            return this;
        }

        public LLMConfig build() {
            return new LLMConfig(this);
        }
    }
}