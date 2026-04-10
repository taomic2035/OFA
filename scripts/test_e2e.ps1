# OFA 端到端测试脚本 (Windows PowerShell)
# 测试 Center REST API 的核心功能

param(
    [string]$CenterUrl = "http://localhost:8080",
    [int]$Timeout = 30
)

# 颜色输出函数
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Success "OFA E2E Test Suite"
Write-Host "Center URL: $CenterUrl"
Write-Host ""

# 等待 Center 就绪
function Wait-ForCenter {
    Write-Warning "Waiting for Center..."
    for ($i = 1; $i -le 30; $i++) {
        try {
            $response = Invoke-WebRequest -Uri "$CenterUrl/health" -TimeoutSec 5 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                Write-Success "Center is ready"
                return $true
            }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    Write-Error "Center not ready after 30s"
    return $false
}

# 测试函数
function Test-Case {
    param(
        [string]$Name,
        [string]$Url,
        [string]$Method = "GET",
        [string]$Data = "",
        [int]$Expected = 200
    )

    Write-Warning "Testing: $Name"

    try {
        if ($Method -eq "GET") {
            $response = Invoke-WebRequest -Uri $Url -Method GET -UseBasicParsing
        } else {
            $response = Invoke-WebRequest -Uri $Url -Method POST `
                -ContentType "application/json" `
                -Body $Data `
                -UseBasicParsing
        }

        if ($response.StatusCode -eq $Expected) {
            Write-Success "  ✓ Status: $($response.StatusCode)"
            $bodyPreview = $response.Content.Substring(0, [Math]::Min(100, $response.Content.Length))
            Write-Host "  Response: $bodyPreview"
            return $true
        } else {
            Write-Error "  ✗ Status: $($response.StatusCode) (expected $Expected)"
            Write-Host "  Response: $($response.Content)"
            return $false
        }
    } catch {
        Write-Error "  ✗ Error: $_"
        return $false
    }
}

# 测试计数
$passed = 0
$failed = 0

# 运行测试
function Run-Test {
    param(
        [string]$Name,
        [string]$Url,
        [string]$Method = "GET",
        [string]$Data = "",
        [int]$Expected = 200
    )

    if (Test-Case -Name $Name -Url $Url -Method $Method -Data $Data -Expected $Expected) {
        $passed++
    } else {
        $failed++
    }
}

# 等待 Center
Wait-ForCenter

Write-Host ""
Write-Success "=== Test Suite ==="
Write-Host ""

# 1. 健康检查
Run-Test -Name "Health Check" -Url "$CenterUrl/health" -Expected 200

# 2. 系统信息
Run-Test -Name "System Info" -Url "$CenterUrl/api/v1/system/info" -Expected 200

# 3. 创建身份
Run-Test -Name "Create Identity" -Url "$CenterUrl/api/v1/identities" -Method POST `
    -Data '{"id":"e2e_001","name":"E2E测试用户","personality":{"openness":0.6}}' -Expected 200

# 4. 获取身份
Run-Test -Name "Get Identity" -Url "$CenterUrl/api/v1/identities/e2e_001" -Expected 200

# 5. 设备注册
Run-Test -Name "Register Device" -Url "$CenterUrl/api/v1/devices" -Method POST `
    -Data '{"agent_id":"device_e2e","identity_id":"e2e_001","device_type":"mobile"}' -Expected 200

# 6. 行为上报
Run-Test -Name "Report Behavior" -Url "$CenterUrl/api/v1/behaviors" -Method POST `
    -Data '{"agent_id":"device_e2e","identity_id":"e2e_001","type":"interaction"}' -Expected 200

# 7. 情绪触发
Run-Test -Name "Trigger Emotion" -Url "$CenterUrl/api/v1/emotions/trigger" -Method POST `
    -Data '{"identity_id":"e2e_001","emotion_type":"joy","intensity":0.7}' -Expected 200

# 8. TTS 声音列表
Run-Test -Name "TTS Voices" -Url "$CenterUrl/api/v1/tts/voices" -Expected 200

# 9. Agent 列表
Run-Test -Name "Agent List" -Url "$CenterUrl/api/v1/agents" -Expected 200

# 10. Skill 列表
Run-Test -Name "Skill List" -Url "$CenterUrl/api/v1/skills" -Expected 200

# 结果汇总
Write-Host ""
Write-Success "=== Test Summary ==="
Write-Host "  Passed: $passed"
Write-Host "  Failed: $failed"
Write-Host ""

if ($failed -eq 0) {
    Write-Success "All tests passed!"
    exit 0
} else {
    Write-Error "Some tests failed!"
    exit 1
}