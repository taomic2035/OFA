package com.ofa.agent.ai;

import androidx.annotation.NonNull;

import java.util.Arrays;
import java.util.HashSet;
import java.util.Set;

/**
 * Inference Configuration - configures AI inference parameters.
 */
public class InferenceConfig {

    // Default values
    private static final int DEFAULT_MAX_SEQUENCE_LENGTH = 128;
    private static final int DEFAULT_IMAGE_WIDTH = 224;
    private static final int DEFAULT_IMAGE_HEIGHT = 224;
    private static final int DEFAULT_NUM_THREADS = 4;
    private static final long DEFAULT_MODEL_CACHE_SIZE = 100 * 1024 * 1024; // 100MB

    // Text inference settings
    private int maxSequenceLength = DEFAULT_MAX_SEQUENCE_LENGTH;
    private String tokenizerType = "wordpiece";
    private boolean lowercase = true;

    // Image inference settings
    private int imageWidth = DEFAULT_IMAGE_WIDTH;
    private int imageHeight = DEFAULT_IMAGE_HEIGHT;
    private boolean normalizeImage = true;

    // Inference settings
    private int numThreads = DEFAULT_NUM_THREADS;
    private boolean useGpu = false;
    private boolean useNnapi = false;
    private long modelCacheSize = DEFAULT_MODEL_CACHE_SIZE;

    // Model requirements
    private Set<String> requiredModels = new HashSet<>();
    private boolean strictMode = true;

    /**
     * Default constructor
     */
    public InferenceConfig() {}

    /**
     * Get default configuration
     */
    @NonNull
    public static InferenceConfig getDefault() {
        return new InferenceConfig();
    }

    // ===== Getters =====

    public int getMaxSequenceLength() {
        return maxSequenceLength;
    }

    public String getTokenizerType() {
        return tokenizerType;
    }

    public boolean isLowercase() {
        return lowercase;
    }

    public int getImageWidth() {
        return imageWidth;
    }

    public int getImageHeight() {
        return imageHeight;
    }

    public boolean isNormalizeImage() {
        return normalizeImage;
    }

    public int getNumThreads() {
        return numThreads;
    }

    public boolean isUseGpu() {
        return useGpu;
    }

    public boolean isUseNnapi() {
        return useNnapi;
    }

    public long getModelCacheSize() {
        return modelCacheSize;
    }

    @NonNull
    public Set<String> getRequiredModels() {
        return new HashSet<>(requiredModels);
    }

    public boolean isStrictMode() {
        return strictMode;
    }

    // ===== Builder =====

    @NonNull
    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private final InferenceConfig config = new InferenceConfig();

        @NonNull
        public Builder maxSequenceLength(int length) {
            config.maxSequenceLength = length;
            return this;
        }

        @NonNull
        public Builder tokenizerType(@NonNull String type) {
            config.tokenizerType = type;
            return this;
        }

        @NonNull
        public Builder lowercase(boolean lowercase) {
            config.lowercase = lowercase;
            return this;
        }

        @NonNull
        public Builder imageSize(int width, int height) {
            config.imageWidth = width;
            config.imageHeight = height;
            return this;
        }

        @NonNull
        public Builder normalizeImage(boolean normalize) {
            config.normalizeImage = normalize;
            return this;
        }

        @NonNull
        public Builder numThreads(int threads) {
            config.numThreads = threads;
            return this;
        }

        @NonNull
        public Builder useGpu(boolean useGpu) {
            config.useGpu = useGpu;
            return this;
        }

        @NonNull
        public Builder useNnapi(boolean useNnapi) {
            config.useNnapi = useNnapi;
            return this;
        }

        @NonNull
        public Builder modelCacheSize(long sizeBytes) {
            config.modelCacheSize = sizeBytes;
            return this;
        }

        @NonNull
        public Builder addRequiredModel(@NonNull String modelId) {
            config.requiredModels.add(modelId);
            return this;
        }

        @NonNull
        public Builder requiredModels(@NonNull String... modelIds) {
            config.requiredModels.addAll(Arrays.asList(modelIds));
            return this;
        }

        @NonNull
        public Builder strictMode(boolean strict) {
            config.strictMode = strict;
            return this;
        }

        @NonNull
        public InferenceConfig build() {
            return config;
        }
    }

    // ===== Preset Configurations =====

    /**
     * Lightweight config for intent classification only
     */
    @NonNull
    public static InferenceConfig lightweight() {
        return builder()
            .requiredModels("intent_classifier")
            .numThreads(2)
            .modelCacheSize(50 * 1024 * 1024)
            .build();
    }

    /**
     * Standard config with intent and slot models
     */
    @NonNull
    public static InferenceConfig standard() {
        return builder()
            .requiredModels("intent_classifier", "slot_extractor")
            .numThreads(4)
            .build();
    }

    /**
     * Full config with all models
     */
    @NonNull
    public static InferenceConfig full() {
        return builder()
            .requiredModels("intent_classifier", "slot_extractor", "ui_element")
            .numThreads(4)
            .useGpu(true)
            .build();
    }

    /**
     * Vision-focused config
     */
    @NonNull
    public static InferenceConfig vision() {
        return builder()
            .requiredModels("ui_element")
            .imageSize(320, 320)
            .useGpu(true)
            .build();
    }
}