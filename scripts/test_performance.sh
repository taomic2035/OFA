#!/bin/bash
# OFA 性能测试脚本
# 测试 Center REST API 性能和各组件性能

set -e

# 配置
CENTER_URL="${CENTER_URL:-http://localhost:8080}"
DURATION="${DURATION:-30s}"
CONCURRENCY="${CONCURRENCY:-10}"
REQUESTS="${REQUESTS:-1000}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}OFA 性能测试套件${NC}"
echo "Center URL: ${CENTER_URL}"
echo "持续时间: ${DURATION}"
echo "并发数: ${CONCURRENCY}"
echo "请求数: ${REQUESTS}"
echo ""

# 等待 Center 就绪
wait_for_center() {
    echo -e "${YELLOW}等待 Center就绪...${NC}"
    for i in {1..30}; do
        if curl -s "${CENTER_URL}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}Center 已就绪${NC}"
            return 0
        fi
        sleep 1
    done
    echo -e "${RED}Center 未就绪${NC}"
    return 1
}

# 运行 Go 基准测试
run_go_benchmarks() {
    echo -e "${BLUE}=== Go 基准测试 ===${NC}"
    echo ""

    cd src/center

    # 缓存性能测试
    echo -e "${YELLOW}缓存性能测试${NC}"
    go test -bench=BenchmarkLocalCache -benchmem ./pkg/cache/...

    # 身份存储性能测试
    echo -e "${YELLOW}身份存储性能测试${NC}"
    go test -bench=BenchmarkIdentity -benchmem ./internal/identity/...

    cd -
    echo ""
}

# 运行压力测试
run_stress_test() {
    echo -e "${BLUE}=== 压力测试 ===${NC}"
    echo ""

    # 使用 Go 测试运行性能测试
    cd src/center

    echo -e "${YELLOW}本地缓存压力测试${NC}"
    go test -v -run TestLocalCachePerformance ./pkg/cache/...

    echo -e "${YELLOW}身份存储压力测试${NC}"
    go test -v -run TestIdentityStorePerformance ./internal/identity/...

    echo -e "${YELLOW}并发身份同步测试${NC}"
    go test -v -run TestConcurrentIdentitySync ./internal/identity/...

    cd -
    echo ""
}

# 运行 HTTP 负载测试
run_http_load_test() {
    echo -e "${BLUE}=== HTTP 负载测试 ===${NC}"
    echo ""

    # 健康检查负载测试
    echo -e "${YELLOW}健康检查端点负载测试${NC}"
    echo "URL: ${CENTER_URL}/health"
    echo "并发: ${CONCURRENCY}, 请求数: ${REQUESTS}"

    start_time=$(date +%s.%N)

    for i in $(seq 1 ${CONCURRENCY}); do
        (
            for j in $(seq 1 $((REQUESTS / CONCURRENCY))); do
                curl -s "${CENTER_URL}/health" > /dev/null 2>&1
            done
        ) &
    done

    wait

    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    throughput=$(echo "${REQUESTS} / ${duration}" | bc)

    echo -e "${GREEN}负载测试完成${NC}"
    echo "总请求数: ${REQUESTS}"
    echo "持续时间: ${duration}s"
    echo "吞吐量: ${throughput} req/s"
    echo ""
}

# 运行 API 响应时间测试
run_api_latency_test() {
    echo -e "${BLUE}=== API 响应时间测试 ===${NC}"
    echo ""

    # 创建身份响应时间
    echo -e "${YELLOW}创建身份 API${NC}"
    for i in {1..10}; do
        start=$(date +%s.%N)
        curl -s -X POST "${CENTER_URL}/api/v1/identities" \
            -H "Content-Type: application/json" \
            -d "{\"id\":\"perf_test_${i}\",\"name\":\"测试用户${i}\"}" \
            > /dev/null 2>&1
        end=$(date +%s.%N)
        latency=$(echo "$end - $start" | bc)
        echo " 请求${i}: ${latency}s"
    done
    echo ""

    # 获取身份响应时间
    echo -e "${YELLOW}获取身份 API${NC}"
    for i in {1..10}; do
        start=$(date +%s.%N)
        curl -s "${CENTER_URL}/api/v1/identities/perf_test_${i}" > /dev/null 2>&1
        end=$(date +%s.%N)
        latency=$(echo "$end - $start" | bc)
        echo "请求${i}: ${latency}s"
    done
    echo ""
}

# 结果汇总
print_summary() {
    echo -e "${GREEN}=== 性能测试汇总 ===${NC}"
    echo ""
    echo "测试项:"
    echo "  - Go 基准测试"
    echo "  - 压力测试"
    echo "  - HTTP 负载测试"
    echo "  - API 响应时间测试"
    echo ""
    echo "建议:"
    echo "  - 缓存 ops/sec > 100000"
    echo "  - 身份存储 ops/sec > 1000"
    echo "  - API 响应时间 < 100ms"
    echo "  - 吞吐量 > 500 req/s"
    echo ""
}

# 主流程
main() {
    wait_for_center

    run_go_benchmarks
    run_stress_test
    run_http_load_test
    run_api_latency_test

    print_summary

    echo -e "${GREEN}性能测试完成!${NC}"
}

main