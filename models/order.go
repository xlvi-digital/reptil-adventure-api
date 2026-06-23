package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	OrderInvoice    string         `gorm:"type:varchar(100);unique;not null" json:"order_invoice"`
	CustomerID      uint           `gorm:"default:1" json:"customer_id"` // Opsional jika belum ada sistem login user
	CustomerName    string         `gorm:"type:varchar(255);not null" json:"customer_name"`
	CustomerEmail   string         `gorm:"type:varchar(255);not null" json:"customer_email"`
	ShippingAddress string         `gorm:"type:text;not null" json:"shipping_address"`
	Courier         string         `gorm:"type:varchar(50);not null" json:"courier"`
	ShippingCost    float64        `gorm:"type:numeric" json:"shipping_cost"`
	ProductTotal    float64        `gorm:"type:numeric" json:"product_total"`
	GrandTotal      float64        `gorm:"type:numeric" json:"grand_total"`
	Status          string         `gorm:"type:varchar(50);default:'PENDING'" json:"status"` // PENDING, PAID, SHIPPED, CANCELLED
	SnapToken       string         `gorm:"type:text" json:"snap_token,omitempty"`
	TrackingNumber  string         `gorm:"type:varchar(100)" json:"tracking_number,omitempty"` // Nomor Resi dari Admin
	CartItems       datatypes.JSON `gorm:"type:jsonb" json:"cart_items"` // Menyimpan snapshot keranjang belanja riil
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}