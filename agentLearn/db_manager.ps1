# SQLite 数据库管理脚本

param(
    [string]$action = "init",
    [string]$dbPath = "./data.db"
)

function Show-Menu {
    Write-Host "`n================================" -ForegroundColor Cyan
    Write-Host "  SQLite 数据库管理工具" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "1. 初始化数据库（创建表和示例数据）" -ForegroundColor Green
    Write-Host "2. 查看数据库表结构" -ForegroundColor Green
    Write-Host "3. 查询用户数据" -ForegroundColor Green
    Write-Host "4. 查询日志数据" -ForegroundColor Green
    Write-Host "5. 查询任务数据" -ForegroundColor Green
    Write-Host "6. 查询指标数据" -ForegroundColor Green
    Write-Host "7. 清空所有数据" -ForegroundColor Yellow
    Write-Host "8. 删除数据库文件" -ForegroundColor Red
    Write-Host "0. 退出" -ForegroundColor Red
    Write-Host ""
}

function Initialize-Database {
    Write-Host "`n正在初始化数据库..." -ForegroundColor Cyan
    if (Test-Path $dbPath) {
        Write-Host "警告：数据库文件已存在，将被覆盖" -ForegroundColor Yellow
        Remove-Item $dbPath -Force
    }
    
    & go run cmd/init_db/main.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ 数据库初始化成功！" -ForegroundColor Green
    } else {
        Write-Host "`n❌ 数据库初始化失败！" -ForegroundColor Red
    }
}

function Show-Tables {
    Write-Host "`n📋 数据库表结构:" -ForegroundColor Cyan
    sqlite3 $dbPath ".schema"
}

function Query-Users {
    Write-Host "`n👥 用户数据:" -ForegroundColor Cyan
    sqlite3 -header -column $dbPath "SELECT id, username, email, age, created_at FROM users ORDER BY id;"
}

function Query-Logs {
    Write-Host "`n📝 日志数据:" -ForegroundColor Cyan
    sqlite3 -header -column $dbPath "SELECT id, level, message, source, created_at FROM logs ORDER BY id DESC LIMIT 10;"
}

function Query-Tasks {
    Write-Host "`n✅ 任务数据:" -ForegroundColor Cyan
    sqlite3 -header -column $dbPath @"
SELECT 
    id, 
    title, 
    description, 
    status,
    CASE priority
        WHEN 1 THEN '高'
        WHEN 2 THEN '中'
        WHEN 3 THEN '低'
        ELSE '未知'
    END as priority_cn,
    created_at 
FROM tasks 
ORDER BY priority, id;
"@
}

function Query-Metrics {
    Write-Host "`n📊 指标数据:" -ForegroundColor Cyan
    sqlite3 -header -column $dbPath "SELECT id, name, value, unit, timestamp FROM metrics ORDER BY id DESC LIMIT 10;"
}

function Clear-AllData {
    Write-Host "`n⚠️  警告：此操作将清空所有数据！" -ForegroundColor Yellow
    $confirm = Read-Host "确认要清空所有数据吗？(y/n)"
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        sqlite3 $dbPath "DELETE FROM metrics;"
        sqlite3 $dbPath "DELETE FROM tasks;"
        sqlite3 $dbPath "DELETE FROM logs;"
        sqlite3 $dbPath "DELETE FROM users;"
        Write-Host "✅ 所有数据已清空" -ForegroundColor Green
    } else {
        Write-Host "❌ 操作已取消" -ForegroundColor Yellow
    }
}

function Delete-Database {
    Write-Host "`n⚠️  警告：此操作将删除数据库文件！" -ForegroundColor Yellow
    $confirm = Read-Host "确认要删除数据库文件吗？(y/n)"
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        if (Test-Path $dbPath) {
            Remove-Item $dbPath -Force
            Write-Host "✅ 数据库文件已删除" -ForegroundColor Green
        } else {
            Write-Host "ℹ️  数据库文件不存在" -ForegroundColor Yellow
        }
    } else {
        Write-Host "❌ 操作已取消" -ForegroundColor Yellow
    }
}

# 主程序
if ($action -eq "init") {
    Initialize-Database
    exit
}

# 交互式菜单
while ($true) {
    Show-Menu
    $choice = Read-Host "请选择操作 (0-8)"
    
    switch ($choice) {
        "1" { Initialize-Database }
        "2" { 
            if (Test-Path $dbPath) {
                Show-Tables 
            } else {
                Write-Host "❌ 数据库文件不存在，请先初始化" -ForegroundColor Yellow
            }
        }
        "3" { 
            if (Test-Path $dbPath) {
                Query-Users 
            } else {
                Write-Host "❌ 数据库文件不存在，请先初始化" -ForegroundColor Yellow
            }
        }
        "4" { 
            if (Test-Path $dbPath) {
                Query-Logs 
            } else {
                Write-Host "❌ 数据库文件不存在，请先初始化" -ForegroundColor Yellow
            }
        }
        "5" { 
            if (Test-Path $dbPath) {
                Query-Tasks 
            } else {
                Write-Host "❌ 数据库文件不存在，请先初始化" -ForegroundColor Yellow
            }
        }
        "6" { 
            if (Test-Path $dbPath) {
                Query-Metrics 
            } else {
                Write-Host "❌ 数据库文件不存在，请先初始化" -ForegroundColor Yellow
            }
        }
        "7" { 
            if (Test-Path $dbPath) {
                Clear-AllData 
            } else {
                Write-Host "❌ 数据库文件不存在" -ForegroundColor Yellow
            }
        }
        "8" { Delete-Database }
        "0" { 
            Write-Host "`n👋 再见！" -ForegroundColor Cyan
            break 
        }
        default { 
            Write-Host "❌ 无效的选择，请重新输入" -ForegroundColor Yellow 
        }
    }
    
    if ($choice -ne "0") {
        Write-Host "`n按任意键继续..." -ForegroundColor Gray
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    }
}
