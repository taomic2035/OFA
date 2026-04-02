# OFA 更新日志

所有重要的变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [1.1.0] - 2026-04-02 💬 Social Notification System

### 新增 - Android SDK Social Notification System

智能社交通知系统，根据现代社交习惯自动选择最佳沟通渠道：

| 组件 | 功能 |
|------|------|
| MessageClassifier | 消息分类器 (10种类型) |
| ChannelSelector | 渠道选择器 (9种渠道) |
| MessageSender | 消息发送器 (多渠道支持) |
| SocialOrchestrator | 社交编排器 (统一入口) |
| ContactAdapter | 联系人适配器 (通讯录集成) |
| SocialTool | 社交工具集 (10个MCP工具) |

**MessageClassifier 消息类型：**
| 类型 | 示例 | 自动选择渠道 |
|------|------|--------------|
| `invitation` | 约你吃饭 | 微信 |
| `urgent` | 服务器宕机！ | 电话 |
| `reminder` | 记得缴费 | 短信 |
| `guide` | 旅游攻略 | 小红书私信 |
| `payment` | 还我50块 | 支付宝 |
| `business` | 明天开会 | 钉钉 |
| `casual` | 好久不见 | 微信 |
| `greeting` | 你好 | 微信 |
| `location` | 我在咖啡厅 | 微信 |

**ChannelSelector 支持渠道：**
| 渠道 | 包名 | 能力 |
|------|------|------|
| 微信 | com.tencent.mm | 文本/语音/图片/视频/位置/支付/群聊 |
| 电话 | built-in | 语音 |
| 短信 | built-in | 文本/群发 |
| 支付宝 | com.eg.android.AlipayGphone | 文本/图片/支付 |
| 抖音 | com.ss.android.ugc.aweme | 文本/图片/视频 |
| 小红书 | com.xingin.xhs | 文本/图片/视频 |
| 钉钉 | com.alibaba.android.rimet | 文本/语音/图片/视频/位置 |
| 企业微信 | com.tencent.wework | 文本/语音/图片/视频/位置 |
| QQ | com.tencent.mobileqq | 文本/语音/图片/视频/位置 |

**SocialOrchestrator 特性：**
- 智能消息分类 (关键词+模式匹配)
- 紧急度评估 (4级: low/medium/high/critical)
- 渠道自动选择 (基于现代社交习惯)
- 多渠道降级 (自动切换备用渠道)
- 用户偏好学习 (记住用户选择)
- 联系人集成 (读取通讯录社交账号)

**SocialTool MCP工具：**
- `social.send` - 智能发送消息
- `social.invite` - 发送邀请 (约吃饭)
- `social.urgent` - 发送紧急消息
- `social.guide` - 分享攻略
- `social.payment` - 支付提醒
- `social.classify` - 分析消息类型
- `social.contact.find` - 查找联系人
- `social.contact.search` - 搜索联系人
- `social.channel.list` - 列出可用渠道
- `social.stats` - 获取发送统计

**现代社交习惯映射：**
- 约吃饭 → 微信 (方便讨论确认)
- 紧急重要 → 电话 (即时响应)
- 攻略分享 → 小红书私信 (内容平台)
- 支付提醒 → 支付宝 (金融场景)
- 工作通知 → 钉钉/企业微信 (职场场景)
- 日常聊天 → 微信/抖音 (社交平台)

新增文件：
- `sdk/src/main/java/com/ofa/agent/social/MessageClassifier.java` - 消息分类
- `sdk/src/main/java/com/ofa/agent/social/ChannelSelector.java` - 渠道选择
- `sdk/src/main/java/com/ofa/agent/social/MessageSender.java` - 消息发送
- `sdk/src/main/java/com/ofa/agent/social/SocialOrchestrator.java` - 社交编排
- `sdk/src/main/java/com/ofa/agent/social/adapter/ContactAdapter.java` - 联系人适配
- `sdk/src/main/java/com/ofa/agent/social/SocialTool.java` - 社交工具
- `sdk/src/main/java/com/ofa/agent/sample/SocialNotificationSample.java` - 使用示例

---

## [1.0.9] - 2026-04-02 🧠 AI Agent Enhancement

### 新增 - Android SDK AI Agent System

智能AI增强系统，提供本地推理和智能决策能力：

| 组件 | 功能 |
|------|------|
| LocalAIEngine | 本地AI推理引擎 (TFLite) |
| ModelManager | 模型加载、缓存、更新 |
| InferenceConfig | 推理配置 (线程/GPU/NNAPI) |
| LocalIntentClassifier | 本地意图分类 |
| MultiArmedBandit | 多臂老虎机算法 (MAB) |
| SmartDecisionEngine | 智能决策引擎 |
| UIElementRecognizer | UI元素智能识别 |
| OperationRecommender | 操作智能推荐 |
| AIEnhancedOrchestrator | AI增强编排器 |

**LocalAIEngine 特性：**
- 文本推理 (意图分类、槽位提取)
- 图像推理 (UI元素识别)
- 异步推理支持
- 多模型管理

**InferenceConfig 预设：**
- `lightweight` - 轻量配置 (仅意图)
- `standard` - 标准配置 (意图+槽位)
- `full` - 完整配置 (全部模型)
- `vision` - 视觉配置 (UI识别)

**MultiArmedBandit 策略：**
- Epsilon-Greedy (简单探索)
- UCB (Upper Confidence Bound)
- Thompson Sampling (贝叶斯采样)

**SmartDecisionEngine 决策类型：**
- 店铺选择 (`shop_selection`)
- 支付方式 (`payment_method`)
- 重试策略 (`retry_strategy`)
- 时机选择 (`timing`)

**OperationRecommender 特性：**
- 基于页面的推荐
- 基于序列的推荐
- 基于历史的推荐
- 持久化学习

**AIEnhancedOrchestrator 特性：**
- 自然语言输入处理
- 智能操作执行
- 决策优化
- 视觉分析

新增文件：
- `sdk/src/main/java/com/ofa/agent/ai/LocalAIEngine.java` - 本地AI引擎
- `sdk/src/main/java/com/ofa/agent/ai/ModelManager.java` - 模型管理
- `sdk/src/main/java/com/ofa/agent/ai/InferenceConfig.java` - 推理配置
- `sdk/src/main/java/com/ofa/agent/ai/AIEnhancedOrchestrator.java` - AI增强编排
- `sdk/src/main/java/com/ofa/agent/ai/intent/LocalIntentClassifier.java` - 意图分类
- `sdk/src/main/java/com/ofa/agent/ai/decision/MultiArmedBandit.java` - MAB算法
- `sdk/src/main/java/com/ofa/agent/ai/decision/SmartDecisionEngine.java` - 智能决策
- `sdk/src/main/java/com/ofa/agent/ai/vision/UIElementRecognizer.java` - UI识别
- `sdk/src/main/java/com/ofa/agent/ai/recommendation/OperationRecommender.java` - 操作推荐

---

## [1.0.8] - 2026-04-02 🔗 Integration & Optimization

### 新增 - Android SDK Integration Layer (Phase 5)

完整的集成优化层，整合所有自动化组件：

| 组件 | 功能 |
|------|------|
| AutomationOrchestrator | 统一编排入口 |
| MemoryAwareAutomation | 记忆感知自动化 |
| IntentTriggeredAutomation | 意图触发自动化 |
| SkillAutomationBridge | 技能桥接 |
| ErrorRecovery | 错误恢复 (6种策略) |
| RetryPolicy | 重试策略 (指数退避) |
| PerformanceMonitor | 性能监控 |
| AutomationLogger | 操作日志 |

**MemoryAwareAutomation 特性：**
- 记住用户偏好 (店铺、商品、地址、支付方式)
- 智能搜索建议
- 自动应用历史选项
- 订单历史记录

**IntentTriggeredAutomation 特性：**
- 自然语言 → 自动化操作
- 意图处理器注册
- 异步执行支持
- 置信度过滤

**SkillAutomationBridge 特性：**
- 技能 → 适配器/模板
- 步骤类型支持 (TOOL/DELAY/CONDITION/ASSIGN)
- 上下文变量传递
- 执行记录

**ErrorRecovery 策略：**

| 策略 | 场景 |
|------|------|
| ScrollToFind | 元素未找到 |
| WaitAndRetry | 超时 |
| DismissOverlay | 弹窗阻挡 |
| WaitForPage | 页面加载中 |
| HandlePermission | 权限问题 |
| HandleNetwork | 网络错误 |

**RetryPolicy 预设：**
- `noRetry()` - 不重试
- `quick()` - 快速重试 (3次, 500ms起)
- `standard()` - 标准重试 (3次, 1s起)
- `aggressive()` - 激进重试 (5次, 2s起)
- `network()` - 网络优化 (5次, 抖动30%)
- `ui()` - UI优化 (3次, 500ms起)

**PerformanceMonitor 特性：**
- 操作耗时统计
- 成功率追踪
- 慢操作告警
- 报告生成

**AutomationLogger 特性：**
- 多级别日志 (DEBUG/INFO/WARN/ERROR)
- 文件持久化
- 自动轮转
- JSON导出

新增文件：
- `sdk/src/main/java/com/ofa/agent/automation/AutomationOrchestrator.java` - 统一编排
- `sdk/src/main/java/com/ofa/agent/automation/integration/MemoryAwareAutomation.java` - 记忆集成
- `sdk/src/main/java/com/ofa/agent/automation/integration/IntentTriggeredAutomation.java` - 意图集成
- `sdk/src/main/java/com/ofa/agent/automation/integration/SkillAutomationBridge.java` - 技能集成
- `sdk/src/main/java/com/ofa/agent/automation/recovery/ErrorRecovery.java` - 错误恢复
- `sdk/src/main/java/com/ofa/agent/automation/recovery/RetryPolicy.java` - 重试策略
- `sdk/src/main/java/com/ofa/agent/automation/monitor/PerformanceMonitor.java` - 性能监控
- `sdk/src/main/java/com/ofa/agent/automation/monitor/AutomationLogger.java` - 操作日志

---

## [1.0.7] - 2026-04-02 🔧 ROM System Layer

### 新增 - Android SDK ROM System Layer (Phase 4)

系统级自动化能力，提供静默安装、权限授权、后台保活等 ROM 内置功能：

| 组件 | 功能 |
|------|------|
| SystemPermissionManager | 权限检测、能力评估、优雅降级 |
| SilentInstaller | 静默安装/卸载 (多方法支持) |
| KeepAliveManager | 后台保活 (5种策略) |
| SystemAutomationEngine | 系统级自动化引擎 |
| HybridAutomationEngine | 混合引擎 (无障碍+系统级) |
| SystemTool | 系统级工具定义 |

**权限管理特性：**
- 检测系统级权限 (INSTALL_PACKAGES, WRITE_SECURE_SETTINGS 等)
- 检测 Root 访问权限
- 检测 System App 状态
- 自动评估能力层级 (BASIC/ENHANCED/SYSTEM_LEVEL)
- 生成能力报告

**静默安装特性：**
- Package Manager 方式 (System App)
- Root Shell 方式
- PackageInstaller Session 方式
- 用户引导方式 (降级)
- 自动选择最佳方法

**保活策略：**

| 策略 | 说明 |
|------|------|
| ForegroundService | 前台服务 + 持久通知 |
| WakeLock | CPU 唤醒锁 |
| BatteryOptimization | 电池优化白名单 |
| SystemApp | 系统应用特权 |
| RootKeepAlive | Root 修改进程策略 |

**新增工具：**

| 工具 | 功能 |
|------|------|
| system.install | 静默安装 APK |
| system.uninstall | 静默卸载应用 |
| system.grantPermission | 静默授权 |
| system.setSecureSetting | 修改安全设置 |
| system.enableAccessibility | 启用无障碍服务 |
| system.keepAlive | 启用/禁用保活 |
| system.getCapability | 获取系统能力报告 |

**HybridAutomationEngine 特性：**
- 自动选择最佳引擎 (系统级/无障碍)
- 能力感知切换
- 优雅降级机制
- 统一操作接口

新增文件：
- `sdk/src/main/java/com/ofa/agent/automation/system/SystemPermissionManager.java` - 权限管理
- `sdk/src/main/java/com/ofa/agent/automation/system/SilentInstaller.java` - 静默安装
- `sdk/src/main/java/com/ofa/agent/automation/system/KeepAliveManager.java` - 保活管理
- `sdk/src/main/java/com/ofa/agent/automation/system/SystemAutomationEngine.java` - 系统引擎
- `sdk/src/main/java/com/ofa/agent/automation/hybrid/HybridAutomationEngine.java` - 混合引擎
- `sdk/src/main/java/com/ofa/agent/automation/SystemTool.java` - 系统工具
- `sdk/src/main/java/com/ofa/agent/sample/SystemAutomationSample.java` - 使用示例

---

## [1.0.6] - 2026-04-02 📱 App Adapter Layer

### 新增 - Android SDK App Adapter Layer (Phase 3)

应用适配层，为主流应用提供预定义操作：

| 适配器 | 包名 | 支持操作 |
|--------|------|----------|
| MeituanAdapter | com.sankuai.meituan | 美团外卖全流程 |
| ElemeAdapter | me.ele | 饿了么全流程 |
| TaobaoAdapter | com.taobao.taobao | 淘宝购物全流程 |
| JDAdapter | com.jingdong.app.mall | 京东购物全流程 |

**适配器支持的操作：**

| 操作 | 功能 |
|------|------|
| search | 搜索商品/店铺 |
| selectShop | 选择店铺 |
| selectProduct | 选择商品 |
| configureOptions | 配置规格（颜色、尺码、口味等） |
| addToCart | 加入购物车 |
| goToCart | 进入购物车 |
| goToCheckout | 去结算 |
| selectAddress | 选择收货地址 |
| submitOrder | 提交订单 |
| pay | 支付（支付宝/微信/银行卡） |
| getOrderStatus | 获取订单状态 |

**页面检测：**
- 自动识别当前页面类型（首页/搜索/商品详情/购物车/结算/订单）
- 智能判断操作可行性

**AppAdapterManager 特性：**
- 自动检测最适合的适配器
- 置信度评分机制
- 统一操作接口
- 动态注册/注销适配器

**操作模板系统：**

| 模板 | 用途 |
|------|------|
| food_delivery | 点外卖完整流程 |
| shopping | 购物下单完整流程 |
| search_and_add | 搜索并加入购物车 |

**模板特性：**
- 参数化操作序列
- 默认参数支持
- 必填参数验证
- 条件步骤（可选执行）
- 等待时间控制

新增文件：
- `sdk/src/main/java/com/ofa/agent/automation/adapter/AppAdapter.java` - 适配器接口
- `sdk/src/main/java/com/ofa/agent/automation/adapter/AppAdapterManager.java` - 适配器管理
- `sdk/src/main/java/com/ofa/agent/automation/adapter/BaseAppAdapter.java` - 基础适配器
- `sdk/src/main/java/com/ofa/agent/automation/adapter/food/MeituanAdapter.java` - 美团适配器
- `sdk/src/main/java/com/ofa/agent/automation/adapter/food/ElemeAdapter.java` - 饿了么适配器
- `sdk/src/main/java/com/ofa/agent/automation/adapter/shopping/TaobaoAdapter.java` - 淘宝适配器
- `sdk/src/main/java/com/ofa/agent/automation/adapter/shopping/JDAdapter.java` - 京东适配器
- `sdk/src/main/java/com/ofa/agent/automation/template/OperationTemplate.java` - 操作模板
- `sdk/src/main/java/com/ofa/agent/automation/template/TemplateRegistry.java` - 模板注册表

---

## [1.0.5] - 2026-04-02 🚀 Automation Enhanced

### 新增 - Android SDK UI Automation Phase 2

UI 自动化增强层，提供高级操作能力：

| 组件 | 功能 |
|------|------|
| ScrollHelper | 滚动辅助（边界检测、智能滚动） |
| PageMonitor | 页面变化监听、稳定性检测 |
| ScreenCapture | 屏幕截图（MediaProjection API） |
| ActionRecorder | 操作录制 |
| ActionReplay | 操作回放 |
| SimpleOcrHelper | 简单 OCR 辅助（占位实现） |

**新增工具：**

| 工具 | 功能 |
|------|------|
| ui.pullToRefresh | 下拉刷新 |
| ui.scrollToTop | 滚动到顶部 |
| ui.scrollToBottom | 滚动到底部 |
| ui.capture | 截图 |
| ui.waitForStable | 等待页面稳定 |
| ui.startRecord | 开始录制 |
| ui.stopRecord | 停止录制 |
| ui.replay | 回放操作 |
| ui.findText | OCR 文字查找 |

**ScrollHelper 特性：**
- 智能滚动查找（自动检测滚动边界）
- 滚动到顶部/底部
- 下拉刷新
- 边界检测

**PageMonitor 特性：**
- 页面变化监听
- 页面稳定性检测
- 包名变化监听
- 页面历史记录

**ScreenCapture 特性：**
- 基于 MediaProjection 的截图
- 区域截图
- 图片比对

**录制回放特性：**
- 操作录制（支持截图）
- JSON 格式保存
- 操作回放（支持原时序）
- 错误处理

新增文件：
- `sdk/src/main/java/com/ofa/agent/automation/advanced/ScrollHelper.java` - 滚动辅助
- `sdk/src/main/java/com/ofa/agent/automation/advanced/PageMonitor.java` - 页面监控
- `sdk/src/main/java/com/ofa/agent/automation/advanced/ScreenCapture.java` - 屏幕截图
- `sdk/src/main/java/com/ofa/agent/automation/advanced/ActionRecorder.java` - 操作录制
- `sdk/src/main/java/com/ofa/agent/automation/advanced/ActionReplay.java` - 操作回放
- `sdk/src/main/java/com/ofa/agent/automation/vision/SimpleOcrHelper.java` - OCR辅助

---

## [1.0.4] - 2026-04-02 🤖 Automation System

### 新增 - Android SDK UI Automation

基于 AccessibilityService 的 UI 自动化系统，支持跨应用操作：

| 组件 | 功能 |
|------|------|
| AutomationEngine | 自动化引擎接口 |
| AccessibilityEngine | 无障碍服务实现 |
| OFAAccessibilityService | 无障碍服务 |
| NodeFinder | UI节点查找器 |
| GesturePerformer | 手势执行器 |
| AutomationManager | 统一管理器 |
| UITool | UI操作工具集 |

**工具定义：**

| 工具 | 功能 |
|------|------|
| ui.click | 点击元素（坐标/文本） |
| ui.longClick | 长按元素 |
| ui.swipe | 滑动手势（方向/坐标） |
| ui.input | 文本输入 |
| ui.find | 查找元素 |
| ui.wait | 等待元素出现 |
| ui.scrollFind | 滚动查找元素 |

**核心特性：**
- 坐标点击、文本点击、选择器点击
- 四向滑动、自定义滑动路径
- 剪贴板输入、ACTION_SET_TEXT输入
- 元素查找（文本、ID、类名、描述）
- 等待元素、等待页面稳定
- 滚动查找（支持最大滚动次数）

**能力层级：**

| 层级 | 说明 |
|------|------|
| BASIC | 基础点击、查找 |
| ENHANCED | 手势执行、滚动查找 |
| FULL_ACCESSIBILITY | 完整无障碍能力 |
| SYSTEM_LEVEL | 系统级能力（需Root） |

新增文件：
- `sdk/src/main/java/com/ofa/agent/automation/AutomationEngine.java` - 引擎接口
- `sdk/src/main/java/com/ofa/agent/automation/AutomationResult.java` - 操作结果
- `sdk/src/main/java/com/ofa/agent/automation/AutomationConfig.java` - 配置
- `sdk/src/main/java/com/ofa/agent/automation/BySelector.java` - 元素选择器
- `sdk/src/main/java/com/ofa/agent/automation/AutomationNode.java` - UI节点
- `sdk/src/main/java/com/ofa/agent/automation/AutomationCapability.java` - 能力层级
- `sdk/src/main/java/com/ofa/agent/automation/AutomationListener.java` - 事件监听
- `sdk/src/main/java/com/ofa/agent/automation/Direction.java` - 滑动方向
- `sdk/src/main/java/com/ofa/agent/automation/ScreenDimension.java` - 屏幕尺寸
- `sdk/src/main/java/com/ofa/agent/automation/AutomationManager.java` - 管理器
- `sdk/src/main/java/com/ofa/agent/automation/UITool.java` - UI工具
- `sdk/src/main/java/com/ofa/agent/automation/accessibility/OFAAccessibilityService.java` - 无障碍服务
- `sdk/src/main/java/com/ofa/agent/automation/accessibility/AccessibilityEngine.java` - 引擎实现
- `sdk/src/main/java/com/ofa/agent/automation/accessibility/NodeFinder.java` - 节点查找
- `sdk/src/main/java/com/ofa/agent/automation/accessibility/GesturePerformer.java` - 手势执行
- `sdk/src/main/res/xml/accessibility_config.xml` - 服务配置
- `sdk/src/main/java/com/ofa/agent/sample/AutomationSample.java` - 使用示例

---

## [1.0.3] - 2026-04-02 🧠 Memory System

### 新增 - Android SDK Memory System

三层用户记忆系统，让系统越来越懂用户：

| 组件 | 层级 | 功能 |
|------|------|------|
| MemoryCache | L1 | 内存缓存 (LRU策略, 毫秒级访问) |
| MemoryDatabase | L2 | Room数据库 (持久化存储) |
| MemoryArchive | L3 | 文件归档 (冷数据备份/导入导出) |

核心特性：
- **智能推荐**: 综合使用频率、最近使用、时间衰减计算推荐分数
- **自动补全**: 根据部分输入推荐完整值
- **偏好记忆**: 记住用户选择，下次自动推荐
- **导入导出**: JSON格式备份和恢复用户记忆

新增文件：
- `sdk/src/main/java/com/ofa/agent/memory/MemoryCache.java` - L1缓存
- `sdk/src/main/java/com/ofa/agent/memory/MemoryEntity.java` - Room实体
- `sdk/src/main/java/com/ofa/agent/memory/MemoryDao.java` - Room DAO
- `sdk/src/main/java/com/ofa/agent/memory/MemoryDatabase.java` - Room数据库
- `sdk/src/main/java/com/ofa/agent/memory/MemoryArchive.java` - L3归档
- `sdk/src/main/java/com/ofa/agent/memory/UserMemoryManager.java` - 三层集成管理器
- `sdk/src/main/java/com/ofa/agent/sample/MemorySample.java` - 使用示例

依赖更新：
- Room Database 2.6.1 (持久化存储)

---

## [1.0.2] - 2026-04-01 🎯 Skill System

### 新增 - Android SDK Skill Orchestration

技能编排系统，支持用户创建自定义多步骤自动化：

| 步骤类型 | 功能 |
|---------|------|
| TOOL | 执行工具调用 |
| INTENT | 触发意图识别 |
| DELAY | 延时等待 |
| WAIT_FOR | 等待条件满足 |
| CONDITION | 条件分支判断 |
| ASSIGN | 变量赋值 |
| INPUT | 获取用户输入 |
| CONFIRM | 请求用户确认 |
| NOTIFY | 发送通知 |
| PARALLEL | 并行执行 |
| LOOP | 循环执行 |
| SUB_SKILL | 调用子技能 |

核心组件：
- `SkillDefinition` - 技能定义（步骤、触发器、输入输出）
- `SkillContext` - 执行上下文（状态、变量、回调）
- `CompositeSkillExecutor` - 技能执行器
- `SkillRegistry` - 技能注册表（持久化）
- `FoodDeliverySkills` - 奶茶点单示例技能

内置技能示例 - 点奶茶流程：
```
1. 启动美团/淘宝闪购APP
2. 切换到外卖页面
3. 选择地址
4. 搜索奶茶
5. 选择商品（型号、甜度、糖度、小料）
6. 提交订单
7. 支付
8. 跟踪外卖进度
9. 快到了提醒用户
```

## [1.0.1] - 2026-04-01 🧠 Intent System

### 新增 - Android SDK Intent Understanding

意图理解系统，读懂用户自然语言指令：

| 功能 | 说明 |
|------|------|
| 模式匹配 | 正则表达式识别意图 |
| 关键词检测 | 多关键词组合匹配 |
| Slot提取 | 自动提取参数（地址、时间、数量等） |
| 置信度评分 | 多引擎结果综合评分 |

22个内置意图：
- 查询类: weather_query, stock_query, news_query, traffic_query...
- 操作类: app_launch, app_close, call_contact, send_message...
- 设置类: setting_change, alarm_set, reminder_set...
- 媒体类: music_play, video_play, photo_take...

核心组件：
- `IntentEngine` - 意图识别引擎
- `IntentDefinition` - 意图定义
- `UserIntent` - 解析结果
- `IntentRegistry` - 意图注册表
- `IntentToolMapper` - 意图→工具映射
- `TaskExecutor` - 任务执行器

---

### 新增 - Web Dashboard

基于 Vue 3 + TypeScript + Vite 的 Web 管理控制台:

| 页面 | 功能 |
|------|------|
| 控制台 | 系统概览、统计卡片、实时活动流、性能指标 |
| 智能体管理 | Agent 列表、搜索过滤、详情弹窗、删除操作 |
| 任务管理 | 任务列表、新建表单、状态筛选、统计条 |
| 系统监控 | 指标卡片、WebSocket 实时更新、任务进度条 |
| 消息中心 | 消息发送、广播、发送历史、快捷操作 |
| 系统设置 | 连接配置、显示设置 |

### 新增 - AI深度集成

| 模块 | 功能 | 文件 |
|------|------|------|
| LLM管理器 | 多LLM提供商(OpenAI/Claude/本地模型)、统一接口 | `pkg/llm/manager.go` |
| LLM适配器 | OpenAI/Claude/Ollama适配器、流式支持 | `pkg/llm/adapters.go` |
| Prompt管理 | 模板注册渲染、变量验证 | `pkg/llm/prompt.go` |
| LLM Agent | 工具调用、记忆管理、Agent注册表 | `pkg/llm/agent.go` |
| 向量存储 | 内存向量存储、RAG检索 | `pkg/llm/vector.go` |

### 新增 - 代码生成

| 模块 | 功能 | 文件 |
|------|------|------|
| 代码生成器 | 模板引擎、多语言格式化 | `pkg/codegen/generator.go` |
| API生成器 | 模型/处理器/路由/OpenAPI生成 | `pkg/codegen/api.go` |
| SDK生成器 | Go/TypeScript/Python SDK | `pkg/codegen/sdk.go` |

### 新增 - Agent协作

| 模块 | 功能 | 文件 |
|------|------|------|
| 协作管理器 | 7种协作类型、生命周期管理 | `pkg/collab/manager.go` |
| 任务编排器 | 顺序/并行/管道/MapReduce执行 | `pkg/collab/orchestrator.go` |
| 任务分配器 | 5种分配策略、Agent评分 | `pkg/collab/allocator.go` |

### 新增 - 去中心化

| 模块 | 功能 | 文件 |
|------|------|------|
| 去中心化管理器 | 多网络类型、节点管理 | `pkg/decentralized/manager.go` |
| Peer管理 | TCP连接、消息广播 | `pkg/decentralized/peer.go` |
| 共识引擎 | PBFT/Raft/PoA/PoS共识 | `pkg/decentralized/consensus.go` |

### 新增 - WASM与插件

| 模块 | 功能 | 文件 |
|------|------|------|
| WASM运行时 | wazero运行时、内存限制 | `pkg/wasm/runtime.go` |
| 插件管理器 | 生命周期管理、钩子系统 | `pkg/plugin/manager.go` |

### 新增 - 多平台SDK

| SDK | 语言 | 特性 |
|-----|------|------|
| Python SDK | Python 3.8+ | asyncio异步、HTTP/WS/gRPC |
| Rust SDK | Rust 1.70+ | tokio异步、安全内存管理 |
| Node.js SDK | TypeScript | 多连接类型、完整技能系统 |
| C++ SDK | C++17 | 嵌入式友好、CMake构建 |

### 新增 - 企业级特性

- **安全**: mTLS、端到端加密、JWT认证、RBAC权限
- **高可用**: 多数据中心、故障转移、灰度发布
- **可观测性**: Prometheus指标、分布式追踪、日志聚合
- **多租户**: 租户隔离、资源配额、计费系统

---

## [0.8.0] - 2026-03-30

### 新增
- 安全审计工具 (`pkg/audit/security.go`)
- OpenAPI文档生成器 (`pkg/openapi/generator.go`)
- 性能基准测试报告 (`pkg/benchmark/report.go`)
- 部署指南文档 (`docs/DEPLOYMENT.md`)

### 改进
- 优化消息路由性能
- 增强缓存命中率

---

## [0.7.0] - 2026-03-30

### 新增
- **通信**: P2P通信、消息路由、广播、NATS队列
- **基础设施**: Redis缓存、PostgreSQL、etcd配置中心
- **场景验证**: 跨设备协同、智能家居、分布式AI、隐私计算

---

## [0.6.0] - 2026-03-30

### 新增
- Lite Agent SDK (手表/手环)
- IoT Agent SDK (智能家居)
- 端到端加密 (X25519/AES-GCM)
- 文件分片传输

---

## [0.5.0] - 2026-03-30

### 新增
- AI助手集成
- 智能调度器
- NLP处理器
- 自动修复
- 智能巡检

---

## [0.4.0] - 2026-03-30

### 新增
- Agent Store 应用商店
- 云服务管理
- 多云提供商支持

---

## [0.3.0] - 2026-03-30

### 新增
- 边缘计算支持
- AI模型推理
- 联邦学习
- 分布式推理

---

## [0.2.0] - 2026-03-30

### 新增
- 多Center集群
- 能力市场
- Web SDK
- Desktop SDK
- 工作流引擎
- RBAC权限

---

## [0.1.0] - 2026-03-28

### 新增
- **Center服务** - gRPC(9090) + REST(8080)双协议
- **Agent客户端** - Go语言实现
- **调度策略** - capability_first, load_balance, latency_first, power_aware, hybrid
- **内置技能** - text.process, json.process, calculator, echo
- **存储层** - 内存存储

### 测试
- 调度器测试: 6/6 通过
- 执行器测试: 14/14 通过

---

## 版本路线图

```
0.1.0 → ... → 0.9.0 → 1.0.1 → 1.0.2 → 1.0.3 → 1.0.4 → 1.0.5 → 1.0.6 → 1.0.7 → 1.0.8 → 1.0.9 → 1.1.0
原型         Beta    Intent   Skill   Memory  Auto v1  Auto v2  Adapter  ROM     Integ   AI Agent  Social
✅           ✅      ✅       ✅       ✅       ✅       ✅       ✅       ✅       ✅       ✅        ✅
```

| 版本 | 里程碑 | 状态 |
|------|--------|------|
| 0.1.0 | 架构原型 | ✅ |
| ... | ... | ✅ |
| **0.9.0** | **Beta** | ✅ |
| **1.0.1** | **Intent System** | ✅ |
| **1.0.2** | **Skill System** | ✅ |
| **1.0.3** | **Memory System** | ✅ |
| **1.0.4** | **Automation v1 (Basic)** | ✅ |
| **1.0.5** | **Automation v2 (Enhanced)** | ✅ |
| **1.0.6** | **App Adapter Layer** | ✅ |
| **1.0.7** | **ROM System Layer** | ✅ |
| **1.0.8** | **Integration & Optimization** | ✅ |
| **1.0.9** | **AI Agent Enhancement** | ✅ |
| **1.1.0** | **Social Notification System** | ✅ 当前 |
| 1.0.0 | 正式发布 | 🔜 计划中 |

---

## 项目统计

| 指标 | 数值 |
|------|------|
| Go源文件 | 119+ |
| Android SDK | 120+ Java类 |
| 内置意图 | 22 |
| 步骤类型 | 12 |
| SDK平台 | 10 |
| 内置技能 | 7+ |
| UI工具 | 14 |
| 系统工具 | 7 |
| 社交工具 | 10 |
| App适配器 | 4 |
| 操作模板 | 3 |
| 保活策略 | 5 |
| 恢复策略 | 6 |
| 重试预设 | 6 |
| AI组件 | 9 |
| 决策策略 | 3 |
| 社交渠道 | 9 |
| 消息类型 | 10 |

---

*更新时间: 2026-04-02*