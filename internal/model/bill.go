package model

import (
	"time"

	"gorm.io/gorm"
)

// Bill 账单模型
type Bill struct {
	ID         uint64         `gorm:"primarykey;comment:账单ID" json:"id"`
	UserID     uint64         `gorm:"not null;index:idx_user_id;comment:所属用户ID;foreignKey:UserID" json:"user_id"` // 关联用户，加索引
	CategoryID uint64         `gorm:"not null;index:idx_user_category;comment:分类ID" json:"category_id"`           // 关联分类，复合索引
	Amount     float64        `gorm:"not null;type:decimal(10,2);comment:金额只有正数，入账前需检查" json:"amount"`
	Type       bool           `gorm:"not null;size:10;index:idx_user_type;comment:类型（income(1)/expense(0)）" json:"type"` // 收入/支出
	Remark     string         `gorm:"size:255;comment:备注" json:"remark"`
	CreatedAt  time.Time      `gorm:"index:idx_user_created_at;comment:创建时间" json:"created_at"` // 按时间筛选索引
	UpdatedAt  time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"` // 软删除
}

// TableName 指定表名
func (b *Bill) TableName() string {
	return "bills"
}
