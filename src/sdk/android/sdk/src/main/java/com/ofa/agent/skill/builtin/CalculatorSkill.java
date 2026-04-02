package com.ofa.agent.skill.builtin;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.ofa.agent.skill.SkillExecutor;
import com.ofa.agent.skill.SkillExecutionException;

import java.nio.charset.StandardCharsets;

/**
 * 计算器技能 - 支持基础数学运算
 * 输入格式: "a op b" 或 "op a"
 * 例如: "5 + 3", "sqrt 16", "10 * 2"
 */
public class CalculatorSkill implements SkillExecutor {

    private static final String SKILL_ID = "calculator";
    private final Gson gson = new Gson();

    @Override
    public String getSkillId() {
        return SKILL_ID;
    }

    @Override
    public String getSkillName() {
        return "Calculator";
    }

    @Override
    public String getCategory() {
        return "math";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            String expr = new String(input, StandardCharsets.UTF_8).trim();
            String[] parts = expr.split("\\s+");

            double result;

            if (parts.length == 2) {
                // 单操作数: "sqrt 16", "sin 30"
                String op = parts[0].toLowerCase();
                double a = Double.parseDouble(parts[1]);

                switch (op) {
                    case "sqrt":
                        result = Math.sqrt(a);
                        break;
                    case "sin":
                        result = Math.sin(Math.toRadians(a));
                        break;
                    case "cos":
                        result = Math.cos(Math.toRadians(a));
                        break;
                    case "tan":
                        result = Math.tan(Math.toRadians(a));
                        break;
                    case "abs":
                        result = Math.abs(a);
                        break;
                    case "log":
                        result = Math.log10(a);
                        break;
                    case "ln":
                        result = Math.log(a);
                        break;
                    case "floor":
                        result = Math.floor(a);
                        break;
                    case "ceil":
                        result = Math.ceil(a);
                        break;
                    case "round":
                        result = Math.round(a);
                        break;
                    default:
                        throw new SkillExecutionException(SKILL_ID, "Unknown operation: " + op);
                }

            } else if (parts.length == 3) {
                // 双操作数: "5 + 3", "10 * 2"
                double a = Double.parseDouble(parts[0]);
                String op = parts[1].toLowerCase();
                double b = Double.parseDouble(parts[2]);

                switch (op) {
                    case "add", "+":
                        result = a + b;
                        break;
                    case "sub", "-":
                        result = a - b;
                        break;
                    case "mul", "*", "x":
                        result = a * b;
                        break;
                    case "div", "/":
                        if (b == 0) {
                            throw new SkillExecutionException(SKILL_ID, "Division by zero");
                        }
                        result = a / b;
                        break;
                    case "mod", "%":
                        result = a % b;
                        break;
                    case "pow":
                        result = Math.pow(a, b);
                        break;
                    case "max":
                        result = Math.max(a, b);
                        break;
                    case "min":
                        result = Math.min(a, b);
                        break;
                    default:
                        throw new SkillExecutionException(SKILL_ID, "Unknown operation: " + op);
                }

            } else {
                throw new SkillExecutionException(SKILL_ID, "Invalid expression format. Use: 'a op b' or 'op a'");
            }

            JsonObject output = new JsonObject();
            output.addProperty("expression", expr);
            output.addProperty("result", result);

            return gson.toJson(output).getBytes(StandardCharsets.UTF_8);

        } catch (NumberFormatException e) {
            throw new SkillExecutionException(SKILL_ID, "Invalid number format", e);
        } catch (Exception e) {
            throw new SkillExecutionException(SKILL_ID, "Calculation failed", e);
        }
    }
}