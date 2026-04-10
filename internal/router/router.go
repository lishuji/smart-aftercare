package router

import (
	"smart-aftercare/internal/handler"
	"smart-aftercare/internal/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// 静态文件：前端页面
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
	))

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
