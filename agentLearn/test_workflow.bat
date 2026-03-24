@echo off
echo ================================
echo 工作流工具测试
echo ================================
echo.

REM 检查配置文件
if not exist "configs\config.json" (
    echo 创建配置文件...
    copy "configs\config.example.json" "configs\config.json"
)

REM 检查环境变量
if not exist ".env" (
    echo 警告：.env 文件不存在，请设置 API Key
    echo.
)

echo 编译项目...
go build ./...
if errorlevel 1 (
    echo 编译失败！
    pause
    exit /b 1
)

echo.
echo 可用的工作流:
echo   1. report_gen.yaml - 报表生成工作流
echo   2. data_export.yaml - 数据导出工作流
echo.

set /p choice="请选择要运行的工作流 (1 或 2，默认 1): "

if "%choice%"=="" set choice=1

if "%choice%"=="1" (
    echo.
    echo 运行报表生成工作流...
    go run cmd\workflow\main.go configs\workflows\report_gen.yaml
) else if "%choice%"=="2" (
    echo.
    echo 运行数据导出工作流...
    go run cmd\workflow\main.go configs\workflows\data_export.yaml
) else (
    echo 无效的选择
)

echo.
echo 按任意键退出...
pause >nul
