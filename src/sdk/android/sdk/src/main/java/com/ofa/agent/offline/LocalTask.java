package com.ofa.agent.offline;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.UUID;

/**
 * 离线任务
 */
public class LocalTask {
    private final String id;
    private final String skillId;
    private final byte[] input;
    private byte[] output;
    private TaskStatus status;
    private String error;
    private long createdAt;
    private long completedAt;
    private int retryCount;
    private int maxRetries;
    private boolean syncPending;

    public LocalTask(@NonNull String skillId, @Nullable byte[] input) {
        this.id = "local-" + UUID.randomUUID().toString().substring(0, 8);
        this.skillId = skillId;
        this.input = input;
        this.status = TaskStatus.PENDING;
        this.createdAt = System.currentTimeMillis();
        this.maxRetries = 3;
        this.syncPending = true;
    }

    public String getId() {
        return id;
    }

    public String getSkillId() {
        return skillId;
    }

    public byte[] getInput() {
        return input;
    }

    public byte[] getOutput() {
        return output;
    }

    public void setOutput(byte[] output) {
        this.output = output;
    }

    public TaskStatus getStatus() {
        return status;
    }

    public void setStatus(TaskStatus status) {
        this.status = status;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }

    public long getCreatedAt() {
        return createdAt;
    }

    public long getCompletedAt() {
        return completedAt;
    }

    public void setCompletedAt(long completedAt) {
        this.completedAt = completedAt;
    }

    public int getRetryCount() {
        return retryCount;
    }

    public void incrementRetry() {
        this.retryCount++;
    }

    public int getMaxRetries() {
        return maxRetries;
    }

    public void setMaxRetries(int maxRetries) {
        this.maxRetries = maxRetries;
    }

    public boolean isSyncPending() {
        return syncPending;
    }

    public void setSyncPending(boolean syncPending) {
        this.syncPending = syncPending;
    }

    public boolean canRetry() {
        return retryCount < maxRetries;
    }

    public long getDuration() {
        if (completedAt > 0) {
            return completedAt - createdAt;
        }
        return System.currentTimeMillis() - createdAt;
    }
}