package models

// Model Provinsi
type Province struct {
	ID        string    `gorm:"primaryKey;size:2;column:id" json:"id"`
	Name      string    `gorm:"size:100;not null;column:name" json:"nama"` // di DB jadi kolom 'name'
	Regencies []Regency `gorm:"foreignKey:ProvinceID;references:ID"`
}

// Model Kabupaten / Kota
type Regency struct {
	ID         string     `gorm:"primaryKey;size:4;column:id" json:"id"`
	ProvinceID string     `gorm:"size:2;not null;column:province_id" json:"id_provinsi"` // di DB jadi 'province_id'
	Name       string     `gorm:"size:100;not null;column:name" json:"nama"`           // di DB jadi 'name'
	Districts  []District `gorm:"foreignKey:RegencyID;references:ID"`
}

// Model Kecamatan
type District struct {
	ID        string `gorm:"primaryKey;size:7;column:id" json:"id"`
	RegencyID string `gorm:"size:4;not null;column:regency_id" json:"id_kabupaten"` // di DB jadi 'regency_id'
	Name      string `gorm:"size:100;not null;column:name" json:"nama"`              // Paksa membaca json atau gunakan tag json:"nama" jika file kecamatan menggunakan key 'nama'
}