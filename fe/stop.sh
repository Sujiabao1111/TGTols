#!/bin/sh

# ============================================
# TOLS 前端 (Next.js) 停止脚本
# ============================================
# 功能：停止正在运行的前端进程
# 用法：./stop.sh
# ============================================

echo "========================================"
echo "  TOLS 前端 (Next.js) 停止脚本"
echo "========================================"

if [ -f "app.pid" ]; then
    PID=$(cat app.pid)
    echo "    正在停止进程 PID=$PID..."

    # 先正常终止
    kill $PID 2>/dev/null
    sleep 1

    # 还在就强制终止
    if ps -p $PID >/dev/null 2>&1; then
        kill -9 $PID 2>/dev/null
    fi

    rm -f app.pid
    echo "    已停止"
else
    echo "    未找到 PID 文件，尝试按进程名停止..."
    pkill -f "next start" 2>/dev/null
    pkill -f "server.js" 2>/dev/null
    echo "    已尝试停止"
fi

echo "========================================"
