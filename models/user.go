package models

import (
	"time"
)

type Role struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(50);not null;unique" json:"name"`
}

type ShippingAddress struct {
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint   `gorm:"not null;unique" json:"user_id"` // One-to-One ke User
	Province      string `gorm:"type:varchar(100)" json:"province"`
	City          string `gorm:"type:varchar(100)" json:"city"`
	District      string `gorm:"type:varchar(100)" json:"district"`
	Village       string `gorm:"type:varchar(100)" json:"village"`
	PostalCode    string `gorm:"type:varchar(10)" json:"postal_code"`
	Coordinates   string `gorm:"type:varchar(255)" json:"coordinates"`
	MapReference  string `gorm:"type:varchar(255)" json:"map_reference"`
	ManualDetails string `gorm:"type:text" json:"manual_details"`
}

type User struct {
	ID              uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID          uint            `gorm:"not null" json:"role_id"`
	Role            Role            `gorm:"foreignKey:RoleID" json:"role"`
	Name            string          `gorm:"type:varchar(100);not null" json:"name"`
	Email           string          `gorm:"type:varchar(100);not null;unique" json:"email"`
	Password        string          `gorm:"type:varchar(255);not null" json:"-"`
	
	// 🌟 PERUBAHAN & TAMBAHAN FIELD BARU UNTUK UI:
	Phone           string          `gorm:"type:varchar(20)" json:"phone"`       // Diubah ke string agar nomor "08..." aman
	Gender          string          `gorm:"type:varchar(20)" json:"gender"`      // Menampung "Laki-laki" / "Perempuan"
	BirthDate       string          `gorm:"type:varchar(20)" json:"birth_date"`  // Menggunakan string agar format "YYYY-MM-DD" dari HTML date input langsung singkron
	ProfilePicture  string          `gorm:"type:varchar(255);default:'default-avatar.png'" json:"profile_picture"` // Menyimpan nama file/URL foto profil
	
	// Relasikan ShippingAddress ke dalam User (One-to-One)
	ShippingAddress ShippingAddress `gorm:"foreignKey:UserID" json:"shipping_address"`
	
	CreatedAt       time.Time       `json:"created_at"`
}