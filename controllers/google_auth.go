package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"reptil-adventure-api/config"
	"reptil-adventure-api/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Konfigurasi Google OAuth menggunakan data Kredensial kamu
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/api/v1/auth/google/callback",
	ClientID:     "476178384350-cllgitnqfs4ilg84t1cjpl76qcd0hs5m.apps.googleusercontent.com",
	ClientSecret: "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",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// 1. Handler untuk melempar user ke Google Sign-In Page
func GoogleLoginHandler(c *gin.Context) {
	// Generate URL resmi login Google
	url := googleOauthConfig.AuthCodeURL("random_state_string")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// 2. Handler Callback setelah Google sukses memverifikasi user
func GoogleCallbackHandler(c *gin.Context) {
	state := c.Query("state")
	if state != "random_state_string" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State token tidak cocok!"})
		return
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menukar code dengan token"})
		return
	}

	// Ambil data profil dari Google API menggunakan token akses
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data user dari Google"})
		return
	}
	defer response.Body.Close()

	var googleUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca profile data Google"})
		return
	}

	// 🔒 PROSES UPSERT: Cari email, jika belum terdaftar, buat akun baru otomatis
	var user models.User
	err = config.DB.Where("email = ?", googleUser.Email).First(&user).Error

	if err != nil {
		// Jika email belum ada di DB, daftarkan sebagai Customer baru (RoleID: 4)
		user = models.User{
			Name:     googleUser.Name,
			Email:    googleUser.Email,
			Password: "", // Kosong karena autentikasi dikelola Google
			RoleID:   4,  // 🔒 Aman, otomatis terkunci sebagai Customer!
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftarkan user baru"})
			return
		}
	}

	// Ambil JWT Token internal dari fungsi bawaan sistem kamu
	jwtToken, err := config.GenerateToken(user.ID, user.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat session token"})
		return
	}

	// 🚀 REDIRECT BALIK KE FRONTEND REACT (Port 3000) dengan membawa token di URL
	c.Redirect(http.StatusSeeOther, "http://localhost:3000/login?token="+jwtToken)
}