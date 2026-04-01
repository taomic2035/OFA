package com.ofa.agent.skill.builtin;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.skill.SkillExecutionException;

import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.Locale;
import java.util.TimeZone;

/**
 * 时间戳技能 - 时间格式化和转换
 * 输入格式: "now", "format", "timestamp <ms>", "utc"
 */
public class TimestampSkill implements SkillExecutor {

    private static final String SKILL_ID = "timestamp";
    private final Gson gson = new Gson();

    @Override
    public String getSkillId() {
        return SKILL_ID;
    }

    @Override
    public String getSkillName() {
        return "Timestamp";
    }

    @Override
    public String getCategory() {
        return "time";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            String cmd = input != null && input.length > 0
                    ? new String(input, StandardCharsets.UTF_8).trim().toLowerCase()
                    : "now";

            Date now = new Date();
            SimpleDateFormat formatter = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault());
            JsonObject output = new JsonObject();

            switch (cmd) {
                case "now":
                    output.addProperty("timestamp", now.getTime());
                    output.addProperty("formatted", formatter.format(now));
                    output.addProperty("iso8601", formatISO8601(now));
                    break;

                case "utc":
                    SimpleDateFormat utcFormatter = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.US);
                    utcFormatter.setTimeZone(TimeZone.getTimeZone("UTC"));
                    output.addProperty("timestamp", now.getTime());
                    output.addProperty("utc", utcFormatter.format(now));
                    break;

                case "format":
                    formatter = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss.SSS Z", Locale.getDefault());
                    output.addProperty("formatted", formatter.format(now));
                    break;

                default:
                    // 尝试解析时间戳
                    if (cmd.startsWith("timestamp ")) {
                        try {
                            long ts = Long.parseLong(cmd.substring(10).trim());
                            Date date = new Date(ts);
                            output.addProperty("input", ts);
                            output.addProperty("formatted", formatter.format(date));
                        } catch (NumberFormatException e) {
                            output.addProperty("error", "Invalid timestamp");
                        }
                    } else {
                        // 默认返回当前时间
                        output.addProperty("timestamp", now.getTime());
                        output.addProperty("formatted", formatter.format(now));
                    }
            }

            output.addProperty("timezone", TimeZone.getDefault().getID());
            output.addProperty("command", cmd);

            return gson.toJson(output).getBytes(StandardCharsets.UTF_8);

        } catch (Exception e) {
            throw new SkillExecutionException(SKILL_ID, "Timestamp operation failed", e);
        }
    }

    private String formatISO8601(Date date) {
        SimpleDateFormat iso = new SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", Locale.US);
        iso.setTimeZone(TimeZone.getTimeZone("UTC"));
        return iso.format(date);
    }
}