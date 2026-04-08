# Start Center and test REST API
$ErrorActionPreference = "Continue"

Write-Host "=== OFA Center REST API Test ===" -ForegroundColor Cyan

# Build center
Write-Host "`n[1] Building Center..." -ForegroundColor Yellow
Set-Location "D:\vibecoding\OFA\src\center"
$env:PATH = "D:\Go\go\bin;$env:PATH"
& go build -o center.exe ./cmd/center 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Build successful!" -ForegroundColor Green

# Start center in background
Write-Host "`n[2] Starting Center service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\center.exe" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 3

# Test health endpoint
Write-Host "`n[3] Testing /health endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)" -ForegroundColor White
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Test TTS voices endpoint
Write-Host "`n[4] Testing /api/v1/tts/voices endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/tts/voices" -UseBasicParsing -TimeoutSec 5
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    $voices = $response.Content | ConvertFrom-Json
    Write-Host "Voices count: $($voices.voices.Count)" -ForegroundColor White
    if ($voices.voices.Count -gt 0) {
        Write-Host "First 3 voices:" -ForegroundColor White
        $voices.voices | Select-Object -First 3 | ForEach-Object {
            Write-Host "  - $($_.name) ($($_.voice_id))"
        }
    }
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Test TTS synthesize endpoint
Write-Host "`n[5] Testing /api/v1/tts/synthesize endpoint..." -ForegroundColor Yellow
try {
    $body = @{
        text = "你好，我是OFA数字人助手"
        voice_id = "zh_female_tianmei"
        format = "mp3"
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/tts/synthesize" `
        -Method POST `
        -Body $body `
        -ContentType "application/json" `
        -UseBasicParsing `
        -TimeoutSec 10

    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    $result = $response.Content | ConvertFrom-Json
    Write-Host "Success: $($result.success)" -ForegroundColor White
    Write-Host "Provider: $($result.provider)" -ForegroundColor White
    Write-Host "Voice used: $($result.voice_used)" -ForegroundColor White
    Write-Host "Duration: $($result.duration_ms)ms" -ForegroundColor White
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

# Test TTS identity voice endpoint
Write-Host "`n[6] Testing TTS identity voice endpoints..." -ForegroundColor Yellow

# Set identity voice
try {
    $body = @{ voice_id = "zh_female_meilinvyou_uranus_bigtts" } | ConvertTo-Json
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/tts/identity/test-user/voice" `
        -Method PUT `
        -Body $body `
        -ContentType "application/json" `
        -UseBasicParsing `
        -TimeoutSec 5
    Write-Host "Set voice: $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "Set voice error: $_" -ForegroundColor Red
}

# Get identity voice
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/tts/identity/test-user/voice" `
        -UseBasicParsing -TimeoutSec 5
    Write-Host "Get voice: $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "Get voice error: $_" -ForegroundColor Red
}

# Stop center
Write-Host "`n[7] Stopping Center..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
Write-Host "Center stopped" -ForegroundColor Green

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan