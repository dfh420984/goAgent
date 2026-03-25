@echo off
chcp 65001 >nul
echo.
echo 🚀 启动 Agent 服务器（带 MCP 集成）
echo ===================================
echo.

REM 检查 Node.js
where npx >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 未找到 Node.js/npx
    echo.
    echo 💡 MCP 功能需要 Node.js，请安装：https://nodejs.org/
    echo 💡 或者移除 mcp-servers.json 以禁用 MCP 功能
    echo.
    pause
    exit /b 1
)

echo ✅ Node.js 已安装
echo.

REM 检查 MCP 配置
if not exist "mcp-servers.json" (
    echo ⚠️  未找到 mcp-servers.json，MCP 功能将被禁用
    echo.
) else (
    echo 📋 检测到 MCP 配置文件
    echo.
)

echo 🏃 编译并启动服务器...
echo.
go run cmd/server/main.go

echo.
pause
