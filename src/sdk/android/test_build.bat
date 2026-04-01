@echo off
REM OFA Android SDK Build Test Script for Windows
REM Run: test_build.bat

echo === OFA Android SDK Build Test ===
echo.

REM Check Java
echo 1. Checking environment...
java -version 2>&1 | findstr "version"
echo.

REM Go to project directory
cd /d "%~dp0"
echo 2. Project directory: %CD%
echo.

REM Count files
echo 3. Checking source files...
for /f %%i in ('dir /b /s src\main\java\com\ofa\agent\mcp\*.java 2^>nul ^| find /c /v ""') do set MCP_COUNT=%%i
for /f %%i in ('dir /b /s src\main\java\com\ofa\agent\tool\*.java 2^>nul ^| find /c /v ""') do set TOOL_COUNT=%%i
for /f %%i in ('dir /b /s src\main\java\com\ofa\agent\ai\*.java 2^>nul ^| find /c /v ""') do set AI_COUNT=%%i

echo    MCP files: %MCP_COUNT%
echo    Tool files: %TOOL_COUNT%
echo    AI files: %AI_COUNT%
echo.

REM Run Gradle build
echo 4. Running Gradle build...
if exist gradlew (
    call gradlew assembleDebug
) else (
    if exist gradle (
        gradle assembleDebug
    ) else (
        echo Gradle not found. Please install Gradle or use Android Studio.
        exit /b 1
    )
)

if %ERRORLEVEL% equ 0 (
    echo.
    echo === BUILD SUCCESS ===
    echo APK location: build\outputs\apk\debug\
) else (
    echo.
    echo === BUILD FAILED ===
    exit /b 1
)