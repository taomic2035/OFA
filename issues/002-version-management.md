# Issue: 版本号管理优化

**创建时间**: 2026-03-31
**状态**: ✅ Resolved
**优先级**: Medium
**解决时间**: 2026-03-31

---

## 问题分析

当前版本号分散在 50+ 个位置，导致维护困难。每次版本更新需要修改大量文件。

### 当前版本号分布

| 类别 | 文件数 | 问题 |
|------|--------|------|
| Go 源码注释 | 26 | 不应包含版本号 |
| Markdown 文档 | 15+ | 重复且难以维护 |
| package.json | 4 | SDK 应有独立版本 |
| 配置文件 | 4 | 可合并 |
| Docker/K8s | 3 | 镜像版本应独立 |

---

## 设计原则

### 单一真相来源 (Single Source of Truth)

```
src/center/pkg/version/version.go  ← 唯一的版本定义
```

其他地方通过以下方式获取版本号：
- Go 代码：`version.Version`
- API：`GET /health` 返回版本
- Dashboard：从 API 动态获取
- 构建时：`-ldflags` 注入

---

## 版本号保留策略

### ✅ 必须保留（核心位置）

| 文件 | 说明 |
|------|------|
| `src/center/pkg/version/version.go` | **唯一版本定义** |
| `CHANGELOG.md` | 版本历史记录 |
| `SECURITY.md` | 安全支持版本范围 |

### ✅ 可保留（有独立意义）

| 文件 | 说明 |
|------|------|
| `src/dashboard/package.json` | Dashboard 独立发布版本 |
| `src/sdk/python/pyproject.json` | Python SDK 独立版本 |
| `src/sdk/nodejs/*/package.json` | Node.js SDK 独立版本 |
| `src/sdk/web/package.json` | Web SDK 独立版本 |
| `skills/core/*/skill.yaml` | 技能独立版本（功能版本，非项目版本） |

### ❌ 应移除（冗余位置）

| 类别 | 文件数 | 处理方式 |
|------|--------|----------|
| Go 源码注释 | 26 | 移除版本号，保留功能描述 |
| `docs/OPERATION_GUIDE.md` | 2处 | 移除示例中的版本号 |
| `docs/API.md` | 3处 | 移除示例中的版本号 |
| `docs/USER_GUIDE.md` | 2处 | 移除示例中的版本号 |
| `docs/DEPLOYMENT.md` | 2处 | 使用 `latest` 或变量 |
| `README.md` | 2处 | 简化，只保留标题 |
| `configs/center.yaml` | 1处 | 移除，运行时从代码获取 |
| `deployments/kubernetes.yaml` | 1处 | 移除，使用镜像 tag |

---

## 实施计划

### Phase 1: 移除 Go 源码注释中的版本号

将：
```go
// Package llm - 大语言模型集成
// 0.9.0 Beta: LLM深度集成
```

改为：
```go
// Package llm - 大语言模型集成
// 提供多 LLM 提供商的统一接口，支持流式响应和嵌入
```

### Phase 2: 简化文档中的版本号

- `README.md`: 只在标题保留 `v0.9 Beta`
- `docs/USER_GUIDE.md`: 标题保留版本
- 其他文档示例使用通用格式

### Phase 3: 优化配置文件

- `configs/center.yaml`: 移除 version 字段
- Dashboard: 从 `/health` API 动态获取版本显示

### Phase 4: 更新构建脚本

构建时注入版本号：
```bash
go build -ldflags "-X main.Version=$(git describe --tags)"
```

---

## 预期效果

| 指标 | 当前 | 优化后 |
|------|------|--------|
| 版本号位置 | 50+ | ~10 |
| 版本更新工作量 | 修改 50+ 文件 | 修改 1-2 文件 |
| 版本不一致风险 | 高 | 低 |

---

## 验收标准

1. `grep -r "0.9.0" --include="*.go"` 只返回 `version.go`
2. 版本号只定义在 `pkg/version/version.go`
3. Dashboard 从 API 获取版本号
4. 文档示例不包含具体版本号