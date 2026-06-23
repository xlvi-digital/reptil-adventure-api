package config

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Kata kunci rahasia untuk mengunci token JWT (Jangan sampai bocor di produksi!)
var JWT_KEY = []byte("kunci_rahasia_reptil_adventure_xlvi_99")

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
	
	// Menandatangani token menggunakan kata kunci rahasia kita
	return token.SignedString(JWT_KEY)
}