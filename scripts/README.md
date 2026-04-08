# OFA 脚本目录

本目录包含项目的构建、测试和部署脚本。

## 脚本列表

### 测试脚本

| 脚本 | 说明 |
|------|------|
| `test_api.ps1` | API 功能测试 (健康检查、Agent列表、技能列表、任务测试) |
| `test_metrics.ps1` | Prometheus 指标测试 |
| `test_tts_api.ps1` | TTS REST API 测试 (语音合成、声音列表、身份声音映射) |
| `test_tts_api.bat` | TTS REST API 测试 (Windows批处理版本) |

### 构建脚本

| 脚本 | 说明 |
|------|------|
| `build_android.ps1` | Android SDK 构建 (需要 JAVA_HOME, ANDROID_HOME) |
| `gen_proto.sh` | Protocol Buffers 代码生成 |
| `start.bat` | 启动脚本 (Windows) |
| `start.sh` | 启动脚本 (Linux/macOS) |
| `ofa.bat` | OFA 命令行工具 (Windows) |

## 使用方法

### 前提条件
1. Center 服务已启动: `.\build\center.exe`
2. 环境变量已配置 (Android 构建需要 JAVA_HOME, ANDROID_HOME)

### 运行测试

```powershell
# API 功能测试
powershell -File test_api.ps1

# Prometheus 指标测试
powershell -File test_metrics.ps1

# TTS REST API 测试
powershell -File test_tts_api.ps1
```

### Android SDK 构建

```powershell
powershell -File build_android.ps1
# 或指定构建任务
powershell -File build_android.ps1 -Task "assembleDebug"
```