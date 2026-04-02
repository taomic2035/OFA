package com.ofa.agent.skill.builtin;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.skill.SkillExecutionException;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

/**
 * 哈希技能 - 计算字符串的哈希值
 * 输入格式: "algorithm:text" 或 "text" (默认 SHA-256)
 * 支持: md5, sha1, sha256, sha512
 */
public class HashSkill implements SkillExecutor {

    private static final String SKILL_ID = "hash";
    private final Gson gson = new Gson();

    @Override
    public String getSkillId() {
        return SKILL_ID;
    }

    @Override
    public String getSkillName() {
        return "Hash";
    }

    @Override
    public String getCategory() {
        return "crypto";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            if (input == null || input.length == 0) {
                throw new SkillExecutionException(SKILL_ID, "Empty input");
            }

            String inputStr = new String(input, StandardCharsets.UTF_8);
            String algorithm = "SHA-256";
            String text;

            // 解析算法: "md5:hello" 或 "hello"
            int colonIndex = inputStr.indexOf(':');
            if (colonIndex > 0) {
                String algo = inputStr.substring(0, colonIndex).toLowerCase();
                text = inputStr.substring(colonIndex + 1);

                switch (algo) {
                    case "md5":
                        algorithm = "MD5";
                        break;
                    case "sha1":
                        algorithm = "SHA-1";
                        break;
                    case "sha256":
                        algorithm = "SHA-256";
                        break;
                    case "sha512":
                        algorithm = "SHA-512";
                        break;
                    default:
                        text = inputStr; // 未知算法，使用整个输入
                }
            } else {
                text = inputStr;
            }

            // 计算哈希
            MessageDigest digest = MessageDigest.getInstance(algorithm);
            byte[] hashBytes = digest.digest(text.getBytes(StandardCharsets.UTF_8));
            String hexHash = bytesToHex(hashBytes);

            JsonObject output = new JsonObject();
            output.addProperty("algorithm", algorithm);
            output.addProperty("input", text.length() > 50 ? text.substring(0, 50) + "..." : text);
            output.addProperty("inputLength", text.length());
            output.addProperty("hash", hexHash);
            output.addProperty("hashLength", hexHash.length());

            return gson.toJson(output).getBytes(StandardCharsets.UTF_8);

        } catch (NoSuchAlgorithmException e) {
            throw new SkillExecutionException(SKILL_ID, "Algorithm not available", e);
        } catch (Exception e) {
            throw new SkillExecutionException(SKILL_ID, "Hash calculation failed", e);
        }
    }

    private String bytesToHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }
}