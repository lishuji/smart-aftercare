package model

import (
	"time"

	"gorm.io/gorm"
)

// Document 文档元数据模型
type Document struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Filename   string         `gorm:"type:varchar(255);not null;comment:文件名" json:"filename"`
	FileType   string         `gorm:"type:varchar(20);not null;default:pdf;comment:文件类型" json:"file_type"`
	Brand      string         `gorm:"type:varchar(100);not null;index;comment:品牌" json:"brand"`
	Model      string         `gorm:"type:varchar(100);not null;index;comment:型号" json:"model"`
	Uploader   string         `gorm:"type:varchar(100);comment:上传人" json:"uploader"`
	UploadTime time.Time      `gorm:"comment:上传时间" json:"upload_time"`
	Status     string         `gorm:"type:varchar(20);not null;default:pending;comment:状态(pending/processing/processed/failed)" json:"status"`
	PageCount  int            `gorm:"default:0;comment:页数" json:"page_count"`
	SliceCount int            `gorm:"default:0;comment:切片数量" json:"slice_count"`
	FileSize   int64          `gorm:"default:0;comment:文件大小(字节)" json:"file_size"`
	FilePath   string         `gorm:"type:varchar(500);comment:文件存储路径" json:"file_path"`
	Remark     string         `gorm:"type:text;comment:备注" json:"remark"`
}

// TableName 自定义表名
func (Document) TableName() string {
	return "documents"
}
