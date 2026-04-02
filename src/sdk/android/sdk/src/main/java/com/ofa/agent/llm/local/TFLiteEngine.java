package com.ofa.agent.llm.local;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.llm.LLMConfig;

import org.tensorflow.lite.Interpreter;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.FloatBuffer;

/**
 * TensorFlow Lite 推理引擎
 * 用于本地 LLM 模型推理
 */
public class TFLiteEngine {

    private static final String TAG = "TFLiteEngine";

    private final Context context;
    private final LLMConfig config;

    private Interpreter interpreter;
    private int[] lastOutput;
    private boolean gpuEnabled;

    public TFLiteEngine(@NonNull Context context, @NonNull LLMConfig config) {
        this.context = context;
        this.config = config;
        this.gpuEnabled = config.isGpuAccel();
    }

    /**
     * 加载模型
     */
    public void loadModel(@NonNull String modelPath) throws IOException {
        Log.i(TAG, "Loading model from: " + modelPath);

        MappedByteBuffer buffer = loadModelFile(modelPath);

        Interpreter.Options options = new Interpreter.Options();
        options.setNumThreads(config.getThreads());

        // GPU 加速 (可选)
        if (gpuEnabled) {
            try {
                // 尝试启用 GPU
                // options.addDelegate(new GpuDelegate());
                Log.i(TAG, "GPU acceleration enabled");
            } catch (Exception e) {
                Log.w(TAG, "GPU not available, using CPU: " + e.getMessage());
                gpuEnabled = false;
            }
        }

        interpreter = new Interpreter(buffer, options);
        Log.i(TAG, "Model loaded successfully");
    }

    /**
     * 生成文本
     */
    public int[] generate(int[] inputTokens, int maxNewTokens) {
        if (interpreter == null) {
            throw new IllegalStateException("Model not loaded");
        }

        int[] currentTokens = inputTokens.clone();
        lastOutput = new int[maxNewTokens];
        int outputIndex = 0;

        for (int i = 0; i < maxNewTokens; i++) {
            // 准备输入
            float[][] input = prepareInput(currentTokens);

            // 推理
            float[][] output = new float[1][getVocabSize()];
            interpreter.run(input, output);

            // 采样下一个 token
            int nextToken = sample(output[0], config.getTemperature());
            lastOutput[outputIndex++] = nextToken;

            // 检查是否结束
            if (isEndToken(nextToken)) {
                break;
            }

            // 更新输入
            currentTokens = appendToken(currentTokens, nextToken);
        }

        // 裁剪输出
        int[] result = new int[outputIndex];
        System.arraycopy(lastOutput, 0, result, 0, outputIndex);
        lastOutput = result;

        return result;
    }

    /**
     * 流式生成
     */
    public void streamGenerate(int[] inputTokens, int maxNewTokens, TokenCallback callback) {
        if (interpreter == null) {
            throw new IllegalStateException("Model not loaded");
        }

        int[] currentTokens = inputTokens.clone();
        lastOutput = new int[maxNewTokens];
        int outputIndex = 0;

        for (int i = 0; i < maxNewTokens; i++) {
            float[][] input = prepareInput(currentTokens);
            float[][] output = new float[1][getVocabSize()];
            interpreter.run(input, output);

            int nextToken = sample(output[0], config.getTemperature());
            lastOutput[outputIndex++] = nextToken;

            callback.onToken(nextToken, false);

            if (isEndToken(nextToken)) {
                callback.onToken(-1, true);
                break;
            }

            currentTokens = appendToken(currentTokens, nextToken);
        }

        // 裁剪输出
        int[] result = new int[outputIndex];
        System.arraycopy(lastOutput, 0, result, 0, outputIndex);
        lastOutput = result;
    }

    /**
     * 获取嵌入向量
     */
    public float[] embed(int[] tokens) {
        if (interpreter == null) {
            return new float[0];
        }

        // 获取模型的隐藏层输出作为嵌入
        // 实际实现取决于模型结构
        float[][] input = prepareInput(tokens);
        float[][] output = new float[1][getEmbeddingDim()];
        // interpreter.run(input, output);

        return output[0];
    }

    /**
     * 获取最后的输出
     */
    public int[] getLastOutput() {
        return lastOutput;
    }

    /**
     * 配置更新
     */
    public void configure(LLMConfig config) {
        // 动态更新配置 (如 temperature)
    }

    /**
     * 关闭引擎
     */
    public void close() {
        if (interpreter != null) {
            interpreter.close();
            interpreter = null;
        }
    }

    // ===== Private Methods =====

    private MappedByteBuffer loadModelFile(String path) throws IOException {
        File file = new File(path);
        FileInputStream fis = new FileInputStream(file);
        FileChannel channel = fis.getChannel();
        long startOffset = channel.position();
        long declaredLength = file.length();
        return channel.map(FileChannel.MapMode.READ_ONLY, startOffset, declaredLength);
    }

    private float[][] prepareInput(int[] tokens) {
        // 将 token ID 转换为模型输入格式
        // 实际格式取决于模型结构
        float[][] input = new float[1][tokens.length];
        for (int i = 0; i < tokens.length; i++) {
            input[0][i] = tokens[i];
        }
        return input;
    }

    private int sample(float[] logits, float temperature) {
        // 温度采样
        if (temperature <= 0) {
            temperature = 1.0f;
        }

        // 应用温度
        float maxLogit = Float.NEGATIVE_INFINITY;
        for (float logit : logits) {
            if (logit > maxLogit) maxLogit = logit;
        }

        float sum = 0;
        float[] probs = new float[logits.length];
        for (int i = 0; i < logits.length; i++) {
            probs[i] = (float) Math.exp((logits[i] - maxLogit) / temperature);
            sum += probs[i];
        }

        // 归一化
        for (int i = 0; i < probs.length; i++) {
            probs[i] /= sum;
        }

        // 随机采样
        double r = Math.random();
        float cumulative = 0;
        for (int i = 0; i < probs.length; i++) {
            cumulative += probs[i];
            if (r < cumulative) {
                return i;
            }
        }

        return logits.length - 1;
    }

    private boolean isEndToken(int token) {
        // 检查是否是结束 token (EOS)
        // 实际值取决于模型
        return token == 2 || token == 1; // 常见的 EOS token ID
    }

    private int[] appendToken(int[] tokens, int newToken) {
        int[] result = new int[tokens.length + 1];
        System.arraycopy(tokens, 0, result, 0, tokens.length);
        result[tokens.length] = newToken;
        return result;
    }

    private int getVocabSize() {
        // 返回词表大小
        // 实际值取决于模型
        return 32000; // 常见的词表大小
    }

    private int getEmbeddingDim() {
        // 返回嵌入维度
        return 768;
    }

    /**
     * Token 回调
     */
    public interface TokenCallback {
        void onToken(int token, boolean finished);
    }
}