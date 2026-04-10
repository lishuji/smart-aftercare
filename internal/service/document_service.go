package service

import (
	"smart-aftercare/config"
	"smart-aftercare/internal/model"
	"smart-aftercare/internal/repository"
	"smart-aftercare/internal/util"
	"smart-aftercare/pkg/logger"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DocumentService 文档处理服务
type DocumentService struct {
	mysqlRepo  *repository.MySQLRepo
	milvusRepo *repository.MilvusRepo
	minioRepo  *repository.MinioRepo
	cfg        *config.Config
}

// NewDocumentService 创建文档处理服务
func NewDocumentService(
	mysqlRepo *repository.MySQLRepo,
	milvusRepo *repository.MilvusRepo,
	minioRepo *repository.MinioRepo,
	cfg *config.Config,
) *DocumentService {
	return &DocumentService{
		mysqlRepo:  mysqlRepo,
		milvusRepo: milvusRepo,
		minioRepo:  minioRepo,
		cfg:        cfg,
	}
}

// UploadAndProcess 上传并处理家电说明书
func (s *DocumentService) UploadAndProcess(filePath, brand, modelName, uploader string) (*model.Document, error) {
	// 1. 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 2. 创建文档记录（状态：processing）
	doc := &model.Document{
		Filename:   filepath.Base(filePath),
		FileType:   strings.TrimPrefix(filepath.Ext(filePath), "."),
		Brand:      brand,
		Model:      modelName,
		Uploader:   uploader,
		UploadTime: time.Now(),
		Status:     "processing",
		FileSize:   fileInfo.Size(),
		FilePath:   filePath,
	}

	if err := s.mysqlRepo.CreateDocument(doc); err != nil {
		return nil, fmt.Errorf("创建文档记录失败: %w", err)
	}

	// 3. 异步处理文档（不阻塞上传响应）
	go func() {
		if err := s.processDocument(doc, filePath, brand, modelName); err != nil {
			logger.Errorf("文档处理失败[%d]: %v", doc.ID, err)
			_ = s.mysqlRepo.UpdateDocumentStatus(doc.ID, "failed")
			doc.Remark = err.Error()
			_ = s.mysqlRepo.UpdateDocument(doc)
		}
	}()

	return doc, nil
}

// processDocument 处理文档核心流程：解析 → 切片 → 向量化 → 入库
func (s *DocumentService) processDocument(doc *model.Document, filePath, brand, modelName string) error {
	logger.Infof("开始处理文档: %s (ID: %d)", doc.Filename, doc.ID)

	// 1. 解析 PDF 文件，提取文本和图片
	pages, err := s.parsePDF(filePath)
	if err != nil {
		return fmt.Errorf("PDF解析失败: %w", err)
	}

	doc.PageCount = len(pages)
	logger.Infof("PDF解析完成，共 %d 页", len(pages))

	// 2. 提取章节结构
	chapters := util.ParseChapters(pages)
	logger.Infof("提取到 %d 个章节", len(chapters))

	// 3. 文本切片 + 添加元数据
	var slices []*repository.VectorSlice
	for _, page := range pages {
		// 获取当前页所属章节
		chapter := util.GetCurrentChapter(page.PageNum, chapters)

		// 提取文本（含 OCR）
		fullText := util.ExtractTextWithOCR(page.Text, page.Images)

		// 按 300 字切片，30 字重叠
		textSlices := util.SplitText(fullText, 300, 30)
		for _, sliceText := range textSlices {
			if strings.TrimSpace(sliceText) == "" {
				continue
			}
			slice := &repository.VectorSlice{
				Content: sliceText,
				Metadata: map[string]string{
					"brand":   brand,
					"model":   modelName,
					"chapter": chapter.Title,
					"page":    strconv.Itoa(page.PageNum),
					"source":  doc.Filename,
				},
			}
			slices = append(slices, slice)
		}

		// 4. 处理图片：上传到 MinIO 并关联元数据
		if len(page.Images) > 0 {
			s.processPageImages(page, brand, modelName, chapter, slices)
		}
	}

	doc.SliceCount = len(slices)
	logger.Infof("文本切片完成，共 %d 个切片", len(slices))

	// 5. 向量化并写入 Milvus
	if len(slices) > 0 {
		if err := s.milvusRepo.InsertSlices(slices, s.cfg.Doubao.APIKey, s.cfg.Doubao.EmbeddingModel); err != nil {
			return fmt.Errorf("向量化入库失败: %w", err)
		}
		logger.Info("向量化入库完成")
	}

	// 6. 上传原始文件到 MinIO
	objectKey := fmt.Sprintf("documents/%s/%s/%s", brand, modelName, doc.Filename)
	if _, err := s.minioRepo.UploadFile(filePath, objectKey); err != nil {
		logger.Warnf("原始文件上传MinIO失败: %v", err)
	}

	// 7. 更新文档状态为已处理
	doc.Status = "processed"
	if err := s.mysqlRepo.UpdateDocument(doc); err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	logger.Infof("文档处理完成: %s (ID: %d, 切片: %d)", doc.Filename, doc.ID, len(slices))
	return nil
}

// parsePDF 解析 PDF 文件，提取每页的文本和图片
func (s *DocumentService) parsePDF(filePath string) ([]util.PageContent, error) {
	// TODO: 集成 unipdf 实现完整的 PDF 解析
	// 当前为简化实现：读取 PDF 文件并按页解析
	//
	// 完整实现流程：
	// 1. 使用 unipdf/model.NewPdfReaderFromFile(filePath) 打开 PDF
	// 2. 遍历每页，使用 extractor.New(page) 提取文本
	// 3. 提取每页的 XObject 图片
	// 4. 返回 PageContent 列表

	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("PDF文件不存在: %w", err)
	}

	// 简化实现：读取文件并创建单页内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取PDF文件失败: %w", err)
	}

	// 返回占位内容（实际应按页解析）
	pages := []util.PageContent{
		{
			PageNum: 1,
			Text:    string(data), // 简化处理，实际需要用 PDF 库解析
			Images:  nil,
		},
	}

	logger.Warnf("使用简化PDF解析，需集成unipdf库以支持完整解析")
	return pages, nil
}

// processPageImages 处理页面中的图片
func (s *DocumentService) processPageImages(
	page util.PageContent,
	brand, modelName string,
	chapter *util.Chapter,
	slices []*repository.VectorSlice,
) {
	for imgIdx, imgData := range page.Images {
		if len(imgData) == 0 {
			continue
		}

		// 生成对象键
		objectKey := fmt.Sprintf("screenshots/%s/%s/page%d_img%d.jpg",
			brand, modelName, page.PageNum, imgIdx)

		// 保存到临时文件并上传
		tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_page%d_img%d.jpg",
			modelName, page.PageNum, imgIdx))

		if err := os.WriteFile(tmpPath, imgData, 0o644); err != nil {
			logger.Warnf("保存临时图片失败: %v", err)
			continue
		}
		defer os.Remove(tmpPath)

		imgURL, err := s.minioRepo.UploadFile(tmpPath, objectKey)
		if err != nil {
			logger.Warnf("上传图片到MinIO失败: %v", err)
			continue
		}

		// 关联图片 URL 到对应的切片元数据
		pageStr := strconv.Itoa(page.PageNum)
		for _, slice := range slices {
			if slice.Metadata["page"] == pageStr && slice.Metadata["chapter"] == chapter.Title {
				slice.Metadata["image_url"] = imgURL
				break
			}
		}
	}
}

// GetDocumentByID 根据 ID 获取文档
func (s *DocumentService) GetDocumentByID(id uint) (*model.Document, error) {
	return s.mysqlRepo.GetDocumentByID(id)
}

// ListDocuments 列出文档
func (s *DocumentService) ListDocuments(brand, modelName string, page, pageSize int) ([]model.Document, int64, error) {
	return s.mysqlRepo.ListDocuments(brand, modelName, page, pageSize)
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(id uint) error {
	return s.mysqlRepo.DeleteDocument(id)
}
