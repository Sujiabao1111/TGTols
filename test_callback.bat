@echo off
chcp 65001 >nul
title 支付回调测试工具

REM ========================================
REM 支付回调测试脚本 (Windows)
REM 用法: test_callback.bat [订单号] [金额] [状态]
REM 示例: test_callback.bat ORD2025020412300012345678 50000.00 1
REM ========================================

REM 参数
set "ORDER_ID=%~1"
set "AMOUNT=%~2"
set "STATUS=%~3"

REM 默认值
if "%ORDER_ID%"=="" set "ORDER_ID=ORD2025020412300012345678"
if "%AMOUNT%"=="" set "AMOUNT=50000.00"
if "%STATUS%"=="" set "STATUS=1"

echo ========================================
echo 🧪 支付回调测试
echo ========================================
echo 订单号: %ORDER_ID%
echo 金额: %AMOUNT%
echo 状态: %STATUS% (1=成功, 2=失败)
echo.

REM 检查 Go 环境
where go >nul 2>nul
if errorlevel 1 (
    echo ❌ 错误: 未找到 Go 环境
    echo 请确保 Go 已安装并添加到 PATH
    pause
    exit /b 1
)

REM 进入测试程序目录
cd /d "%~dp0be\cmd\test_callback" 2>nul
if errorlevel 1 (
    echo ❌ 错误: 无法进入目录 be\cmd\test_callback
    echo 请确保在正确的目录下运行此脚本
    pause
    exit /b 1
)

echo 📦 编译测试程序...
go build -o test_callback.exe main.go
if errorlevel 1 (
    echo ❌ 编译失败
    pause
    exit /b 1
)

echo.
echo 📤 发送回调请求...
echo.

REM 运行测试程序，传递参数
test_callback.exe "%ORDER_ID%" "%AMOUNT%" "%STATUS%"

echo.
pause
