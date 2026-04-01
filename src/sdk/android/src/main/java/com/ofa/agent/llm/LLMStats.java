package com.ofa.agent.llm;

import androidx.annotation.NonNull;

/**
 * LLM 统计信息
 */
public class LLMStats {

    private final long totalRequests;
    private final long successRequests;
    private final long failedRequests;
    private final long totalInputTokens;
    private final long totalOutputTokens;
    private final long totalLatencyMs;
    private final double avgLatencyMs;
    private final long lastRequestTime;

    public LLMStats() {
        this.totalRequests = 0;
        this.successRequests = 0;
        this.failedRequests = 0;
        this.totalInputTokens = 0;
        this.totalOutputTokens = 0;
        this.totalLatencyMs = 0;
        this.avgLatencyMs = 0;
        this.lastRequestTime = 0;
    }

    public LLMStats(long totalRequests, long successRequests, long failedRequests,
                    long totalInputTokens, long totalOutputTokens, long totalLatencyMs,
                    long lastRequestTime) {
        this.totalRequests = totalRequests;
        this.successRequests = successRequests;
        this.failedRequests = failedRequests;
        this.totalInputTokens = totalInputTokens;
        this.totalOutputTokens = totalOutputTokens;
        this.totalLatencyMs = totalLatencyMs;
        this.avgLatencyMs = totalRequests > 0 ? (double) totalLatencyMs / totalRequests : 0;
        this.lastRequestTime = lastRequestTime;
    }

    public long getTotalRequests() { return totalRequests; }

    public long getSuccessRequests() { return successRequests; }

    public long getFailedRequests() { return failedRequests; }

    public long getTotalInputTokens() { return totalInputTokens; }

    public long getTotalOutputTokens() { return totalOutputTokens; }

    public long getTotalTokens() { return totalInputTokens + totalOutputTokens; }

    public long getTotalLatencyMs() { return totalLatencyMs; }

    public double getAvgLatencyMs() { return avgLatencyMs; }

    public long getLastRequestTime() { return lastRequestTime; }

    public double getSuccessRate() {
        return totalRequests > 0 ? (double) successRequests / totalRequests : 0;
    }

    /**
     * Builder for incremental updates
     */
    public static class Builder {
        private long totalRequests;
        private long successRequests;
        private long failedRequests;
        private long totalInputTokens;
        private long totalOutputTokens;
        private long totalLatencyMs;
        private long lastRequestTime;

        public Builder from(@NonNull LLMStats stats) {
            this.totalRequests = stats.totalRequests;
            this.successRequests = stats.successRequests;
            this.failedRequests = stats.failedRequests;
            this.totalInputTokens = stats.totalInputTokens;
            this.totalOutputTokens = stats.totalOutputTokens;
            this.totalLatencyMs = stats.totalLatencyMs;
            this.lastRequestTime = stats.lastRequestTime;
            return this;
        }

        public Builder addRequest(boolean success, int inputTokens, int outputTokens, long latencyMs) {
            this.totalRequests++;
            if (success) {
                this.successRequests++;
            } else {
                this.failedRequests++;
            }
            this.totalInputTokens += inputTokens;
            this.totalOutputTokens += outputTokens;
            this.totalLatencyMs += latencyMs;
            this.lastRequestTime = System.currentTimeMillis();
            return this;
        }

        public LLMStats build() {
            return new LLMStats(totalRequests, successRequests, failedRequests,
                    totalInputTokens, totalOutputTokens, totalLatencyMs, lastRequestTime);
        }
    }
}