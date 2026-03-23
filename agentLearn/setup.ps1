# 复制环境变量配置
Copy-Item .env.example .env -ErrorAction SilentlyContinue

# 提示用户配置 API Key
Write-Host "================================" -ForegroundColor Cyan
Write-Host "GoAgent TaskRunner 快速启动" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "请编辑 .env 文件，设置你的 API Key:" -ForegroundColor Yellow
Write-Host "  OPENAI_API_KEY=your_api_key_here" -ForegroundColor Yellow
Write-Host ""
Write-Host "或者使用国内 API 提供商:" -ForegroundColor Yellow
Write-Host "  DEEPSEEK_API_KEY=your_deepseek_key" -ForegroundColor Yellow
Write-Host "  DEEPSEEK_BASE_URL=https://api.deepseek.com/v1" -ForegroundColor Yellow
Write-Host ""
Write-Host "配置完成后，运行以下命令启动服务：" -ForegroundColor Green
Write-Host "  .\start.ps1" -ForegroundColor Green
Write-Host ""
