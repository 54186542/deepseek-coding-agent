#!/usr/bin/env pwsh
# dca — deepseek coding agent 启动脚本
# 用法: .\dca.ps1 [额外参数...]

# 从用户环境变量加载 API key（如果当前进程没有的话）
if (-not $env:DEEPSEEK_API_KEY) {
    $env:DEEPSEEK_API_KEY = [Environment]::GetEnvironmentVariable("DEEPSEEK_API_KEY", "User")
}

# 默认模型和根目录
$model = "deepseek-v4-flash"
$root = (Get-Location).Path

Write-Host "🚀 dca — deepseek coding agent" -ForegroundColor Cyan
Write-Host "   model: $model" -ForegroundColor DarkGray
Write-Host "   root:  $root" -ForegroundColor DarkGray

# 用 cmd /c 包装，避免 PowerShell 5.1 对 mise exec -- 的参数解析问题
$argsStr = if ($args) { ($args | ForEach-Object { "`"$_`"" }) -join " " } else { "" }
$cmd = "mise exec -- go run . --model $model --root $root $argsStr"
cmd /c $cmd
