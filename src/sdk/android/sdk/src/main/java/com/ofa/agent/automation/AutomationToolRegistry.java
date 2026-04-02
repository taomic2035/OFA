package com.ofa.agent.automation;

import android.content.Context;

import androidx.annotation.NonNull;

import com.ofa.agent.tool.ToolRegistry;

/**
 * Automation Tool Registry - 注册所有自动化相关工具
 */
public class AutomationToolRegistry {

    /**
     * 注册所有自动化工具到 ToolRegistry
     *
     * @param registry 工具注册中心
     * @param context Android Context
     */
    public static void registerAll(@NonNull ToolRegistry registry, @NonNull Context context) {
        // UI 自动化工具
        registry.register(new UITool(context));

        // 支付工具
        registry.register(new PaymentTool(context));

        // 订单工具
        registry.register(new OrderTool(context));

        // 系统级工具 (如果可用)
        SystemTool systemTool = new SystemTool(context);
        if (systemTool.isAvailable()) {
            registry.register(systemTool);
        }
    }

    /**
     * 注册基础 UI 工具
     */
    public static void registerUITools(@NonNull ToolRegistry registry, @NonNull Context context) {
        registry.register(new UITool(context));
    }

    /**
     * 注册支付相关工具
     */
    public static void registerPaymentTools(@NonNull ToolRegistry registry, @NonNull Context context) {
        registry.register(new PaymentTool(context));
    }

    /**
     * 注册订单相关工具
     */
    public static void registerOrderTools(@NonNull ToolRegistry registry, @NonNull Context context) {
        registry.register(new OrderTool(context));
    }
}