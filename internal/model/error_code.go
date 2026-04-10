package model

import (
	"time"

	"gorm.io/gorm"
)

// ErrorCode 故障代码模型
type ErrorCode struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `gorm:"type:varchar(50);not null;index;comment:故障代码" json:"code"`
	Brand     string         `gorm:"type:varchar(100);not null;index;comment:品牌" json:"brand"`
	Model     string         `gorm:"type:varchar(100);not null;index;comment:型号" json:"model"`
	Reason    string         `gorm:"type:text;not null;comment:故障原因" json:"reason"`
	Solution  string         `gorm:"type:text;not null;comment:解决方案" json:"solution"`
	Category  string         `gorm:"type:varchar(50);comment:故障分类(电路/制冷/排水等)" json:"category"`
	Severity  string         `gorm:"type:varchar(20);default:medium;comment:严重程度(low/medium/high/critical)" json:"severity"`
	Source    string         `gorm:"type:varchar(255);comment:数据来源" json:"source"`
}

// TableName 自定义表名
func (ErrorCode) TableName() string {
	return "error_codes"
}
