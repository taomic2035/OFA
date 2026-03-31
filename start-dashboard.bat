@echo off
echo ========================================
echo   OFA Dashboard 启动脚本
echo ========================================
echo.

:: Kill existing processes on ports
echo 清理端口...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do taskkill /F /PID %%a 2>nul
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :9090 ^| findstr LISTENING') do taskkill /F /PID %%a 2>nul
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :3000 ^| findstr LISTENING') do taskkill /F /PID %%a 2>nul
timeout /t 1 /nobreak >nul

:: Start Center
echo 启动 Center 服务...
cd /d D:\vibecoding\OFA\src\center
start "OFA Center" cmd /C "..\..\build\center.exe"
timeout /t 3 /nobreak >nul

:: Check Center
echo 检查 Center 服务...
curl -s http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo [错误] Center 服务启动失败
    pause
    exit /b 1
)
echo [成功] Center 服务已启动 (端口 8080, 9090)

:: Start Dashboard
echo 启动 Dashboard 前端...
cd /d D:\vibecoding\OFA\src\dashboard
echo.
echo ========================================
echo   服务已启动:
echo   - Center: http://localhost:8080
echo   - Dashboard: http://localhost:3000
echo ========================================
echo.
echo 按任意键停止服务...
npm run dev
