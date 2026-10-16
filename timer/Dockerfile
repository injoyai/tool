# ===== 构建阶段 =====
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Go 代理(加速国内构建,可覆盖: docker build --build-arg GOPROXY=...)
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

# 先拷贝依赖,利用 docker 层缓存
COPY go.mod go.sum* ./
RUN go mod download

# 拷贝源码(web/ 会通过 //go:embed 打包进二进制,无需额外拷贝前端资源)
COPY . .

# 静态编译(纯 Go sqlite: modernc.org/sqlite,无需 cgo)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# ===== 运行阶段 =====
FROM alpine:latest

# tzdata: 让 time.Local 按 TZ 环境变量生效(日志时间正确)
# ca-certificates: HTTP/Webhook 任务访问 HTTPS 需要
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app
COPY --from=builder /build/server .

# 数据目录(SQLite 数据库)与配置目录(config.yaml)
RUN mkdir -p /app/data/database /app/config

# 时区(可覆盖: docker run -e TZ=Asia/Shanghai)
ENV TZ=Asia/Shanghai
EXPOSE 8078

# 持久化: 数据库文件、配置文件
VOLUME ["/app/data", "/app/config"]

ENTRYPOINT ["./server"]

# ===== 使用说明 =====
# 构建: docker build -t injoyai/timer:v2 .
# 运行(免登录):  docker run -d -p 8078:8078 -v timer-data:/app/data injoyai/timer:v2
# 运行(带密码):  docker run -d -p 8078:8078 -e password=你的密码 -v timer-data:/app/data injoyai/timer:v2
# 可用环境变量:
#   TZ        时区,默认 Asia/Shanghai
#   port      监听端口,默认 8078
#   db        数据库路径,默认 ./data/database/timer.db
#   password  登录密码(同 PASSWORD),留空则免登录
# 配置文件: 挂载到 /app/config/config.yaml (可选,环境变量优先级更高)
