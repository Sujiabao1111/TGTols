#!/bin/sh
set -e

cp /app/config.docker.template.yaml /app/config.docker.yaml

DB_HOST=${DB_HOST:-host.docker.internal}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-root}
DB_PASSWORD=${DB_PASSWORD:-guesswhat}
DB_NAME=${DB_NAME:-star}

sed -i "s/DB_HOST_PLACEHOLDER/${DB_HOST}/g" /app/config.docker.yaml
sed -i "s/DB_PORT_PLACEHOLDER/${DB_PORT}/g" /app/config.docker.yaml
sed -i "s/DB_USER_PLACEHOLDER/${DB_USER}/g" /app/config.docker.yaml
sed -i "s/DB_PASSWORD_PLACEHOLDER/${DB_PASSWORD}/g" /app/config.docker.yaml
sed -i "s/DB_NAME_PLACEHOLDER/${DB_NAME}/g" /app/config.docker.yaml

echo "[BG Setup] MySQL: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
exec "$@"
