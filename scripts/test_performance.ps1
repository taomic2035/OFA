# OFA 性能测试脚本 (Windows PowerShell)
# 测试 Center REST API 性能和各组件性能

param(
    [string]$CenterUrl = "http://localhost:8080",
    [int]$Concurrency = 10,
    [int]$Requests = 1000,
    [int]$DurationSeconds = 30
)

# 颜色输出函数
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Info { Write-Host $args -ForegroundColor Blue }

Write-Success "OFA 性能测试套件"
Write-Host "Center URL: $CenterUrl"
Write-Host "并发数: $Concurrency"
Write-Host "请求数: $Requests"
Write-Host ""

# 等待 Center 就绪
function Wait-ForCenter {
    Write-Warning "等待 Center就绪..."
    for ($i = 1; $i -le 30; $i++) {
        try {
            $response = Invoke-WebRequest -Uri "$CenterUrl/health" -TimeoutSec 5 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                Write-Success "Center 已就绪"
                return $true
            }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    Write-Error "Center 未就绪"
    return $false
}

# 运行 Go 基准测试
function Run-GoBenchmarks {
    Write-Info "=== Go 基准测试 ==="
    Write-Host ""

    Push-Location "src/center"

    # 缓存性能测试
    Write-Warning "缓存性能测试"
    go test -bench=BenchmarkLocalCache -benchmem ./pkg/cache/...

    # 身份存储性能测试
    Write-Warning "身份存储性能测试"
    go test -bench=BenchmarkIdentity -benchmem ./internal/identity/...

    Pop-Location
    Write-Host ""
}

# 运行压力测试
function Run-StressTest {
    Write-Info "=== 压力测试 ==="
    Write-Host ""

    Push-Location "src/center"

    Write-Warning "本地缓存压力测试"
    go test -v -run TestLocalCachePerformance ./pkg/cache/...

    Write-Warning "身份存储压力测试"
    go test -v -run TestIdentityStorePerformance ./internal/identity/...

    Write-Warning "并发身份同步测试"
    go test -v -run TestConcurrentIdentitySync ./internal/identity/...

    Pop-Location
    Write-Host ""
}

# 运行 HTTP 负载测试
function Run-HttpLoadTest {
    Write-Info "=== HTTP 负载测试 ==="
    Write-Host ""

    Write-Warning "健康检查端点负载测试"
    Write-Host "URL: $CenterUrl/health"
    Write-Host "并发: $Concurrency, 请求数: $Requests"

    $startTime = Get-Date
    $jobs = @()

    for ($i = 1; $i -le $Concurrency; $i++) {
        $jobs += Start-Job -ScriptBlock {
            param($url, $count)
            for ($j = 1; $j -le $count; $j++) {
                try {
                    Invoke-WebRequest -Uri $url -TimeoutSec 5 -UseBasicParsing | Out-Null
                } catch {}
            }
        } -ArgumentList "$CenterUrl/health", ($Requests / $Concurrency)
    }

    Wait-Job -Job $jobs | Out-Null
    Remove-Job -Job $jobs

    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalSeconds
    $throughput = $Requests / $duration

    Write-Success "负载测试完成"
    Write-Host "总请求数: $Requests"
    Write-Host "持续时间: $duration s"
    Write-Host "吞吐量: $throughput req/s"
    Write-Host ""
}

# 运行 API 响应时间测试
function Run-ApiLatencyTest {
    Write-Info "=== API 响应时间测试 ==="
    Write-Host ""

    # 创建身份响应时间
    Write-Warning "创建身份 API"
    for ($i = 1; $i -le 10; $i++) {
        $start = Get-Date
        try {
            Invoke-WebRequest -Uri "$CenterUrl/api/v1/identities" -Method POST `
                -ContentType "application/json" `
                -Body "{\"id\":\"perf_test_$i\",\"name\":\"测试用户$i\"}" `
                -UseBasicParsing | Out-Null
        } catch {}
        $end = Get-Date
        $latency = ($end - $start).TotalMilliseconds
        Write-Host "请求$i: $latency ms"
    }
    Write-Host ""

    # 获取身份响应时间
    Write-Warning "获取身份 API"
    for ($i = 1; $i -le 10; $i++) {
        $start = Get-Date
        try {
            Invoke-WebRequest -Uri "$CenterUrl/api/v1/identities/perf_test_$i" -UseBasicParsing | Out-Null
        } catch {}
        $end = Get-Date
        $latency = ($end - $start).TotalMilliseconds
        Write-Host "请求$i: $latency ms"
    }
    Write-Host ""
}

# 结果汇总
function Print-Summary {
    Write-Success "=== 性能测试汇总 ==="
    Write-Host ""
    Write-Host "测试项:"
    Write-Host "  - Go 基准测试"
    Write-Host "  - 压力测试"
    Write-Host "  - HTTP 负载测试"
    Write-Host "  - API 响应时间测试"
    Write-Host ""
    Write-Host "建议:"
    Write-Host "  - 缓存 ops/sec > 100000"
    Write-Host "  - 身份存储 ops/sec > 1000"
    Write-Host "  - API 响应时间 < 100ms"
    Write-Host "  - 吞吐量 > 500 req/s"
    Write-Host ""
}

# 主流程
function Main {
    Wait-ForCenter

    Run-GoBenchmarks
    Run-StressTest
    Run-HttpLoadTest
    Run-ApiLatencyTest

    Print-Summary

    Write-Success "性能测试完成!"
}

Main