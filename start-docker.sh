#!/bin/bash

# ============================================
# TOLS 项目 Docker 一键启动脚本 (Linux/Mac)
# ============================================

set -e

echo "============================================"
echo "   TOLS 项目 Docker 一键启动脚本"
echo "============================================"
echo ""

# 检查 Docker
echo "[1/4] 检查 Docker..."
if ! command -v docker &> /dev/null; then
    echo "[错误] 未检测到 Docker，请先安装"
    echo "  Linux: curl -fsSL https://get.docker.com | sh"
    echo "  Mac:   brew install docker"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "[错误] Docker 未运行，请启动 Docker 服务"
    echo "  Linux: sudo systemctl start docker"
    exit 1
fi
echo "    Docker 运行正常"

# 检查 .env 文件
echo ""
echo "[2/4] 检查环境变量配置..."
if [ ! -f .env ]; then
    echo "    未找到 .env 文件，正在从模板创建..."
    if [ -f .env.docker ]; then
        cp .env.docker .env
        echo "    已创建 .env 文件，请编辑填入以下配置："
        echo "       - DB_HOST: MySQL 地址（本地填 host.docker.internal）"
        echo "       - DB_PASSWORD: MySQL 密码"
        echo "       - REDIS_HOST: Redis 地址"
        echo "       - Clerk API Keys"
        echo "    编辑完成后请重新运行此脚本"
        ${EDITOR:-nano} .env
        exit 0
    else
        echo "    错误：找不到 .env.docker 模板文件"
        exit 1
    fi
else
    echo "    .env 文件已存在"
fi

# 构建并启动
echo ""
echo "[3/4] 构建并启动服务..."
echo "    首次构建可能需要 5-10 分钟，请耐心等待..."
docker-compose up -d --build

echo ""
echo "[4/4] 检查服务状态..."
docker-compose ps

echo ""
echo "============================================"
echo "   启动完成！"
echo "============================================"
echo ""
echo "访问地址："
echo "  - 用户前端:    http://localhost:3000"
echo "  - 主后端 API:  http://localhost:3555"
echo ""
echo "前提条件（确保已就绪）："
echo "  - MySQL 已运行并可访问"
echo "  - Redis 已运行并可访问"
echo "  - 数据库 star 已创建并导入 tols.sql"
echo ""
echo "常用命令："
echo "  docker-compose logs -f fe    查看前端日志"
echo "  docker-compose logs -f be    查看后端日志"
echo "  docker-compose stop          停止服务"
echo "  docker-compose start         启动服务"
echo "  docker-compose down          停止并删除容器"
echo ""
