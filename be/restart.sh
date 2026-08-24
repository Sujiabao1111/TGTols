#!/bin/bash

# ============================================
# TOLS 后端 (Go) 一键重启脚本
# ============================================
# 功能：编译 + 停止旧进程 + 启动新进程
# 用法：./restart.sh
# ============================================

# 1. 给同名二进制文件添加执行权限
echo "[1/3] 添加脚本执行权限..."
chmod a+x ${PWD##*/} 2>/dev/null
echo "    ok"

# 2. 停止旧进程
echo "[2/3] 停止旧进程..."
if [ -f "app.pid" ]; then
    PID=$(cat app.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo "    发现运行中的进程 PID=$PID，正在停止..."
        kill $PID
        sleep 1
        # 还在就强制终止
        if ps -p $PID > /dev/null 2>&1; then
            echo "    强制停止..."
            kill -9 $PID
        fi
    fi
    rm -f app.pid
    echo "    旧进程已清理"
else
    echo "    无运行中的进程"
fi

# 3. 启动新进程
echo "[3/3] 启动新进程..."
nohup ./${PWD##*/} > all.log 2>&1 &
APP_PID=$!
echo $APP_PID > app.pid

sleep 1
if ps -p $APP_PID > /dev/null 2>&1; then
    echo "========================================"
    echo "  重启成功！"
    echo "  PID:    $APP_PID"
    echo "  端口:   3555"
    echo "  日志:   all.log"
    echo "========================================"
else
    echo "    [错误] 启动失败，请检查 all.log"
    rm -f app.pid
    exit 1
fi
