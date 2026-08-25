package model

import "time"

// ResourcePack 平台资源包定义
type ResourcePack struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`         // 资源包名称
	TotalCount  int       `json:"total_count"`  // 认证次数
	Price       float64   `json:"price"`        // 售价（元）
	Stock       int       `json:"stock"`        // 库存：-1 不限量，>=0 限量剩余库存
	Status      int       `json:"status"`       // 状态：1-上架 0-下架
	Description string    `json:"description"`  // 描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserResourcePack 用户已购资源包
type UserResourcePack struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	PackID         int64     `json:"pack_id"`
	PackName       string    `json:"pack_name"` // 资源包名称（快照）
	TotalCount     int       `json:"total_count"`
	RemainingCount int       `json:"remaining_count"`
	Status         int       `json:"status"` // 1-有效 0-已耗尽/禁用
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
