package main

import (
	"smart-aftercare/config"
	_ "smart-aftercare/docs"
	"smart-aftercare/internal/handler"
	"smart-aftercare/internal/repository"
	"smart-aftercare/internal/router"
	"smart-aftercare/internal/service"
	"smart-aftercare/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title           智能售后服务系统 API
// @version         1.0
// @description     基于 RAG 的智能家电售后问答与文档管理系统。支持文档上传解析、智能问答、故障代码查询等功能。
// @termsOfService  http://swagger.io/terms/

// @contact.name   Smart Aftercare Team
// @contact.email  support@smart-aftercare.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000
// @BasePath  /api

func main() {
	// 1. 初始化日志
	logger.Init("./logs")
	logger.Info("========== 智能售后服务系统启动 ==========")

	// 2. 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("配置加载失败：", err)
	}
	logger.Infof("配置加载完成，端口: %s", cfg.Server.Port)

	// 3. 初始化依赖（数据库、向量库、缓存、存储）
	mysqlRepo, err := repository.NewMySQLRepo(cfg.MySQL)
	if err != nil {
		logger.Fatal("MySQL初始化失败：", err)
	}
	defer mysqlRepo.Close()
	logger.Info("MySQL连接成功")

	// 自动迁移数据库表
	if err := mysqlRepo.AutoMigrate(); err != nil {
		logger.Fatal("数据库表迁移失败：", err)
	}
	logger.Info("数据库表迁移完成")

	milvusRepo, err := repository.NewMilvusRepo(cfg.Milvus)
	if err != nil {
		logger.Fatal("Milvus初始化失败：", err)
	}
	defer milvusRepo.Close()
	logger.Info("Milvus连接成功")

	redisRepo := repository.NewRedisRepo(cfg.Redis)
	if err := redisRepo.Ping(context.Background()); err != nil {
		logger.Fatal("Redis连接失败：", err)
	}
	defer redisRepo.Close()
	logger.Info("Redis连接成功")

	minioRepo, err := repository.NewMinIORepo(cfg.Minio)
	if err != nil {
		logger.Fatal("MinIO初始化失败：", err)
	}
	logger.Info("MinIO连接成功")

	// 4. 初始化业务服务
	documentService := service.NewDocumentService(mysqlRepo, milvusRepo, minioRepo, cfg)
	ragService := service.NewRagService(milvusRepo, redisRepo, &cfg.Doubao)
	errorCodeService := service.NewErrorCodeService(mysqlRepo, ragService)

	// 5. 初始化处理器
	docHandler := handler.NewDocumentHandler(documentService)
	qaHandler := handler.NewQAHandler(ragService, errorCodeService, mysqlRepo)
	healthHandler := handler.NewHealthHandler(mysqlRepo, redisRepo, milvusRepo)

	// 6. 初始化路由
	r := router.Setup(docHandler, qaHandler, healthHandler)

	// 7. 创建上传目录
	os.MkdirAll("./uploads", 0o755)
	os.MkdirAll("./logs", 0o755)

	// 8. 启动 HTTP 服务（支持优雅关闭）
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 在 goroutine 中启动服务
	go func() {
		logger.Infof("服务启动: http://0.0.0.0:%s", cfg.Server.Port)
		logger.Info("API 接口:")
		logger.Info("  POST /api/document/upload  - 文档上传")
		logger.Info("  GET  /api/documents        - 文档列表")
		logger.Info("  GET  /api/document/:id     - 文档详情")
		logger.Info("  POST /api/qa               - 智能问答")
		logger.Info("  POST /api/qa/error-code    - 故障代码查询")
		logger.Info("  GET  /api/health           - 健康检查")
		logger.Info("  GET  /api/health/ready     - 就绪检查")
		logger.Info("  GET  /api/stats            - 系统统计")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务启动失败：", err)
		}
	}()

	// 9. 等待中断信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务...")

	// 给 5 秒时间处理已有请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务关闭异常：", err)
	}

	logger.Info("========== 服务已关闭 ==========")
}
