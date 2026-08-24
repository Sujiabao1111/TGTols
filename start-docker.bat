@echo off
chcp 65001 >nul
echo ============================================
echo    TOLS 项目 Docker 一键启动脚本
echo ============================================
echo.

REM 检查 Docker 是否安装
docker --version >nul 2>&1
if errorlevel 1 (
    echo [错误] 未检测到 Docker，请先安装 Docker Desktop
    echo 下载地址: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo [1/4] 检查 Docker 运行状态...
docker info >nul 2>&1
if errorlevel 1 (
    echo [错误] Docker 未运行，请启动 Docker Desktop
    pause
    exit /b 1
)
echo     Docker 运行正常

REM 检查 .env 文件
if not exist .env (
    echo.
    echo [2/4] 检测到未配置环境变量，正在创建...
    if exist .env.docker (
        copy .env.docker .env >nul
        echo     已从 .env.docker 复制模板到 .env
        echo     !!! 请编辑 .env 文件，填入以下配置：
        echo        - DB_HOST: MySQL 地址（本地填 host.docker.internal）
        echo        - DB_PASSWORD: MySQL 密码
        echo        - REDIS_HOST: Redis 地址
        echo        - Clerk API Keys
        notepad .env
        echo.
        echo 配置完成后请重新运行此脚本
        pause
        exit /b 0
    ) else (
        echo     错误：找不到 .env.docker 模板文件
        pause
        exit /b 1
    )
) else (
    echo [2/4] 环境变量文件 .env 已存在
)

echo [3/4] 构建并启动服务...
echo     首次构建可能需要 5-10 分钟，请耐心等待...
echo.
docker-compose up -d --build

if errorlevel 1 (
    echo.
    echo [错误] 启动失败，请检查上方错误信息
    pause
    exit /b 1
)

echo.
echo [4/4] 检查服务状态...
timeout /t 3 /nobreak >nul
docker-compose ps

echo.
echo ============================================
echo    启动完成！
echo ============================================
echo.
echo 访问地址：
echo   - 用户前端:    http://localhost:3000
echo   - 主后端 API:  http://localhost:3555
echo.
echo 前提条件（确保已就绪）：
echo   - MySQL 已运行并可访问
echo   - Redis 已运行并可访问
echo   - 数据库 star 已创建并导入 tols.sql
echo.
echo 常用命令：
echo   docker-compose logs -f fe    查看前端日志
echo   docker-compose logs -f be    查看后端日志
echo   docker-compose stop          停止服务
echo   docker-compose start         启动服务
echo   docker-compose down          停止并删除容器
echo.
pause
