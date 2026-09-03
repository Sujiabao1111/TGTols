#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
FE_DIR="${FE_DIR:-$ROOT_DIR/fe}"
PORT="${PORT:-5000}"
PID_FILE="${PID_FILE:-$FE_DIR/.next-start.pid}"
LOG_FILE="${LOG_FILE:-$FE_DIR/fe.log}"
START_CMD="${START_CMD:-node ./node_modules/next/dist/bin/next start}"

if [ ! -d "$FE_DIR" ]; then
  echo "前端目录不存在: $FE_DIR" >&2
  exit 1
fi

if [ ! -f "$FE_DIR/package.json" ]; then
  echo "未找到 package.json: $FE_DIR/package.json" >&2
  exit 1
fi

if ! command -v npm >/dev/null 2>&1; then
  echo "未找到 npm，请先安装 Node.js。" >&2
  exit 1
fi

if ! command -v node >/dev/null 2>&1; then
  echo "未找到 node，请先安装 Node.js。" >&2
  exit 1
fi

echo "==> 构建前端"
(cd "$FE_DIR" && npm run build)

find_port_pids() {
  if command -v lsof >/dev/null 2>&1; then
    lsof -tiTCP:"$PORT" -sTCP:LISTEN 2>/dev/null || true
    return
  fi

  if command -v fuser >/dev/null 2>&1; then
    fuser "$PORT"/tcp 2>/dev/null | tr ' ' '\n' | sed '/^$/d' || true
    return
  fi

  if command -v ss >/dev/null 2>&1; then
    ss -ltnp 2>/dev/null | awk -v port=":$PORT" '
      index($4, port) {
        while (match($0, /pid=[0-9]+/)) {
          pid = substr($0, RSTART + 4, RLENGTH - 4)
          print pid
          $0 = substr($0, RSTART + RLENGTH)
        }
      }
    ' || true
  fi
}

find_cmd_pids() {
  if command -v pgrep >/dev/null 2>&1; then
    pgrep -f "$FE_DIR/.*next(.*/dist/bin/next)? start|$FE_DIR/.*npm run start|$FE_DIR/.*node .*next.*start" 2>/dev/null || true
    return
  fi

  ps -ef | awk -v dir="$FE_DIR" '
    index($0, dir) && (
      index($0, "next start") ||
      index($0, "next/dist/bin/next start") ||
      index($0, "npm run start")
    ) && index($0, "awk") == 0 {
      print $2
    }
  ' || true
}

collect_pids() {
  {
    if [ -f "$PID_FILE" ]; then
      cat "$PID_FILE" 2>/dev/null || true
    fi
    find_port_pids
    find_cmd_pids
  } | sed '/^$/d' | sort -u
}

PIDS=$(collect_pids || true)

if [ -n "$PIDS" ]; then
  echo "==> 停止旧前端进程: $PIDS"
  for pid in $PIDS; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
  done

  i=0
  while [ "$i" -lt 10 ]; do
    alive=0
    for pid in $PIDS; do
      if kill -0 "$pid" 2>/dev/null; then
        alive=1
        break
      fi
    done

    if [ "$alive" -eq 0 ]; then
      break
    fi

    sleep 1
    i=$((i + 1))
  done

  for pid in $PIDS; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "==> 强制结束进程: $pid"
      kill -9 "$pid" 2>/dev/null || true
    fi
  done
else
  echo "==> 未发现旧前端进程"
fi

rm -f "$PID_FILE"

echo "==> 启动新前端"
(
  cd "$FE_DIR"
  nohup env PORT="$PORT" sh -c "exec $START_CMD" >>"$LOG_FILE" 2>&1 &
  echo $! >"$PID_FILE"
)

NEW_PID=$(cat "$PID_FILE")
sleep 2

if kill -0 "$NEW_PID" 2>/dev/null; then
  echo "前端已启动"
  echo "PID: $NEW_PID"
  echo "PORT: $PORT"
  echo "LOG: $LOG_FILE"
else
  echo "前端启动失败，请查看日志: $LOG_FILE" >&2
  exit 1
fi
