# ============================================
# TOLS 主后端 (Go + Fiber) Docker 构建文件
# ============================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 设置国内代理加速依赖下载
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

# 先复制 go.mod 和 go.sum，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制全部源码
COPY . .

# 编译为 Linux 可执行文件（静态链接，无 CGO）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app main.go

# ============================================
# 运行阶段：使用最小镜像
# ============================================
FROM alpine:latest

# 安装时区和 CA 证书
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制编译好的可执行文件
COPY --from=builder /app/app .

# 复制 Docker 专用配置模板（含占位符，启动时由 entrypoint 替换）
COPY --from=builder /app/config.docker.json ./config.docker.json

# 复制入口脚本
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 3555

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./app"]
