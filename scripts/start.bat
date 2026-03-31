@echo off
REM OFA Quick Start Script for Windows

echo ==========================================
echo   OFA - Omni Federated Agents
echo ==========================================
echo.

REM Check command
if "%1"=="" goto help
if "%1"=="start" goto start
if "%1"=="stop" goto stop
if "%1"=="status" goto status
if "%1"=="logs" goto logs
if "%1"=="build" goto build
if "%1"=="test" goto test
if "%1"=="help" goto help

:help
echo Usage: %0 [command]
echo.
echo Commands:
echo   start   Start all services
echo   stop    Stop all services
echo   status  Show service status
echo   logs    View logs
echo   build   Build services
echo   test    Run tests
echo   help    Show this help
goto end

:start
echo Starting services...
docker-compose -f deployments\docker-compose.yaml up -d
echo.
echo Services started!
echo   - REST API: http://localhost:8080
echo   - gRPC API: localhost:9090
goto end

:stop
echo Stopping services...
docker-compose -f deployments\docker-compose.yaml down
echo Services stopped.
goto end

:status
echo Service status:
docker-compose -f deployments\docker-compose.yaml ps
goto end

:logs
docker-compose -f deployments\docker-compose.yaml logs -f
goto end

:build
echo Building services...
if exist build\center.exe (
    del build\center.exe
)
if exist build\agent.exe (
    del build\agent.exe
)

REM Check Go
where go >nul 2>nul
if %errorlevel%==0 (
    echo Building Center...
    cd src\center
    go build -o ..\..\build\center.exe .\cmd\center
    cd ..\..

    echo Building Agent...
    cd src\agent\go
    go build -o ..\..\..\build\agent.exe .\cmd\agent
    cd ..\..\..

    echo Build complete!
) else (
    echo Go not installed, using Docker build...
    docker-compose -f deployments\docker-compose.yaml build
)
goto end

:test
echo Running tests...
where go >nul 2>nul
if %errorlevel%==0 (
    echo Center tests...
    cd src\center
    go test -v ./...
    cd ..\..

    echo Agent tests...
    cd src\agent\go
    go test -v ./...
    cd ..\..\..

    echo Tests complete!
) else (
    echo Go not installed!
)
goto end

:end