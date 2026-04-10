package handler

import (
	"smart-aftercare/internal/repository"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	mysqlRepo  *repository.MySQLRepo
	redisRepo  *repository.RedisRepo
	milvusRepo *repository.MilvusRepo
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(
	mysqlRepo *repository.MySQLRepo,
	redisRepo *repository.RedisRepo,
	milvusRepo *repository.MilvusRepo,
) *HealthHandler {
	return &HealthHandler{
		mysqlRepo:  mysqlRepo,
		redisRepo:  redisRepo,
		milvusRepo: milvusRepo,
	}
}

// Health 基础健康检查
// GET /api/health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "smart-aftercare",
	})
}

// ReadinessCheck 就绪检查（检查所有依赖服务连接）
// GET /api/health/ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx := context.Background()
	status := gin.H{
		"status": "ready",
	}
	allHealthy := true

	// 检查 MySQL
	if err := h.mysqlRepo.Ping(); err != nil {
		status["mysql"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		status["mysql"] = "healthy"
	}

	// 检查 Redis
	if err := h.redisRepo.Ping(ctx); err != nil {
		status["redis"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		status["redis"] = "healthy"
	}

	// 检查 Milvus
	stats, err := h.milvusRepo.GetCollectionStats(ctx)
	if err != nil {
		status["milvus"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		status["milvus"] = "healthy"
		status["milvus_stats"] = stats
	}

	if !allHealthy {
		status["status"] = "degraded"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// Stats 获取系统统计信息
// GET /api/stats
func (h *HealthHandler) Stats(c *gin.Context) {
	stats, err := h.mysqlRepo.GetDocumentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
	})
}
