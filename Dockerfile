# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖（git、gcc 等，OCR 需要 Tesseract）
RUN apk add --no-cache git gcc musl-dev tesseract-ocr tesseract-ocr-data-chi-sim

# 复制 go.mod 和 go.sum，先下载依赖（利用 Docker 缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译（启用 CGO 以支持 Tesseract）
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o smart-aftercare ./cmd/api

# ==================== 运行阶段 ====================
FROM alpine:3.18

# 安装必要依赖
RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-chi-sim ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 复制编译产物
COPY --from=builder /app/smart-aftercare .

# 复制配置文件
COPY --from=builder /app/config ./config

# 创建必要目录
RUN mkdir -p ./uploads ./logs

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8000/api/health || exit 1

# 启动命令
CMD ["./smart-aftercare"]
