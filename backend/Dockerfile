# ========================
# Stage 1: Builder
# ========================
FROM golang:1.21-alpine AS builder

# 安装必要的编译工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用（禁用CGO以获得静态二进制）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o kyc_server \
    ./cmd/server

# ========================
# Stage 2: Runtime
# ========================
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# 设置工作目录
WORKDIR /app

# 从builder复制二进制文件
COPY --from=builder /app/kyc_server .

# 创建日志目录并修改所有者为 app 用户
# 关键：镜像中若不存在 /app/logs，Docker 挂载命名卷时会以 root 属主创建该目录，
# 导致以非 root 用户（app）运行时无法写入日志（permission denied）。
RUN mkdir -p /app/logs && chown -R app:app /app

# 切换到非root用户
USER app

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/api/v1/public/config || exit 1

# 启动应用
CMD ["./kyc_server"]
