# OFA 测试脚本 (PowerShell)
# 用于测试Center和Agent的功能

$baseUrl = "http://localhost:8080"

Write-Host "=== OFA API 测试 ===" -ForegroundColor Green

# 1. 健康检查
Write-Host "`n1. 健康检查" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "状态: $($health.status)" -ForegroundColor Cyan
    Write-Host "版本: $($health.version)" -ForegroundColor Cyan
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
    Write-Host "请确保Center已启动: .\build\center.exe" -ForegroundColor Red
    exit 1
}

# 2. 获取系统信息
Write-Host "`n2. 系统信息" -ForegroundColor Yellow
try {
    $info = Invoke-RestMethod -Uri "$baseUrl/api/v1/system/info" -Method GET
    Write-Host "在线Agent数: $($info.agent_count)" -ForegroundColor Cyan
    Write-Host "待处理任务: $($info.task_count)" -ForegroundColor Cyan
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

# 3. 获取Agent列表
Write-Host "`n3. Agent列表" -ForegroundColor Yellow
try {
    $agents = Invoke-RestMethod -Uri "$baseUrl/api/v1/agents" -Method GET
    Write-Host "总数: $($agents.total)" -ForegroundColor Cyan
    foreach ($agent in $agents.agents) {
        Write-Host "  - $($agent.agent_id): 状态=$($agent.status)" -ForegroundColor White
    }
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

# 4. 获取技能列表
Write-Host "`n4. 技能列表" -ForegroundColor Yellow
try {
    $skills = Invoke-RestMethod -Uri "$baseUrl/api/v1/skills" -Method GET
    Write-Host "可用技能:" -ForegroundColor Cyan
    foreach ($skill in $skills.skills) {
        Write-Host "  - $($skill.id): $($skill.name) ($($skill.category))" -ForegroundColor White
    }
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

# 5. 提交文本处理任务
Write-Host "`n5. 提交任务测试" -ForegroundColor Yellow
$inputJson = '{"text":"hello world","operation":"uppercase"}'
$inputBase64 = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($inputJson))

$taskBody = @{
    skill_id = "text.process"
    input = $inputBase64
    priority = 0
} | ConvertTo-Json

try {
    $task = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks" -Method POST -ContentType "application/json" -Body $taskBody
    Write-Host "任务已提交: $($task.task_id)" -ForegroundColor Cyan

    # 查询任务状态
    Start-Sleep -Seconds 1
    $status = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks/$($task.task_id)" -Method GET
    if ($status.success) {
        Write-Host "任务状态: $($status.task.status)" -ForegroundColor Cyan
        if ($status.task.output) {
            $output = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($status.task.output))
            Write-Host "输出: $output" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

# 6. 计算器任务测试
Write-Host "`n6. 计算器任务测试" -ForegroundColor Yellow
$calcInput = '{"operation":"add","a":10,"b":25}'
$calcBase64 = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($calcInput))

$calcBody = @{
    skill_id = "calculator"
    input = $calcBase64
} | ConvertTo-Json

try {
    $calcTask = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks" -Method POST -ContentType "application/json" -Body $calcBody
    Write-Host "计算器任务已提交: $($calcTask.task_id)" -ForegroundColor Cyan

    Start-Sleep -Seconds 1
    $calcStatus = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks/$($calcTask.task_id)" -Method GET
    if ($calcStatus.success) {
        Write-Host "任务状态: $($calcStatus.task.status)" -ForegroundColor Cyan
        if ($calcStatus.task.output) {
            $output = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($calcStatus.task.output))
            Write-Host "输出: $output" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green