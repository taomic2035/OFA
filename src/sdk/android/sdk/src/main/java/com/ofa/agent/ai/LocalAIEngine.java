package com.ofa.agent.ai;

import android.content.Context;
import android.graphics.Bitmap;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Local AI Engine - provides on-device AI inference capabilities.
 * Supports text and image inference with multiple model backends.
 */
public class LocalAIEngine {

    private static final String TAG = "LocalAIEngine";

    private final Context context;
    private final ExecutorService executor;
    private final ModelManager modelManager;

    private InferenceConfig config;
    private boolean initialized = false;
    private InferenceListener listener;

    /**
     * Inference result
     */
    public static class InferenceResult {
        public final boolean success;
        public final String prediction;
        public final float confidence;
        public final Map<String, Float> probabilities;
        public final Map<String, String> extras;
        public final long inferenceTimeMs;
        public final String error;

        public InferenceResult(boolean success, String prediction, float confidence,
                               Map<String, Float> probabilities, Map<String, String> extras,
                               long inferenceTimeMs, String error) {
            this.success = success;
            this.prediction = prediction;
            this.confidence = confidence;
            this.probabilities = probabilities;
            this.extras = extras;
            this.inferenceTimeMs = inferenceTimeMs;
            this.error = error;
        }

        public static InferenceResult success(String prediction, float confidence,
                                               Map<String, Float> probabilities) {
            return new InferenceResult(true, prediction, confidence, probabilities,
                null, 0, null);
        }

        public static InferenceResult error(String error) {
            return new InferenceResult(false, null, 0, null, null, 0, error);
        }
    }

    /**
     * Inference listener
     */
    public interface InferenceListener {
        void onInferenceComplete(@NonNull String modelId, @NonNull InferenceResult result);
        void onModelError(@NonNull String modelId, @NonNull String error);
    }

    /**
     * Create local AI engine
     */
    public LocalAIEngine(@NonNull Context context) {
        this.context = context;
        this.executor = Executors.newSingleThreadExecutor();
        this.modelManager = new ModelManager(context);
    }

    /**
     * Initialize engine with config
     */
    public void initialize(@NonNull InferenceConfig config) {
        Log.i(TAG, "Initializing Local AI Engine...");

        this.config = config;

        // Load required models
        for (String modelId : config.getRequiredModels()) {
            boolean loaded = modelManager.loadModel(modelId);
            if (!loaded && config.isStrictMode()) {
                Log.e(TAG, "Failed to load required model: " + modelId);
                return;
            }
        }

        initialized = true;
        Log.i(TAG, "Local AI Engine initialized with " + modelManager.getLoadedCount() + " models");
    }

    /**
     * Initialize with default config
     */
    public void initialize() {
        initialize(InferenceConfig.getDefault());
    }

    /**
     * Check if engine is ready
     */
    public boolean isReady() {
        return initialized && modelManager.hasModels();
    }

    /**
     * Set inference listener
     */
    public void setListener(@Nullable InferenceListener listener) {
        this.listener = listener;
    }

    /**
     * Run text inference
     */
    @NonNull
    public InferenceResult inferText(@NonNull String text) {
        return inferText("intent_classifier", text);
    }

    /**
     * Run text inference with specific model
     */
    @NonNull
    public InferenceResult inferText(@NonNull String modelId, @NonNull String text) {
        if (!initialized) {
            return InferenceResult.error("Engine not initialized");
        }

        long startTime = System.currentTimeMillis();

        try {
            // Preprocess text
            float[] input = preprocessText(text);

            // Run inference
            float[] output = modelManager.runInference(modelId, input);

            // Postprocess
            InferenceResult result = postprocessOutput(modelId, output);

            // Update timing
            long inferenceTime = System.currentTimeMillis() - startTime;

            // Notify listener
            if (listener != null) {
                listener.onInferenceComplete(modelId, result);
            }

            Log.d(TAG, "Text inference completed in " + inferenceTime + "ms");
            return result;

        } catch (Exception e) {
            Log.e(TAG, "Text inference failed: " + e.getMessage());

            if (listener != null) {
                listener.onModelError(modelId, e.getMessage());
            }

            return InferenceResult.error(e.getMessage());
        }
    }

    /**
     * Run image inference
     */
    @NonNull
    public InferenceResult inferImage(@NonNull Bitmap image) {
        return inferImage("ui_element", image);
    }

    /**
     * Run image inference with specific model
     */
    @NonNull
    public InferenceResult inferImage(@NonNull String modelId, @NonNull Bitmap image) {
        if (!initialized) {
            return InferenceResult.error("Engine not initialized");
        }

        long startTime = System.currentTimeMillis();

        try {
            // Preprocess image
            float[][][][] input = preprocessImage(image);

            // Run inference
            float[] output = modelManager.runInference(modelId, input);

            // Postprocess
            InferenceResult result = postprocessOutput(modelId, output);

            // Update timing
            long inferenceTime = System.currentTimeMillis() - startTime;

            // Notify listener
            if (listener != null) {
                listener.onInferenceComplete(modelId, result);
            }

            Log.d(TAG, "Image inference completed in " + inferenceTime + "ms");
            return result;

        } catch (Exception e) {
            Log.e(TAG, "Image inference failed: " + e.getMessage());

            if (listener != null) {
                listener.onModelError(modelId, e.getMessage());
            }

            return InferenceResult.error(e.getMessage());
        }
    }

    /**
     * Run inference asynchronously
     */
    public void inferTextAsync(@NonNull String modelId, @NonNull String text,
                                @NonNull InferenceCallback callback) {
        executor.execute(() -> {
            InferenceResult result = inferText(modelId, text);
            callback.onResult(result);
        });
    }

    /**
     * Inference callback
     */
    public interface InferenceCallback {
        void onResult(@NonNull InferenceResult result);
    }

    /**
     * Preprocess text for model input
     */
    private float[] preprocessText(@NonNull String text) {
        // Tokenize and convert to float array
        // This is a simplified version - real implementation would use proper tokenizer
        String[] tokens = text.toLowerCase().split("\\s+");
        float[] input = new float[config.getMaxSequenceLength()];

        for (int i = 0; i < Math.min(tokens.length, input.length); i++) {
            // Simple hash-based token ID
            input[i] = Math.abs(tokens[i].hashCode() % 10000);
        }

        return input;
    }

    /**
     * Preprocess image for model input
     */
    private float[][][][] preprocessImage(@NonNull Bitmap image) {
        int targetWidth = config.getImageWidth();
        int targetHeight = config.getImageHeight();

        // Resize image
        Bitmap resized = Bitmap.createScaledBitmap(image, targetWidth, targetHeight, true);

        // Convert to float array
        float[][][][] input = new float[1][targetHeight][targetWidth][3];

        for (int y = 0; y < targetHeight; y++) {
            for (int x = 0; x < targetWidth; x++) {
                int pixel = resized.getPixel(x, y);
                input[0][y][x][0] = ((pixel >> 16) & 0xFF) / 255.0f; // R
                input[0][y][x][1] = ((pixel >> 8) & 0xFF) / 255.0f;  // G
                input[0][y][x][2] = (pixel & 0xFF) / 255.0f;         // B
            }
        }

        return input;
    }

    /**
     * Postprocess model output
     */
    @NonNull
    private InferenceResult postprocessOutput(@NonNull String modelId, float[] output) {
        // Find max probability
        int maxIndex = 0;
        float maxProb = output[0];
        float total = 0;

        for (int i = 0; i < output.length; i++) {
            total += output[i];
            if (output[i] > maxProb) {
                maxProb = output[i];
                maxIndex = i;
            }
        }

        // Normalize probabilities
        java.util.Map<String, Float> probabilities = new java.util.HashMap<>();
        String[] labels = modelManager.getLabels(modelId);

        if (labels != null) {
            for (int i = 0; i < Math.min(labels.length, output.length); i++) {
                probabilities.put(labels[i], output[i] / total);
            }
        }

        String prediction = labels != null && maxIndex < labels.length ?
            labels[maxIndex] : String.valueOf(maxIndex);

        return InferenceResult.success(prediction, maxProb / total, probabilities);
    }

    /**
     * Get model manager
     */
    @NonNull
    public ModelManager getModelManager() {
        return modelManager;
    }

    /**
     * Get loaded model IDs
     */
    @NonNull
    public String[] getLoadedModels() {
        return modelManager.getLoadedModelIds();
    }

    /**
     * Check if model is loaded
     */
    public boolean isModelLoaded(@NonNull String modelId) {
        return modelManager.isModelLoaded(modelId);
    }

    /**
     * Get inference stats
     */
    @NonNull
    public String getStats() {
        return "LocalAIEngine{" +
            "initialized=" + initialized +
            ", modelsLoaded=" + modelManager.getLoadedCount() +
            '}';
    }

    /**
     * Shutdown engine
     */
    public void shutdown() {
        Log.i(TAG, "Shutting down Local AI Engine...");

        modelManager.unloadAll();
        executor.shutdown();
        initialized = false;

        Log.i(TAG, "Local AI Engine shutdown complete");
    }
}