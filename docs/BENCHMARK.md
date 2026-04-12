# OFA Center 性能压测文档 (v9.0.0)

## 一、压测框架概述

OFA Center 提供完整的性能压测框架，支持：
- WebSocket 连接压测
- 身份操作压测
- 场景检测压测
- Chat/LLM 对话压测
- TTS 音频生成压测

## 二、压测配置

### 默认配置

| 压测类型 | 持续时间 | 并发数 | 热身时间 |
|---------|---------|--------|---------|
| WebSocket 连接 | 30s | 10 | 5s |
| 身份操作 | 30s | 20 | 5s |
| Chat 对话 | 60s | 5 | 10s |
| 场景检测 | 30s | 15 | 5s |

### 性能阈值

| 指标 | 默认阈值 |
|------|---------|
| 平均延迟 | < 100ms |
| P99 延迟 | < 500ms |
| 吞吐量 | > 100 req/s |
| 错误率 | < 1% |

## 三、运行压测

### 使用 Go 测试

```bash
cd src/center

# 运行所有压测
go test -bench=. ./internal/benchmark/...

# 运行特定压测
go test -bench=WebSocket ./internal/benchmark/...
go test -bench=Identity ./internal/benchmark/...
go test -bench=Chat ./internal/benchmark/...

# 指定压测时间
go test -bench=. -benchtime=60s ./internal/benchmark/...
```

### 使用脚本

```bash
# 启动 Center 服务
go run cmd/center/main.go

# 运行压测脚本
./scripts/benchmark/run_benchmark.sh
```

## 四、压测结果

### 结果指标

| 指标 | 说明 |
|------|------|
| Total Requests | 总请求数 |
| Success Count | 成功数 |
| Failure Count | 失败数 |
| Error Rate | 错误率 (%) |
| Throughput | 吞吐量 |
| Avg Latency | 平均延迟 |
| Min Latency | 最小延迟 |
| Max Latency | 最大延迟 |
| P50 Latency | 50分位延迟 |
| P90 Latency | 90分位延迟 |
| P99 Latency | 99分位延迟 |

### 结果示例

```
========================================
Benchmark Results: WebSocket_Connection
========================================
Duration:       30s
Total Requests: 3000
Success:        2985
Failures:       15
Error Rate:     0.50%
Throughput:     100.00 req/s

Latency Distribution:
  Average:      25ms
  Min:          5ms
  Max:          150ms
  P50:          20ms
  P90:          45ms
  P99:          80ms
========================================
```

## 五、性能优化建议

### WebSocket 性能

1. **连接池**: 使用预创建连接减少握手开销
2. **消息批处理**: 合并小消息减少网络往返
3. **压缩**: 启用 WebSocket 压缩减少带宽

### 身份操作性能

1. **缓存**: 使用 Redis 缓存热点身份
2. **批量操作**: 支持批量创建/更新身份
3. **索引**: 优化数据库索引

### Chat/LLM 性能

1. **流式响应**: 使用 SSE 流式返回
2. **缓存**: 缓存常见问答
3. **异步处理**: 异步队列处理请求

### 场景检测性能

1. **定时检测**: 降低检测频率
2. **缓存场景**: 缓存检测结果
3. **事件驱动**: 仅在状态变化时触发

## 六、监控集成

压测结果可集成到 Prometheus/Grafana：

```yaml
# Prometheus 指标
ofa_benchmark_requests_total{operation="websocket"}
ofa_benchmark_latency_avg{operation="identity"}
ofa_benchmark_throughput{operation="chat"}
ofa_benchmark_error_rate{operation="scene"}
```

---

*文档版本: v9.0.0*
*更新时间: 2026-04-12*
