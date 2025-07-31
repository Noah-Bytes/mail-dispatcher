# 构建阶段
FROM golang:1.24.5 AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mail-dispatcher cmd/main.go

# 运行阶段
FROM golang:1.24.5

# 安装 ca-certificates 用于 HTTPS 请求
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# 创建非 root 用户
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -s /bin/bash appuser

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/mail-dispatcher .

# 创建日志目录
RUN mkdir -p logs && chown -R appuser:appgroup logs

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./mail-dispatcher"] 