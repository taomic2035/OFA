# OFA Android SDK 构建脚本 (PowerShell)

param(
    [string]$Task = "assembleRelease"
)

$ErrorActionPreference = "Stop"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "OFA Android SDK Build" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

# 设置环境变量 (可根据实际情况修改)
if (-not $env:JAVA_HOME) { $env:JAVA_HOME = "D:\Java\jdk-17" }
if (-not $env:ANDROID_HOME) { $env:ANDROID_HOME = "D:\Android\Sdk" }
$env:Path = "$env:JAVA_HOME\bin;$env:ANDROID_HOME\platform-tools;$env:Path"

Write-Host "[环境变量]" -ForegroundColor Yellow
Write-Host "JAVA_HOME = $env:JAVA_HOME"
Write-Host "ANDROID_HOME = $env:ANDROID_HOME"

# 检查 Java
Write-Host "`n[检查 Java]" -ForegroundColor Yellow
& "$env:JAVA_HOME\bin\java.exe" -version

# 检查 Android SDK
Write-Host "`n[检查 Android SDK]" -ForegroundColor Yellow
if (Test-Path "$env:ANDROID_HOME\platforms") {
    Write-Host "Android SDK OK" -ForegroundColor Green
} else {
    Write-Host "ERROR: Android SDK not found" -ForegroundColor Red
    exit 1
}

# 运行构建
Write-Host "`n开始构建: $Task" -ForegroundColor Cyan
& ".\gradlew.bat" $Task --no-daemon

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=====================================" -ForegroundColor Green
    Write-Host "构建成功!" -ForegroundColor Green
    Write-Host "=====================================" -ForegroundColor Green

    $aar = Get-ChildItem "sdk\build\outputs\aar\*.aar" -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($aar) {
        Write-Host "AAR: $($aar.FullName)" -ForegroundColor Cyan
    }
} else {
    Write-Host "`n构建失败，退出码: $LASTEXITCODE" -ForegroundColor Red
}