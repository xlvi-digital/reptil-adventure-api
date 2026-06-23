package models

import (
	"time"

	"gorm.io/datatypes"
)

type Product struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id_database"`
	ProductID    string         `gorm:"column:product_id;type:varchar(50);not null;unique" json:"id"`
	Name        string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	// Subtitle     string         `gorm:"column:subtitle;type:varchar(255)" json:"subtitle"`
	
	// 🆕 PERBAIKAN: Sekarang mendukung banyak gambar menggunakan tipe JSONB
	Image        datatypes.JSON `gorm:"column:image;type:jsonb" json:"image"` 
	
	// AccentColor  string         `gorm:"column:accent_color;type:varchar(20)" json:"accentColor"`
	Description  string         `gorm:"column:description;type:text" json:"description"`
	Price        float64        `gorm:"column:price;type:numeric;not null" json:"price"`
	Stock        float64        `gorm:"column:stock;type:numeric;not null" json:"stock"`
	Colors       datatypes.JSON `gorm:"column:colors;type:jsonb" json:"colors"`
	Sizes        datatypes.JSON `gorm:"column:sizes;type:jsonb" json:"sizes"`
	
	CategoryID   uint           `gorm:"column:category_id" json:"category_id"`
	Category     *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}