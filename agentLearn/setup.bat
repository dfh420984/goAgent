@echo off
echo ================================
echo GoAgent TaskRunner 快速启动
echo ================================
echo.

REM 复制环境变量配置
if not exist .env (
    echo 创建 .env 文件...
    copy .env.example .env
)

REM 提示用户
echo.
echo 请编辑 .env 文件，设置你的 API Key:
echo   OPENAI_API_KEY=your_api_key_here
echo.
echo 配置完成后，双击 start.bat 启动服务
pause
