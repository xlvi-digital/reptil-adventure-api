package config

import (
	"fmt"
	"log"
	"os"
	"reptil-adventure-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// 🚀 Ambil string koneksi utuh dari environment variable Supabase
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback ke konfigurasi lokal jika DATABASE_URL di PC kamu belum diset
		host := "localhost"
		user := "postgres"
		password := "Reptilbukanhewan26"
		dbName := "reptil_adventure"
		port := "5432"
		sslMode := "disable"
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", host, user, password, dbName, port, sslMode)
	}

	// 🚀 Buka koneksi menggunakan driver postgres GORM dengan konfigurasi khusus PgBouncer Supabase
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // 🔥 WAJIB true agar tidak error bentrok dengan transaction pooler (PgBouncer) Supabase
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	
	if err != nil {
		log.Fatal("Gagal terhubung ke database PostgreSQL Supabase: ", err)
	}

	// EKSEKUSI MIGRASI SATU PER SATU AGAR AMAN DAN TIDAK REBUTAN CONTEXT
	fmt.Println("⏳ Melakukan migrasi struktur tabel...")
	err = database.AutoMigrate(
		&models.Role{}, 
		&models.User{}, 
		&models.Category{}, 
		&models.Product{}, 
		&models.Order{}, 
		&models.Province{}, // Mendaftarkan model Provinsi
		&models.Regency{},  // Mendaftarkan model Kabupaten
		&models.District{}, // Mendaftarkan model Kecamatan
		&models.ShippingAddress{},
	)
	if err != nil {
		log.Fatal("Gagal Auto Migration: ", err)
	}

	fmt.Println("🚀 KONEKSI DAN SELURUH TABEL DATABASE BERHASIL DIMIGRASI!")
	DB = database

	// LOGIKA SEEDING ROLES
	expectedRoles := map[uint]string{
		1: "Developer",
		2: "Owner",
		3: "Admin",
		4: "Customer",
	}

	for id, name := range expectedRoles {
		var role models.Role
		if err := DB.First(&role, id).Error; err != nil {
			DB.Create(&models.Role{ID: id, Name: name})
			fmt.Printf("🌱 Role %s (ID: %d) berhasil ditanam otomatis!\n", name, id)
		}
	}

	// SEEDING KATEGORI OTOMATIS
	var countCategory int64
	DB.Model(&models.Category{}).Count(&countCategory)
	if countCategory == 0 {
		categories := []models.Category{
			{Name: "Tas & Carrier", Slug: "tas-carrier"},
			{Name: "Tenda & Shelter", Slug: "tenda-shelter"},
			{Name: "Pakaian", Slug: "pakaian"},
			{Name: "Aksesoris", Slug: "aksesoris"},
		}
		DB.Create(&categories)
		fmt.Println("🌱 Data Kategori E-Commerce awal berhasil ditanam otomatis!")
	}

	// 📍 PANGGIL SEEDER WILAYAH INDONESIA DI SINI
	SeedWilayah()

	fmt.Println("🚀 KONEKSI DAN SELURUH RE-ALOKASI TABEL BERHASIL!")
}