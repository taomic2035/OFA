# OFA Metrics 端点测试脚本

Write-Host "=== OFA Prometheus Metrics 测试 ===" -ForegroundColor Cyan
Write-Host ""

# 测试健康检查
Write-Host "1. 健康检查" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -ErrorAction Stop
    Write-Host "状态: $($health.status)" -ForegroundColor Green
    Write-Host "版本: $($health.version)" -ForegroundColor Green
} catch {
    Write-Host "错误: Center服务未启动或端口8080不可访问" -ForegroundColor Red
    Write-Host "请先启动Center: .\build\center.exe" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# 测试metrics端点
Write-Host "2. Prometheus Metrics 端点" -ForegroundColor Yellow
try {
    $metrics = Invoke-WebRequest -Uri "http://localhost:8080/metrics" -UseBasicParsing -ErrorAction Stop
    Write-Host "状态码: $($metrics.StatusCode)" -ForegroundColor Green

    # 解析并显示关键指标
    $content = $metrics.Content
    Write-Host ""
    Write-Host "=== Agent 指标 ===" -ForegroundColor Cyan
    $content -split "`n" | Where-Object { $_ -match "^ofa_agents" } | ForEach-Object { Write-Host $_ }

    Write-Host ""
    Write-Host "=== Task 指标 ===" -ForegroundColor Cyan
    $content -split "`n" | Where-Object { $_ -match "^ofa_tasks" } | ForEach-Object { Write-Host $_ }

    Write-Host ""
    Write-Host "=== 系统指标 ===" -ForegroundColor Cyan
    $content -split "`n" | Where-Object { $_ -match "^ofa_(request|health|grpc)" } | ForEach-Object { Write-Host $_ }

    Write-Host ""
    Write-Host "=== Go运行时指标 (示例) ===" -ForegroundColor Cyan
    $content -split "`n" | Where-Object { $_ -match "^go_goroutines" } | ForEach-Object { Write-Host $_ }

} catch {
    Write-Host "错误: 无法获取metrics端点" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "使用方法:" -ForegroundColor Yellow
Write-Host "  1. 启动Center服务: .\build\center.exe"
Write-Host "  2. 运行此脚本: powershell -File .\scripts\test_metrics.ps1"
Write-Host "  3. 配置Prometheus抓取: http://localhost:8080/metrics"