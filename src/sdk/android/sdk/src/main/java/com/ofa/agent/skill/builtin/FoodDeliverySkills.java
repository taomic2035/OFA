package com.ofa.agent.skill.builtin;

import androidx.annotation.NonNull;

import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillRegistry;
import com.ofa.agent.skill.SkillStep;

/**
 * 外卖订购技能
 * 包含点奶茶等常见外卖场景
 */
public class FoodDeliverySkills {

    /**
     * 注册所有外卖技能
     */
    public static void registerAll(@NonNull SkillRegistry registry) {
        registerOrderBubbleTeaSkill(registry);
        registerTrackDeliverySkill(registry);
    }

    /**
     * 点奶茶技能
     *
     * 流程：
     * 1. 确认订单信息
     * 2. 打开外卖APP（美团/饿了么）
     * 3. 搜索奶茶店
     * 4. 选择商品和规格
     * 5. 确认地址
     * 6. 提交订单
     * 7. 支付
     * 8. 跟踪配送
     */
    @NonNull
    public static SkillDefinition createOrderBubbleTeaSkill() {
        return new SkillDefinition.Builder()
                .id("food_order.bubble_tea")
                .name("点奶茶")
                .description("通过外卖APP订购奶茶，支持选择甜度、糖度、小料等")
                .category("food_delivery")
                .version("1.0.0")
                .author("OFA")
                .tag("外卖")
                .tag("奶茶")
                .tag("自动化")

                // 输入参数
                .input("drinkName", "饮品名称，如：珍珠奶茶、杨枝甘露")
                .input(" sweetness", "甜度：全糖、七分糖、五分糖、三分糖、无糖")
                .input("temperature", "温度：热、温、常温、去冰、少冰、正常冰")
                .input("toppings", "小料列表，如：珍珠、椰果、布丁")
                .input("size", "杯型：小杯、中杯、大杯")
                .input("quantity", "数量，默认1")
                .input("shop", "店铺名称（可选）")
                .input("address", "收货地址（可选，使用默认地址）")
                .input("app", "外卖APP：meituan(美团) 或 eleme(饿了么)")

                // 触发器
                .trigger("intent", "点.*奶茶")
                .trigger("intent", "买.*奶茶")
                .trigger("intent", "要.*奶茶")

                // 步骤定义
                .step(new SkillStep.Builder()
                        .id("confirm_order")
                        .name("确认订单")
                        .type(SkillStep.StepType.CONFIRM)
                        .param("message", "确认订购：${quantity}杯 ${drinkName}，${sweetness}，${temperature}？")
                        .timeout(60000)
                        .nextStep("open_app")
                        .description("向用户确认订单信息")
                        .build())

                .step(new SkillStep.Builder()
                        .id("open_app")
                        .name("打开外卖APP")
                        .type(SkillStep.StepType.TOOL)
                        .action("app.launch")
                        .param("packageName", "${app == 'meituan' ? 'com.sankuai.meituan' : 'me.ele'}")
                        .timeout(10000)
                        .retry(2, 2000)
                        .nextStep("wait_app_load")
                        .description("启动外卖APP")
                        .build())

                .step(new SkillStep.Builder()
                        .id("wait_app_load")
                        .name("等待APP加载")
                        .type(SkillStep.StepType.DELAY)
                        .param("duration", 3000)
                        .nextStep("search_shop")
                        .description("等待APP启动完成")
                        .build())

                .step(new SkillStep.Builder()
                        .id("search_shop")
                        .name("搜索店铺")
                        .type(SkillStep.StepType.TOOL)
                        .action("ui.search")
                        .param("query", "${shop != null ? shop : drinkName}")
                        .timeout(10000)
                        .onError("manual_search")
                        .nextStep("select_shop")
                        .description("搜索奶茶店铺")
                        .build())

                .step(new SkillStep.Builder()
                        .id("select_shop")
                        .name("选择店铺")
                        .type(SkillStep.StepType.INPUT)
                        .param("prompt", "请选择一家店铺（说出序号或店铺名）")
                        .param("variable", "selectedShop")
                        .timeout(60000)
                        .nextStep("select_drink")
                        .optional(true)
                        .description("用户选择店铺")
                        .build())

                .step(new SkillStep.Builder()
                        .id("select_drink")
                        .name("选择饮品")
                        .type(SkillStep.StepType.TOOL)
                        .action("ui.click")
                        .param("target", "${drinkName}")
                        .timeout(10000)
                        .onError("manual_select")
                        .nextStep("select_options")
                        .description("选择要购买的饮品")
                        .build())

                .step(new SkillStep.Builder()
                        .id("select_options")
                        .name("选择规格")
                        .type(SkillStep.StepType.TOOL)
                        .action("ui.configure")
                        .param("sweetness", "${sweetness}")
                        .param("temperature", "${temperature}")
                        .param("size", "${size}")
                        .param("toppings", "${toppings}")
                        .param("quantity", "${quantity}")
                        .timeout(30000)
                        .onError("manual_configure")
                        .nextStep("add_to_cart")
                        .description("选择甜度、温度、杯型、小料等")
                        .build())

                .step(new SkillStep.Builder()
                        .id("add_to_cart")
                        .name("加入购物车")
                        .type(SkillStep.StepType.TOOL)
                        .action("ui.click")
                        .param("target", "加入购物车")
                        .timeout(10000)
                        .nextStep("check_address")
                        .description("将商品加入购物车")
                        .build())

                .step(new SkillStep.Builder()
                        .id("check_address")
                        .name("确认地址")
                        .type(SkillStep.StepType.CONDITION)
                        .action("${address != null}")
                        .branch("${address != null}", "set_address")
                        .branch("${address == null}", "go_to_checkout")
                        .description("检查是否需要设置地址")
                        .build())

                .step(new SkillStep.Builder()
                        .id("set_address")
                        .name("设置地址")
                        .type(SkillStep.StepType.TOOL)
                        .action("ui.setAddress")
                        .param("address", "${address}")
                        .timeout(15000)
                        .nextStep("go_to_checkout")
                        .description("设置收货地址")
                        .build())

                .step(new SkillStep.Builder()
                        .id("go_to_checkout")
                        .name("去结算")
                        .type(SkillStep.StepType.TOOL)
                        .action("ui.click")
                        .param("target", "去结算")
                        .timeout(10000)
                        .nextStep("confirm_payment")
                        .description("进入结算页面")
                        .build())

                .step(new SkillStep.Builder()
                        .id("confirm_payment")
                        .name("确认支付")
                        .type(SkillStep.StepType.CONFIRM)
                        .param("message", "订单已生成，确认支付？")
                        .timeout(60000)
                        .nextStep("pay")
                        .onError("cancel_order")
                        .description("确认支付")
                        .build())

                .step(new SkillStep.Builder()
                        .id("pay")
                        .name("支付")
                        .type(SkillStep.StepType.TOOL)
                        .action("payment.pay")
                        .timeout(120000)
                        .nextStep("wait_confirmation")
                        .description("完成支付")
                        .build())

                .step(new SkillStep.Builder()
                        .id("wait_confirmation")
                        .name("等待订单确认")
                        .type(SkillStep.StepType.WAIT_FOR)
                        .action("${orderConfirmed}")
                        .timeout(30000)
                        .nextStep("save_order_info")
                        .description("等待商家确认订单")
                        .build())

                .step(new SkillStep.Builder()
                        .id("save_order_info")
                        .name("保存订单信息")
                        .type(SkillStep.StepType.ASSIGN)
                        .param("orderId", "${lastOrderId}")
                        .param("estimatedTime", "${estimatedDeliveryTime}")
                        .nextStep("notify_success")
                        .description("保存订单ID和预计送达时间")
                        .build())

                .step(new SkillStep.Builder()
                        .id("notify_success")
                        .name("通知成功")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "下单成功")
                        .param("message", "您的${drinkName}已下单，预计${estimatedTime}送达")
                        .nextStep("start_tracking")
                        .description("发送成功通知")
                        .build())

                .step(new SkillStep.Builder()
                        .id("start_tracking")
                        .name("开始跟踪配送")
                        .type(SkillStep.StepType.SUB_SKILL)
                        .action("food_order.track_delivery")
                        .param("orderId", "${orderId}")
                        .optional(true)
                        .description("跟踪配送状态")
                        .build())

                // 错误处理分支
                .step(new SkillStep.Builder()
                        .id("manual_search")
                        .name("手动搜索")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "需要手动操作")
                        .param("message", "请手动搜索店铺并选择饮品")
                        .nextStep("select_options")
                        .description("提示用户手动操作")
                        .build())

                .step(new SkillStep.Builder()
                        .id("manual_select")
                        .name("手动选择")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "需要手动操作")
                        .param("message", "请手动选择饮品")
                        .nextStep("select_options")
                        .description("提示用户手动操作")
                        .build())

                .step(new SkillStep.Builder()
                        .id("manual_configure")
                        .name("手动配置")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "需要手动操作")
                        .param("message", "请手动选择规格")
                        .nextStep("add_to_cart")
                        .description("提示用户手动操作")
                        .build())

                .step(new SkillStep.Builder()
                        .id("cancel_order")
                        .name("取消订单")
                        .type(SkillStep.StepType.TOOL)
                        .action("order.cancel")
                        .optional(true)
                        .description("取消订单")
                        .build())

                .estimatedTimeMs(300000) // 预计5分钟
                .build();
    }

    /**
     * 跟踪配送技能
     */
    @NonNull
    public static SkillDefinition createTrackDeliverySkill() {
        return new SkillDefinition.Builder()
                .id("food_order.track_delivery")
                .name("跟踪配送")
                .description("跟踪外卖配送状态，快到时提醒用户")
                .category("food_delivery")
                .version("1.0.0")

                .input("orderId", "订单ID")

                // 步骤
                .step(new SkillStep.Builder()
                        .id("check_status")
                        .name("检查配送状态")
                        .type(SkillStep.StepType.TOOL)
                        .action("order.getStatus")
                        .param("orderId", "${orderId}")
                        .timeout(10000)
                        .nextStep("evaluate_status")
                        .description("查询订单配送状态")
                        .build())

                .step(new SkillStep.Builder()
                        .id("evaluate_status")
                        .name("判断状态")
                        .type(SkillStep.StepType.CONDITION)
                        .action("${status}")
                        .branch("${status == 'delivered'}", "notify_arrived")
                        .branch("${status == 'arriving'}", "notify_coming")
                        .branch("${status == 'preparing'}", "wait_and_check")
                        .branch("${status == 'picked_up'}", "notify_picked_up")
                        .description("根据状态决定下一步")
                        .build())

                .step(new SkillStep.Builder()
                        .id("notify_picked_up")
                        .name("通知已取餐")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "骑手已取餐")
                        .param("message", "您的订单正在配送中，预计${estimatedMinutes}分钟送达")
                        .nextStep("wait_and_check")
                        .description("通知骑手已取餐")
                        .build())

                .step(new SkillStep.Builder()
                        .id("notify_coming")
                        .name("通知即将送达")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "即将送达")
                        .param("message", "骑手距离约${distance}米，请准备取餐")
                        .param("priority", "high")
                        .nextStep("wait_arrival")
                        .description("通知即将送达")
                        .build())

                .step(new SkillStep.Builder()
                        .id("wait_arrival")
                        .name("等待送达")
                        .type(SkillStep.StepType.WAIT_FOR)
                        .action("${status == 'delivered'}")
                        .timeout(600000) // 10分钟
                        .nextStep("notify_arrived")
                        .description("等待送达")
                        .build())

                .step(new SkillStep.Builder()
                        .id("notify_arrived")
                        .name("通知已送达")
                        .type(SkillStep.StepType.NOTIFY)
                        .param("title", "订单已送达")
                        .param("message", "您的外卖已送达，请及时取餐")
                        .param("priority", "high")
                        .description("通知已送达")
                        .build())

                .step(new SkillStep.Builder()
                        .id("wait_and_check")
                        .name("等待后再次检查")
                        .type(SkillStep.StepType.DELAY)
                        .param("duration", 60000) // 1分钟
                        .nextStep("check_status")
                        .description("等待后再次检查状态")
                        .build())

                .estimatedTimeMs(1800000) // 预计30分钟
                .build();
    }

    /**
     * 注册点奶茶技能
     */
    public static void registerOrderBubbleTeaSkill(@NonNull SkillRegistry registry) {
        registry.register(createOrderBubbleTeaSkill());
    }

    /**
     * 注册配送跟踪技能
     */
    public static void registerTrackDeliverySkill(@NonNull SkillRegistry registry) {
        registry.register(createTrackDeliverySkill());
    }
}