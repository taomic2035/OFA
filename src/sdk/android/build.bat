@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

echo ========================================
echo OFA Android SDK Build Script
echo ========================================
echo.

:: 设置环境变量 (可根据实际情况修改)
if not defined JAVA_HOME set "JAVA_HOME=D:\Java\jdk-17"
if not defined ANDROID_HOME set "ANDROID_HOME=D:\Android\Sdk"
set "PATH=%JAVA_HOME%\bin;%ANDROID_HOME%\platform-tools;%ANDROID_HOME%\cmdline-tools\latest\bin;%PATH%"

:: 验证 Java
echo [1/4] Checking Java...
"%JAVA_HOME%\bin\java" -version 2>&1
if errorlevel 1 (
    echo ERROR: Java not found at %JAVA_HOME%
    exit /b 1
)

:: 验证 Android SDK
echo [2/4] Checking Android SDK...
if not exist "%ANDROID_HOME%\platforms" (
    echo ERROR: Android SDK not found at %ANDROID_HOME%
    exit /b 1
)

:: 编译
echo [3/4] Building SDK (Release)...
call gradlew.bat assembleRelease --no-daemon

echo.
echo [4/4] Build Result:
echo ========================================
if exist "sdk\build\outputs\aar\sdk-release.aar" (
    echo BUILD SUCCESS!
    echo Output: sdk\build\outputs\aar\sdk-release.aar
) else (
    echo BUILD FAILED
)
echo ========================================