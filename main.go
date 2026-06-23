package main

import (
	"os"
	"reptil-adventure-api/config"
	"reptil-adventure-api/routes"
)

func main() {
	// Pastikan folder uploads otomatis dibuat di root directory jika belum ada
	_ = os.MkdirAll("./uploads", os.ModePerm)

	// 1. Jalankan koneksi database & Auto Migration
	config.ConnectDatabase()

	// 2. Memanggil konfigurasi rute URL API dari folder routes
	r := routes.SetupRouter()
	
	// 3. Menyalakan server web di port 8080
	r.Run(":8080")
}