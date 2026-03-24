@echo off
echo ================================
echo SQLite 数据库初始化
echo ================================
echo.

REM 检查是否存在数据库文件
if exist "data.db" (
    echo ⚠️  警告：数据库文件已存在
    set /p confirm="是否要重新创建？(y/n): "
    if /i "%confirm%"=="y" (
        echo 删除旧数据库文件...
        del /q data.db
    ) else (
        echo 操作已取消
        pause
        exit /b 0
    )
)

echo.
echo 正在创建数据库...
go run cmd\init_db\main.go

if errorlevel 1 (
    echo.
    echo ❌ 数据库创建失败！
    pause
    exit /b 1
)

echo.
echo ================================
echo ✅ 数据库初始化完成！
echo ================================
echo.
echo 📁 数据库文件：data.db
echo.
echo 可用的管理命令:
echo   - PowerShell: .\db_manager.ps1
echo   - 直接查询：sqlite3 data.db
echo.

pause
