@echo off
echo ================================
echo 启动 GoAgent TaskRunner
echo ================================
echo.

REM 检查配置文件
if not exist "configs\config.json" (
    echo 创建配置文件...
    copy "configs\config.example.json" "configs\config.json"
)

REM 启动后端
echo 启动后端服务...
start "GoAgent Backend" cmd /k "go run cmd\server\main.go"

REM 等待后端启动
echo 等待后端启动...
timeout /t 5 /nobreak >nul

REM 进入前端目录
cd frontend

REM 检查 node_modules
if not exist "node_modules" (
    echo 安装前端依赖...
    call npm install
)

REM 启动前端
echo 启动前端开发服务器...
call npm run dev

echo.
echo 按任意键停止服务...
pause >nul
