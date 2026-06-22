package model

import (
	"time"
)

// ProductImage 商品图片
type ProductImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"index" json:"product_id"`         // 商品ID
	URL       string    `gorm:"size:500" json:"image_url"`       // 图片URL
	SortOrder int       `gorm:"default:0" json:"sort_order"`     // 排序顺序
	IsPrimary bool      `gorm:"default:false" json:"is_primary"` // 是否主图
	CreatedAt time.Time `json:"created_at"`
}
