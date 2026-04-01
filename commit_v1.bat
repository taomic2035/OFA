@echo off
chcp 65001 >nul
cd /d D:\vibecoding\OFA

echo ========================================
echo OFA Android SDK v1.0.0 Git Commit
echo ========================================
echo.

echo [1/4] Checking git status...
git status
echo.

echo [2/4] Adding files...
git add src/sdk/android/src/main/java/com/ofa/agent/mcp/
git add src/sdk/android/src/main/java/com/ofa/agent/tool/
git add src/sdk/android/src/main/java/com/ofa/agent/ai/
git add src/sdk/android/src/main/java/com/ofa/agent/OFAAgent.java
git add src/sdk/android/build.gradle
git add src/sdk/android/README.md
git add src/sdk/android/CHANGELOG.md
git add src/sdk/android/docs/MCP_TOOLS_GUIDE.md
git add src/sdk/android/test_build.bat
git add src/sdk/android/test_build.sh
git add README.md
echo Files added.
echo.

echo [3/4] Creating commit...
git commit -m "feat(android-sdk): add MCP/Tool integration v1.0.0

- Add MCP protocol layer (MCPServer, MCPClient, ToolDefinition)
- Add Tool system (ToolRegistry, ToolExecutor, PermissionManager)
- Add 30+ built-in tools (system, device, data, AI)
- Add AI Agent interface with OpenAI-compatible adapter
- Integrate offline execution with L1-L4 support
- Update OFAAgent with tool calling capabilities

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
echo.

echo [4/4] Pushing to remote...
git push origin main
echo.

echo ========================================
echo Done!
echo ========================================
pause