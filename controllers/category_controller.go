package controllers

import (
	"net/http"
	"reptil-adventure-api/config"
	"reptil-adventure-api/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 1. READ: Mengambil semua data kategori (Sudah ada di kode Anda)
func GetAllCategoriesHandler(c *gin.Context) {
	var categories []models.Category
	
	if err := config.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kategori"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// 2. CREATE: Membuat kategori baru (Mengatasi eror 404 POST Anda)
func CreateCategoryHandler(c *gin.Context) {
	var input models.Category

	// Validasi data JSON yang dikirim dari React Frontend
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	// Validasi input kosong
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama kategori tidak boleh kosong"})
		return
	}

	// Simpan ke database PostgreSQL melalui GORM
	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kategori ke database"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// 3. UPDATE: Mengubah nama kategori berdasarkan ID
func UpdateCategoryHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID kategori tidak valid"})
		return
	}

	var category models.Category
	// Cari apakah kategori tersebut eksis di database
	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	var input models.Category
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	// Update kolom name saja ke database
	if err := config.DB.Model(&category).Update("name", input.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kategori"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// 4. DELETE: Menghapus kategori berdasarkan ID
func DeleteCategoryHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID kategori tidak valid"})
		return
	}

	var category models.Category
	// Pastikan kategori ada sebelum dihapus
	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	// Hapus data dari PostgreSQL
	if err := config.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kategori"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kategori berhasil dihapus"})
}