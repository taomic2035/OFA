package com.ofa.agent.skill.builtin;

import com.ofa.agent.offline.OfflineManager;

/**
 * 内置离线技能注册器
 * 将所有支持离线执行的技能注册到 OfflineManager
 */
public class OfflineSkills {

    /**
     * 注册所有内置离线技能
     * @param manager 离线管理器
     */
    public static void registerAll(OfflineManager manager) {
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