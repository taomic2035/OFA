package com.ofa.agent.llm.local;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.FileInputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;

/**
 * 分词器
 * 支持常见的分词算法 (BPE, SentencePiece 等)
 */
public class Tokenizer {

    private static final String TAG = "Tokenizer";

    private final Map<String, Integer> encoder = new HashMap<>();
    private final Map<Integer, String> decoder = new HashMap<>();
    private final Map<String, Integer> specialTokens = new HashMap<>();

    private int vocabSize;
    private int padTokenId = 0;
    private int eosTokenId = 2;
    private int bosTokenId = 1;

    /**
     * 默认分词器 (简单字符级)
     */
    public Tokenizer() {
        initDefaultVocab();
    }

    /**
     * 从文件加载分词器
     */
    public Tokenizer(@NonNull String tokenizerPath) {
        try {
            loadTokenizer(tokenizerPath);
        } catch (Exception e) {
            Log.w(TAG, "Failed to load tokenizer, using default: " + e.getMessage());
            initDefaultVocab();
        }
    }

    /**
     * 编码文本为 token ID 序列
     */
    public int[] encode(@NonNull String text) {
        // 简单实现：基于空格和字符分割
        // 实际实现应使用 BPE 或 SentencePiece

        text = text.trim();
        if (text.isEmpty()) {
            return new int[]{bosTokenId};
        }

        // 预处理
        text = preprocess(text);

        // 分词
        String[] words = text.split("\\s+");
        int[] tokens = new int[words.length + 2];
        tokens[0] = bosTokenId;

        for (int i = 0; i < words.length; i++) {
            String word = words[i];
            Integer tokenId = encoder.get(word);
            if (tokenId != null) {
                tokens[i + 1] = tokenId;
            } else {
                // 未登录词，按字符分割
                tokens[i + 1] = encodeUnknown(word);
            }
        }

        tokens[tokens.length - 1] = eosTokenId;
        return tokens;
    }

    /**
     * 解码 token ID 序列为文本
     */
    public String decode(@NonNull int[] tokens) {
        StringBuilder sb = new StringBuilder();

        for (int tokenId : tokens) {
            if (tokenId == bosTokenId || tokenId == eosTokenId || tokenId == padTokenId) {
                continue;
            }

            String token = decoder.get(tokenId);
            if (token != null) {
                sb.append(token);
                if (!token.endsWith(" ") && !token.startsWith(" ")) {
                    sb.append(" ");
                }
            }
        }

        return postprocess(sb.toString());
    }

    /**
     * 获取词表大小
     */
    public int getVocabSize() {
        return vocabSize;
    }

    /**
     * 获取特殊 token ID
     */
    public int getPadTokenId() { return padTokenId; }
    public int getEosTokenId() { return eosTokenId; }
    public int getBosTokenId() { return bosTokenId; }

    // ===== Private Methods =====

    private void loadTokenizer(String path) throws Exception {
        try (FileInputStream fis = new FileInputStream(path);
             BufferedReader reader = new BufferedReader(new InputStreamReader(fis, StandardCharsets.UTF_8))) {

            StringBuilder sb = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                sb.append(line);
            }

            JSONObject config = new JSONObject(sb.toString());

            // 加载词表
            JSONObject vocab = config.optJSONObject("vocab");
            if (vocab != null) {
                JSONArray names = vocab.names();
                if (names != null) {
                    for (int i = 0; i < names.length(); i++) {
                        String token = names.getString(i);
                        int id = vocab.getInt(token);
                        encoder.put(token, id);
                        decoder.put(id, token);
                    }
                }
            }

            // 加载特殊 token
            JSONObject special = config.optJSONObject("special_tokens");
            if (special != null) {
                if (special.has("pad_token")) {
                    padTokenId = special.getInt("pad_token");
                }
                if (special.has("eos_token")) {
                    eosTokenId = special.getInt("eos_token");
                }
                if (special.has("bos_token")) {
                    bosTokenId = special.getInt("bos_token");
                }
            }

            vocabSize = encoder.size();
            Log.i(TAG, "Tokenizer loaded: vocab_size=" + vocabSize);
        }
    }

    private void initDefaultVocab() {
        // 初始化一个简单的默认词表
        // 用于演示和测试

        int id = 0;
        encoder.put("<pad>", id); decoder.put(id++, "<pad>");
        encoder.put("<s>", id); decoder.put(id++, "<s>");
        encoder.put("</s>", id); decoder.put(id++, "</s>");
        encoder.put("<unk>", id); decoder.put(id++, "<unk>");

        // 常用词
        String[] commonWords = {
                "the", "a", "an", "is", "are", "was", "were", "be", "been", "being",
                "have", "has", "had", "do", "does", "did", "will", "would", "could", "should",
                "i", "you", "he", "she", "it", "we", "they", "me", "him", "her",
                "this", "that", "these", "those", "what", "which", "who", "when", "where", "why",
                "hello", "world", "yes", "no", "ok", "please", "thank", "sorry",
                "get", "set", "run", "make", "take", "give", "find", "tell", "ask", "work",
                "time", "day", "year", "way", "thing", "man", "woman", "child", "world", "life"
        };

        for (String word : commonWords) {
            encoder.put(word, id);
            decoder.put(id, word);
            id++;
        }

        // 添加数字和标点
        for (int i = 0; i < 10; i++) {
            String num = String.valueOf(i);
            encoder.put(num, id);
            decoder.put(id, num);
            id++;
        }

        String[] puncts = {".", ",", "!", "?", ":", ";", "'", "\"", "-", "(", ")", "[", "]"};
        for (String p : puncts) {
            encoder.put(p, id);
            decoder.put(id, p);
            id++;
        }

        vocabSize = id;
        padTokenId = 0;
        bosTokenId = 1;
        eosTokenId = 2;
    }

    private String preprocess(String text) {
        // 文本预处理
        return text.toLowerCase().trim();
    }

    private String postprocess(String text) {
        // 文本后处理
        return text.replaceAll("\\s+", " ").trim();
    }

    private int encodeUnknown(String word) {
        // 处理未登录词
        // 简单实现：使用字符编码
        int hash = 0;
        for (char c : word.toCharArray()) {
            hash = hash * 31 + c;
        }
        return Math.abs(hash) % 10000 + 1000; // 映射到保留区
    }
}