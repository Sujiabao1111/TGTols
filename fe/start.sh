#!/bin/bash
PORT="${PORT:-5000}"

# ============================================
# TOLS 前端 (Next.js) 一键启动脚本
# ============================================
# 功能：构建 + 停止旧进程 + 启动新进程
# 用法：./start.sh
# ============================================

echo "========================================"
echo "  TOLS 前端 (Next.js) 启动脚本"
echo "========================================"

# 1. 构建生产版本
echo "[1/3] 构建生产版本..."
npm run build > build.log 2>&1
if [ $? -ne 0 ]; then
    echo "[错误] 构建失败，请检查 build.log"
    tail -20 build.log
    exit 1
fi
echo "    构建成功"

# 2. 停止旧进程
echo "[2/3] 检查并停止旧进程..."
if [ -f app.pid ]; then
    OLD_PID=$(cat app.pid)
    if ps -p $OLD_PID > /dev/null 2>&1; then
        echo "    发现运行中的进程 PID=$OLD_PID，正在停止..."
        kill $OLD_PID
        sleep 2
        # 如果还在运行，强制杀掉
        if ps -p $OLD_PID > /dev/null 2>&1; then
            echo "    强制停止..."
            kill -9 $OLD_PID
        fi
    fi
    rm -f app.pid
    echo "    旧进程已清理"
else
    # 兜底：按进程名查找并停止
    pkill -f "next start" 2>/dev/null
    pkill -f "server.js" 2>/dev/null
    echo "    已清理残留进程"
fi

# 3. 启动新进程
echo "[3/3] 启动服务..."

nohup env PORT="$PORT" npm run start > all.log 2>&1 &
echo "    已启动"

APP_PID=$!
echo $APP_PID > app.pid

sleep 2
if ps -p $APP_PID > /dev/null 2>&1; then
    echo "========================================"
    echo "  启动成功！"
    echo "  PID:    $APP_PID"
    echo "  端口:   $PORT"
    echo "  日志:   all.log"
    echo "  停止:   ./stop.sh"
    echo "========================================"
else
    echo "[错误] 启动失败，请检查 all.log"
    tail -20 all.log
    rm -f app.pid
    exit 1
fi
