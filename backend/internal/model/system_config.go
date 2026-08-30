package model

import "time"

// SystemConfig 系统配置（键值对，值以 JSON 文本存储）
type SystemConfig struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigKey   string    `json:"config_key" gorm:"size:50;not null;uniqueIndex"` // 配置键
	ConfigValue string    `json:"config_value" gorm:"type:text;not null"`         // 配置值（JSON格式）
	Description string    `json:"description" gorm:"size:255"`                    // 配置说明
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
