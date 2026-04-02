package com.ofa.agent.sample;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.skill.CompositeSkillExecutor;
import com.ofa.agent.skill.SkillContext;
import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillRegistry;
import com.ofa.agent.skill.SkillResult;
import com.ofa.agent.skill.SkillStep;
import com.ofa.agent.tool.BuiltInTools;
import com.ofa.agent.tool.ToolRegistry;

import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * 技能系统使用示例
 * 展示如何创建和执行自定义技能
 */
public class SkillSample {

    private static final String TAG = "SkillSample";

    private final Context context;
    private final ToolRegistry toolRegistry;
    private final SkillRegistry skillRegistry;
    private final CompositeSkillExecutor executor;

    public SkillSample(@NonNull Context context) {
        this.context = context.getApplicationContext();

        // 初始化工具注册表
        this.toolRegistry = new ToolRegistry(context);
        BuiltInTools.registerAll(context, toolRegistry);

        // 初始化技能注册表
        this.skillRegistry = SkillRegistry.getInstance(context);

        // 初始化技能执行器
        this.executor = new CompositeSkillExecutor(context, toolRegistry);

        Log.i(TAG, "Skill system initialized");
    }

    /**
     * 示例1：创建简单的自定义技能
     */
    public void createSimpleSkill() {
        // 创建一个简单的"早安问候"技能
        SkillDefinition goodMorningSkill = new SkillDefinition.Builder()
                .id("custom.good_morning")
                .name("早安问候")
                .description("获取天气、播报日程、播放新闻")
                .category("routine")
                .tag("早晨")
                .tag("日常")

                // 步骤1：获取天气
                .step(new SkillStep.Builder()
                        .id("get_weather")
                        .name("获取天气")
                        .type(SkillStep.StepType.TOOL)
                        .action("weather.get")
                        .param("location", "current")
                        .nextStep("get_schedule")
                        .build())

                // 步骤2：获取日程
                .step(new SkillStep.Builder()
                        .id("get_schedule")
                        .name("获取今日日程")
                        .type(SkillStep.StepType.TOOL)
                        .action("calendar.today")
                        .nextStep("get_news")
                        .build())

                // 步骤3：播放新闻
                .step(new SkillStep.Builder()
                        .id("get_news")
                        .name("播报新闻")
                        .type(SkillStep.StepType.TOOL)
                        .action("news.play")
                        .param("category", "headline")
                        .param("count", 5)
                        .build())

                // 触发器
                .trigger("voice", "早安")
                .trigger("voice", "早上好")
                .trigger("schedule", "07:00")

                .estimatedTimeMs(60000)
                .build();

        // 保存到注册表
        skillRegistry.saveSkill(goodMorningSkill);
        Log.i(TAG, "Created skill: " + goodMorningSkill.getId());
    }

    /**
     * 示例2：创建带条件的技能
     */
    public void createConditionalSkill() {
        // 创建一个"出门前检查"技能
        SkillDefinition checkBeforeLeaveSkill = new SkillDefinition.Builder()
                .id("custom.check_before_leave")
                .name("出门前检查")
                .description("检查天气、交通，提醒带伞或穿雨衣")
                .category("routine")

                // 步骤1：获取天气
                .step(new SkillStep.Builder()
                        .id("get_weather")
                        .name("获取天气")
                        .type(SkillStep.StepType.TOOL)
                        .action("weather.get")
                        .nextStep("check_rain")
                        .build())

                // 步骤2：检查是否下雨
                .step(new SkillStep.Builder()
                        .id("check_rain")
                        .name("检查下雨")
                        .type(SkillStep.StepType.CONDITION)
                        .action("${weather}")
                        .branch("${weather == 'rain'}", "remind_umbrella")
                        .branch("${weather == 'sunny'}", "check_traffic")
                        .branch("${weather == 'cloudy'}", "check_traffic")
                        .build())

                // 步骤3a：提醒带伞
                .step(new SkillStep.Builder()
                        .id("remind_umbrella")
                        .name("提醒带伞")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "天气提醒")
                        .param("message", "今天有雨，记得带伞")
                        .nextStep("check_traffic")
                        .build())

                // 步骤4：检查交通
                .step(new SkillStep.Builder()
                        .id("check_traffic")
                        .name("检查交通")
                        .type(SkillStep.StepType.TOOL)
                        .action("traffic.getStatus")
                        .param("destination", "${destination}")
                        .build())

                .trigger("voice", "出门")
                .trigger("voice", "我要出门")

                .input("destination", "目的地")
                .estimatedTimeMs(30000)
                .build();

        skillRegistry.saveSkill(checkBeforeLeaveSkill);
        Log.i(TAG, "Created skill: " + checkBeforeLeaveSkill.getId());
    }

    /**
     * 示例3：创建带用户交互的技能
     */
    public void createInteractiveSkill() {
        // 创建一个"订餐"技能
        SkillDefinition orderFoodSkill = new SkillDefinition.Builder()
                .id("custom.order_food")
                .name("订餐")
                .description("帮助用户选择并订购午餐")
                .category("food")

                // 步骤1：询问想吃什么的类型
                .step(new SkillStep.Builder()
                        .id("ask_food_type")
                        .name("询问食物类型")
                        .type(SkillStep.StepType.INPUT)
                        .param("prompt", "今天想吃什么类型的？(中餐/西餐/日料/快餐)")
                        .param("variable", "foodType")
                        .timeout(60000)
                        .nextStep("recommend_restaurants")
                        .build())

                // 步骤2：推荐餐厅
                .step(new SkillStep.Builder()
                        .id("recommend_restaurants")
                        .name("推荐餐厅")
                        .type(SkillStep.StepType.TOOL)
                        .action("restaurant.search")
                        .param("type", "${foodType}")
                        .param("location", "nearby")
                        .nextStep("select_restaurant")
                        .build())

                // 步骤3：选择餐厅
                .step(new SkillStep.Builder()
                        .id("select_restaurant")
                        .name("选择餐厅")
                        .type(SkillStep.StepType.INPUT)
                        .param("prompt", "为您推荐了以下餐厅，请选择一家（说出序号）")
                        .param("variable", "selectedRestaurant")
                        .timeout(60000)
                        .nextStep("confirm_order")
                        .build())

                // 步骤4：确认订单
                .step(new SkillStep.Builder()
                        .id("confirm_order")
                        .name("确认订单")
                        .type(SkillStep.StepType.CONFIRM)
                        .param("message", "确认在${selectedRestaurant}下单？")
                        .timeout(60000)
                        .nextStep("submit_order")
                        .onError("cancel")
                        .build())

                // 步骤5：提交订单
                .step(new SkillStep.Builder()
                        .id("submit_order")
                        .name("提交订单")
                        .type(SkillStep.StepType.TOOL)
                        .action("order.submit")
                        .nextStep("notify_success")
                        .build())

                // 步骤6：通知成功
                .step(new SkillStep.Builder()
                        .id("notify_success")
                        .name("通知成功")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "下单成功")
                        .param("message", "订单已提交，预计30分钟送达")
                        .build())

                // 取消分支
                .step(new SkillStep.Builder()
                        .id("cancel")
                        .name("取消")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "已取消")
                        .param("message", "订单已取消")
                        .build())

                .trigger("voice", "订餐")
                .trigger("voice", "点外卖")
                .trigger("schedule", "11:30")

                .estimatedTimeMs(120000)
                .build();

        skillRegistry.saveSkill(orderFoodSkill);
        Log.i(TAG, "Created skill: " + orderFoodSkill.getId());
    }

    /**
     * 示例4：执行技能
     */
    public void executeSkill() {
        // 获取技能定义
        SkillDefinition skill = skillRegistry.getSkill("food_order.bubble_tea");
        if (skill == null) {
            Log.w(TAG, "Skill not found");
            return;
        }

        // 准备输入参数
        Map<String, Object> inputs = new java.util.HashMap<>();
        inputs.put("drinkName", "珍珠奶茶");
        inputs.put("sweetness", "五分糖");
        inputs.put("temperature", "少冰");
        inputs.put("size", "中杯");
        inputs.put("quantity", 1);
        inputs.put("app", "meituan");

        // 创建执行上下文
        SkillContext ctx = new SkillContext(skill.getId(), context);
        ctx.setCallback(new SkillContext.Callback() {
            @Override
            public void onStepStart(@NonNull String stepId, @NonNull SkillStep step) {
                Log.d(TAG, "Step started: " + step.getName());
            }

            @Override
            public void onStepComplete(@NonNull String stepId, @NonNull SkillContext.StepResult result) {
                Log.d(TAG, "Step completed: " + stepId + ", success=" + result.success);
            }

            @Override
            public void onStatusChange(@NonNull SkillContext.ExecutionStatus oldStatus,
                                        @NonNull SkillContext.ExecutionStatus newStatus) {
                Log.d(TAG, "Status changed: " + oldStatus + " -> " + newStatus);
            }

            @Override
            public void onProgress(int progress, String message) {
                Log.d(TAG, "Progress: " + progress + "% - " + message);
            }

            @Override
            public void onComplete(@NonNull SkillResult result) {
                if (result.isSuccess()) {
                    Log.i(TAG, "Skill completed successfully in " + result.getExecutionTimeMs() + "ms");
                } else {
                    Log.e(TAG, "Skill failed: " + result.getError());
                }
            }

            @Override
            public void onError(@NonNull String stepId, @NonNull String error) {
                Log.e(TAG, "Error at step " + stepId + ": " + error);
            }
        });

        // 设置用户交互处理器
        ctx.setInteractionHandler(new SkillContext.UserInteractionHandler() {
            @Override
            public void requestInput(@NonNull String prompt, @NonNull SkillContext.InputCallback callback) {
                // 在实际应用中，这里应该弹出输入对话框
                Log.i(TAG, "Request input: " + prompt);
                // 模拟用户输入
                callback.onInput("默认输入");
            }

            @Override
            public void requestConfirm(@NonNull String message, @NonNull SkillContext.ConfirmCallback callback) {
                // 在实际应用中，这里应该弹出确认对话框
                Log.i(TAG, "Request confirm: " + message);
                // 模拟用户确认
                callback.onConfirm(true);
            }

            @Override
            public void requestChoice(@NonNull String prompt, @NonNull String[] options,
                                       @NonNull SkillContext.ChoiceCallback callback) {
                // 在实际应用中，这里应该弹出选择对话框
                Log.i(TAG, "Request choice: " + prompt);
                // 模拟用户选择第一个
                callback.onChoice(0, options[0]);
            }
        });

        // 执行技能
        CompletableFuture<SkillResult> future = executor.execute(skill, inputs, ctx);

        // 等待完成（可选）
        future.thenAccept(result -> {
            Log.i(TAG, "Execution finished: " + result.toJson());
        });
    }

    /**
     * 示例5：通过语音触发技能
     */
    public void triggerByVoice(@NonNull String voiceInput) {
        // 根据语音匹配技能
        SkillDefinition matchedSkill = skillRegistry.matchTrigger(voiceInput);

        if (matchedSkill != null) {
            Log.i(TAG, "Matched skill: " + matchedSkill.getName());
            // 执行技能...
        } else {
            Log.w(TAG, "No skill matched for: " + voiceInput);
        }
    }

    /**
     * 示例6：列出所有技能
     */
    public void listSkills() {
        Log.i(TAG, "=== Registered Skills ===");
        for (SkillDefinition skill : skillRegistry.getAllSkills()) {
            Log.i(TAG, "- " + skill.getName() + " (" + skill.getId() + ")");
            Log.d(TAG, "  Category: " + skill.getCategory());
            Log.d(TAG, "  Steps: " + skill.getSteps().size());
            Log.d(TAG, "  Triggers: " + skill.getTriggers().size());
        }
        Log.i(TAG, "Total: " + skillRegistry.getSkillCount() + " skills");
    }

    /**
     * 示例7：搜索技能
     */
    public void searchSkills(@NonNull String query) {
        Log.i(TAG, "Searching for: " + query);
        for (SkillDefinition skill : skillRegistry.searchSkills(query)) {
            Log.i(TAG, "- " + skill.getName());
        }
    }

    /**
     * 运行演示
     */
    public void runDemo() {
        Log.i(TAG, "=== Skill System Demo ===");

        // 创建示例技能
        createSimpleSkill();
        createConditionalSkill();
        createInteractiveSkill();

        // 列出所有技能
        listSkills();

        // 搜索技能
        searchSkills("外卖");
    }

    /**
     * 关闭资源
     */
    public void shutdown() {
        executor.shutdown();
    }
}