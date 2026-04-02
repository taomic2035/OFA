package com.ofa.agent.skill.builtin;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonElement;
import com.google.gson.JsonParser;
import com.google.gson.JsonObject;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.skill.SkillExecutionException;

import java.nio.charset.StandardCharsets;

/**
 * JSON 格式化技能 - JSON 美化和验证
 * 输入: JSON 字符串
 * 输出: 格式化的 JSON
 */
public class JSONFormatSkill implements SkillExecutor {

    private static final String SKILL_ID = "json.format";
    private final Gson prettyGson = new GsonBuilder().setPrettyPrinting().create();

    @Override
    public String getSkillId() {
        return SKILL_ID;
    }

    @Override
    public String getSkillName() {
        return "JSON Formatter";
    }

    @Override
    public String getCategory() {
        return "data";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            if (input == null || input.length == 0) {
                throw new SkillExecutionException(SKILL_ID, "Empty input");
            }

            String jsonStr = new String(input, StandardCharsets.UTF_8).trim();

            // 解析 JSON
            JsonElement element = JsonParser.parseString(jsonStr);

            JsonObject output = new JsonObject();

            // 格式化输出
            String formatted = prettyGson.toJson(element);
            output.addProperty("formatted", formatted);
            output.addProperty("valid", true);
            output.addProperty("type", getJsonType(element));
            output.addProperty("size", formatted.length());

            return output.toString().getBytes(StandardCharsets.UTF_8);

        } catch (Exception e) {
            JsonObject output = new JsonObject();
            output.addProperty("valid", false);
            output.addProperty("error", e.getMessage());
            return output.toString().getBytes(StandardCharsets.UTF_8);
        }
    }

    private String getJsonType(JsonElement element) {
        if (element.isJsonObject()) return "object";
        if (element.isJsonArray()) return "array";
        if (element.isJsonPrimitive()) {
            if (element.getAsJsonPrimitive().isString()) return "string";
            if (element.getAsJsonPrimitive().isNumber()) return "number";
            if (element.getAsJsonPrimitive().isBoolean()) return "boolean";
        }
        if (element.isJsonNull()) return "null";
        return "unknown";
    }
}