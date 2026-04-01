@echo off
REM OFA Android SDK v1.0.0 Git Commit Script
REM Run this script from the OFA root directory

echo === OFA Android SDK v1.0.0 Git Commit ===
echo.

REM Check if we're in a git repository
git rev-parse --is-inside-work-tree >nul 2>&1
if errorlevel 1 (
    echo Initializing git repository...
    git init
    git remote add origin https://github.com/taomic2035/OFA.git
)

echo 1. Staging files...
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

echo.
echo 2. Creating commit...
git commit -m "feat(android-sdk): add MCP/Tool integration v1.0.0

- Add MCP protocol layer (MCPServer, MCPClient, ToolDefinition)
- Add Tool system (ToolRegistry, ToolExecutor, PermissionManager)
- Add 30+ built-in tools (system, device, data, AI)
- Add AI Agent interface with OpenAI-compatible adapter
- Integrate offline execution with L1-L4 support
- Update OFAAgent with tool calling capabilities

New features:
- MCPServer for AI agent tool interactions
- 30+ device tools (camera, bluetooth, wifi, sensors, etc.)
- Data tools (contacts, calendar, media)
- AI tools (speech synthesis)
- ToolCallingAdapter for OpenAI function calling format

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"

echo.
echo 3. Pushing to remote...
git push origin main

echo.
echo === Done! ===