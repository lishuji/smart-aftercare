package service

import (
	"smart-aftercare/config"
	"smart-aftercare/internal/repository"
	"smart-aftercare/internal/util"
	"smart-aftercare/pkg/doubao"
	"smart-aftercare/pkg/logger"
	"context"
	"time"
)

// RagService RAG 检索+生成服务
type RagService struct {
	milvusRepo  *repository.MilvusRepo
	redisRepo   *repository.RedisRepo
	doubaoCfg   *config.DoubaoConfig
	doubaoClient *doubao.Client
}

// NewRagService 创建 RAG 服务
func NewRagService(
	milvusRepo *repository.MilvusRepo,
	redisRepo *repository.RedisRepo,
	doubaoCfg *config.DoubaoConfig,
) *RagService {
	return &RagService{
		milvusRepo:  milvusRepo,
		redisRepo:   redisRepo,
		doubaoCfg:   doubaoCfg,
		doubaoClient: doubao.NewClient(doubaoCfg.APIKey, doubaoCfg.ChatModel),
	}
}

// QAResult 问答结果
type QAResult struct {
	Answer   string   `json:"answer"`
	Sources  []string `json:"sources"`
	Images   []string `json:"images"`
	CacheHit bool     `json:"cache_hit"`
}

// QA 智能问答核心流程（缓存查询 → 检索 → 生成）
func (s *RagService) QA(ctx context.Context, query, brand, modelName string) (*QAResult, error) {
	// 1. 构建缓存键
	cacheKey := "rag:qa:" + modelName + ":" + query

	// 2. 查询缓存
	cachedAnswer, cachedSources, cachedImages, err := s.redisRepo.GetQACache(cacheKey)
	if err == nil && cachedAnswer != "" {
		logger.Info("命中缓存: ", cacheKey)
		return &QAResult{
			Answer:   cachedAnswer,
			Sources:  cachedSources,
			Images:   cachedImages,
			CacheHit: true,
		}, nil
	}

	// 3. 构建检索过滤条件（型号优先）
	filter := buildFilter(brand, modelName)

	// 4. 提取关键词
	keywords := util.ExtractApplianceKeywords(query)
	logger.Infof("提取关键词: %v", keywords)

	// 5. 双重检索：关键词检索 + 向量检索
	keywordResults, err := s.milvusRepo.SearchByKeywords(ctx, keywords, filter, 2)
	if err != nil {
		logger.Warnf("关键词检索失败: %v", err)
		keywordResults = nil
	}

	vectorResults, err := s.milvusRepo.SearchByVector(ctx, query,
		s.doubaoCfg.APIKey, s.doubaoCfg.EmbeddingModel, filter, 3)
	if err != nil {
		logger.Errorf("向量检索失败: %v", err)
		return nil, err
	}

	// 6. 合并结果并按章节优先级排序
	combinedResults := util.MergeAndRankResults(keywordResults, vectorResults)
	if len(combinedResults) == 0 {
		return &QAResult{
			Answer: "未查询到相关信息，请确认型号是否正确或尝试换一种方式提问。",
		}, nil
	}

	// 7. 构造上下文和来源信息
	contextText := util.BuildContextText(combinedResults)
	sources := util.FormatSources(combinedResults)
	images := util.CollectImageURLs(combinedResults)

	// 8. 大模型生成回答
	prompt := util.GenerateAppliancePrompt(query, contextText, modelName)
	answer, err := s.doubaoClient.Generate(ctx, prompt, s.doubaoCfg.Temperature, s.doubaoCfg.MaxToken)
	if err != nil {
		logger.Errorf("大模型生成失败: %v", err)
		return nil, err
	}

	// 9. 缓存结果（1 小时过期）
	if err := s.redisRepo.SetQACache(cacheKey, answer, sources, images, 1*time.Hour); err != nil {
		logger.Warnf("缓存设置失败: %v", err)
	}

	return &QAResult{
		Answer:  answer,
		Sources: sources,
		Images:  images,
	}, nil
}

// buildFilter 构建 Milvus 过滤表达式
func buildFilter(brand, modelName string) string {
	var conditions []string

	if modelName != "" {
		conditions = append(conditions, `model == "`+modelName+`"`)
	}
	if brand != "" {
		conditions = append(conditions, `brand == "`+brand+`"`)
	}

	if len(conditions) == 0 {
		return ""
	}

	return joinConditions(conditions, " and ")
}

// joinConditions 连接多个条件
func joinConditions(conditions []string, sep string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += sep
		}
		result += c
	}
	return result
}
