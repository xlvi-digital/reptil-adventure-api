package routes

import (
	"reptil-adventure-api/controllers"
	"reptil-adventure-api/middleware"

	"github.com/gin-gonic/gin"
)

// Fungsi pembantu CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Gunakan CORS
	r.Use(CORSMiddleware())
	r.Static("/uploads", "./uploads")

	v1 := r.Group("/api/v1")
	{
		// 🔓 RUTE PUBLIK (Bisa diakses siapa saja tanpa login)
		v1.GET("/ping", controllers.PingHandler)
		v1.POST("/auth/register", controllers.RegisterHandler)
		v1.POST("/auth/login", controllers.LoginHandler)
		
		// 🛒 Produk & Kategori untuk halaman toko Customer
		v1.GET("/products", controllers.GetAllProductsHandler) // Hanya GET yang publik
		v1.GET("/categories", controllers.GetAllCategoriesHandler)

		// 🆕 Endpoint untuk Transaksi Customer
		v1.GET("/provinces", controllers.GetProvinces)
		v1.GET("/regencies", controllers.GetRegencies)
		v1.GET("/districts", controllers.GetDistricts)

		v1.GET("/auth/google", controllers.GoogleLoginHandler)
		v1.GET("/auth/google/callback", controllers.GoogleCallbackHandler)
		// v1.GET("/orders/customer/:customer_id", controllers.GetCustomerOrdersHandler)

		// 🚀 DAFTARKAN WEBHOOK MIDTRANS DI SINI (Jalur Publik):
		v1.POST("/payments/notification", controllers.MidtransNotificationHandler)

		// 🔒 GRUP RUTE CUSTOMER TERKUNCI (Tambahkan Blok Ini!)
        // Rute-rute inilah yang dipanggil oleh halaman UserAccount.jsx di frontend Anda
        userProtected := v1.Group("/user")
        userProtected.Use(middleware.AuthMiddleware()) // Mengunci rute dengan pengecekan JWT Token
        {
            userProtected.GET("/profile", controllers.GetUserProfile)  // Mengatasi 404 GET /api/v1/user/profile
            userProtected.PUT("/profile", controllers.UpdateUserProfile) // Untuk simpan/update informasi profil & alamat
            userProtected.GET("/orders", controllers.GetCustomerOrders) // Mengatasi GET /api/v1/user/orders
			userProtected.POST("/upload-avatar", controllers.UploadAvatarHandler)
			// 🚀 PINDAHKAN KE SINI AGAR TERPROTEKSI JWT & BISA BACA userID DENGAN AMAN!
    		userProtected.POST("/orders", controllers.CreateOrder)
        }

		// 🔒 GRUP RUTE TERKUNCI (Hanya Admin yang bisa mengelola data)
		adminProtected := v1.Group("/admin")
		adminProtected.Use(middleware.AuthMiddleware())
		{
			adminProtected.GET("/dashboard", func(c *gin.Context) {
				roleID := c.MustGet("role_id")
				userID := c.MustGet("userID")

				c.JSON(205, gin.H{
					"message": "Selamat! Anda berhasil menembus gerbang keamanan JWT Middleware!",
					"user_id": userID,
					"role_id": roleID,
					"status":  "authorized",
				})
			})

			// 🚀 KELOLA PRODUK (Hanya Admin)
			adminProtected.POST("/products", controllers.CreateProductHandler)
			adminProtected.PUT("/products/:id", controllers.UpdateProductHandler)
			adminProtected.DELETE("/products/:id", controllers.DeleteProductHandler)

			// 🚀 TAMBAHKAN BARIS INI:
			adminProtected.POST("/products/import", controllers.ImportProductsFromExcel)
			
			// 🚀 KELOLA KATEGORI (Hanya Admin)
			adminProtected.POST("/categories", controllers.CreateCategoryHandler) 
			adminProtected.PUT("/categories/:id", controllers.UpdateCategoryHandler)
			adminProtected.DELETE("/categories/:id", controllers.DeleteCategoryHandler)

			
			// 📦 🚀 🆕 KELOLA PESANAN / ORDERS (Hanya Admin)
			adminProtected.GET("/orders", controllers.GetOrders)               // Get list & filter pesanan
			adminProtected.GET("/orders/:invoice", controllers.GetOrderByInvoice) // Detail satu pesanan
			adminProtected.PUT("/orders/:invoice/status", controllers.UpdateOrderStatus) // Ubah status bayar cepat
			adminProtected.PUT("/orders/:invoice/shipping", controllers.UpdateOrderShipping) // Input kurir & nomor resi
			adminProtected.DELETE("/orders/:invoice", controllers.DeleteOrder)  // Hapus / batalkan pesanan
		}
	}

	return r
}