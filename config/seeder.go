package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reptil-adventure-api/models"
)

func SeedWilayah() {
	var count int64
	// Proteksi data ganda agar tidak duplikat saat server di-restart
	DB.Model(&models.Province{}).Count(&count)
	if count > 0 {
		fmt.Println("💡 Data wilayah Indonesia sudah ada di database. Seeding dilewati.")
		return
	}

	fmt.Println("⏳ [Seeder] Memulai migrasi data wilayah tingkat nasional...")

	// 1. IMPORT PROVINSI
	provFile, err := os.ReadFile("data_wilayah/provinsi.json")
	if err != nil {
		fmt.Println("❌ Gagal membaca file provinsi.json:", err)
		return
	}
	var provinces []models.Province
	json.Unmarshal(provFile, &provinces)
	if err := DB.Create(&provinces).Error; err != nil {
		fmt.Println("❌ Gagal menyimpan Provinsi:", err)
		return
	}
	fmt.Printf("✅ %d Provinsi berhasil ditanam.\n", len(provinces))

	// 2. IMPORT KABUPATEN (Suntik nama file sebagai province_id)
	regFiles, err := os.ReadDir("data_wilayah/kabupaten")
	if err != nil {
		fmt.Println("❌ Gagal membuka folder kabupaten:", err)
		return
	}
	for _, file := range regFiles {
		if filepath.Ext(file.Name()) == ".json" {
			// Mengambil ID Provinsi dari nama file (cth: "11.json" -> "11")
			provinceIDFromFile := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]

			content, _ := os.ReadFile(filepath.Join("data_wilayah/kabupaten", file.Name()))
			var regencies []models.Regency
			json.Unmarshal(content, &regencies)

			if len(regencies) > 0 {
				// Suntikkan isi field ProvinceID secara manual
				for i := range regencies {
					regencies[i].ProvinceID = provinceIDFromFile
				}
				DB.Create(&regencies)
			}
		}
	}
	fmt.Println("✅ Seluruh data Kabupaten sukses ditanam dengan relasi Provinsi.")

	// 3. IMPORT KECAMATAN (Suntik nama file sebagai regency_id)
	distFiles, err := os.ReadDir("data_wilayah/kecamatan")
	if err != nil {
		fmt.Println("❌ Gagal membuka folder kecamatan:", err)
		return
	}
	for _, file := range distFiles {
		if filepath.Ext(file.Name()) == ".json" {
			// Mengambil ID Kabupaten dari nama file (cth: "1101.json" -> "1101")
			regencyIDFromFile := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]

			content, _ := os.ReadFile(filepath.Join("data_wilayah/kecamatan", file.Name()))
			var districts []models.District
			json.Unmarshal(content, &districts)

			if len(districts) > 0 {
				// Suntikkan isi field RegencyID secara manual
				for i := range districts {
					districts[i].RegencyID = regencyIDFromFile
				}
				DB.Create(&districts)
			}
		}
	}
	fmt.Println("✅ Seluruh data Kecamatan sukses ditanam dengan relasi Kabupaten.")
	fmt.Println("🎉 [Seeder Sempurna] Pengisian database wilayah Indonesia sukses total 100%!")
}