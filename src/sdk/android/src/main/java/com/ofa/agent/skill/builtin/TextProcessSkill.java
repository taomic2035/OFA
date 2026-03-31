package com.ofa.agent.skill.builtin;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.skill.SkillExecutionException;

import java.nio.charset.StandardCharsets;
import java.util.Locale;

/**
 * Built-in text processing skill
 */
public class TextProcessSkill implements SkillExecutor {

    private static final String SKILL_ID = "text.process";
    private final Gson gson = new Gson();

    @Override
    public String getSkillId() {
        return SKILL_ID;
    }

    @Override
    public String getSkillName() {
        return "Text Process";
    }

    @Override
    public String getCategory() {
        return "text";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            String inputStr = new String(input, StandardCharsets.UTF_8);
            JsonObject request = gson.fromJson(inputStr, JsonObject.class);

            String text = request.has("text") ? request.get("text").getAsString() : "";
            String operation = request.has("operation") ? request.get("operation").getAsString() : "uppercase";

            String result;
            switch (operation.toLowerCase(Locale.ROOT)) {
                case "uppercase":
                    result = text.toUpperCase(Locale.ROOT);
                    break;
                case "lowercase":
                    result = text.toLowerCase(Locale.ROOT);
                    break;
                case "reverse":
                    result = new StringBuilder(text).reverse().toString();
                    break;
                case "length":
                    JsonObject lengthResult = new JsonObject();
                    lengthResult.addProperty("result", text.length());
                    return gson.toJson(lengthResult).getBytes(StandardCharsets.UTF_8);
                default:
                    throw new SkillExecutionException(SKILL_ID,
                            SkillExecutionException.ErrorCode.INVALID_INPUT,
                            "Unknown operation: " + operation);
            }

            JsonObject response = new JsonObject();
            response.addProperty("result", result);
            return gson.toJson(response).getBytes(StandardCharsets.UTF_8);

        } catch (SkillExecutionException e) {
            throw e;
        } catch (Exception e) {
            throw new SkillExecutionException(SKILL_ID, "Text processing failed", e);
        }
    }
}