package config

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 🚀 KITA KEMBALIKAN VARIABEL INI SECARA GLOBAL AGAR BISA DIBACA OLEH MIDDLEWARE
var JWT_KEY = []byte("kunci_rahasia_reptil_adventure_xlvi_99")

// GetJWTKey bertugas mengambil kunci rahasia yang dinamis
func GetJWTKey() []byte {
	// Jika ada variabel JWT_SECRET di environment (seperti saat live nanti), pakai itu
	if envKey := os.Getenv("JWT_SECRET"); envKey != "" {
		return []byte(envKey)
	}
	// Jika di lokal (kosong), pakai fallback bawaan di atas
	return JWT_KEY
}

// GenerateToken bertugas mencetak token JWT jika login sukses
func GenerateToken(userID uint, roleID uint) (string, error) {
	// Menentukan data apa saja yang mau dititipkan di dalam token (Claims)
	claims := jwt.MapClaims{
		"user_id": userID,
		"role_id": roleID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token hangus dalam 24 jam
	}

	// Proses pembuatan token dengan algoritma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// 🚀 Menandatangani token menggunakan fungsi kunci dinamis
	return token.SignedString(GetJWTKey())
}