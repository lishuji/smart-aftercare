package repository

import (
	"smart-aftercare/config"
	"smart-aftercare/internal/model"
	"smart-aftercare/pkg/logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// MySQLRepo MySQL 数据访问层
type MySQLRepo struct {
	db *gorm.DB
}

// NewMySQLRepo 创建 MySQL 数据访问层
func NewMySQLRepo(cfg config.MySQLConfig) (*MySQLRepo, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &MySQLRepo{db: db}, nil
}

// AutoMigrate 自动迁移数据库表结构
func (m *MySQLRepo) AutoMigrate() error {
	return m.db.AutoMigrate(
		&model.Document{},
		&model.ErrorCode{},
		&model.User{},
		&model.QueryLog{},
	)
}

// GetDB 获取底层 GORM DB 实例（谨慎使用）
func (m *MySQLRepo) GetDB() *gorm.DB {
	return m.db
}

// ==================== Document 操作 ====================

// CreateDocument 创建文档记录
func (m *MySQLRepo) CreateDocument(doc *model.Document) error {
	if err := m.db.Create(doc).Error; err != nil {
		return fmt.Errorf("创建文档记录失败: %w", err)
	}
	return nil
}

// GetDocumentByID 根据 ID 查询文档
func (m *MySQLRepo) GetDocumentByID(id uint) (*model.Document, error) {
	var doc model.Document
	if err := m.db.First(&doc, id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// UpdateDocumentStatus 更新文档状态
func (m *MySQLRepo) UpdateDocumentStatus(id uint, status string) error {
	return m.db.Model(&model.Document{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateDocument 更新文档信息
func (m *MySQLRepo) UpdateDocument(doc *model.Document) error {
	return m.db.Save(doc).Error
}

// ListDocuments 分页查询文档列表
func (m *MySQLRepo) ListDocuments(brand, modelName string, page, pageSize int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64

	query := m.db.Model(&model.Document{})
	if brand != "" {
		query = query.Where("brand = ?", brand)
	}
	if modelName != "" {
		query = query.Where("model = ?", modelName)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// DeleteDocument 删除文档（软删除）
func (m *MySQLRepo) DeleteDocument(id uint) error {
	return m.db.Delete(&model.Document{}, id).Error
}

// ==================== ErrorCode 操作 ====================

// GetErrorCodeByCodeAndModel 根据故障代码和型号查询
func (m *MySQLRepo) GetErrorCodeByCodeAndModel(code, modelName string) (*model.ErrorCode, error) {
	var ec model.ErrorCode
	query := m.db.Where("code = ?", code)
	if modelName != "" {
		query = query.Where("model = ?", modelName)
	}
	if err := query.First(&ec).Error; err != nil {
		return nil, err
	}
	return &ec, nil
}

// CreateErrorCode 创建故障代码
func (m *MySQLRepo) CreateErrorCode(ec *model.ErrorCode) error {
	return m.db.Create(ec).Error
}

// BatchCreateErrorCodes 批量创建故障代码
func (m *MySQLRepo) BatchCreateErrorCodes(codes []model.ErrorCode) error {
	if len(codes) == 0 {
		return nil
	}
	return m.db.CreateInBatches(codes, 100).Error
}

// ListErrorCodes 分页查询故障代码
func (m *MySQLRepo) ListErrorCodes(brand, modelName string, page, pageSize int) ([]model.ErrorCode, int64, error) {
	var codes []model.ErrorCode
	var total int64

	query := m.db.Model(&model.ErrorCode{})
	if brand != "" {
		query = query.Where("brand = ?", brand)
	}
	if modelName != "" {
		query = query.Where("model = ?", modelName)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("code ASC").Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// ==================== QueryLog 操作 ====================

// CreateQueryLog 创建查询日志
func (m *MySQLRepo) CreateQueryLog(log *model.QueryLog) error {
	return m.db.Create(log).Error
}

// SaveQueryLogAsync 异步保存查询日志（不阻塞主流程）
func (m *MySQLRepo) SaveQueryLogAsync(ql *model.QueryLog) {
	go func() {
		if err := m.CreateQueryLog(ql); err != nil {
			logger.Warnf("保存查询日志失败: %v", err)
		}
	}()
}

// Ping 检查数据库连接
func (m *MySQLRepo) Ping() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close 关闭数据库连接
func (m *MySQLRepo) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDocumentStats 获取文档统计信息
func (m *MySQLRepo) GetDocumentStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalDocs int64
	m.db.Model(&model.Document{}).Count(&totalDocs)
	stats["total_documents"] = totalDocs

	var processedDocs int64
	m.db.Model(&model.Document{}).Where("status = ?", "processed").Count(&processedDocs)
	stats["processed_documents"] = processedDocs

	var totalErrorCodes int64
	m.db.Model(&model.ErrorCode{}).Count(&totalErrorCodes)
	stats["total_error_codes"] = totalErrorCodes

	// 统计今日查询数
	today := time.Now().Format("2006-01-02")
	var todayQueries int64
	m.db.Model(&model.QueryLog{}).Where("DATE(created_at) = ?", today).Count(&todayQueries)
	stats["today_queries"] = todayQueries

	return stats, nil
}
