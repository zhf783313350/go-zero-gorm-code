# 构建阶段
FROM golang:1.26.0-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置Go代理（可选，用于中国地区加速）
ENV GOPROXY=https://goproxy.cn,direct

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/access-control .

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/access-control .
COPY --from=builder /app/etc /app/etc

# 暴露端口
EXPOSE 8080

# 健康检查（使用 TCP 端口检测）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 8080 || exit 1

# 运行应用（使用 Docker 专用配置）
CMD ["./access-control", "-f", "etc/config-docker.yaml"]
