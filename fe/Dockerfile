# ============================================
# TOLS 前端 (Next.js) Docker 构建文件
# ============================================
# 使用 Node.js 20 Alpine 作为构建环境
FROM node:20-alpine AS builder

WORKDIR /app

# 先复制 package 文件，利用 Docker 缓存层
COPY package*.json ./

# 安装依赖
RUN npm ci

# 复制全部源码
COPY . .

# 构建参数：这些会在构建时传入
# 后端 API 地址（浏览器可访问的地址）
ARG NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL}

# Clerk 公钥（构建时内联到前端代码中）
ARG NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
ENV NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=${NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY}

ARG NEXT_PUBLIC_TWA_RETURN_URL=https://t.me/ppnetbet_bot/PPNetApp
ENV NEXT_PUBLIC_TWA_RETURN_URL=${NEXT_PUBLIC_TWA_RETURN_URL}

# 构建生产版本
RUN npm run build

# ============================================
# 运行阶段：使用更小的镜像
# ============================================
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=5000

# 从构建阶段复制 standalone 输出
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 5000

CMD ["node", "server.js"]
