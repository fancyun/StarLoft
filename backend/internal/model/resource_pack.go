package model

import "time"

// ResourcePack 平台资源包定义
type ResourcePack struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;not null"` // 资源包名称
	TotalCount  int       `json:"total_count" gorm:"not null"`   // 认证次数
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null"` // 售价（元）
	Stock       int       `json:"stock" gorm:"not null;default:-1"` // 库存：-1 不限量，>=0 限量剩余库存
	Status      int       `json:"status" gorm:"type:tinyint;not null;default:1;index"` // 状态：1-上架 0-下架
	Description string    `json:"description" gorm:"size:255"` // 描述
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserResourcePack 用户已购资源包
type UserResourcePack struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         int64     `json:"user_id" gorm:"not null;index;index:idx_user_status,priority:1"`
	PackID         int64     `json:"pack_id" gorm:"not null"`
	PackName       string    `json:"pack_name" gorm:"size:100;not null"` // 资源包名称（快照）
	TotalCount     int       `json:"total_count" gorm:"not null"`
	RemainingCount int       `json:"remaining_count" gorm:"not null"`
	Status         int       `json:"status" gorm:"type:tinyint;not null;default:1;index:idx_user_status,priority:2"` // 1-有效 0-已耗尽/禁用
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
