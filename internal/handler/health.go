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
// @Summary      健康检查
// @Description  返回服务基础健康状态
// @Tags         系统监控
// @Produce      json
// @Success      200  {object}  handler.HealthResponse  "服务健康"
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "smart-aftercare",
	})
}

// ReadinessCheck 就绪检查（检查所有依赖服务连接）
// @Summary      就绪检查
// @Description  检查所有依赖服务（MySQL、Redis、Milvus）的连接状态
// @Tags         系统监控
// @Produce      json
// @Success      200  {object}  handler.ReadinessResponse  "所有服务就绪"
// @Failure      503  {object}  handler.ReadinessResponse  "部分服务不可用"
// @Router       /health/ready [get]
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
// @Summary      系统统计
// @Description  获取文档数量、查询次数等系统统计信息
// @Tags         系统监控
// @Produce      json
// @Success      200  {object}  handler.StatsResponse  "统计信息"
// @Failure      500  {object}  handler.ErrorResponse  "获取统计信息失败"
// @Router       /stats [get]
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
