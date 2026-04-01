@echo off
cd /d D:\vibecoding\OFA

echo ========================================
echo OFA Git Status Check
echo ========================================

echo.
echo [1] Checking if .git exists...
if exist .git (
    echo .git directory EXISTS
) else (
    echo .git directory DOES NOT EXIST
    echo.
    echo [2] Initializing git repository...
    git init
    git remote add origin https://github.com/taomic2035/OFA.git
)

echo.
echo [3] Git status:
git status

echo.
echo [4] Git remote:
git remote -v

echo.
echo [5] Git branch:
git branch -a

echo.
echo [6] Git log:
git log --oneline -5 2>&1 || echo No commits yet

echo.
echo [7] Adding files...
git add -A

echo.
echo [8] Committing...
git commit -m "feat(android-sdk): add MCP/Tool integration v1.0.0" -m "- MCP protocol, Tools, AI interface" -m "Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"

echo.
echo [9] Pushing...
git branch -M main
git push -u origin main --force

echo.
echo [10] Final status:
git status
git log --oneline -1

echo.
echo ========================================
echo Done!
echo ========================================
pause