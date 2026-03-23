# GoAgent TaskRunner 启动脚本
Write-Host "================================" -ForegroundColor Cyan
Write-Host "启动 GoAgent TaskRunner" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# 检查配置文件
if (-not (Test-Path "configs\config.json")) {
    Write-Host "创建配置文件..." -ForegroundColor Yellow
    Copy-Item "configs\config.example.json" "configs\config.json" -ErrorAction SilentlyContinue
}

# 检查环境变量
if (-not (Test-Path ".env")) {
    Write-Host "警告：.env 文件不存在，请确保已设置环境变量" -ForegroundColor Red
    Write-Host ""
}

# 启动后端服务
Write-Host "启动后端服务..." -ForegroundColor Green
$backendJob = Start-Job -ScriptBlock {
    Set-Location $using:PWD
    go run ./cmd/server/main.go
}

Write-Host "等待后端启动..." -ForegroundColor Green
Start-Sleep -Seconds 3

# 检查后端是否启动成功
$backendStatus = Get-Job -Id $backendJob.Id
if ($backendStatus.State -eq "Failed") {
    Write-Host "后端启动失败！" -ForegroundColor Red
    Receive-Job -Id $backendJob.Id
    exit 1
}

Write-Host "后端服务已启动 (端口 8080)" -ForegroundColor Green
Write-Host ""

# 启动前端
Write-Host "启动前端开发服务器..." -ForegroundColor Green
Set-Location frontend

# 检查 node_modules
if (-not (Test-Path "node_modules")) {
    Write-Host "安装前端依赖..." -ForegroundColor Yellow
    npm install
}

# 启动 Vite
npm run dev

# 清理
Write-Host ""
Write-Host "停止后端服务..." -ForegroundColor Yellow
Stop-Job -Id $backendJob
Remove-Job -Id $backendJob
Write-Host "再见！" -ForegroundColor Cyan
