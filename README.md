# 智能售后服务系统（Smart Aftercare）

基于 Golang 技术栈的家电售后智能问答系统，采用 RAG（检索增强生成）架构，支持家电说明书上传、智能问答、故障代码查询等功能。

## 🏗️ 项目架构

```
smart-aftercare/
├── cmd/api/              # 服务入口
│   └── main.go           # 启动 Gin 服务、初始化依赖
├── internal/             # 内部业务逻辑
│   ├── handler/          # 接口处理器（路由实现）
│   │   ├── document.go   # 文档上传/解析接口
│   │   ├── qa.go         # 问答/故障代码查询接口
│   │   └── health.go     # 健康检查接口
│   ├── service/          # 核心业务服务
│   │   ├── document_service.go  # 文档处理（加载→切片→向量化）
│   │   ├── rag_service.go       # RAG 检索+生成服务
│   │   └── error_code_service.go # 故障代码匹配服务
│   ├── model/            # 数据模型（GORM）
│   │   ├── document.go   # 文档元数据模型
│   │   ├── error_code.go # 故障代码模型
│   │   └── user.go       # 用户模型 + 查询日志
│   ├── repository/       # 数据访问层
│   │   ├── milvus_repo.go # 向量库操作
│   │   ├── mysql_repo.go  # MySQL 操作
│   │   ├── redis_repo.go  # Redis 缓存操作
│   │   └── minio_repo.go  # MinIO 存储操作
│   ├── router/           # 路由注册
│   │   └── router.go
│   ├── middleware/        # 中间件
│   │   └── middleware.go
│   └── util/             # 工具函数
│       ├── ocr.go        # OCR 识别工具
│       ├── slice.go      # 文本切片工具
│       ├── prompt.go     # Prompt 模板 + 关键词提取
│       └── result.go     # 检索结果合并与排序
├── config/               # 配置
│   ├── config.go         # 配置加载
│   └── config.yaml       # 服务配置文件
├── pkg/                  # 公共依赖
│   ├── doubao/           # 豆包大模型 SDK 封装
│   │   └── client.go
│   └── logger/           # 日志工具
│       └── logger.go
├── db/                   # 数据库脚本
│   └── init.sql          # 初始化 SQL（含示例故障代码）
├── docker-compose.yml    # 服务编排
├── Dockerfile            # 应用镜像
├── Makefile              # 构建命令
├── go.mod                # Go 模块定义
└── .env                  # 环境变量（本地开发）
```

## 🛠️ 技术栈

| 模块 | 技术选型 | 说明 |
|------|----------|------|
| 后端框架 | Gin | 轻量高性能 HTTP 框架 |
| 数据库 ORM | GORM + MySQL 8.0 | 文档元数据、故障代码表 |
| 向量库 | Milvus + milvus-sdk-go | 向量存储与检索 |
| 缓存 | Redis + go-redis | 热门问答缓存 |
| 对象存储 | MinIO + minio-go | 文档和截图存储 |
| 大模型 | 豆包（HTTP API） | Embedding + Chat |
| 配置管理 | Viper | 支持 YAML、环境变量 |
| 日志 | Logrus | 结构化日志 |
| 并发处理 | Goroutine + Channel | 原生高并发 |
| 部署 | Docker + Docker Compose | 容器化部署 |

## 🚀 快速开始

### 前置条件

- Go 1.21+
- Docker & Docker Compose
- 豆包 API Key（用于大模型调用）

### 1. 启动依赖服务

```bash
# 启动 MySQL、Milvus、Redis、MinIO
make deps-up
```

### 2. 配置

修改 `config/config.yaml` 中的豆包 API Key：

```yaml
doubao:
  api_key: "your-actual-api-key"
```

### 3. 运行服务

```bash
# 本地运行
make run

# 或编译后运行
make build
./build/smart-aftercare
```

### 4. Docker 全量部署

```bash
# 构建并启动所有服务
make docker-up

# 查看日志
make docker-logs
```

## 📡 API 接口

### 文档上传

```bash
curl -X POST http://localhost:8000/api/document/upload \
  -F "file=@说明书.pdf" \
  -F "brand=美的" \
  -F "model=KFR-35GW" \
  -F "uploader=admin"
```

### 智能问答

```bash
curl -X POST http://localhost:8000/api/qa \
  -H "Content-Type: application/json" \
  -d '{
    "query": "怎么调睡眠模式？",
    "model": "美的KFR-35GW",
    "brand": "美的"
  }'
```

### 故障代码查询

```bash
curl -X POST http://localhost:8000/api/qa/error-code \
  -H "Content-Type: application/json" \
  -d '{
    "code": "E1",
    "model": "美的KFR-35GW"
  }'
```

### 健康检查

```bash
# 基础检查
curl http://localhost:8000/api/health

# 就绪检查（含依赖服务状态）
curl http://localhost:8000/api/health/ready

# 系统统计
curl http://localhost:8000/api/stats
```

## 🔧 开发命令

```bash
make help          # 查看所有命令
make build         # 编译应用
make run           # 本地运行
make test          # 运行测试
make lint          # 代码检查
make deps-up       # 启动依赖服务（开发模式）
make docker-up     # Docker 全量启动
make docker-down   # 停止所有服务
make docker-logs   # 查看应用日志
make mod-tidy      # 整理 Go 依赖
```

## 📋 迁移注意事项

1. **GORM AutoMigrate**：启动时自动创建/更新表结构，无需手动执行 DDL
2. **PDF 解析**：需集成 `unipdf` 库以支持完整 PDF 文本/图片提取（当前为简化实现）
3. **OCR**：依赖系统 Tesseract 库，Docker 镜像已内置
4. **并发处理**：使用 Goroutine 替代 Python asyncio，大模型调用通过 `context` 控制超时

## 📄 License

Internal Use Only
