FROM alpine:latest

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 直接复制我们在宿主机本地用 Go 1.26 编译好的 Linux 跨平台二进制文件
COPY zero-app /app/zero-app

EXPOSE 8080               