package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reptil-adventure-api/config"
	"reptil-adventure-api/models"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// PingHandler (Fungsi Lama)
func PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Halo xlvi-digital! Respons dari AuthController berhasil dipanggil!",
		"status":  "success",
	})
}

// RegisterHandler (SUDAH DIAMANKAN) untuk mendaftarkan Customer Baru
func RegisterHandler(c *gin.Context) {
	// 1. Buat struktur input tanpa membawa field "role_id" dari frontend
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
	}

	// 2. Ikat (Bind) data JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. ENKRIPSI PASSWORD
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengamankan password"})
		return
	}

	// 4. Masukkan data ke struktur objek User GORM
	// 🔒 Kunci RoleID secara manual (Misal: 4 adalah ID untuk role Customer)
	// Jadi, siapapun yang register dari luar tidak akan bisa menyamar menjadi Admin (ID: 3)
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		RoleID:   4, // DIKUNCI HANYA UNTUK CUSTOMER
		// Phone & Address disesuaikan dengan field di struct models.User Anda jika ada:
		// Phone: input.Phone,
		// Address: input.Address,
	}

	// 5. Simpan permanen ke database PostgreSQL
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email sudah terdaftar!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Akun Reptil Adventure berhasil didaftarkan!",
		"status":  "success",
	})
}

// LoginHandler (MENGEMBALIKAN DATA USER UNTUK FRONTEND RECT)
func LoginHandler(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// 1. Ikat input JSON dari user
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Cari user di database berdasarkan email
	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau Password salah!"})
		return
	}

	// 3. COCOKKAN PASSWORD
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau Password salah!"})
		return
	}

	// 4. GENERATE TOKEN JWT
	token, err := config.GenerateToken(user.ID, user.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token akses"})
		return
	}

	// Tentukan penamaan role string berdasarkan RoleID untuk kebutuhan redirect di React
	var roleName string
	switch user.RoleID {
	case 1:
		roleName = "owner"
	case 2:
		roleName = "developer"
	case 3:
		roleName = "admin"
	default:
		roleName = "customer"
	}

	// 5. Kirim token & data profile user esensial ke frontend React Anda
	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil! Selamat datang kembali.",
		"token":   token,
		"status":  "success",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  roleName, // Mengirim teks gampang dibaca frontend ("admin"/"customer")
			"address_info": gin.H{
				"recipient_name": user.Name,
				// sesuaikan sub-objek ini dengan kolom detail alamat di model database Anda jika ada
				"detail":      "Alamat belum diatur", 
				"district":    "-",
				"city":        "-",
				"province":    "-",
				"postal_code": "-",
			},
		},
	})
}

func GetUserProfile(c *gin.Context) {
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	var actualUserID uint
	switch v := userIDFromContext.(type) {
	case uint:
		actualUserID = v
	case int:
		actualUserID = uint(v)
	case float64:
		actualUserID = uint(v)
	}

	var user models.User
	if err := config.DB.Preload("ShippingAddress").First(&user, actualUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// 🚀 PASTIKAN KEY "profile_picture" DI BAWAH INI ADA DAN SAMA DENGAN MODEL
	c.JSON(http.StatusOK, gin.H{
		"id":              user.ID,
		"name":            user.Name,
		"email":           user.Email,
		"phone":           user.Phone,
		"gender":          user.Gender,
		"birth_date":      user.BirthDate,
		"profile_picture": user.ProfilePicture, // 👈 INI WAJIB ADA!
		"shipping_address": gin.H{
			"province":       user.ShippingAddress.Province,
			"city":           user.ShippingAddress.City,
			"district":       user.ShippingAddress.District,
			"village":        user.ShippingAddress.Village,
			"postal_code":    user.ShippingAddress.PostalCode,
			"coordinates":    user.ShippingAddress.Coordinates,
			"map_reference":  user.ShippingAddress.MapReference,
			"manual_details": user.ShippingAddress.ManualDetails,
		},
	})
}


func UpdateUserProfile(c *gin.Context) {
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	var actualUserID uint
	switch v := userIDFromContext.(type) {
	case uint:
		actualUserID = v
	case int:
		actualUserID = uint(v)
	case float64:
		actualUserID = uint(v)
	}

	// 1. Ambil data user yang ada di DB saat ini beserta alamatnya
	var user models.User
	if err := config.DB.Preload("ShippingAddress").First(&user, actualUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// 2. Struct untuk membaca request dari React
	var input struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Gender       string `json:"gender"`
		BirthDate    string `json:"birth_date"`
		AddressInfo  struct {
			Province      string `json:"province"`
			City          string `json:"city"`
			District      string `json:"district"`
			Village       string `json:"village"`
			PostalCode    string `json:"postal_code"`
			Coordinates   string `json:"coordinates"`
			MapReference  string `json:"map_reference"`
			ManualDetails string `json:"manual_details"`
		} `json:"shipping_address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	// 3. Timpa data lama dengan data baru dari input Form
	user.Name = input.Name
	user.Phone = input.Phone
	user.Gender = input.Gender
	user.BirthDate = input.BirthDate

	// 4. Update data alamat (GORM otomatis mengelola One-to-One dengan Save)
	user.ShippingAddress.UserID = actualUserID
	user.ShippingAddress.Province = input.AddressInfo.Province
	user.ShippingAddress.City = input.AddressInfo.City
	user.ShippingAddress.District = input.AddressInfo.District
	user.ShippingAddress.Village = input.AddressInfo.Village
	user.ShippingAddress.PostalCode = input.AddressInfo.PostalCode
	user.ShippingAddress.Coordinates = input.AddressInfo.Coordinates
	user.ShippingAddress.MapReference = input.AddressInfo.MapReference
	user.ShippingAddress.ManualDetails = input.AddressInfo.ManualDetails

	// 5. Simpan seluruh perubahan objek User ke Database
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data di database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Informasi akun Anda berhasil diperbarui!",
	})
}
                                   
func GetCustomerOrders(c *gin.Context) {
	// 1. Ambil userID dari context JWT
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid atau kedaluwarsa"})
		return
	}

	// 2. Lakukan konversi tipe data (type assertion) dengan aman ke uint atau int
	var actualUserID uint
	switch v := userIDFromContext.(type) {
	case uint:
		actualUserID = v
	case int:
		actualUserID = uint(v)
	case float64:
		actualUserID = uint(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Format User ID tidak valid"})
		return
	}

	var orders []models.Order
	
	// 3. Jalankan query ke database menggunakan actualUserID yang tipenya sudah pasti
	// ⚠️ PENTING: Pastikan nama kolom di database Anda adalah "user_id". 
	// Jika di struct models.Order Anda menggunakan nama kolom lain (misal CustomerID), ganti menjadi "customer_id = ?"
	// Baris 266 di auth_controller.go
	if err := config.DB.Where("customer_id = ?", actualUserID).Order("id desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat riwayat transaksi: " + err.Error(),
		})
		return
	}

	// 4. Kembalikan array kosong jika belum ada transaksi, atau array berisi data jika ada
	if orders == nil {
		orders = []models.Order{}
	}

	c.JSON(http.StatusOK, orders)
}

func UploadAvatarHandler(c *gin.Context) {
	// 1. Ambil userID dari JWT context
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	var actualUserID uint
	switch v := userIDFromContext.(type) {
	case uint:
		actualUserID = v
	case int:
		actualUserID = uint(v)
	case float64:
		actualUserID = uint(v)
	}

	// 2. Ambil file dari form-data (key harus "avatar" sesuai frontend)
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan: " + err.Error()})
		return
	}

	// 3. Generate nama file unik agar tidak bentrok
	extension := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("avatar-%d-%d%s", actualUserID, time.Now().Unix(), extension)

	// 4. Simpan file fisik ke folder './uploads'
	// Folder 'uploads' harus dibuat terlebih dahulu di root project Anda
	dst := filepath.Join("uploads", newFilename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file di server"})
		return
	}

	// 5. Update kolom profile_picture di database untuk user tersebut
	if err := config.DB.Model(&models.User{}).Where("id = ?", actualUserID).Update("profile_picture", newFilename).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui database: " + err.Error()})
		return
	}

	// 6. Kembalikan response sukses beserta nama filenya
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Foto profil berhasil diperbarui!",
		"filename": newFilename,
	})
}

