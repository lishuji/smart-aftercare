package handler

import (
	"smart-aftercare/internal/model"
	"smart-aftercare/internal/repository"
	"smart-aftercare/internal/service"
	"smart-aftercare/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// QAHandler 问答/故障代码查询接口处理器
type QAHandler struct {
	ragService       *service.RagService
	errorCodeService *service.ErrorCodeService
	mysqlRepo        *repository.MySQLRepo
}

// NewQAHandler 创建问答处理器
func NewQAHandler(ragService *service.RagService, errorCodeService *service.ErrorCodeService, mysqlRepo *repository.MySQLRepo) *QAHandler {
	return &QAHandler{
		ragService:       ragService,
		errorCodeService: errorCodeService,
		mysqlRepo:        mysqlRepo,
	}
}

// QARequest 问答请求
type QARequest struct {
	Query string `json:"query" binding:"required"`
	Model string `json:"model"`
	Brand string `json:"brand"`
}

// QA 智能问答接口
// POST /api/qa
func (h *QAHandler) QA(c *gin.Context) {
	startTime := time.Now()

	var req QARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 调用 RAG 服务
	result, err := h.ragService.QA(c.Request.Context(), req.Query, req.Brand, req.Model)
	if err != nil {
		logger.Errorf("QA查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败，请稍后重试",
		})
		return
	}

	// 异步保存查询日志
	duration := time.Since(startTime).Milliseconds()
	h.mysqlRepo.SaveQueryLogAsync(&model.QueryLog{
		Query:     req.Query,
		Brand:     req.Brand,
		Model:     req.Model,
		QueryType: "qa",
		Answer:    result.Answer,
		CacheHit:  result.CacheHit,
		Duration:  duration,
		UserIP:    c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"answer":    result.Answer,
			"sources":   result.Sources,
			"images":    result.Images,
			"cache_hit": result.CacheHit,
			"duration":  duration,
		},
	})
}

// ErrorCodeRequest 故障代码查询请求
type ErrorCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	Model string `json:"model"`
}

// QueryErrorCode 故障代码查询接口
// POST /api/qa/error-code
func (h *QAHandler) QueryErrorCode(c *gin.Context) {
	startTime := time.Now()

	var req ErrorCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 调用故障代码服务
	result, err := h.errorCodeService.QueryErrorCode(c.Request.Context(), req.Code, req.Model)
	if err != nil {
		logger.Errorf("故障代码查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败，请稍后重试",
		})
		return
	}

	// 异步保存查询日志
	duration := time.Since(startTime).Milliseconds()
	h.mysqlRepo.SaveQueryLogAsync(&model.QueryLog{
		Query:     "故障代码:" + req.Code,
		Model:     req.Model,
		QueryType: "error_code",
		Answer:    result.Answer,
		Duration:  duration,
		UserIP:    c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"code":    result.Code,
			"answer":  result.Answer,
			"sources": result.Sources,
			"images":  result.Images,
			"from_db":  result.FromDB,
			"from_rag": result.FromRAG,
			"duration": duration,
		},
	})
}
