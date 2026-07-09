# 第一阶段：用标准的 Go 1.26 镜像在容器内部编译源码
FROM golang:1.26-alpine AS builder
WORKDIR /build

# 配置国内代理加速拉包
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

# 复制云端服务器刚拉下来的最新源码
COPY . .
RUN go build -o /build/zero-app main.go

# 第二阶段：纯净运行环境
FROM alpine:latest

# 🔥 核心修复：执行 apk 之前，先将 Alpine 的官方海外源替换为国内阿里云源，彻底解决卡死和超时问题
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update --no-cache && \
    apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 把第一阶段现场编译出来的全新二进制文件拷过来
COPY --from=builder /build/zero-app /app/zero-app
COPY etc /app/etc

CMD ["./zero-app", "-f", "etc/access-control-api.yaml"]