@echo off
rem dca — deepseek coding agent 启动脚本 (CMD)
rem 用法: dca.bat [额外参数...]

setlocal enabledelayedexpansion

rem 加载 API key（从注册表读取用户环境变量）
for /f "skip=2 tokens=3*" %%a in ('reg query HKEY_CURRENT_USER\Environment /v DEEPSEEK_API_KEY 2^>nul') do set "DEEPSEEK_API_KEY=%%a %%b"
if not defined DEEPSEEK_API_KEY (
    echo [错误] 未设置 DEEPSEEK_API_KEY 环境变量
    echo 请运行: [Environment]::SetEnvironmentVariable("DEEPSEEK_API_KEY", "sk-xxx", "User")
    pause
    exit /b 1
)

set "MODEL=deepseek-v4-flash"
set "ROOT=%CD%"

echo [dca] model: %MODEL%
echo [dca] root:  %ROOT%

mise exec -- go run . --model %MODEL% --root %ROOT% %*
