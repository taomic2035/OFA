package com.ofa.agent.intent;

import androidx.annotation.NonNull;

import java.util.ArrayList;
import java.util.List;
import java.util.regex.Pattern;

/**
 * 意图注册表
 * 提供预定义的常用意图
 */
public class IntentRegistry {

    private final IntentEngine engine;

    public IntentRegistry(@NonNull IntentEngine engine) {
        this.engine = engine;
    }

    /**
     * 注册所有预定义意图
     */
    public void registerDefaultIntents() {
        registerSystemIntents();
        registerCommunicationIntents();
        registerMediaIntents();
        registerDeviceIntents();
        registerNavigationIntents();
        registerAppIntents();
    }

    /**
     * 系统控制意图
     */
    public void registerSystemIntents() {
        // 打开设置
        engine.register(new IntentDefinition.Builder()
                .id("system.open_settings")
                .category("system")
                .action("open_settings")
                .description("打开系统设置")
                .keywords("设置", "setting", "preferences", "选项", "配置")
                .pattern("打开.*设置|设置.*打开|进入.*设置")
                .defaultConfidence(0.8)
                .build());

        // 关闭应用
        engine.register(new IntentDefinition.Builder()
                .id("system.close_app")
                .category("system")
                .action("close_app")
                .description("关闭应用")
                .keywords("关闭", "退出", "close", "exit", "quit", "结束")
                .pattern("关闭.*应用|退出.*程序|结束.*运行")
                .slot("app_name", "string", "应用名称", false)
                .defaultConfidence(0.75)
                .build());

        // 重启
        engine.register(new IntentDefinition.Builder()
                .id("system.restart")
                .category("system")
                .action("restart")
                .description("重启设备或应用")
                .keywords("重启", "restart", "重新启动")
                .pattern("重启|重新启动")
                .slot("target", "string", "重启目标", false)
                .defaultConfidence(0.8)
                .build());

        // 音量控制
        engine.register(new IntentDefinition.Builder()
                .id("system.volume")
                .category("system")
                .action("volume")
                .description("调节音量")
                .keywords("音量", "声音", "volume", "大声", "小声", "静音")
                .pattern("(调|设置|增大|减小).*音量|音量.*(大|小|高|低)|静音|取消静音")
                .slotWithPattern("action", "string", "(增大|调大|大|高|up)", false)
                .slotWithPattern("action", "string", "(减小|调小|小|低|down)", false)
                .slotWithPattern("level", "number", "(\\d+)", false)
                .defaultConfidence(0.8)
                .build());
    }

    /**
     * 通讯意图
     */
    public void registerCommunicationIntents() {
        // 打电话
        engine.register(new IntentDefinition.Builder()
                .id("communication.call")
                .category("communication")
                .action("call")
                .description("拨打电话")
                .keywords("电话", "拨打", "call", "打给", "联系")
                .pattern("打.*电话|拨打|给.*打电话|联系.*")
                .slotWithPattern("contact", "string", "(给|打给|联系)\\s*([\\w\\s]+)", true)
                .slotWithPattern("phone_number", "string", "(\\d{11}|\\d{3}-\\d{8})", false)
                .confirmationTemplate("确认拨打 {contact} 的电话?")
                .defaultConfidence(0.85)
                .build());

        // 发短信
        engine.register(new IntentDefinition.Builder()
                .id("communication.sms")
                .category("communication")
                .action("sms")
                .description("发送短信")
                .keywords("短信", "消息", "sms", "text", "发送")
                .pattern("发.*短信|发送.*消息|给.*发消息")
                .slotWithPattern("contact", "string", "给\\s*([\\w\\s]+)", true)
                .slotWithPattern("content", "string", "内容[是为]\\s*(.+)", true)
                .confirmationTemplate("确认发送短信给 {contact}：{content}")
                .defaultConfidence(0.85)
                .build());

        // 发邮件
        engine.register(new IntentDefinition.Builder()
                .id("communication.email")
                .category("communication")
                .action("email")
                .description("发送邮件")
                .keywords("邮件", "email", "邮箱", "发送")
                .pattern("发.*邮件|发送.*邮箱|写.*邮件")
                .slotWithPattern("recipient", "string", "([\\w.-]+@[\\w.-]+)", false)
                .slotWithPattern("subject", "string", "主题[是为]\\s*(.+)", false)
                .slotWithPattern("content", "string", "内容[是为]\\s*(.+)", false)
                .defaultConfidence(0.8)
                .build());
    }

    /**
     * 媒体意图
     */
    public void registerMediaIntents() {
        // 拍照
        engine.register(new IntentDefinition.Builder()
                .id("media.capture")
                .category("media")
                .action("capture")
                .description("拍照或录像")
                .keywords("拍照", "照片", "相机", "camera", "录像", "视频")
                .pattern("拍照|拍.*照片|打开.*相机|录像|录.*视频")
                .slot("type", "string", "类型(photo/video)", false)
                .defaultConfidence(0.9)
                .build());

        // 播放音乐
        engine.register(new IntentDefinition.Builder()
                .id("media.play_music")
                .category("media")
                .action("play_music")
                .description("播放音乐")
                .keywords("播放", "音乐", "play", "music", "歌曲", "歌")
                .pattern("播放.*音乐|播放.*歌|听.*歌|放.*音乐")
                .slotWithPattern("song", "string", "播放\\s*([\\w\\s]+)", false)
                .slotWithPattern("artist", "string", "歌手[是为]\\s*(.+)", false)
                .defaultConfidence(0.85)
                .build());

        // 停止播放
        engine.register(new IntentDefinition.Builder()
                .id("media.stop")
                .category("media")
                .action("stop")
                .description("停止播放")
                .keywords("停止", "暂停", "stop", "pause", "结束")
                .pattern("停止.*播放|暂停|别.*播放|关掉.*音乐")
                .defaultConfidence(0.8)
                .build());

        // 查看图片
        engine.register(new IntentDefinition.Builder()
                .id("media.view_image")
                .category("media")
                .action("view_image")
                .description("查看图片")
                .keywords("图片", "照片", "image", "photo", "查看", "看")
                .pattern("查看.*图片|看.*照片|打开.*相册|显示.*图片")
                .slotWithPattern("album", "string", "相册\\s*([\\w\\s]+)", false)
                .defaultConfidence(0.8)
                .build());
    }

    /**
     * 设备控制意图
     */
    public void registerDeviceIntents() {
        // 打开WiFi
        engine.register(new IntentDefinition.Builder()
                .id("device.wifi_on")
                .category("device")
                .action("wifi_on")
                .description("打开WiFi")
                .keywords("wifi", "无线", "网络", "打开")
                .pattern("打开.*wifi|开启.*无线|wifi.*打开")
                .defaultConfidence(0.9)
                .build());

        // 关闭WiFi
        engine.register(new IntentDefinition.Builder()
                .id("device.wifi_off")
                .category("device")
                .action("wifi_off")
                .description("关闭WiFi")
                .keywords("wifi", "无线", "网络", "关闭")
                .pattern("关闭.*wifi|关掉.*无线|wifi.*关闭")
                .defaultConfidence(0.9)
                .build());

        // 打开蓝牙
        engine.register(new IntentDefinition.Builder()
                .id("device.bluetooth_on")
                .category("device")
                .action("bluetooth_on")
                .description("打开蓝牙")
                .keywords("蓝牙", "bluetooth", "打开")
                .pattern("打开.*蓝牙|开启.*蓝牙|蓝牙.*打开")
                .defaultConfidence(0.9)
                .build());

        // 调节亮度
        engine.register(new IntentDefinition.Builder()
                .id("device.brightness")
                .category("device")
                .action("brightness")
                .description("调节屏幕亮度")
                .keywords("亮度", "brightness", "屏幕", "调亮", "调暗")
                .pattern("(调|设置).*亮度|亮度.*(亮|暗|高|低)")
                .slotWithPattern("level", "number", "(\\d+)", false)
                .slotWithPattern("direction", "string", "(亮|高|up|增加)", false)
                .defaultConfidence(0.8)
                .build());

        // 检查电池
        engine.register(new IntentDefinition.Builder()
                .id("device.battery")
                .category("device")
                .action("battery")
                .description("检查电池状态")
                .keywords("电池", "battery", "电量", "剩余", "检查")
                .pattern("电池.*多少|电量.*剩余|检查.*电池|剩余.*电量")
                .defaultConfidence(0.85)
                .build());
    }

    /**
     * 导航意图
     */
    public void registerNavigationIntents() {
        // 导航到某地
        engine.register(new IntentDefinition.Builder()
                .id("navigation.navigate")
                .category("navigation")
                .action("navigate")
                .description("导航到目的地")
                .keywords("导航", "去", "到", "navigate", "地图", "路线")
                .pattern("导航.*到|去.*怎么走|带.*去|路线.*到")
                .slotWithPattern("destination", "string", "(到|去)\\s*([\\w\\s]+)", true)
                .slotWithPattern("origin", "string", "从\\s*([\\w\\s]+)", false)
                .confirmationTemplate("确认导航到 {destination}?")
                .defaultConfidence(0.85)
                .build());

        // 查找位置
        engine.register(new IntentDefinition.Builder()
                .id("navigation.search_location")
                .category("navigation")
                .action("search_location")
                .description("搜索位置")
                .keywords("位置", "在哪", "where", "找", "搜索", "locate")
                .pattern(".*在哪|找.*位置|搜索.*地点|定位.*")
                .slotWithPattern("query", "string", "([\\w\\s]+)\\s*在哪", true)
                .defaultConfidence(0.8)
                .build());

        // 当前位置
        engine.register(new IntentDefinition.Builder()
                .id("navigation.current_location")
                .category("navigation")
                .action("current_location")
                .description("获取当前位置")
                .keywords("当前位置", "我在哪", "where", "位置", "定位")
                .pattern("我在哪|当前位置|我的位置|定位")
                .defaultConfidence(0.85)
                .build());
    }

    /**
     * 应用意图
     */
    public void registerAppIntents() {
        // 打开应用
        engine.register(new IntentDefinition.Builder()
                .id("app.open")
                .category("app")
                .action("open")
                .description("打开应用")
                .keywords("打开", "open", "启动", "运行", "app")
                .pattern("打开.*应用|启动.*|运行.*")
                .slotWithPattern("app_name", "string", "(打开|启动|运行)\\s*([\\w\\s]+)", true)
                .defaultConfidence(0.85)
                .build());

        // 搜索
        engine.register(new IntentDefinition.Builder()
                .id("app.search")
                .category("app")
                .action("search")
                .description("搜索内容")
                .keywords("搜索", "search", "查找", "找", "查询")
                .pattern("搜索.*|查找.*|找.*")
                .slotWithPattern("query", "string", "(搜索|查找)\\s*([\\w\\s]+)", true)
                .slotWithPattern("scope", "string", "在\\s*([\\w\\s]+)\\s*中", false)
                .defaultConfidence(0.8)
                .build());

        // 分享
        engine.register(new IntentDefinition.Builder()
                .id("app.share")
                .category("app")
                .action("share")
                .description("分享内容")
                .keywords("分享", "share", "发送", "转发")
                .pattern("分享.*|转发.*")
                .slotWithPattern("content", "string", "(分享|转发)\\s*([\\w\\s]+)", false)
                .slotWithPattern("target", "string", "到\\s*([\\w\\s]+)", false)
                .defaultConfidence(0.8)
                .build());
    }

    /**
     * 注册自定义意图
     */
    public void registerCustom(@NonNull IntentDefinition definition) {
        engine.register(definition);
    }

    /**
     * 批量注册自定义意图
     */
    public void registerCustomAll(@NonNull List<IntentDefinition> definitions) {
        engine.registerAll(definitions);
    }

    /**
     * 获取意图引擎
     */
    @NonNull
    public IntentEngine getEngine() {
        return engine;
    }
}