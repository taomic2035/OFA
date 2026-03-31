@echo off
REM OFA 快速启动脚本

echo === OFA 快速启动 ===
echo.

REM 检查Go环境
D:\Go\go\bin\go.exe version >nul 2>&1
if errorlevel 1 (
    echo 错误: Go未安装或路径不正确
    echo 请安装Go到 D:\Go\go\
    exit /b 1
)

REM 设置Go代理
set GOPROXY=https://goproxy.cn,direct
set PATH=D:\Go\go\bin;%PATH%

REM 检查参数
if "%1"=="" goto help
if "%1"=="build" goto build
if "%1"=="test" goto test
if "%1"=="run-center" goto run-center
if "%1"=="run-agent" goto run-agent
if "%1"=="clean" goto clean
goto help

:build
echo 构建 Center...
cd /d D:\vibecoding\OFA\src\center
go build -o ..\..\build\center.exe ./cmd/center
if errorlevel 1 (
    echo Center 构建失败
    exit /b 1
)
echo Center 构建成功

echo 构建 Agent...
cd /d D:\vibecoding\OFA\src\agent\go
go build -o ..\..\..\build\agent.exe ./cmd/agent
if errorlevel 1 (
    echo Agent 构建失败
    exit /b 1
)
echo Agent 构建成功
goto end

:test
echo 运行 Center 测试...
cd /d D:\vibecoding\OFA\src\center
go test ./... -v
echo.
echo 运行 Agent 测试...
cd /d D:\vibecoding\OFA\src\agent\go
go test ./... -v
goto end

:run-center
echo 启动 Center...
cd /d D:\vibecoding\OFA
.\build\center.exe
goto end

:run-agent
echo 启动 Agent...
cd /d D:\vibecoding\OFA
.\build\agent.exe
goto end

:clean
echo 清理构建产物...
del /q D:\vibecoding\OFA\build\*.exe 2>nul
echo 清理完成
goto end

:help
echo.
echo 用法: ofa.bat [命令]
echo.
echo 命令:
echo   build       构建 Center 和 Agent
echo   test        运行测试
echo   run-center  启动 Center 服务
echo   run-agent   启动 Agent 客户端
echo   clean       清理构建产物
echo.
echo 示例:
echo   ofa.bat build
echo   ofa.bat test
echo.

:end