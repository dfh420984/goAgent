@echo off
chcp 65001 >nul
echo.
echo 🔧 MCP 工具集成到工作流 - 演示
echo ================================
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
echo 🏃 运行演示程序...
echo.
go run cmd/workflow_mcp_demo/main.go

echo.
pause
