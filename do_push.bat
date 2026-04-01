@echo off
cd /d D:\vibecoding\OFA
echo === Git Status ===
git status --short
echo.
echo === Adding files ===
git add -A
echo.
echo === Committing ===
git commit -m "feat(sdk): add offline capability for C++/Go/Desktop SDK" -m "- C++ SDK: offline/p2p/constraint modules" -m "- Go Agent SDK: offline/p2p/constraint modules" -m "- Desktop SDK: offline/p2p/constraint modules" -m "Co-Authored-By: Claude ^<noreply@anthropic.com^>"
echo.
echo === Pushing ===
git push origin main
echo.
echo === Done ===
pause