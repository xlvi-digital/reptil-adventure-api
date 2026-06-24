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
	// 🚀 Membaca dari Environment Variable (Render), jika kosong gunakan default lokal Anda
	host := os.Getenv("DB_HOST")
	if host == "" { host = "localhost" }

	user := os.Getenv("DB_USER")
	if user == "" { user = "postgres" }

	password := os.Getenv("DB_PASSWORD")
	if password == "" { password = "Reptilbukanhewan26" }

	dbName := os.Getenv("DB_NAME")
	if dbName == "" { dbName = "reptil_adventure" }

	port := os.Getenv("DB_PORT")
	if port == "" { port = "5432" }

	// Render biasanya membutuhkan sslmode=require jika menggunakan cloud database eksternal.
	// Kita buat dinamis: jika di lokal pakai disable, di production disesuaikan.
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" { sslMode = "disable" }

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", host, user, password, dbName, port, sslMode)
	
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal("Gagal terhubung ke database PostgreSQL: ", err)
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

	// LOGIKA SEEDING ROLES (SUDAH DIPERBARUI)
	// LOGIKA SEEDING ROLES (VERSI CEK SATU PER SATU)
	expectedRoles := map[uint]string{
		1: "Developer",
		2: "Owner",
		3: "Admin",
		4: "Customer", // 🌟 Dipastikan terdeteksi di sini
	}

	for id, name := range expectedRoles {
		var role models.Role
		// Cek apakah ID tersebut sudah ada atau belum
		if err := DB.First(&role, id).Error; err != nil {
			// Jika belum ada (record not found), buat baru dengan ID terpasang tetap
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

