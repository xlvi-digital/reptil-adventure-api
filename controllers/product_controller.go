package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reptil-adventure-api/config" // 👈 Ganti dengan package tempat variabel DB global Anda berada
	"reptil-adventure-api/models" // 👈 Sesuaikan dengan nama module/folder models Anda
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"gorm.io/datatypes"
)

// Struct untuk validasi data Form saat Create/Update
type ProductInput struct {
	Name        string  `form:"name" binding:"required"`
	Price       float64 `form:"price" binding:"required"`
	Stock       float64 `form:"stock" binding:"required"`
	CategoryID  uint    `form:"category_id" binding:"required"`
	Description string  `form:"description"`
	Colors      string  `form:"colors"`
	Sizes       string  `form:"sizes"`
}

// Struct skema JSON untuk kolom Image di database
type ImageJsonSchema struct {
	Primary string   `json:"primary"`
	Support []string `json:"support"`
}

// ==========================================
// 1. CREATE PRODUCT (POST /api/v1/products)
// ==========================================
func CreateProductHandler(c *gin.Context) {
	var input ProductInput

	// Bind data text dan numeric dari FormData
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validasi gagal: " + err.Error()})
		return
	}

	// Pastikan folder uploads tersedia
	_ = os.MkdirAll("./uploads", os.ModePerm)

	// --- Proses Upload Gambar Utama ---
	var primaryImagePath string
	primaryFile, err := c.FormFile("image")
	if err == nil {
		ext := filepath.Ext(primaryFile.Filename)
		filename := uuid.New().String() + ext
		primaryImagePath = "/uploads/" + filename
		if err := c.SaveUploadedFile(primaryFile, "."+primaryImagePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar utama"})
			return
		}
	}

	// --- Proses Upload Gambar Pendukung (Gallery) ---
	var supportImages []string
	form, err := c.MultipartForm()
	if err == nil {
		galleryFiles := form.File["gallery"]
		for _, file := range galleryFiles {
			ext := filepath.Ext(file.Filename)
			filename := uuid.New().String() + ext
			path := "/uploads/" + filename
			if err := c.SaveUploadedFile(file, "."+path); err == nil {
				supportImages = append(supportImages, path)
			}
		}
	}

	// Marshalling skema gambar ke format JSON string/bytes
	imageSchema := ImageJsonSchema{
		Primary: primaryImagePath,
		Support: supportImages,
	}
	imageBytes, _ := json.Marshal(imageSchema)

	// Buat objek model GORM baru
	product := models.Product{
		ProductID:   "PROD-" + strings.ToUpper(uuid.New().String()[:8]), // Generate custom string ID
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		CategoryID:  input.CategoryID,
		Image:       datatypes.JSON(imageBytes),
		Colors:      datatypes.JSON([]byte(input.Colors)),
		Sizes:       datatypes.JSON([]byte(input.Sizes)),
	}

	// Simpan ke database (Ganti config.DB sesuai DB global Anda)
	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan ke database: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Produk baru berhasil ditambahkan!", "data": product})
}

// ==========================================
// 2. UPDATE PRODUCT (PUT /api/v1/products/:id)
// ==========================================
func UpdateProductHandler(c *gin.Context) {
	productID := c.Param("id") // Mengambil parameter string ID dari URL (Contoh: 'PROD-XXXX')
	var product models.Product

	// Cari produk berdasarkan product_id string di DB
	if err := config.DB.Where("product_id = ?", productID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var input ProductInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validasi gagal: " + err.Error()})
		return
	}

	// Ambil data gambar lama dari DB terlebih dahulu
	var currentImage ImageJsonSchema
	if len(product.Image) > 0 {
		_ = json.Unmarshal(product.Image, &currentImage)
	}

	// --- Perbarui Gambar Utama jika ada file baru diupload ---
	primaryFile, err := c.FormFile("image")
	if err == nil {
		ext := filepath.Ext(primaryFile.Filename)
		filename := uuid.New().String() + ext
		newPrimaryPath := "/uploads/" + filename
		if err := c.SaveUploadedFile(primaryFile, "."+newPrimaryPath); err == nil {
			// Hapus file fisik gambar lama jika ada
			if currentImage.Primary != "" {
				_ = os.Remove("." + currentImage.Primary)
			}
			currentImage.Primary = newPrimaryPath
		}
	}

	// --- Perbarui Gambar Galeri jika ada file baru diupload ---
	form, err := c.MultipartForm()
	if err == nil {
		galleryFiles := form.File["gallery"]
		if len(galleryFiles) > 0 {
			// Hapus file fisik galeri lama sebelum ditimpa yang baru
			for _, oldPath := range currentImage.Support {
				_ = os.Remove("." + oldPath)
			}
			
			// Reset isi slice dan masukkan file baru
			currentImage.Support = []string{}
			for _, file := range galleryFiles {
				ext := filepath.Ext(file.Filename)
				filename := uuid.New().String() + ext
				path := "/uploads/" + filename
				if err := c.SaveUploadedFile(file, "."+path); err == nil {
					currentImage.Support = append(currentImage.Support, path)
				}
			}
		}
	}

	imageBytes, _ := json.Marshal(currentImage)

	// Update data field model objek
	product.Name = input.Name
	product.Description = input.Description
	product.Price = input.Price
	product.Stock = input.Stock
	product.CategoryID = input.CategoryID
	product.Image = datatypes.JSON(imageBytes)
	product.Colors = datatypes.JSON([]byte(input.Colors))
	product.Sizes = datatypes.JSON([]byte(input.Sizes))

	// Simpan perubahan ke database
	if err := config.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil diperbarui!", "data": product})
}

// ==========================================
// 3. DELETE PRODUCT (DELETE /api/v1/products/:id)
// ==========================================
func DeleteProductHandler(c *gin.Context) {
	productID := c.Param("id")
	var product models.Product

	// Cari data produk di DB terlebih dahulu
	if err := config.DB.Where("product_id = ?", productID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	// --- Opsi Tambahan: Hapus File Gambar Fisik dari Server ---
	var currentImage ImageJsonSchema
	if len(product.Image) > 0 {
		if err := json.Unmarshal(product.Image, &currentImage); err == nil {
			// Hapus gambar utama
			if currentImage.Primary != "" {
				_ = os.Remove("." + currentImage.Primary)
			}
			// Hapus semua gambar galeri pendukung
			for _, path := range currentImage.Support {
				_ = os.Remove("." + path)
			}
		}
	}

	// Hapus baris data dari database
	if err := config.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk dari database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk beserta aset gambarnya berhasil dihapus!"})
}

// ==========================================
// 4. GET ALL PRODUCTS (Tambahan untuk kelengkapan)
// ==========================================
func GetAllProductsHandler(c *gin.Context) {
	var products []models.Product
	if err := config.DB.Preload("Category").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// Helper untuk mendownload gambar dari URL internet dan menyimpannya ke local server
func downloadImageFromURL(url string) (string, error) {
	if url == "" {
		return "", nil
	}

	// 1. Lakukan HTTP GET Request ke URL gambar
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Pastikan response statusnya OK (200) dan tipenya adalah gambar
	if resp.StatusCode != http.StatusOK {
		return "", http.ErrMissingFile
	}

	// 2. Tentukan ekstensi file berdasarkan Content-Type (default .jpg jika tidak terbaca)
	ext := ".jpg"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "image/png") {
		ext = ".png"
	} else if strings.Contains(contentType, "image/jpeg") {
		ext = ".jpg"
	} else if strings.Contains(contentType, "image/webp") {
		ext = ".webp"
	}

	// 3. Generate nama file unik menggunakan UUID
	filename := uuid.New().String() + ext
	localPath := "/uploads/" + filename

	// 4. Buat file kosong di folder lokal server
	out, err := os.Create("." + localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 5. Salin data gambar dari internet ke file lokal
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

func ImportProductsFromExcel(c *gin.Context) {
	// 1. Ambil file dari request multipart form
	fileHeader, err := c.FormFile("excel_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File Excel wajib diunggah!"})
		return
	}

	// 2. Buka file stream tanpa menyimpan file ke disk laptop/server
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file Excel"})
		return
	}
	defer file.Close()

	// 3. Baca file menggunakan excelize
	f, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak valid, pastikan formatnya .xlsx"})
		return
	}
	defer f.Close()

	// 4. Ambil nama sheet pertama (biasanya "Sheet1")
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sheet kosong atau tidak mengandung data"})
		return
	}

	var productsToInsert []models.Product
	var successCount int

	// 5. Looping baris data (Mulai dari indeks 1, karena indeks 0 adalah Header)
	// 5. Looping baris data (Mulai dari indeks 1, karena indeks 0 adalah Header)
	for i, row := range rows {
		if i == 0 {
			continue // Lewati baris header judul
		}

		// 🚀 FIX: Jika baris kosong atau kolom kurang dari 5, langsung lewati
		if len(row) < 5 || row[0] == "" {
			continue 
		}

		name := row[0]
		description := row[1]
		
		// 🚀 FIX: Validasi Harga (jika kosong atau bukan angka, set ke 0)
		priceRaw, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			priceRaw = 0
		}
		
		// 🚀 FIX: Validasi Stok
		stockRaw, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			stockRaw = 0
		}
		
		// 🚀 FIX: Validasi Category ID. Jika kosong/gagal parse, skip baris ini 
		// demi mencegah Foreign Key Error di PostgreSQL
		categoryIDRaw, err := strconv.ParseUint(row[4], 10, 32)
		if err != nil || categoryIDRaw == 0 {
			continue // Lewati baris ini karena ID Kategori wajib valid
		}

		// Ekstraksi data varian warna & ukuran
		var colorsArray []string
		if len(row) > 5 && row[5] != "" {
			for _, v := range strings.Split(row[5], ",") {
				colorsArray = append(colorsArray, strings.TrimSpace(v))
			}
		}

		var sizesArray []string
		if len(row) > 6 && row[6] != "" {
			for _, v := range strings.Split(row[6], ",") {
				sizesArray = append(sizesArray, strings.TrimSpace(v))
			}
		}


		// 🚀 BARU: Ekstraksi URL Gambar Utama dari Kolom H (Indeks 7)
		var primaryImagePath string
		if len(row) > 7 && row[7] != "" {
			// Panggil fungsi helper untuk download gambar otomatis
			downloadedPath, err := downloadImageFromURL(strings.TrimSpace(row[7]))
			if err == nil {
				primaryImagePath = downloadedPath
			}
			// Jika gagal download, primaryImagePath akan tetap kosong "" secara aman
		}

		// Marshalling skema gambar ke format JSON string/bytes
		imageSchema := ImageJsonSchema{
			Primary: primaryImagePath,       // 🚀 Berhasil terisi path lokal otomatis jika URL valid
			Support: []string{},             // Kosongkan gallery bawaan untuk import excel
		}
		
		colorsJSON, _ := json.Marshal(colorsArray)
		sizesJSON, _ := json.Marshal(sizesArray)
		imageJSON, _ := json.Marshal(imageSchema) // 🚀 Gunakan schema baru yang dinamis

		// 6. Bungkus ke dalam objek struct Product
		product := models.Product{
			ProductID:   "PROD-" + strings.ToUpper(uuid.New().String()[:8]),
			Name:        name,
			Description: description,
			Price:       priceRaw,
			Stock:       stockRaw,
			CategoryID:  uint(categoryIDRaw),
			Colors:      datatypes.JSON(colorsJSON), 
			Sizes:       datatypes.JSON(sizesJSON),  
			Image:       datatypes.JSON(imageJSON), // 🚀 Update bagian ini
		}

		productsToInsert = append(productsToInsert, product)
		successCount++
	}

// 7. Bulk Insert ke database PostgreSQL lewat GORM
	// 7. FIX: Simpan menggunakan Transaction + Loop Individual Create
	// Cara ini membuat GORM bisa memicu Auto-Increment uint secara normal di PostgreSQL
	if len(productsToInsert) > 0 {
		// Mulai transaksi database
		tx := config.DB.Begin()
		
		for _, prod := range productsToInsert {
			// Insert satu per satu di dalam jalur transaksi yang sama
			if err := tx.Create(&prod).Error; err != nil {
				tx.Rollback() // Batalkan semua data jika ada satu saja yang gagal
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal menyimpan produk [" + prod.Name + "]: " + err.Error(),
				})
				return
			}
		}
		
		// Jika semua baris sukses tanpa error, komit/simpan permanen ke database
		tx.Commit()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Proses import berhasil!",
		"total":   successCount,
	})
}
