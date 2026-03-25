@echo off
chcp 65001 >nul
echo.
echo 🚀 MCP 集成演示 - 快速启动
echo ===========================
echo.

REM 检查 Node.js
where npx >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 未找到 Node.js/npx
    echo.
    echo 💡 请先安装 Node.js: https://nodejs.org/
    echo.
    pause
    exit /b 1
)

echo ✅ Node.js 已安装
echo.

REM 安装 MCP 服务器（如果尚未安装）
echo 📦 检查 MCP 服务器...
npx @modelcontextprotocol/server-filesystem --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo 📥 正在安装 MCP 文件系统服务器...
    call npm install -g @modelcontextprotocol/server-filesystem
) else (
    echo ✅ MCP 文件系统服务器已安装
)

echo.
echo 🏃 运行演示程序...
echo.
go run cmd/mcp_demo/main.go

echo.
pause
