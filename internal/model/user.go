package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型（可选，用于管理上传者和查询记录）
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"type:varchar(100);uniqueIndex;not null;comment:用户名" json:"username"`
	Nickname  string         `gorm:"type:varchar(100);comment:昵称" json:"nickname"`
	Role      string         `gorm:"type:varchar(20);default:user;comment:角色(admin/user)" json:"role"`
	Status    string         `gorm:"type:varchar(20);default:active;comment:状态(active/disabled)" json:"status"`
}

// TableName 自定义表名
func (User) TableName() string {
	return "users"
}

// QueryLog 查询日志模型（用于记录用户查询和统计）
type QueryLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Query     string    `gorm:"type:text;not null;comment:查询内容" json:"query"`
	Brand     string    `gorm:"type:varchar(100);comment:品牌" json:"brand"`
	Model     string    `gorm:"type:varchar(100);comment:型号" json:"model"`
	QueryType string    `gorm:"type:varchar(20);comment:查询类型(qa/error_code)" json:"query_type"`
	Answer    string    `gorm:"type:text;comment:回答" json:"answer"`
	Sources   string    `gorm:"type:text;comment:来源(JSON)" json:"sources"`
	CacheHit  bool      `gorm:"default:false;comment:是否命中缓存" json:"cache_hit"`
	Duration  int64     `gorm:"comment:响应耗时(毫秒)" json:"duration"`
	UserIP    string    `gorm:"type:varchar(50);comment:用户IP" json:"user_ip"`
}

// TableName 自定义表名
func (QueryLog) TableName() string {
	return "query_logs"
}
