package com.ofa.agent.mcp;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.Map;

/**
 * Prompt Result - rendered prompt output.
 */
public class PromptResult {

    private final String name;
    private final String renderedText;
    private final Map<String, Object> arguments;
    private final boolean success;
    private final String error;

    /**
     * Success constructor
     */
    public PromptResult(@NonNull String name, @NonNull String renderedText,
                        @Nullable Map<String, Object> arguments) {
        this.name = name;
        this.renderedText = renderedText;
        this.arguments = arguments;
        this.success = true;
        this.error = null;
    }

    /**
     * Error constructor
     */
    public PromptResult(@NonNull String name, @NonNull String error) {
        this.name = name;
        this.renderedText = null;
        this.arguments = null;
        this.success = false;
        this.error = error;
    }

    @NonNull
    public String getName() {
        return name;
    }

    @Nullable
    public String getRenderedText() {
        return renderedText;
    }

    @Nullable
    public Map<String, Object> getArguments() {
        return arguments;
    }

    public boolean isSuccess() {
        return success;
    }

    @Nullable
    public String getError() {
        return error;
    }
}