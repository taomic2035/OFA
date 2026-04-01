#!/bin/bash
# OFA Android SDK 构建测试脚本
# 运行方式: ./test_build.sh

set -e

echo "=== OFA Android SDK Build Test ==="
echo ""

# 检查环境
echo "1. Checking environment..."
echo "   Java: $(java -version 2>&1 | head -1)"
echo "   Gradle: $(gradle --version 2>&1 | head -3 || echo 'Not installed')"
echo ""

# 进入项目目录
cd "$(dirname "$0")"
echo "2. Project directory: $(pwd)"
echo ""

# 检查文件
echo "3. Checking source files..."
MCP_COUNT=$(find src/main/java/com/ofa/agent/mcp -name "*.java" 2>/dev/null | wc -l)
TOOL_COUNT=$(find src/main/java/com/ofa/agent/tool -name "*.java" 2>/dev/null | wc -l)
AI_COUNT=$(find src/main/java/com/ofa/agent/ai -name "*.java" 2>/dev/null | wc -l)

echo "   MCP files: $MCP_COUNT"
echo "   Tool files: $TOOL_COUNT"
echo "   AI files: $AI_COUNT"
echo ""

# 运行 Gradle 构建
echo "4. Running Gradle build..."
if gradle assembleDebug; then
    echo ""
    echo "=== BUILD SUCCESS ==="
    echo "APK location: build/outputs/apk/debug/"
else
    echo ""
    echo "=== BUILD FAILED ==="
    exit 1
fi