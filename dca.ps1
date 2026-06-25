#!/usr/bin/env pwsh
# dca — deepseek coding agent 启动脚本
# 用法: .\dca.ps1 [额外参数...]

# 从用户环境变量加载 API key（如果当前进程没有的话）
if (-not $env:DEEPSEEK_API_KEY) {
    $env:DEEPSEEK_API_KEY = [Environment]::GetEnvironmentVariable("DEEPSEEK_API_KEY", "User")
}

# 默认模型和根目录
$model = "deepseek-flash-v4"
$root = (Get-Location).Path

Write-Host "🚀 dca — deepseek coding agent" -ForegroundColor Cyan
Write-Host "   model: $model" -ForegroundColor DarkGray
Write-Host "   root:  $root" -ForegroundColor DarkGray

mise exec -- go run . --model $model --root $root @args
