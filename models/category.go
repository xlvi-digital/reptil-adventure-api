package models

import (
	"time"
)

type Category struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"` // "Tas & Carrier", "Pakaian", dll
	Slug      string    `gorm:"type:varchar(100);not null;unique" json:"slug"` // "tas-carrier", "pakaian", dll
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Has Many: Satu kategori bisa dipakai oleh banyak produk
	Products  []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}