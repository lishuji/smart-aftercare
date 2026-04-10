package handler

import (
	"smart-aftercare/internal/service"
	"smart-aftercare/pkg/logger"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DocumentHandler 文档上传/解析接口处理器
type DocumentHandler struct {
	documentService *service.DocumentService
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(documentService *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// UploadDocumentRequest 上传文档请求参数（FormData）
type UploadDocumentRequest struct {
	Brand    string `form:"brand" binding:"required"`
	Model    string `form:"model" binding:"required"`
	Uploader string `form:"uploader"`
}

// UploadDocument 文档上传接口
// @Summary      上传文档
// @Description  上传家电说明书文档（PDF/DOC/DOCX/TXT），系统将异步解析、切片、向量化入库
// @Tags         文档管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file      formData  file    true   "上传的文件（支持 PDF/DOC/DOCX/TXT，最大 50MB）"
// @Param        brand     formData  string  true   "品牌名称"
// @Param        model     formData  string  true   "型号"
// @Param        uploader  formData  string  false  "上传人"
// @Success      200  {object}  handler.UploadDocumentResponse  "上传成功"
// @Failure      400  {object}  handler.ErrorResponse           "参数错误"
// @Failure      500  {object}  handler.ErrorResponse           "服务器内部错误"
// @Router       /document/upload [post]
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	// 1. 解析表单参数
	var req UploadDocumentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 2. 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请选择要上传的文件",
		})
		return
	}

	// 3. 校验文件类型
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".txt": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的文件类型，仅支持 PDF、DOC、DOCX、TXT",
		})
		return
	}

	// 4. 校验文件大小（最大 50MB）
	maxSize := int64(50 << 20)
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件大小超过限制（最大50MB）",
		})
		return
	}

	// 5. 保存到本地上传目录
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		logger.Errorf("创建上传目录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "服务器内部错误",
		})
		return
	}

	// 使用品牌+型号+原始文件名作为存储路径
	savePath := filepath.Join(uploadDir, req.Brand+"_"+req.Model+"_"+file.Filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		logger.Errorf("保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "文件保存失败",
		})
		return
	}

	// 6. 调用文档处理服务（异步处理）
	doc, err := h.documentService.UploadAndProcess(savePath, req.Brand, req.Model, req.Uploader)
	if err != nil {
		logger.Errorf("文档处理启动失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "文档处理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "文档上传成功，正在后台处理",
		"data":    doc,
	})
}

// GetDocument 获取文档详情
// @Summary      获取文档详情
// @Description  根据文档 ID 获取文档详细信息
// @Tags         文档管理
// @Produce      json
// @Param        id   path      int  true  "文档 ID"
// @Success      200  {object}  handler.DocumentDataResponse  "文档详情"
// @Failure      400  {object}  handler.ErrorResponse         "无效的文档ID"
// @Failure      404  {object}  handler.ErrorResponse         "文档不存在"
// @Router       /document/{id} [get]
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文档ID",
		})
		return
	}

	doc, err := h.documentService.GetDocumentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文档不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": doc,
	})
}

// ListDocuments 文档列表
// @Summary      文档列表
// @Description  分页查询文档列表，支持按品牌、型号筛选
// @Tags         文档管理
// @Produce      json
// @Param        brand      query     string  false  "品牌名称"
// @Param        model      query     string  false  "型号"
// @Param        page       query     int     false  "页码（默认 1）"
// @Param        page_size  query     int     false  "每页条数（默认 10，最大 100）"
// @Success      200  {object}  handler.DocumentListResponse  "文档列表"
// @Failure      500  {object}  handler.ErrorResponse         "查询失败"
// @Router       /documents [get]
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	brand := c.Query("brand")
	modelName := c.Query("model")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	docs, total, err := h.documentService.ListDocuments(brand, modelName, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":      docs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// DeleteDocument 删除文档
// @Summary      删除文档
// @Description  根据文档 ID 删除文档及其关联的向量数据
// @Tags         文档管理
// @Produce      json
// @Param        id   path      int  true  "文档 ID"
// @Success      200  {object}  handler.MessageResponse  "删除成功"
// @Failure      400  {object}  handler.ErrorResponse    "无效的文档ID"
// @Failure      500  {object}  handler.ErrorResponse    "删除失败"
// @Router       /document/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文档ID",
		})
		return
	}

	if err := h.documentService.DeleteDocument(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
