# Issue: 代码编译错误修复

**创建时间**: 2026-03-31
**状态**: ✅ Resolved
**优先级**: High
**解决时间**: 2026-03-31

---

## 问题描述

项目存在多个编译错误，导致 `go build` 和 `go test` 失败。

---

## 问题列表

### 1. duplicate case in switch statement

**文件**: `src/center/internal/store/memory.go:176`

**错误信息**:
```
duplicate case "sqlite" (constant of string type StoreType) in expression switch
```

**原因**:
```go
const (
    StoreMemory StoreType = "memory"
    StoreSQLite StoreType = "sqlite"  // 已经是 "sqlite"
)

switch storeType {
case StoreMemory, "":
    return NewMemoryStore()
case StoreSQLite, "sqlite":  // StoreSQLite == "sqlite"，重复了
    return NewSQLiteStore(cfg)
}
```

**修复方案**: 移除重复的 `"sqlite"` 字符串

---

### 2. Message 结构体字段缺失

**文件**: `src/center/internal/store/sqlite.go` (lines 389-411)

**错误信息**:
```
msg.Type undefined
msg.Status undefined
msg.CreatedAt undefined
```

**当前 Message 结构体** (`internal/models/models.go:75-84`):
```go
type Message struct {
    ID        string
    FromAgent string
    ToAgent   string
    Action    string
    Payload   []byte
    Timestamp time.Time  // 存在
    TTL       int32
    Headers   map[string]string
    // 缺少: Type, Status, CreatedAt
}
```

**修复方案**: 在 Message 结构体中添加缺失字段，或修改 sqlite.go 使用现有字段

---

### 3. 包名冲突

**文件**: `src/center/pkg/ai/` 和 `src/center/pkg/federated/`

**错误信息**:
```
found packages ai and distributed in same directory
found packages federated and securefl in same directory
```

**修复方案**: 统一目录中的 package 声明

---

### 4. Dashboard Node.js 版本兼容性

**文件**: `src/dashboard/`

**错误信息**:
```
vue-tsc 与 Node.js v24.14.0 不兼容
Search string not found: "/supportedTSExtensions = .*(?=;)/"
```

**修复方案**:
- 降级 Node.js 版本到 v20.x 或 v18.x
- 或升级 vue-tsc 到兼容版本

---

## 影响范围

- Center 服务无法编译
- 部分测试无法运行
- Dashboard 无法构建

---

## 修复计划

| 序号 | 问题 | 优先级 | 复杂度 | 状态 |
|------|------|--------|--------|------|
| 1 | duplicate case | High | Low | ✅ 已修复 |
| 2 | Message 字段缺失 | High | Medium | ✅ 已修复 |
| 3 | 包名冲突 | High | Low | ✅ 已修复 |
| 4 | Node.js 兼容性 | Medium | Low | ✅ 已修复 |

---

## 修复记录

### 修复 1: duplicate case
- **文件**: `src/center/internal/store/memory.go:176`
- **修改**: 移除 `case StoreSQLite, "sqlite":` 中重复的 `"sqlite"`
- **修改后**: `case StoreSQLite:`

### 修复 2: Message 字段缺失
- **文件**: `src/center/internal/models/models.go`
- **修改**: 添加 `Type`, `Status`, `CreatedAt` 字段到 Message 结构体
- **新增**: `MessageType` 和 `MessageStatus` 类型定义

### 修复 3: 包名冲突
- **修改**: 将文件移动到正确的子目录
  - `pkg/ai/distributed.go` → `pkg/ai/distributed/`
  - `pkg/ai/tuning.go` → `pkg/ai/tuning/`
  - `pkg/ai/quantize.go` → `pkg/ai/quantize/`
  - `pkg/ai/version.go` → `pkg/ai/modelversion/`
  - `pkg/federated/secure.go` → `pkg/federated/securefl/`

### 修复 4: Node.js 兼容性
- **文件**: `src/dashboard/package.json`
- **修改**: 升级 `vue-tsc` 从 `^1.8.0` 到 `^2.0.0`
- **新增**: `engines.node >= 18.0.0`

---

## 验证结果

```
✅ go build ./pkg/version/... ./internal/models/... ./internal/store/... 成功
✅ go build ./cmd/center 成功
✅ go test ./internal/store/... ./pkg/auth/... ./pkg/errors/... 成功
```

---

## 相关文件

- `src/center/internal/store/memory.go`
- `src/center/internal/store/sqlite.go`
- `src/center/internal/models/models.go`
- `src/center/pkg/ai/`
- `src/center/pkg/federated/`
- `src/dashboard/package.json`