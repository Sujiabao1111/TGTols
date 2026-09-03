#!/bin/sh
set -e

# ============================================
# TOLS 后端容器启动入口脚本
# 根据环境变量动态替换数据库连接配置
# ============================================

# 复制模板配置文件
cp /app/config.docker.json /app/config.json

# 替换 MySQL 连接信息
DB_HOST=${DB_HOST:-host.docker.internal}
DB_PORT=${DB_PORT:-3306}
DB_PASSWORD=${DB_PASSWORD:-guesswhat}
DB_NAME=${DB_NAME:-star}

# 替换 config.json 中的占位符
sed -i "s/DB_HOST_PLACEHOLDER/${DB_HOST}/g" /app/config.json
sed -i "s/DB_PORT_PLACEHOLDER/${DB_PORT}/g" /app/config.json
sed -i "s/DB_PASSWORD_PLACEHOLDER/${DB_PASSWORD}/g" /app/config.json
sed -i "s/DB_NAME_PLACEHOLDER/${DB_NAME}/g" /app/config.json

# 替换 Redis 连接信息
REDIS_HOST=${REDIS_HOST:-host.docker.internal}
REDIS_PORT=${REDIS_PORT:-6379}

sed -i "s/REDIS_HOST_PLACEHOLDER/${REDIS_HOST}/g" /app/config.json
sed -i "s/REDIS_PORT_PLACEHOLDER/${REDIS_PORT}/g" /app/config.json

# BOAN / M7PP credentials are injected at runtime. They must never be exposed
# to the frontend image or committed in config.docker.json.
BOAN_API_KEY=${BOAN_API_KEY:-}
BOAN_API_SECRET=${BOAN_API_SECRET:-}

sed -i "s/BOAN_API_KEY_PLACEHOLDER/${BOAN_API_KEY}/g" /app/config.json
sed -i "s/BOAN_API_SECRET_PLACEHOLDER/${BOAN_API_SECRET}/g" /app/config.json

# Telegram Stars / Mini App credentials.
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}
TELEGRAM_WEBHOOK_SECRET=${TELEGRAM_WEBHOOK_SECRET:-}
sed -i "s|TELEGRAM_BOT_TOKEN_PLACEHOLDER|${TELEGRAM_BOT_TOKEN}|g" /app/config.json
sed -i "s|TELEGRAM_WEBHOOK_SECRET_PLACEHOLDER|${TELEGRAM_WEBHOOK_SECRET}|g" /app/config.json

if [ -z "${BOAN_API_KEY}" ] || [ -z "${BOAN_API_SECRET}" ]; then
  echo "WARNING: BOAN_API_KEY/BOAN_API_SECRET are empty; M7PP game sync will be disabled"
else
  echo "  M7PP: BOAN credentials configured"
fi

echo "========================================"
echo "  数据库配置已生效"
echo "  MySQL: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "  Redis: ${REDIS_HOST}:${REDIS_PORT}"
echo "========================================"

# 执行原始命令
exec "$@"
