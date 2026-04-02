package com.ofa.agent.skill.builtin;

import androidx.annotation.NonNull;

import com.ofa.agent.offline.OfflineManager;
import com.ofa.agent.skill.SkillRegistry;

/**
 * 内置离线技能注册器
 * 将所有支持离线执行的技能注册到 OfflineManager 和 SkillRegistry
 */
public class OfflineSkills {

    /**
     * 注册所有内置离线技能到 SkillRegistry
     * @param registry 技能注册表
     */
    public static void registerAll(@NonNull SkillRegistry registry) {
        // 基础技能
        registry.register(new EchoSkill());
        registry.register(new TextProcessSkill());

        // 数学/数据处理
        registry.register(new CalculatorSkill());
        registry.register(new JSONFormatSkill());

        // 时间工具
        registry.register(new TimestampSkill());

        // 安全工具
        registry.register(new HashSkill());

        // 外卖技能
        FoodDeliverySkills.registerAll(registry);
    }

    /**
     * 注册所有内置离线技能到 OfflineManager
     * @param manager 离线管理器
     */
    public static void registerAll(@NonNull OfflineManager manager) {
        // 基础技能
        manager.registerSkill("echo", new EchoSkill(), true);
        manager.registerSkill("text.process", new TextProcessSkill(), true);

        // 数学/数据处理
        manager.registerSkill("calculator", new CalculatorSkill(), true);
        manager.registerSkill("json.format", new JSONFormatSkill(), true);

        // 时间工具
        manager.registerSkill("timestamp", new TimestampSkill(), true);

        // 安全工具
        manager.registerSkill("hash", new HashSkill(), true);
    }
}