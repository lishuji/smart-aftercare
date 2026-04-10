package repository

import (
	"smart-aftercare/config"
	"smart-aftercare/pkg/doubao"
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	// VectorDim 向量维度（豆包 Embedding-v1 为 768 维）
	VectorDim = 768
)

// VectorSlice 文本切片+元数据
type VectorSlice struct {
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Score    float32           `json:"score,omitempty"`
}

// MilvusRepo Milvus 向量库数据访问层
type MilvusRepo struct {
	client     client.Client
	collection string
	dim        int
}

// NewMilvusRepo 创建 Milvus 数据访问层
func NewMilvusRepo(cfg config.MilvusConfig) (*MilvusRepo, error) {
	ctx := context.Background()

	// 连接 Milvus
	c, err := client.NewGrpcClient(ctx, cfg.Address())
	if err != nil {
		return nil, fmt.Errorf("连接Milvus失败: %w", err)
	}

	// 检查集合是否存在
	exists, err := c.HasCollection(ctx, cfg.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("检查集合失败: %w", err)
	}

	if !exists {
		if err := createCollection(ctx, c, cfg.CollectionName); err != nil {
			return nil, err
		}
	}

	// 加载集合到内存
	if err := c.LoadCollection(ctx, cfg.CollectionName, false); err != nil {
		return nil, fmt.Errorf("加载集合失败: %w", err)
	}

	return &MilvusRepo{
		client:     c,
		collection: cfg.CollectionName,
		dim:        VectorDim,
	}, nil
}

// createCollection 创建 Milvus 集合
func createCollection(ctx context.Context, c client.Client, collectionName string) error {
	// 定义 schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: fmt.Sprintf("%d", VectorDim),
				},
			},
			{
				Name:     "content",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "4000",
				},
			},
			{
				Name:     "brand",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "100",
				},
			},
			{
				Name:     "model",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "100",
				},
			},
			{
				Name:     "chapter",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "200",
				},
			},
			{
				Name:     "page",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "20",
				},
			},
			{
				Name:     "source",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "500",
				},
			},
			{
				Name:     "image_url",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "1000",
				},
			},
		},
	}

	// 创建集合
	if err := c.CreateCollection(ctx, schema, 2); err != nil {
		return fmt.Errorf("创建集合失败: %w", err)
	}

	// 创建向量索引（IVF_FLAT，适合中小规模数据）
	idx, err := entity.NewIndexIvfFlat(entity.L2, 128)
	if err != nil {
		return fmt.Errorf("创建索引参数失败: %w", err)
	}
	if err := c.CreateIndex(ctx, collectionName, "vector", idx, false); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	return nil
}

// InsertSlices 文本切片向量化并插入 Milvus
func (m *MilvusRepo) InsertSlices(slices []*VectorSlice, apiKey, embeddingModel string) error {
	if len(slices) == 0 {
		return nil
	}

	// 1. 批量向量化
	contents := make([]string, len(slices))
	for i, slice := range slices {
		contents[i] = slice.Content
	}

	// 分批向量化（每批不超过 16 条，避免 API 限流）
	batchSize := 16
	allVectors := make([][]float32, 0, len(slices))
	for i := 0; i < len(contents); i += batchSize {
		end := i + batchSize
		if end > len(contents) {
			end = len(contents)
		}
		batch := contents[i:end]
		vectors, err := doubao.GenerateEmbeddings(apiKey, embeddingModel, batch)
		if err != nil {
			return fmt.Errorf("向量化失败(batch %d-%d): %w", i, end, err)
		}
		allVectors = append(allVectors, vectors...)
	}

	// 2. 构造插入数据
	vectorCol := entity.NewColumnFloatVector("vector", VectorDim, allVectors)
	contentCol := entity.NewColumnVarChar("content", extractField(slices, "content"))
	brandCol := entity.NewColumnVarChar("brand", extractMetadata(slices, "brand"))
	modelCol := entity.NewColumnVarChar("model", extractMetadata(slices, "model"))
	chapterCol := entity.NewColumnVarChar("chapter", extractMetadata(slices, "chapter"))
	pageCol := entity.NewColumnVarChar("page", extractMetadata(slices, "page"))
	sourceCol := entity.NewColumnVarChar("source", extractMetadata(slices, "source"))
	imageURLCol := entity.NewColumnVarChar("image_url", extractMetadata(slices, "image_url"))

	// 3. 插入 Milvus
	ctx := context.Background()
	_, err := m.client.Insert(ctx, m.collection, "",
		vectorCol, contentCol, brandCol, modelCol, chapterCol, pageCol, sourceCol, imageURLCol,
	)
	if err != nil {
		return fmt.Errorf("向量数据插入失败: %w", err)
	}

	// 4. 刷新集合
	if err := m.client.Flush(ctx, m.collection, false); err != nil {
		return fmt.Errorf("刷新集合失败: %w", err)
	}

	return nil
}

// SearchByVector 向量检索
func (m *MilvusRepo) SearchByVector(ctx context.Context, query string, apiKey, embeddingModel, filter string, topK int) ([]*VectorSlice, error) {
	// 1. Query 向量化
	vector, err := doubao.GenerateEmbedding(apiKey, embeddingModel, query)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	// 2. 构造检索参数
	sp, err := entity.NewIndexIvfFlatSearchParam(10)
	if err != nil {
		return nil, fmt.Errorf("创建检索参数失败: %w", err)
	}

	// 3. 构造查询向量
	vectors := []entity.Vector{entity.FloatVector(vector)}

	// 4. 执行检索
	outputFields := []string{"content", "brand", "model", "chapter", "page", "source", "image_url"}
	results, err := m.client.Search(
		ctx,
		m.collection,
		nil,          // partition names
		filter,       // expr filter
		outputFields, // output fields
		vectors,      // query vectors
		"vector",     // vector field name
		entity.L2,    // metric type
		topK,         // topK
		sp,           // search param
	)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}

	// 5. 解析结果
	return parseSearchResults(results), nil
}

// SearchByKeywords 关键词检索（基于元数据+文本匹配）
func (m *MilvusRepo) SearchByKeywords(ctx context.Context, keywords []string, filter string, topK int) ([]*VectorSlice, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	// 构造关键词查询表达式
	keywordExpr := ""
	for i, kw := range keywords {
		if i > 0 {
			keywordExpr += " or "
		}
		keywordExpr += fmt.Sprintf(`content like "%%%s%%"`, kw)
	}

	// 合并过滤条件
	finalExpr := keywordExpr
	if filter != "" {
		finalExpr = filter + " and (" + keywordExpr + ")"
	}

	// 执行查询
	outputFields := []string{"content", "brand", "model", "chapter", "page", "source", "image_url"}
	queryResult, err := m.client.Query(
		ctx,
		m.collection,
		nil, // partition names
		finalExpr,
		outputFields,
	)
	if err != nil {
		return nil, fmt.Errorf("关键词检索失败: %w", err)
	}

	// 解析结果
	return parseQueryResults(queryResult, topK), nil
}

// ==================== 辅助函数 ====================

// extractField 提取切片的 Content 字段
func extractField(slices []*VectorSlice, _ string) []string {
	result := make([]string, len(slices))
	for i, s := range slices {
		result[i] = s.Content
	}
	return result
}

// extractMetadata 提取切片的元数据字段
func extractMetadata(slices []*VectorSlice, key string) []string {
	result := make([]string, len(slices))
	for i, s := range slices {
		if v, ok := s.Metadata[key]; ok {
			result[i] = v
		} else {
			result[i] = ""
		}
	}
	return result
}

// parseSearchResults 解析向量检索结果
func parseSearchResults(results []client.SearchResult) []*VectorSlice {
	var slices []*VectorSlice

	for _, result := range results {
		for i := 0; i < result.ResultCount; i++ {
			slice := &VectorSlice{
				Metadata: make(map[string]string),
			}

			// 提取各字段
			for _, field := range result.Fields {
				col, ok := field.(*entity.ColumnVarChar)
				if !ok {
					continue
				}
				val, err := col.ValueByIdx(i)
				if err != nil {
					continue
				}
				switch col.Name() {
				case "content":
					slice.Content = val
				case "brand":
					slice.Metadata["brand"] = val
				case "model":
					slice.Metadata["model"] = val
				case "chapter":
					slice.Metadata["chapter"] = val
				case "page":
					slice.Metadata["page"] = val
				case "source":
					slice.Metadata["source"] = val
				case "image_url":
					if val != "" {
						slice.Metadata["image_url"] = val
					}
				}
			}

			// 获取分数
			if i < len(result.Scores) {
				slice.Score = result.Scores[i]
			}

			slices = append(slices, slice)
		}
	}

	return slices
}

// parseQueryResults 解析关键词查询结果
func parseQueryResults(columns []entity.Column, topK int) []*VectorSlice {
	if len(columns) == 0 {
		return nil
	}

	// 获取结果行数
	rowCount := 0
	for _, col := range columns {
		rowCount = col.Len()
		break
	}

	if topK > 0 && rowCount > topK {
		rowCount = topK
	}

	var slices []*VectorSlice
	for i := 0; i < rowCount; i++ {
		slice := &VectorSlice{
			Metadata: make(map[string]string),
		}

		for _, col := range columns {
			varcharCol, ok := col.(*entity.ColumnVarChar)
			if !ok {
				continue
			}
			val, err := varcharCol.ValueByIdx(i)
			if err != nil {
				continue
			}
			switch varcharCol.Name() {
			case "content":
				slice.Content = val
			case "brand":
				slice.Metadata["brand"] = val
			case "model":
				slice.Metadata["model"] = val
			case "chapter":
				slice.Metadata["chapter"] = val
			case "page":
				slice.Metadata["page"] = val
			case "source":
				slice.Metadata["source"] = val
			case "image_url":
				if val != "" {
					slice.Metadata["image_url"] = val
				}
			}
		}

		slices = append(slices, slice)
	}

	return slices
}

// Close 关闭 Milvus 连接
func (m *MilvusRepo) Close() error {
	return m.client.Close()
}

// GetCollectionStats 获取集合统计信息
func (m *MilvusRepo) GetCollectionStats(ctx context.Context) (map[string]string, error) {
	return m.client.GetCollectionStatistics(ctx, m.collection)
}
