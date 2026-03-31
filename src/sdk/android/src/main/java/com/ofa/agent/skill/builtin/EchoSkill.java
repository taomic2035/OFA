package com.ofa.agent.skill.builtin;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.skill.SkillExecutionException;

import java.nio.charset.StandardCharsets;
import java.util.Locale;

/**
 * Built-in echo skill for testing
 */
public class EchoSkill implements SkillExecutor {

    private static final String SKILL_ID = "echo";
    private final Gson gson = new Gson();

    @Override
    public String getSkillId() {
        return SKILL_ID;
    }

    @Override
    public String getSkillName() {
        return "Echo";
    }

    @Override
    public String getCategory() {
        return "utility";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            String inputStr = new String(input, StandardCharsets.UTF_8);

            JsonObject result = new JsonObject();
            result.addProperty("echo", inputStr);
            result.addProperty("length", input.length);

            return gson.toJson(result).getBytes(StandardCharsets.UTF_8);

        } catch (Exception e) {
            throw new SkillExecutionException(SKILL_ID, "Echo failed", e);
        }
    }
}