@echo off
cd /d D:\vibecoding\OFA\src\center
set PATH=D:\Go\go\bin;%PATH%
set GOPROXY=https://goproxy.cn,direct
echo ===== Scheduler Tests ===== > D:\vibecoding\OFA\src\center\test_output.txt
go test ./internal/scheduler/... -v >> D:\vibecoding\OFA\src\center\test_output.txt 2>&1
echo. >> D:\vibecoding\OFA\src\center\test_output.txt
echo ===== Version Build ===== >> D:\vibecoding\OFA\src\center\test_output.txt
go build ./pkg/version/... >> D:\vibecoding\OFA\src\center\test_output.txt 2>&1
echo Build result: %ERRORLEVEL% >> D:\vibecoding\OFA\src\center\test_output.txt
echo. >> D:\vibecoding\OFA\src\center\test_output.txt
echo ===== Center Build ===== >> D:\vibecoding\OFA\src\center\test_output.txt
go build ./cmd/center >> D:\vibecoding\OFA\src\center\test_output.txt 2>&1
echo Build result: %ERRORLEVEL% >> D:\vibecoding\OFA\src\center\test_output.txt
echo DONE >> D:\vibecoding\OFA\src\center\test_output.txt