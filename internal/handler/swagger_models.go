package handler

import "smart-aftercare/internal/model"

// ==================== 通用响应 ====================

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"参数错误"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:"操作成功"`
}

// ==================== 文档相关响应 ====================

// UploadDocumentResponse 文档上传响应
type UploadDocumentResponse struct {
	Code    int            `json:"code" example:"200"`
	Message string         `json:"message" example:"文档上传成功，正在后台处理"`
	Data    model.Document `json:"data"`
}

// DocumentDataResponse 文档详情响应
type DocumentDataResponse struct {
	Code int            `json:"code" example:"200"`
	Data model.Document `json:"data"`
}

// DocumentListData 文档列表数据
type DocumentListData struct {
	List     []model.Document `json:"list"`
	Total    int64            `json:"total" example:"100"`
	Page     int              `json:"page" example:"1"`
	PageSize int              `json:"page_size" example:"10"`
}

// DocumentListResponse 文档列表响应
type DocumentListResponse struct {
	Code int              `json:"code" example:"200"`
	Data DocumentListData `json:"data"`
}

// ==================== 问答相关响应 ====================

// QAResponseData 问答响应数据
type QAResponseData struct {
	Answer   string   `json:"answer" example:"根据说明书，当空调显示E1故障代码时..."`
	Sources  []string `json:"sources" example:"美的 KFR-35GW（第5页，故障排查）"`
	Images   []string `json:"images" example:"http://minio:9000/screenshots/img1.jpg"`
	CacheHit bool     `json:"cache_hit" example:"false"`
	Duration int64    `json:"duration" example:"1200"`
}

// QAResponse 问答响应
type QAResponse struct {
	Code int            `json:"code" example:"200"`
	Data QAResponseData `json:"data"`
}

// ErrorCodeResponseData 故障代码查询响应数据
type ErrorCodeResponseData struct {
	Code     string   `json:"code" example:"E1"`
	Answer   string   `json:"answer" example:"故障代码E1：室内温度传感器故障\n解决方案：检查传感器连接..."`
	Sources  []string `json:"sources" example:"美的 KFR-35GW 故障代码表"`
	Images   []string `json:"images"`
	FromDB   bool     `json:"from_db" example:"true"`
	FromRAG  bool     `json:"from_rag" example:"false"`
	Duration int64    `json:"duration" example:"50"`
}

// ErrorCodeResponse 故障代码查询响应
type ErrorCodeResponse struct {
	Code int                   `json:"code" example:"200"`
	Data ErrorCodeResponseData `json:"data"`
}

// ==================== 健康检查相关响应 ====================

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"smart-aftercare"`
}

// ReadinessResponse 就绪检查响应
type ReadinessResponse struct {
	Status      string      `json:"status" example:"ready"`
	MySQL       string      `json:"mysql" example:"healthy"`
	Redis       string      `json:"redis" example:"healthy"`
	Milvus      string      `json:"milvus" example:"healthy"`
	MilvusStats interface{} `json:"milvus_stats,omitempty"`
}

// StatsResponse 系统统计响应
type StatsResponse struct {
	Code int         `json:"code" example:"200"`
	Data interface{} `json:"data"`
}
