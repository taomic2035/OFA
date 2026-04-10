#!/bin/bash
# OFA 端到端测试脚本
# 测试 Center REST API 的核心功能

set -e

# 配置
CENTER_URL="${CENTER_URL:-http://localhost:8080}"
TIMEOUT=30

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}OFA E2E Test Suite${NC}"
echo "Center URL: ${CENTER_URL}"
echo ""

# 等待 Center 就绪
wait_for_center() {
    echo -e "${YELLOW}Waiting for Center...${NC}"
    for i in {1..30}; do
        if curl -s "${CENTER_URL}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}Center is ready${NC}"
            return 0
        fi
        sleep 1
    done
    echo -e "${RED}Center not ready after 30s${NC}"
    return 1
}

# 测试函数
test_case() {
    local name="$1"
    local url="$2"
    local method="$3"
    local data="$4"
    local expected="$5"

    echo -e "${YELLOW}Testing: ${name}${NC}"

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "${url}" 2>&1)
    else
        response=$(curl -s -w "\n%{http_code}" -X "${method}" "${url}" \
            -H "Content-Type: application/json" \
            -d "${data}" 2>&1)
    fi

    status=$(echo "$response" | tail -1)
    body=$(echo "$response" | head -n -1)

    if [ "$status" = "$expected" ]; then
        echo -e "${GREEN}  ✓ Status: ${status}${NC}"
        echo "  Response: ${body:0:100}"
    else
        echo -e "${RED}  ✗ Status: ${status} (expected ${expected})${NC}"
        echo "  Response: ${body}"
        return 1
    fi
    return 0
}

# 测试计数
passed=0
failed=0

# 运行测试
run_test() {
    if test_case "$@"; then
        ((passed++))
    else
        ((failed++))
    fi
}

# 等待 Center
wait_for_center

echo ""
echo -e "${GREEN}=== Test Suite ===${NC}"
echo ""

# 1. 健康检查
run_test "Health Check" "${CENTER_URL}/health" "GET" "" "200"

# 2. 系统信息
run_test "System Info" "${CENTER_URL}/api/v1/system/info" "GET" "" "200"

# 3. 创建身份
run_test "Create Identity" "${CENTER_URL}/api/v1/identities" "POST" \
    '{"id":"e2e_001","name":"E2E测试用户","personality":{"openness":0.6}}' "200"

# 4. 获取身份
run_test "Get Identity" "${CENTER_URL}/api/v1/identities/e2e_001" "GET" "" "200"

# 5. 设备注册
run_test "Register Device" "${CENTER_URL}/api/v1/devices" "POST" \
    '{"agent_id":"device_e2e","identity_id":"e2e_001","device_type":"mobile"}' "200"

# 6. 行为上报
run_test "Report Behavior" "${CENTER_URL}/api/v1/behaviors" "POST" \
    '{"agent_id":"device_e2e","identity_id":"e2e_001","type":"interaction"}' "200"

# 7. 情绪触发
run_test "Trigger Emotion" "${CENTER_URL}/api/v1/emotions/trigger" "POST" \
    '{"identity_id":"e2e_001","emotion_type":"joy","intensity":0.7}' "200"

# 8. TTS 声音列表
run_test "TTS Voices" "${CENTER_URL}/api/v1/tts/voices" "GET" "" "200"

# 9. Agent 列表
run_test "Agent List" "${CENTER_URL}/api/v1/agents" "GET" "" "200"

# 10. Skill 列表
run_test "Skill List" "${CENTER_URL}/api/v1/skills" "GET" "" "200"

# 结果汇总
echo ""
echo -e "${GREEN}=== Test Summary ===${NC}"
echo -e "  Passed: ${passed}"
echo -e "  Failed: ${failed}"
echo ""

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi