package middleware

import (
	"net/http"
	"reptil-adventure-api/config"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware bertugas mencegat request dan memeriksa validitas token JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil data Header "Authorization" dari request
		authHeader := c.GetHeader("Authorization")
		
		// 2. Cek apakah token dikirim dengan format "Bearer <token>"
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak! Token tidak ditemukan."})
			c.Abort() // Hentikan request di sini, jangan biarkan masuk ke controller
			return
		}

		// 3. Potong teks "Bearer " untuk mengambil string token murninya saja
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 4. Validasi dan bedah token menggunakan kunci rahasia kita
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return config.JWT_KEY, nil
		})

		// 5. Jika token rusak, kedaluwarsa, atau tidak sah
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau sudah kedaluwarsa!"})
			c.Abort()
			return
		}

		// 6. Jika lolos pemeriksaan, ambil data claims di dalam token
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			// Titipkan data user_id dan role_id ke dalam context Gin
			// agar bisa dibaca oleh fungsi controller selanjutnya
			c.Set("userID", claims["user_id"])
			c.Set("role_id", claims["role_id"])
		}

		// Lanjutkan perjalanan request ke controller utama
		c.Next()
	}
}