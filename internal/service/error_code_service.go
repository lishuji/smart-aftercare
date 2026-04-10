package service

import (
	"smart-aftercare/internal/repository"
	"smart-aftercare/pkg/logger"
	"context"
)

// ErrorCodeService 故障代码服务
type ErrorCodeService struct {
	mysqlRepo  *repository.MySQLRepo
	ragService *RagService
}

// NewErrorCodeService 创建故障代码服务
func NewErrorCodeService(mysqlRepo *repository.MySQLRepo, ragService *RagService) *ErrorCodeService {
	return &ErrorCodeService{
		mysqlRepo:  mysqlRepo,
		ragService: ragService,
	}
}

// ErrorCodeResult 故障代码查询结果
type ErrorCodeResult struct {
	Code     string   `json:"code"`
	Answer   string   `json:"answer"`
	Sources  []string `json:"sources"`
	Images   []string `json:"images"`
	FromDB   bool     `json:"from_db"`   // 是否来自本地数据库
	FromRAG  bool     `json:"from_rag"`  // 是否来自 RAG 检索
}

// QueryErrorCode 故障代码查询（本地表优先，无匹配则走 RAG）
func (s *ErrorCodeService) QueryErrorCode(ctx context.Context, code, modelName string) (*ErrorCodeResult, error) {
	// 1. 本地 MySQL 查询故障代码表
	errorCode, err := s.mysqlRepo.GetErrorCodeByCodeAndModel(code, modelName)
	if err == nil && errorCode != nil {
		// 本地匹配成功
		answer := "故障代码" + code + "：" + errorCode.Reason + "\n解决方案：" + errorCode.Solution
		sources := []string{errorCode.Brand + " " + errorCode.Model + " 故障代码表"}

		return &ErrorCodeResult{
			Code:    code,
			Answer:  answer,
			Sources: sources,
			FromDB:  true,
		}, nil
	}

	logger.Infof("本地无匹配故障代码，走RAG检索: code=%s, model=%s", code, modelName)

	// 2. 降级走 RAG 检索
	ragQuery := "家电故障代码" + code + "是什么意思？如何解决？"
	qaResult, err := s.ragService.QA(ctx, ragQuery, "", modelName)
	if err != nil {
		return nil, err
	}

	return &ErrorCodeResult{
		Code:    code,
		Answer:  qaResult.Answer,
		Sources: qaResult.Sources,
		Images:  qaResult.Images,
		FromRAG: true,
	}, nil
}
