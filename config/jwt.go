package config

import (
	"os" // 🚀 Tambahkan import "os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken bertugas mencetak token JWT jika login sukses
func GenerateToken(userID uint, roleID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role_id": roleID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// 🚀 Ambil kunci rahasia dari environment, jika tidak ada pakai fallback lokal
	jwtKeyString := os.Getenv("JWT_SECRET")
	if jwtKeyString == "" {
		jwtKeyString = "kunci_rahasia_reptil_adventure_xlvi_99"
	}
	var jwtKeyBytes = []byte(jwtKeyString)

	return token.SignedString(jwtKeyBytes)
}