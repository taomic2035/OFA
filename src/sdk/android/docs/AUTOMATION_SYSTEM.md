# OFA Android SDK - UI Automation System

OFA Android SDK 的 UI 自动化系统，基于 AccessibilityService 实现。

## 概述

UI 自动化系统提供了跨应用的 UI 操作能力，支持点击、滑动、输入、查找等操作。主要用于实现如"点奶茶"等复杂的自动化技能。

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                    AutomationManager                      │
│                    (统一管理器)                           │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  AutomationEngine                         │
│                  (引擎接口)                               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│               AccessibilityEngine                         │
│               (无障碍服务实现)                            │
├─────────────────┬─────────────────┬─────────────────────┤
│   NodeFinder    │ GesturePerformer│ OFAAccessibilitySvc │
│   (节点查找)     │  (手势执行)      │  (无障碍服务)        │
└─────────────────┴─────────────────┴─────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                      UITool                               │
│              (工具执行器 - ToolExecutor)                  │
└─────────────────────────────────────────────────────────┘
```

## 能力层级

| 层级 | 常量 | 说明 |
|------|------|------|
| 无能力 | NONE | 无障碍服务未启用 |
| 基础 | BASIC | 基础点击、查找操作 |
| 增强 | ENHANCED | 手势执行、滚动查找 |
| 完整 | FULL_ACCESSIBILITY | 完整无障碍能力 |
| 系统级 | SYSTEM_LEVEL | 需要Root/系统权限 |

## 快速开始

### 1. 配置 AndroidManifest

```xml
<manifest>
    <application>
        <!-- 无障碍服务 -->
        <service
            android:name="com.ofa.agent.automation.accessibility.OFAAccessibilityService"
            android:permission="android.permission.BIND_ACCESSIBILITY_SERVICE"
            android:exported="true">
            <intent-filter>
                <action android:name="android.accessibilityservice.AccessibilityService" />
            </intent-filter>
            <meta-data
                android:name="android.accessibilityservice"
                android:resource="@xml/accessibility_config" />
        </service>
    </application>
</manifest>
```

### 2. 初始化自动化管理器

```java
// 获取管理器实例
AutomationManager manager = AutomationManager.init(context);

// 配置
AutomationConfig config = AutomationConfig.builder()
    .clickTimeout(5000)
    .swipeTimeout(3000)
    .autoRetry(true)
    .maxRetries(3)
    .build();

// 设置监听器
manager.setListener(new AutomationListener() {
    @Override
    public void onEngineAvailable(AutomationCapability capability) {
        Log.i(TAG, "引擎就绪: " + capability.getDescription());
    }

    @Override
    public void onEngineUnavailable(String reason) {
        Log.w(TAG, "引擎不可用: " + reason);
    }

    // ... 其他回调
});

// 启动
manager.start(config);
```

### 3. 引导用户开启无障碍服务

```java
if (!manager.isAccessibilityServiceEnabled()) {
    // 打开无障碍设置页面
    manager.openAccessibilitySettings();
}
```

## 基本操作

### 点击

```java
AutomationEngine engine = manager.getEngine();

// 按坐标点击
AutomationResult result1 = engine.click(500, 300);

// 按文本点击
AutomationResult result2 = engine.click("确定");

// 按选择器点击
BySelector selector = BySelector.text("登录")
    .andClassName("android.widget.Button");
AutomationResult result3 = engine.click(selector);
```

### 长按

```java
// 按坐标长按
engine.longClick(500, 300);

// 按文本长按
engine.longClick("删除");
```

### 滑动

```java
// 按方向滑动
engine.swipe(Direction.DOWN, 0);  // 下滑
engine.swipe(Direction.UP, 0);    // 上滑
engine.swipe(Direction.LEFT, 0);  // 左滑
engine.swipe(Direction.RIGHT, 0); // 右滑

// 按坐标滑动
engine.swipe(500, 1000, 500, 500, 300); // 从(500,1000)滑到(500,500)，耗时300ms
```

### 输入

```java
// 输入文本（需先聚焦输入框）
engine.inputText("Hello World");

// 点击并输入
engine.inputText(500, 300, "Hello World");

// 按选择器输入
engine.inputText(BySelector.id("username"), "user123");
```

### 查找元素

```java
// 按文本查找
AutomationNode node = engine.findElement(BySelector.text("确定"));

// 按ID查找
AutomationNode node = engine.findElement(BySelector.id("com.app:id/button"));

// 按类名查找
AutomationNode node = engine.findElement(BySelector.className("android.widget.Button"));

// 组合条件
BySelector selector = BySelector.textContains("确认")
    .clickable(true)
    .enabled(true);
AutomationNode node = engine.findElement(selector);
```

### 等待

```java
// 等待元素出现
AutomationResult result = engine.waitForElement(
    BySelector.text("加载完成"),
    30000  // 超时30秒
);

// 等待页面稳定
engine.waitForPageStable(5000);
```

### 滚动查找

```java
// 滚动查找（最多滚动10次）
AutomationResult result = engine.scrollFind(
    BySelector.text("目标元素"),
    10  // 最大滚动次数
);
```

## BySelector 选择器

### 静态工厂方法

| 方法 | 说明 |
|------|------|
| `text(String)` | 精确匹配文本 |
| `textContains(String)` | 文本包含 |
| `textStartsWith(String)` | 文本开头 |
| `textEndsWith(String)` | 文本结尾 |
| `id(String)` | 资源ID |
| `className(String)` | 类名 |
| `desc(String)` | 内容描述 |
| `descContains(String)` | 描述包含 |
| `clickable()` | 可点击元素 |
| `scrollable()` | 可滚动元素 |

### 链式调用

```java
BySelector selector = BySelector.text("登录")
    .andClassName("android.widget.Button")
    .clickable(true)
    .enabled(true);
```

## UITool 工具

自动化操作已集成为 ToolExecutor，可通过工具系统调用：

| 工具名 | 参数 | 说明 |
|--------|------|------|
| ui.click | x, y 或 text | 点击 |
| ui.longClick | x, y 或 text | 长按 |
| ui.swipe | direction 或 fromX,fromY,toX,toY | 滑动 |
| ui.input | text, x, y(可选) | 输入 |
| ui.find | text | 查找 |
| ui.wait | text, timeout | 等待 |
| ui.scrollFind | text, maxScrolls | 滚动查找 |

### 示例：通过工具系统调用

```java
Map<String, Object> args = new HashMap<>();
args.put("operation", "click");
args.put("text", "确定");

ToolResult result = uiTool.execute(args, context);
```

## 自动化监听器

```java
public interface AutomationListener {
    // 引擎状态
    void onEngineAvailable(AutomationCapability capability);
    void onEngineUnavailable(String reason);

    // 操作回调
    void onOperationStart(String operation, String target);
    void onOperationComplete(String operation, AutomationResult result);
    void onOperationError(String operation, String error, boolean willRetry);

    // 手势回调
    void onGesturePerformed(String gestureType, int x, int y);

    // 元素查找
    void onElementFound(BySelector selector, AutomationNode node);
    void onElementNotFound(BySelector selector, boolean timedOut);

    // 页面变化
    void onPageChange(String packageName, String activityName);

    // 服务状态
    void onAccessibilityServiceStateChanged(boolean enabled);

    // 截图
    void onScreenshotCaptured(String screenshotPath);
}
```

## 配置选项

```java
AutomationConfig config = AutomationConfig.builder()
    // 超时设置
    .clickTimeout(5000)      // 点击超时
    .swipeTimeout(3000)      // 滑动超时
    .inputTimeout(10000)     // 输入超时
    .waitTimeout(30000)      // 等待超时
    .pageStableTimeout(5000) // 页面稳定超时

    // 手势参数
    .clickDuration(100)       // 点击持续时间
    .longClickDuration(500)   // 长按持续时间
    .swipeDuration(300)       // 滑动持续时间
    .scrollDistance(500)      // 滚动距离

    // 行为设置
    .autoRetry(true)         // 自动重试
    .maxRetries(3)           // 最大重试次数
    .retryDelay(1000)        // 重试间隔
    .enableLogging(true)     // 启用日志
    .build();
```

## 注意事项

### 权限要求

- 用户必须在系统设置中启用无障碍服务
- 路径：设置 → 无障碍 → OFA UI Automation Service

### 最佳实践

1. **检查服务状态**：操作前检查 `manager.isAvailable()`
2. **处理超时**：设置合理的超时时间
3. **错误处理**：实现 `onOperationError` 回调
4. **引导用户**：当服务不可用时引导用户开启

### 已知限制

- 截图功能需要 Android 9+ 或 MediaProjection API
- 某些 ROM 可能限制后台无障碍服务
- 部分 APP 可能阻止无障碍访问

## 与 Skill 系统集成

自动化系统与 Skill 系统无缝集成：

```java
// 在 SkillDefinition 中使用 UI 工具
.step(new SkillStep.Builder()
    .id("click_login")
    .name("点击登录")
    .type(SkillStep.StepType.TOOL)
    .action("ui.click")
    .param("text", "登录")
    .timeout(10000)
    .build())
```

## 示例：点奶茶流程

```java
// 1. 启动APP
engine.click(BySelector.text("美团"));

// 2. 等待加载
engine.waitForPageStable(3000);

// 3. 搜索
engine.inputText(BySelector.id("search_input"), "奶茶");
engine.click("搜索");

// 4. 选择店铺
engine.click(BySelector.textContains("喜茶"));

// 5. 选择商品
engine.scrollFind("多肉葡萄", 5);
engine.click("多肉葡萄");

// 6. 选择规格
engine.click(BySelector.textContains("七分糖"));
engine.click(BySelector.textContains("少冰"));
engine.click("加入购物车");

// 7. 结算
engine.click("去结算");
```

## Phase 2: 高级功能

### ScrollHelper - 滚动辅助

```java
import com.ofa.agent.automation.advanced.ScrollHelper;

ScrollHelper scrollHelper = new ScrollHelper(engine);

// 滚动查找元素
AutomationResult result = scrollHelper.scrollFind(
    BySelector.text("目标元素"),
    Direction.DOWN  // 向下滚动查找
);

// 滚动到顶部
scrollHelper.scrollToTop();

// 滚动到底部
scrollHelper.scrollToBottom();

// 下拉刷新
scrollHelper.pullToRefresh();

// 检查是否可滚动
boolean canScroll = scrollHelper.canScroll(Direction.DOWN);
```

### PageMonitor - 页面监控

```java
import com.ofa.agent.automation.advanced.PageMonitor;

PageMonitor pageMonitor = new PageMonitor(engine);

// 添加监听器
pageMonitor.addListener(new PageMonitor.PageChangeListener() {
    @Override
    public void onPageChanged(String packageName, String activityName) {
        Log.i(TAG, "页面变化: " + packageName);
    }

    @Override
    public void onPageStable() {
        Log.i(TAG, "页面稳定");
    }

    @Override
    public void onPackageChanged(String oldPackage, String newPackage) {
        Log.i(TAG, "应用切换: " + oldPackage + " -> " + newPackage);
    }
});

// 开始监控
pageMonitor.startMonitoring();

// 等待页面稳定
boolean stable = pageMonitor.waitForStable(5000);

// 等待页面变化
boolean changed = pageMonitor.waitForChange(3000);

// 等待特定应用
boolean found = pageMonitor.waitForPackage("com.example.app", 10000);

// 停止监控
pageMonitor.stopMonitoring();
```

### ScreenCapture - 屏幕截图

```java
import com.ofa.agent.automation.advanced.ScreenCapture;

ScreenCapture screenCapture = new ScreenCapture(context);

// 请求权限（在 Activity 中）
screenCapture.requestPermission(activity, REQUEST_CODE);

// 在 onActivityResult 中初始化
@Override
protected void onActivityResult(int requestCode, int resultCode, Intent data) {
    if (requestCode == REQUEST_CODE) {
        screenCapture.initialize(resultCode, data);
    }
}

// 截图到 Bitmap
Bitmap bitmap = screenCapture.captureBitmap();

// 截图到文件
String path = screenCapture.captureToFile(new File("/sdcard/screenshots"));

// 区域截图
Bitmap region = screenCapture.captureRegion(100, 100, 500, 500);

// 比较两张图片
float similarity = ScreenCapture.compareBitmaps(bitmap1, bitmap2);
```

### ActionRecorder - 操作录制

```java
import com.ofa.agent.automation.advanced.ActionRecorder;

ActionRecorder recorder = new ActionRecorder(engine, screenCapture);

// 开始录制
recorder.startRecording("test_recording");

// 暂停/恢复
recorder.pauseRecording();
recorder.resumeRecording();

// 停止录制
List<ActionRecorder.RecordedAction> actions = recorder.stopRecording();

// 保存录制
String savedPath = recorder.saveRecording();

// 加载录制
List<ActionRecorder.RecordedAction> loaded = ActionRecorder.loadRecording(new File(path));
```

### ActionReplay - 操作回放

```java
import com.ofa.agent.automation.advanced.ActionReplay;

ActionReplay replay = new ActionReplay(engine, screenCapture);

// 设置回放参数
replay.setPlaybackDelay(500);      // 动作间隔
replay.setSpeedMultiplier(2.0f);   // 2倍速
replay.setStopOnError(true);       // 出错停止

// 设置监听器
replay.setPlaybackListener(new ActionReplay.PlaybackListener() {
    @Override
    public void onPlaybackStart(int totalActions) {}

    @Override
    public void onActionStart(int index, ActionRecorder.RecordedAction action) {}

    @Override
    public void onActionComplete(int index, ActionRecorder.RecordedAction action,
                                  AutomationResult result) {}

    @Override
    public void onPlaybackComplete(int successCount, int failCount) {}

    @Override
    public void onPlaybackError(int index, String error) {}
});

// 回放录制
ActionReplay.PlaybackResult result = replay.play(actions);

// 从文件回放
ActionReplay.PlaybackResult result = replay.playFromFile(new File(path), true);

// 停止回放
replay.stop();
```

### Phase 2 新增工具

| 工具 | 参数 | 说明 |
|------|------|------|
| ui.pullToRefresh | - | 下拉刷新 |
| ui.scrollToTop | - | 滚动到顶部 |
| ui.scrollToBottom | - | 滚动到底部 |
| ui.capture | savePath (可选) | 截图 |
| ui.waitForStable | timeout (可选) | 等待页面稳定 |
| ui.startRecord | name (可选) | 开始录制 |
| ui.stopRecord | - | 停止录制 |
| ui.replay | file, respectTiming (可选) | 回放操作 |
| ui.findText | text | OCR 文字查找 |

### 使用 Phase 2 工具

```java
Map<String, Object> args = new HashMap<>();
args.put("operation", "pullToRefresh");
ToolResult result = uiTool.execute(args, context);

args.clear();
args.put("operation", "capture");
args.put("savePath", "/sdcard/test.png");
result = uiTool.execute(args, context);

args.clear();
args.put("operation", "startRecord");
args.put("name", "my_recording");
result = uiTool.execute(args, context);

// ... 执行一些操作 ...

args.clear();
args.put("operation", "stopRecord");
result = uiTool.execute(args, context);

// 回放
args.clear();
args.put("operation", "replay");
args.put("file", "/sdcard/ofa/recordings/my_recording.json");
args.put("respectTiming", true);
result = uiTool.execute(args, context);
```

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0.5 | 2026-04-02 | Phase 2: 滚动辅助、页面监控、截图、录制回放 |
| 1.0.4 | 2026-04-02 | Phase 1: 基础 UI 自动化 |

---

*文档更新: 2026-04-02*