# OFA 更新日志

所有重要的变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [v9.0.0-dev] - 2026-03-31 🔧 开发中

### 新增 (Sprint 30) - LLM深度集成 🎯 下一代版本
- **LLM管理器**
  - 多LLM提供商支持(OpenAI/Claude/本地模型)
  - 统一接口抽象
  - 统计和监控
  - `pkg/llm/manager.go` - LLM管理器
- **LLM适配器**
  - OpenAI适配器(支持流式)
  - Claude适配器
  - 本地模型适配器(Ollama兼容)
  - 流式响应支持
  - `pkg/llm/adapters.go` - LLM适配器
- **Prompt管理**
  - 模板注册和渲染
  - 变量验证和默认值
  - 上下文管理
  - 预置模板
  - `pkg/llm/prompt.go` - Prompt管理
- **LLM Agent**
  - 工具调用能力
  - 记忆管理
  - Agent注册表
  - 内置工具(天气/搜索/计算器)
  - `pkg/llm/agent.go` - LLM Agent
- **向量存储**
  - 内存向量存储
  - 余弦相似度搜索
  - RAG检索增强生成
  - 知识库管理
  - `pkg/llm/vector.go` - 向量存储
- **LLM服务**
  - HTTP API接口
  - 流式响应
  - 知识库服务
  - `pkg/llm/service.go` - LLM服务

### 新增 (Sprint 31) - 自动代码生成 🎯 下一代版本
- **代码生成器**
  - 模板引擎支持
  - 多语言格式化器
  - 批量生成
  - 文件头添加
  - `pkg/codegen/generator.go` - 代码生成器
- **API代码生成器**
  - 模型代码生成
  - 处理器代码生成
  - 路由代码生成
  - 测试代码生成
  - OpenAPI文档生成
  - `pkg/codegen/api.go` - API生成器
- **SDK代码生成器**
  - Go SDK生成
  - TypeScript SDK生成
  - Python SDK生成
  - Proto文件生成
  - `pkg/codegen/sdk.go` - SDK生成器
- **文档生成器**
  - Markdown文档生成
  - HTML文档生成
  - OpenAPI规范生成
  - README文档生成
  - `pkg/codegen/doc.go` - 文档生成器

### 新增 (Sprint 32) - 智能Agent协作 🎯 下一代版本
- **协作管理器**
  - 协作生命周期管理(创建/规划/执行/完成)
  - 7种协作类型(顺序/并行/管道/MapReduce/共识/拍卖/协商)
  - 任务依赖管理
  - Agent角色分配
  - 约束条件检查
  - 事件发布订阅
  - `pkg/collab/manager.go` - 协作管理器
- **任务编排器**
  - 顺序执行模式
  - 并行执行模式
  - 管道执行模式
  - MapReduce执行模式
  - 状态追踪器
  - 任务依赖检查
  - `pkg/collab/orchestrator.go` - 任务编排器
- **任务分配器**
  - 5种分配策略(能力/负载均衡/延迟/成本/综合评分)
  - Agent注册表管理
  - 约束条件过滤
  - 动态负载追踪
  - Agent评分系统
  - 重新分配支持
  - `pkg/collab/allocator.go` - 任务分配器
- **结果聚合器**
  - 5种聚合策略(合并/最佳/共识/投票/平均)
  - 任务结果统计
  - 成本计算
  - 成功率计算
  - 聚合统计报告
  - `pkg/collab/aggregator.go` - 结果聚合器
- **Agent协商器**
  - 提议创建和投票机制
  - 加权投票计算
  - 协议生成和执行
  - 5种冲突解决策略(先来先得/优先级/投票/随机/协商)
  - 任务协商流程
  - `pkg/collab/negotiator.go` - Agent协商器

### 新增 (Sprint 33) - 去中心化增强 🎯 下一代版本
- **去中心化管理器**
  - 多网络类型支持(全P2P/混合/联邦/网状)
  - 节点生命周期管理
  - 分布式任务分发
  - 网络健康监控
  - `pkg/decentralized/manager.go` - 去中心化管理器
- **Peer管理**
  - TCP连接管理
  - 消息收发和广播
  - 健康检查和心跳
  - Peer状态追踪
  - `pkg/decentralized/peer.go` - Peer管理器
- **共识引擎**
  - 多共识算法(PBFT/Raft/PoA/PoS/投票)
  - 提案投票机制
  - 加权投票计算
  - 验证者管理
  - `pkg/decentralized/consensus.go` - 共识引擎
- **数据复制**
  - 多复制策略(全量/部分/地理/按需/分片)
  - 数据哈希校验
  - 复制状态追踪
  - 自动修复机制
  - `pkg/decentralized/replication.go` - 数据复制器
- **Peer发现**
  - 多发现方法(DHT/启动节点/广播/Gossip/注册中心)
  - 持续发现机制
  - Peer端点管理
  - 发现日志记录
  - `pkg/decentralized/discovery.go` - Peer发现器
- **数据同步**
  - 多同步模式(立即/批量/定时/按需)
  - 同步任务队列
  - 冲突检测和解决
  - 同步状态追踪
  - `pkg/decentralized/sync.go` - 同步管理器
- **信任管理**
  - 多信任算法(简单/加权/贝叶斯/滑动窗口/声誉)
  - 信任事件记录
  - 信任评分衰减
  - 信任等级划分
  - `pkg/decentralized/trust.go` - 信任管理器

---

## [v8.0.0] - 2026-03-30 🎉 正式发布

### 新增 (Sprint 26) - 最终发布
- **性能基准测试报告**
  - 10项核心性能测试
  - 性能评分系统 (0-100)
  - JSON报告导出
  - 性能优化建议
  - `pkg/benchmark/report.go` - 报告生成器
- **版本管理模块**
  - 版本信息管理
  - 构建信息追踪
  - `pkg/version/version.go` - 版本管理
- **发布文档**
  - RELEASE_NOTES.md - 发布说明
  - 完整的升级指南
  - 已知问题列表

---

## [v8.1.0-dev] - 2026-03-30 🔧 开发中

### 新增 (Sprint 27) - WebAssembly技能支持 🎯 增强版本
- **WASM运行时集成**
  - 基于wazero的WASM运行时
  - 内存和燃料限制
  - 超时控制
  - `pkg/wasm/runtime.go` - WASM运行时
- **WASM技能加载器**
  - 支持从文件/URL/注册表加载
  - 技能缓存
  - 技能注册表
  - SHA256校验
  - `pkg/wasm/loader.go` - 技能加载器
- **WASM安全沙箱**
  - 内存/CPU限制
  - 文件访问控制
  - 网络访问控制
  - 权限管理
  - 安全策略
  - `pkg/wasm/sandbox.go` - 安全沙箱

### 新增 (Sprint 28) - 插件系统增强 🎯 增强版本
- **插件管理器**
  - 插件生命周期管理(注册/启用/启动/停止/注销)
  - 钩子系统(before/after task/skill)
  - 事件发布订阅机制
  - 统计信息和配置导出导入
  - `pkg/plugin/manager.go` - 插件管理器
- **插件加载器**
  - 多来源加载支持(local/http/registry/wasm)
  - 依赖解析和验证
  - 热加载监听
  - 缓存管理
  - 权限验证
  - `pkg/plugin/loader.go` - 插件加载器
- **插件注册表**
  - 插件注册和发现
  - 分类和搜索索引
  - 版本管理
  - 精选/热门/已验证插件
  - 统计信息
  - `pkg/plugin/registry.go` - 插件注册表

### 新增 (Sprint 29) - 多平台SDK扩展 🎯 增强版本
- **Python SDK**
  - 异步Agent实现(asyncio)
  - HTTP/WebSocket/gRPC连接
  - 技能执行器和注册表
  - 内置技能(Echo/Text/Calculator/JSON/List)
  - 协议和消息模块
  - `src/sdk/python/` - Python Agent SDK
- **Rust SDK**
  - 高性能异步实现(tokio)
  - 安全内存管理
  - 多连接类型支持(grpc/http/websocket)
  - 技能系统
  - 内置技能
  - `src/sdk/rust/` - Rust Agent SDK
- **Node.js SDK**
  - TypeScript实现
  - HTTP/WebSocket/gRPC连接
  - 技能执行器和注册表
  - 内置技能
  - 完整类型定义
  - `src/sdk/nodejs/` - Node.js Agent SDK
- **C++ SDK**
  - 现代C++17实现
  - 嵌入式友好设计
  - CMake构建系统
  - 技能系统
  - 内置技能
  - `src/sdk/cpp/` - C++ Agent SDK

---

## [v8.0.0] - 2026-03-30 🎉 正式发布

### 新增 (Sprint 26) - 最终发布
- **性能基准测试报告**
  - 10项核心性能测试
  - 性能评分系统 (0-100)
  - JSON报告导出
  - 性能优化建议
  - `pkg/benchmark/report.go` - 报告生成器
- **版本管理模块**
  - 版本信息管理
  - 构建信息追踪
  - `pkg/version/version.go` - 版本管理
- **发布文档**
  - RELEASE_NOTES.md - 发布说明
  - 完整的升级指南
  - 已知问题列表

### 新增 (Sprint 25) - 发布准备
- **安全审计工具**
  - SSL/TLS配置检查
  - HTTP安全头检查
  - 端点安全性检查
  - 敏感文件检测
  - 安全评分系统
  - `pkg/audit/security.go` - 安全审计器
- **OpenAPI文档生成器**
  - OpenAPI 3.0.3规范支持
  - 自动生成API文档
  - JSON/YAML导出
  - 预置OFA API文档
  - `pkg/openapi/generator.go` - 文档生成器
- **部署指南**
  - 开发环境部署
  - Docker部署
  - Kubernetes生产部署
  - 监控配置
  - 安全配置
  - 故障排查
  - `docs/DEPLOYMENT.md` - 部署文档

### 发布总结
- **功能完整度**: 100%
- **架构符合度**: 95%
- **场景实现度**: 100%
- **测试通过率**: 100%
- **性能评分**: 优秀

---

## [v7.1.0] - 2026-03-30

### 新增 (Sprint 17) - 通信能力补强 🎯 纠偏
- **P2P通信**
  - Agent间直接通信
  - TCP连接管理
  - 心跳检测(Ping/Pong)
  - 消息确认机制(ACK)
  - 超时和重传
  - 组播组管理
  - `pkg/messaging/p2p.go` - P2P管理器
- **消息路由**
  - 路由表管理
  - 路由规则引擎
  - 4种路由类型(直接/转发/广播/负载均衡)
  - Agent能力索引
  - 路由发现
  - `pkg/messaging/router.go` - 消息路由器
- **消息广播**
  - 6种广播模式(全部/在线/类型/区域/能力/排除)
  - 组播组管理
  - 订阅机制
  - 批量发送
  - 投递结果追踪
  - `pkg/messaging/broadcast.go` - 广播管理器

### 新增 (Sprint 18) - 消息持久化 🎯 纠偏
- **NATS消息队列集成**
  - NATS服务器连接管理
  - JetStream持久化上下文
  - 消息发布/订阅
  - 持久化订阅(Durable Subscription)
  - 请求-响应模式
  - 异步发布(Ack回调)
  - 流和消费者管理
  - 连接状态监控
  - `pkg/messaging/queue.go` - NATS管理器
- **消息持久化存储**
  - 多存储类型支持(Memory/SQLite/PostgreSQL/Redis)
  - 消息索引(Subject/From/To/Status)
  - TTL过期管理
  - 消息状态追踪(Pending/Sent/Delivered/Acked/Failed/Expired)
  - 消息查询和过滤
  - 批量存储和导入导出
  - 自动清理过期消息
  - 统计信息收集
  - `pkg/messaging/store.go` - 消息存储

### 新增 (Sprint 19) - 基础设施升级 🎯 纠偏
- **Redis分布式缓存**
  - Redis连接池管理
  - L1本地缓存 + L2 Redis多级缓存
  - 会话存储与管理
  - 消息缓存
  - 批量操作(GetMulti/SetMulti/DeleteMulti)
  - 计数器支持(Increment/Decrement)
  - 分布式锁(SetNX)
  - 过期时间管理
  - 统计与监控
  - `pkg/cache/redis.go` - Redis缓存管理器
- **PostgreSQL数据库支持**
  - 连接池配置
  - Agent/Task数据存储
  - 数据库迁移系统
  - 事务支持(Begin/Commit/Rollback/WithTransaction)
  - 批量插入操作
  - JSONB数据存储
  - 索引优化
  - 健康检查
  - 统计与监控
  - `pkg/store/postgres.go` - PostgreSQL存储

### 新增 (Sprint 20) - 配置中心 🎯 纠偏
- **etcd配置中心**
  - 配置管理(Get/Set/Delete/GetPrefix)
  - 配置监听(Watch/WatchPrefix)
  - 服务注册与发现(Register/Discover/Watch)
  - 租约管理(Grant/Revoke/KeepAlive)
  - 事务支持(Txn/CompareAndSwap)
  - 集群信息查询
  - 连接池管理
  - 健康检查
  - `pkg/config/etcd.go` - etcd配置中心

---

## [v7.1.0-dev] - 2026-03-30

### 新增 (Sprint 24) - 场景验证测试 🎯 场景实现度验证
- **场景验证框架**
  - 4个核心场景测试框架
  - 跨设备协同测试 (P2P连接/消息路由/广播/文件传输/分布式任务)
  - 智能家居联动测试 (MQTT连接/设备影子/设备控制/传感器/自动化规则)
  - 分布式AI推理测试 (模型加载/分布式推理/GPU调度/模型量化/联邦学习)
  - 隐私计算验证 (端到端加密/本地处理/数据隔离/安全聚合/审计日志)
  - 验证报告生成 (JSON/Text/HTML格式)
  - `pkg/scenario/validator.go` - 场景验证器
  - `pkg/scenario/validator_test.go` - 场景测试
  - `pkg/scenario/cli.go` - 命令行工具

---

## [v7.0.0-dev] - 2026-03-30

### 新增 (Sprint 21) - Lite Agent SDK 🎯 Agent扩展
- **Lite Agent SDK** (手表/手环)
  - 低功耗设计(60秒心跳，省电模式)
  - 轻量级消息协议(二进制格式，压缩传输)
  - 传感器管理(心率/计步/GPS/加速度计/温度/光照)
  - 电池管理与功耗优化
  - 多种连接方式(TCP/BLE/WebSocket)
  - 内置技能(心率/计步/位置/通知)
  - 功耗优化器(根据电量自动调整)
  - `src/sdk/lite/agent.go` - Lite Agent核心
  - `src/sdk/lite/protocol.go` - 轻量级协议
  - `src/sdk/lite/example.go` - 使用示例

### 新增 (Sprint 22) - IoT Agent SDK 🎯 Agent扩展
- **IoT Agent SDK** (智能家居)
  - MQTT协议支持(QoS 0/1/2, TLS加密)
  - 设备影子(Desired/Reported/Delta状态)
  - 属性管理与同步
  - 事件发布与遥测数据
  - 命令处理与执行
  - 内置设备类型(智能灯/插座/传感器/门锁/温控器)
  - 设备工厂模式
  - `src/sdk/iot/agent.go` - IoT Agent核心
  - `src/sdk/iot/mqtt.go` - MQTT协议实现
  - `src/sdk/iot/devices.go` - 设备类型定义

### 新增 (Sprint 23) - 安全增强 🎯 功能完善
- **端到端加密**
  - X25519/ECDH密钥交换
  - AES-256-GCM/ChaCha20加密
  - 握手协议(Initiate/Response/Complete)
  - 会话管理与密钥轮换
  - Ed25519数字签名
  - 前向保密支持
  - `pkg/security/e2e.go` - 端到端加密管理器
- **文件分片传输**
  - 大文件分片(默认1MB)
  - 断点续传支持
  - SHA256校验机制
  - 传输状态管理
  - 暂停/恢复/取消操作
  - 流式传输接口
  - `pkg/transfer/chunk.go` - 分片传输管理器

---

## [v6.0.0-dev] - 2026-03-30

### 新增 (Sprint 13)
- **AI助手集成**
  - 自然语言任务理解
  - 意图识别 (7种意图类型)
  - 实体抽取
  - 任务模板匹配
  - 异常诊断助手
  - 智能推荐Agent选择
  - 多轮对话上下文管理
  - `pkg/assistant/assistant.go` - AI助手管理器
- **智能调度器**
  - 基于历史数据的预测调度
  - 多种预测算法 (指数平滑/线性回归/ARIMA/ML)
  - 4种预测类型 (负载/延迟/资源/成本)
  - 5种优化策略 (最小成本/最小延迟/最大资源利用率/最大吞吐量/平衡)
  - Agent评分系统
  - 动态负载预测
  - 模型参数自动调整
  - `pkg/smart/scheduler.go` - 智能调度器

### 新增 (Sprint 14)
- **NLP处理器**
  - 自然语言解析
  - 意图识别 (10种意图类别)
  - 实体抽取 (12种实体类型)
  - 任务参数构建
  - 多语言支持 (中/英/日/韩/德/法)
  - 分词和规范化
  - 关键词和模式匹配
  - 任务模板管理
  - `pkg/nlp/processor.go` - NLP处理器
- **对话管理**
  - 多轮对话支持
  - 对话状态管理 (8种状态)
  - 上下文记忆和实体记忆
  - 澄清询问机制
  - 响应模板系统
  - 反馈收集
  - 会话超时管理
  - 情感分析
  - `pkg/nlp/dialog.go` - 对话管理器

### 新增 (Sprint 15)
- **自动修复**
  - 故障自动检测 (10种故障类型)
  - 自动修复策略 (restart/reconnect/retry/scale/migrate/cleanup)
  - 自愈规则引擎
  - 修复历史记录
  - 冷却期管理
  - 手动修复支持
  - `pkg/auto/repair.go` - 自动修复管理器
- **智能巡检**
  - 定期健康检查 (10种检查类型)
  - 异常预警
  - 性能分析报告
  - 优化建议生成
  - 异常检测 (基线偏差分析)
  - 健康评分系统
  - 巡检调度器
  - `pkg/auto/patrol.go` - 智能巡检管理器
  - 多意图处理
  - 情感分析
  - `pkg/nlp/dialog.go` - 对话管理器

---

## [v5.0.0-dev] - 2026-03-30

### 新增 (Sprint 12)
- **Agent Store 应用商店**
  - 商品管理 (Agent/技能/模板/模型/数据集)
  - 分类管理 (8个预置分类)
  - 搜索索引 (名称/描述/标签)
  - 评分系统 (1-5星评价)
  - 版本管理 (发布/审核/下架)
  - 作者认证
  - 价格模型 (免费/付费/订阅)
  - `pkg/store/agent_store.go` - Agent Store管理器
- **云服务管理**
  - 多云提供商 (AWS/GCP/Azure/阿里/腾讯/本地)
  - 部署管理 (创建/扩展/删除)
  - 自动伸缩 (CPU/内存阈值触发)
  - 健康检查 (30秒周期)
  - 指标收集 (CPU/内存/请求/延迟)
  - 成本计算 (按小时计费)
  - 服务层级 (Starter/Professional/Enterprise)
  - `pkg/cloud/manager.go` - 云服务管理器

---

## [v4.0.0-dev] - 2026-03-30

### 新增 (Sprint 11)
- **高可用**
  - 多数据中心管理 (跨地域部署/健康检查)
  - 故障转移 (自动/手动切换)
  - 数据备份 (全量/增量/定时)
  - 灾备恢复 (RTO/RPO/恢复计划)
  - `pkg/ha/multidc.go` - 多数据中心管理
  - `pkg/ha/backup.go` - 备份恢复
- **灰度发布**
  - 金丝雀部署 (流量百分比/指标监控)
  - 自动回滚 (阈值检测)
  - 流量渐进 (逐步放量)
  - `pkg/ha/canary.go` - 灰度发布
- **服务网格**
  - Istio集成 (VirtualService/DestinationRule)
  - 流量管理 (路由/超时/重试)
  - 熔断器 (连接池/故障隔离)
  - `pkg/ha/mesh.go` - 服务网格
- **审计日志**
  - 操作审计 (事件类型/Actor/变更记录)
  - 合规报告 (按时间/租户/类型)
  - 日志保留 (自动清理)
  - `pkg/audit/logger.go` - 审计日志

### 新增 (Sprint 10)
- **多租户支持**
  - 租户管理 (创建/更新/删除/暂停)
  - 资源配额 (CPU/内存/GPU/存储/API调用)
  - 租户隔离 (数据/网络/资源隔离)
  - 套餐管理 (Free/Basic/Pro/Enterprise)
  - 使用量计费 (按量付费)
  - `pkg/tenant/manager.go` - 租户管理器
  - `pkg/tenant/isolation.go` - 租户隔离
- **可观测性**
  - 分布式追踪 (Span/Trace/上下文传递)
  - 日志聚合 (多级别/租户隔离/持久化)
  - 告警规则 (阈值/静默/通知渠道)
  - `pkg/observability/tracing.go` - 分布式追踪
  - `pkg/observability/logging.go` - 日志聚合
  - `pkg/observability/alerting.go` - 告警管理
- **安全增强**
  - mTLS支持 (双向TLS认证)
  - 证书颁发机构 (自动签发/轮换)
  - 密钥管理 (本地/云KMS集成)
  - 数据加密 (AES-GCM)
  - `pkg/security/mtls.go` - mTLS管理
  - `pkg/security/kms.go` - 密钥管理
- **性能优化**
  - 多级缓存 (L1内存/L2进程/L3分布式)
  - 连接池 (gRPC/HTTP连接复用)
  - 熔断器 (故障隔离/自动恢复)
  - `pkg/performance/cache.go` - 多级缓存
  - `pkg/performance/pool.go` - 连接池

---

## [v3.0.0-dev] - 2026-03-30

### 新增
- **边缘计算**
  - 边缘Center部署 (`pkg/edge/center.go`)
  - 本地任务处理 (离线模式)
  - 边云协同 (任务转发/同步)
  - 数据预处理 (图像/文本/音视频/JSON)
  - `pkg/edge/preprocess.go` - 数据预处理器
- **AI能力**
  - 模型推理技能 (ONNX/GGML/Generic)
  - 模型管理 (注册/加载/卸载)
  - GPU调度 (内存管理/设备选择)
  - 推理队列 (优先级/超时)
  - `pkg/ai/manager.go` - AI管理器
  - `pkg/ai/handlers.go` - 模型处理器
- **联邦学习**
  - 模型训练任务 (多轮训练)
  - 数据隐私保护 (差分隐私/安全聚合/同态加密)
  - 梯度聚合 (FedAvg/FedProx/FedSGD/SCAFFOLD)
  - 客户端管理 (选择/可靠性/心跳)
  - `pkg/federated/learning.go` - 联邦学习管理器
- **分布式推理 (Sprint 9)**
  - 多GPU/多节点推理
  - 分布策略: data_parallel, model_parallel, pipeline, tensor_parallel
  - 节点管理和监控
  - 模型分区和负载均衡
  - `pkg/ai/distributed.go` - 分布式推理管理器
- **模型压缩 (Sprint 9)**
  - 量化类型: INT8, INT4, FP16, BF16, Mixed, Dynamic
  - 模型剪枝 (权重稀疏化)
  - 知识蒸馏 (教师-学生模型)
  - 压缩比和精度损失估算
  - `pkg/ai/quantize.go` - 模型量化器
- **自动调参 (Sprint 9)**
  - 调参策略: grid_search, random_search, bayesian, hyperband, bohb
  - 搜索空间定义 (int/float/string/bool/categorical)
  - 早停机制
  - 并发试验执行
  - `pkg/ai/tuning.go` - 自动调参器
- **模型版本管理 (Sprint 9)**
  - A/B测试 (流量分割/显著性检验)
  - 版本回滚 (历史记录/一键回滚)
  - 版本比较 (指标对比)
  - 状态管理: active, staged, deprecated, archived
  - `pkg/ai/version.go` - 版本注册中心
- **联邦学习安全增强 (Sprint 9)**
  - 安全聚合协议 (密钥生成/秘密分享)
  - 差分隐私噪声 (Gaussian/Laplace机制)
  - 同态加密 (加法同态)
  - 安全审计 (事件日志/签名验证)
  - 安全级别: standard, high, maximum
  - `pkg/federated/secure.go` - 安全增强模块

### 支持的操作
- **数据预处理**: resize, normalize, crop, convert, tokenize, extract, flatten
- **AI推理**: load, infer, unload, get_info
- **联邦学习**: create_task, start_task, submit_update, aggregate
- **分布式推理**: infer, get_stats, list_nodes, register_node
- **模型压缩**: quantize, prune, distill, calibrate, optimize_for_target
- **自动调参**: create_job, start_job, cancel_job, get_best_parameters
- **版本管理**: register_version, activate_version, rollback, create_ab_test

---

## [v2.0.0] - 2026-03-30

### 新增
- **多Center集群**
  - 服务发现与注册 (`pkg/cluster/discovery.go`)
  - 负载均衡 (轮询/最少连接/最少负载/权重/区域感知)
  - 故障转移 (自动检测/恢复/任务重分配)
  - 数据同步 (跨节点同步/版本控制)
  - `pkg/cluster/loadbalancer.go` - 负载均衡器
  - `pkg/cluster/failover.go` - 故障转移管理器
  - `pkg/cluster/sync.go` - 数据同步管理器
- **能力市场**
  - 技能仓库 (本地/远程存储)
  - 版本管理 (语义版本/版本比较)
  - 依赖解析 (自动依赖检测)
  - 安全验证 (Ed25519签名)
  - `pkg/market/capability.go` - 技能仓库管理
  - `pkg/market/version.go` - 版本管理器
- **Web Agent SDK**
  - TypeScript/JavaScript实现
  - WebSocket通信
  - 内置技能: Echo, TextProcess, JSONProcess
  - 自动心跳和重连
  - `src/sdk/web/` - Web SDK源码
- **工作流引擎**
  - 步骤编排 (任务/条件/并行/等待/子工作流)
  - 定时调度 (cron/interval)
  - 事件触发 (消息/事件触发)
  - 错误处理 (失败/跳过/重试)
  - `pkg/workflow/engine.go` - 工作流引擎
  - `pkg/workflow/scheduler.go` - 调度器
- **RBAC权限管理**
  - 角色管理 (创建/更新/删除/继承)
  - 用户管理 (创建/更新/删除/角色分配)
  - 权限检查 (资源/操作/条件)
  - 内置角色 (admin/operator/developer/viewer/agent)
  - HTTP中间件 (认证/RBAC/管理员/所有者)
  - `pkg/rbac/manager.go` - RBAC管理器
  - `pkg/rbac/middleware.go` - HTTP中间件
  - `pkg/rbac/api.go` - REST API
- **流式任务处理**
  - 流管理 (创建/关闭/暂停/恢复)
  - 消息推送 (序列号/时间戳)
  - 订阅机制 (回调/HTTP流)
  - HTTP流式传输 (chunked/flush)
  - `pkg/stream/manager.go` - 流管理器
  - `pkg/stream/api.go` - REST API
- **Desktop Agent SDK**
  - 跨平台支持 (Windows/macOS/Linux)
  - 系统托盘集成
  - 内置技能: SystemInfo, FileOperation, Command, Echo
  - 脚本/二进制/WASM技能加载
  - 自动连接和心跳
  - `src/sdk/desktop/` - Desktop SDK源码

---

## [v1.0.0] - 2026-03-28

### 新增
- **JWT安全认证**
  - Ed25519签名算法
  - Access Token + Refresh Token双令牌机制
  - 认证中间件
  - 权限检查中间件
  - `pkg/auth/jwt.go` - JWT管理器
  - `pkg/auth/middleware.go` - 认证中间件
- **Docker部署支持**
  - 多阶段构建优化镜像大小
  - docker-compose.yml一键部署
  - 可选Prometheus + Grafana监控栈
- **Kubernetes部署配置**
  - Deployment配置
  - Service配置
  - Ingress配置
  - ConfigMap配置管理
- **iOS Agent SDK**
  - Swift 5.9实现，支持iOS 13+/macOS 12+
  - gRPC-Swift通信
  - async/await异步API
  - 内置技能: Echo, TextProcess, Calculator
  - 后台模式支持
  - `src/sdk/ios/` - iOS SDK源码
- **性能测试工具**
  - 并发任务压测
  - 延迟分布统计
  - 吞吐量计算
  - `pkg/benchmark/benchmark.go`

### 变更
- 更新Dockerfile支持SQLite (CGO)
- 添加健康检查端点

### 新增
- **Prometheus监控指标导出**
  - 新增 `/metrics` 端点，支持Prometheus格式指标
  - Agent指标: 总数、在线数、离线数、按类型统计
  - 任务指标: 总数、完成数、失败数、取消数、执行时长
  - 消息指标: 总数、投递数、失败数、延迟
  - 系统指标: HTTP请求时长、健康检查次数、gRPC连接数
  - Go运行时指标: 内存、GC、goroutine等
- **HTTP请求监控中间件**
  - 自动记录所有API请求的耗时和状态码
- **metrics包**
  - `pkg/metrics/metrics.go` - 指标定义和收集
  - `pkg/metrics/metrics_test.go` - 指标单元测试
- **SQLite数据库持久化**
  - 新增SQLite存储后端，支持数据持久化
  - 零配置：内存模式（默认）或SQLite文件存储
  - 支持Agent、Task、Message的完整CRUD操作
  - WAL模式提升并发性能
  - `internal/store/sqlite.go` - SQLite存储实现
  - `internal/store/memory.go` - 内存存储实现
  - `internal/store/interface.go` - 存储接口定义
- **存储层重构**
  - 统一存储接口 StoreInterface
  - 工厂模式支持多种存储后端
  - 配置切换：`database.type: "memory"` 或 `"sqlite"`

### 变更
- 更新 `pkg/rest/server.go`，集成Prometheus指标
- 更新 `go.mod`，添加 prometheus/client_golang 和 mattn/go-sqlite3 依赖
- 重构存储层，支持可插拔存储后端

### 新增 (Android SDK)
- **Android Agent SDK**
  - Java语言实现，支持Android 7.0+
  - gRPC双向流通信
  - 自动心跳和重连机制
  - 可插拔技能系统
  - 内置技能: Echo, TextProcess
  - WorkManager后台任务支持
  - `src/sdk/android/` - Android SDK源码

### 新增 (错误处理)
- **统一错误码定义**
  - 客户端错误 (1xx)
  - Agent错误 (2xx)
  - Task错误 (3xx)
  - Skill错误 (4xx)
  - 消息错误 (5xx)
  - 服务端错误 (9xx)
- **ErrorInfo结构**
  - 错误码、消息、详情、原因链
  - JSON序列化支持
  - 错误包装和解包
- `pkg/errors/errors.go` - 错误处理包
- `pkg/errors/errors_test.go` - 错误处理测试

---

## [v0.1.0] - 2026-03-28

### 新增
- **Center服务**
  - gRPC服务 (端口9090)
  - REST API服务 (端口8080)
  - Agent注册和管理
  - 任务提交和调度
  - 消息路由
- **Agent客户端**
  - Go语言实现
  - 连接Center服务
  - 报告能力
  - 执行任务
- **调度策略**
  - capability_first - 能力优先
  - load_balance - 负载均衡
  - latency_first - 延迟优先
  - power_aware - 功耗感知
  - hybrid - 混合策略
- **内置技能**
  - text.process - 文本处理 (uppercase/lowercase/reverse/length)
  - json.process - JSON处理 (get_keys/get_values/pretty)
  - calculator - 计算器 (add/sub/mul/div/pow/sqrt)
  - echo - 回显测试
- **存储层**
  - 内存存储 (sync.Map)
- **工具脚本**
  - ofa.bat - 快速构建/测试/运行脚本
  - test_api.ps1 - PowerShell API测试脚本
- **文档**
  - README.md
  - docs/OPERATION_GUIDE.md
  - docs/01-FEATURE_SPECIFICATION.md
  - docs/02-TECH_SELECTION.md
  - docs/03-ARCHITECTURE_DESIGN.md
  - docs/04-SOLUTION_DESIGN.md
  - docs/05-TEST_CASES.md
  - docs/06-VERSION_PLAN.md
  - docs/API.md
  - docs/DEVELOPMENT.md
  - BUILD_REPORT.md
  - PROJECT_STATUS.md

### 测试
- 调度器测试: 6个通过
- 执行器测试: 14个通过
- 总计: 20个测试全部通过

---

## 版本路线图

### v0.5 (Alpha版本) - 进行中
- [x] Prometheus监控指标
- [ ] 数据库持久化 (PostgreSQL + Redis)
- [ ] Android Agent SDK
- [ ] 完善错误处理

### v1.0 (MVP版本) - 计划中
- Android/iOS Agent
- 完整消息系统
- 监控告警
- 安全认证

### v2.0 (正式版本) - 计划中
- 多平台Agent (Desktop/Web)
- Center集群
- 能力市场

### v3.0 (高级特性) - 计划中
- 边缘计算
- AI能力
- 联邦学习