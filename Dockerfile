# 使用官方Go镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git make

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nidavellir main.go

# 使用轻量级镜像作为运行环境
FROM alpine:latest

# 安装ca证书
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -g 1001 -S nidavellir && \
    adduser -u 1001 -S nidavellir -G nidavellir

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/nidavellir .
COPY --from=builder /app/configs ./configs

# 更改文件所有者
RUN chown -R nidavellir:nidavellir /app

# 切换到非root用户
USER nidavellir

# 暴露端口
EXPOSE 8080 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 启动应用
CMD ["./nidavellir"]