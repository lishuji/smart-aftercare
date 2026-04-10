package router

import (
	"smart-aftercare/internal/handler"
	"smart-aftercare/internal/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup 初始化路由
func Setup(
	docHandler *handler.DocumentHandler,
	qaHandler *handler.QAHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())

	// 跨域配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 设置上传文件大小限制（50MB）
	r.MaxMultipartMemory = 50 << 20

	// API 路由组
	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/health", healthHandler.Health)
		api.GET("/health/ready", healthHandler.ReadinessCheck)
		api.GET("/stats", healthHandler.Stats)

		// 文档管理
		api.POST("/document/upload", docHandler.UploadDocument)
		api.GET("/document/:id", docHandler.GetDocument)
		api.GET("/documents", docHandler.ListDocuments)
		api.DELETE("/document/:id", docHandler.DeleteDocument)

		// 问答
		api.POST("/qa", qaHandler.QA)
		api.POST("/qa/error-code", qaHandler.QueryErrorCode)
	}

	return r
}
