# Makefile for smart-aftercare

APP_NAME := smart-aftercare
BUILD_DIR := ./build
CMD_DIR := ./cmd/api

# Go 编译参数
GO := go
GOFLAGS := -v
LDFLAGS := -s -w

.PHONY: all build run clean test lint docker-build docker-up docker-down help

## help: 显示帮助信息
help:
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

## build: 编译应用
build:
	@echo ">>> 编译 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo ">>> 编译完成: $(BUILD_DIR)/$(APP_NAME)"

## run: 本地运行
run:
	$(GO) run $(CMD_DIR)/main.go

## test: 运行测试
test:
	$(GO) test ./... -v -cover

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo ">>> 覆盖率报告: coverage.html"

## lint: 代码检查
lint:
	@which golangci-lint > /dev/null || (echo "请安装 golangci-lint" && exit 1)
	golangci-lint run ./...

## clean: 清理编译产物
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo ">>> 清理完成"

## docker-build: 构建 Docker 镜像
docker-build:
	docker build -t $(APP_NAME):latest .

## docker-up: 启动所有服务
docker-up:
	docker-compose up -d
	@echo ">>> 服务已启动"
	@echo ">>> API: http://localhost:8000"
	@echo ">>> MinIO Console: http://localhost:9001"

## docker-down: 停止所有服务
docker-down:
	docker-compose down
	@echo ">>> 服务已停止"

## docker-logs: 查看应用日志
docker-logs:
	docker-compose logs -f smart-aftercare

## deps-up: 仅启动依赖服务（开发模式）
deps-up:
	docker-compose up -d mysql milvus redis minio
	@echo ">>> 依赖服务已启动"

## deps-down: 停止依赖服务
deps-down:
	docker-compose down

## mod-tidy: 整理 Go 依赖
mod-tidy:
	$(GO) mod tidy

## proto: 生成 protobuf（预留）
proto:
	@echo ">>> 暂无 protobuf 定义"

all: clean build
