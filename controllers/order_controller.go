package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/datatypes"

	"reptil-adventure-api/config"
	"reptil-adventure-api/models"
)

// Struct Input untuk request checkout/order baru dari pelanggan
type OrderInput struct {
	CustomerName   string          `json:"customer_name" binding:"required"`
	CustomerEmail  string          `json:"customer_email" binding:"required"`
	CustomerPhone  string          `json:"customer_phone" binding:"required"`
	ProvinceName   string          `json:"province_name" binding:"required"`
	CityName       string          `json:"city_name" binding:"required"`
	DistrictName   string          `json:"district_name" binding:"required"`
	VillageName    string          `json:"village_name" binding:"required"`
	PostalCode     string          `json:"postal_code" binding:"required"`
	DetailAddress  string          `json:"detail_address" binding:"required"`
	MapCoordinates string          `json:"map_coordinates"`
	RawMapAddress  string          `json:"raw_map_address"`
	Courier        string          `json:"courier" binding:"required"`
	ShippingCost   float64         `json:"shipping_cost"`
	CartItems      []CartItemInput `json:"cart_items" binding:"required"`
}

type CartItemInput struct {
	// 🚀 DIUBAH KE STRING: Menerima string SKU dari frontend (misal: "PROD-35930505")
	ProductID string `json:"product_id"` 
	Quantity  int    `json:"quantity"`
}

// Struct Input khusus pembaruan data pengiriman oleh Admin
type UpdateShippingInput struct {
	Courier       string `json:"courier" binding:"required"`
	ReceiptNumber string `json:"receipt_number" binding:"required"`
}

// Struct Input khusus pembaruan status transaksi cepat oleh Admin
type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// ==========================================
// 1. POST: /api/orders (Pembuatan Pesanan Baru + Midtrans Snap)
// ==========================================
func CreateOrder(c *gin.Context) {
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Sesi berakhir, silakan login kembali"})
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

	var input OrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var productTotal int64 = 0 
	var detailItemsForSnap []midtrans.ItemDetails

	// =================================================================
// SILAKAN GANTI BLOK LOOP PROSES VALIDASI PRODUK DI ORDER CONTROLLER:
// =================================================================

// 3. VALIDASI PRODUK & HITUNG TOTAL HARGA
for _, item := range input.CartItems {
    var product models.Product
    
    // 🚀 PERBAIKAN AKURAT: Cari berdasarkan kolom `product_id` (Varchar SKU), bukan Primary Key ID
    if err := config.DB.Where("product_id = ?", item.ProductID).First(&product).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "message": fmt.Sprintf("Produk dengan SKU '%s' tidak ditemukan di database", item.ProductID),
        })
        return
    }

    itemPriceInt := int64(product.Price)
    itemTotal := itemPriceInt * int64(item.Quantity)
    productTotal += itemTotal

    productName := product.Name
    if len(productName) > 45 {
        productName = productName[:42] + "..."
    }

    detailItemsForSnap = append(detailItemsForSnap, midtrans.ItemDetails{
        // Gunakan product.ProductID string ("PROD-XXXX") agar sinkron ke Midtrans
        ID:    product.ProductID, 
        Name:  productName, 
        Price: itemPriceInt,
        Qty:   int32(item.Quantity),
    })
}

	shippingCostInt := int64(input.ShippingCost)
	detailItemsForSnap = append(detailItemsForSnap, midtrans.ItemDetails{
		ID:    "ONGKIR-01",
		Name:  fmt.Sprintf("Ongkir %s", input.Courier), 
		Price: shippingCostInt,
		Qty:   1,
	})

	grandTotalInt := productTotal + shippingCostInt
	invoiceNumber := fmt.Sprintf("INV-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

	midtrans.ServerKey = "Mid-server-UCDHGr0iCmPc_gtM097jECLD"
	midtrans.Environment = midtrans.Sandbox

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  invoiceNumber,
			GrossAmt: grandTotalInt, 
		},
		Items: &detailItemsForSnap,
		CustomerDetail: &midtrans.CustomerDetails{
			FName: input.CustomerName,
			Email: input.CustomerEmail,
			Phone: input.CustomerPhone,
			BillAddr: &midtrans.CustomerAddress{
				FName:       input.CustomerName,
				Phone:       input.CustomerPhone,
				Address:     input.DetailAddress,
				City:        input.CityName,
				Postcode:    input.PostalCode,
				CountryCode: "IDN",
			},
			ShipAddr: &midtrans.CustomerAddress{
				FName:       input.CustomerName,
				Phone:       input.CustomerPhone,
				Address:     fmt.Sprintf("%s, Kel. %s, Kec. %s", input.DetailAddress, input.VillageName, input.DistrictName),
				City:        input.CityName,
				Postcode:    input.PostalCode,
				CountryCode: "IDN",
			},
		},
	}

	snapToken, errSnap := snap.CreateTransactionToken(snapReq)
	if errSnap != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal membuat token Midtrans: " + errSnap.Message})
		return
	}

	fullAddressString := fmt.Sprintf(
		"Penerima: %s\nNo. HP: %s\nAlamat: %s, Kel. %s, Kec. %s, %s, %s, %s\nKoordinat GPS: %s",
		input.CustomerName,
		input.CustomerPhone,
		input.DetailAddress,
		input.VillageName,
		input.DistrictName,
		input.CityName,
		input.ProvinceName,
		input.PostalCode,
		input.MapCoordinates,
	)

	cartBytes, _ := json.Marshal(input.CartItems)

	order := models.Order{
		OrderInvoice:    invoiceNumber,
		CustomerID:      actualUserID, 
		CustomerName:    input.CustomerName,
		CustomerEmail:   input.CustomerEmail,
		ShippingAddress: fullAddressString,
		Courier:         input.Courier,
		ShippingCost:    float64(shippingCostInt),
		ProductTotal:    float64(productTotal),
		GrandTotal:      float64(grandTotalInt),
		Status:          "PENDING", 
		TrackingNumber:  "-",       
		CartItems:       datatypes.JSON(cartBytes),
		SnapToken:       snapToken,
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan ke database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Order berhasil dibuat",
		"order_invoice": order.OrderInvoice,
		"snap_token":    snapToken,
	})
}

// ==========================================
// 2. GET: /api/admin/orders (Membaca Semua Riwayat Pesanan)
// ==========================================
func GetOrders(c *gin.Context) {
	var orders []models.Order
	statusFilter := c.Query("status")

	query := config.DB.Order("created_at DESC")
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil daftar pesanan: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// ==========================================
// 3. GET: /api/admin/orders/:invoice (Detail Satu Pesanan)
// ==========================================
func GetOrderByInvoice(c *gin.Context) {
	var order models.Order
	invoice := c.Param("invoice")

	if err := config.DB.Where("order_invoice = ?", invoice).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Nomor invoice pesanan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// ==========================================
// 4. PUT: /api/admin/orders/:invoice/status (Update Status Utama)
// ==========================================
func UpdateOrderStatus(c *gin.Context) {
	var order models.Order
	invoice := c.Param("invoice")

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := config.DB.Where("order_invoice = ?", invoice).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pesanan tidak ditemukan"})
		return
	}

	if err := config.DB.Model(&order).Update("status", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal memperbarui status transaksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status pesanan berhasil diubah menjadi " + input.Status})
}

// ==========================================
// 5. PUT: /api/admin/orders/:invoice/shipping (Update Resi & Kurir)
// ==========================================
func UpdateOrderShipping(c *gin.Context) {
	var order models.Order
	invoice := c.Param("invoice")

	var input UpdateShippingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := config.DB.Where("order_invoice = ?", invoice).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pesanan tidak ditemukan"})
		return
	}

	finalStatus := order.Status
	if input.ReceiptNumber != "-" && input.ReceiptNumber != "" {
		finalStatus = "SHIPPED"
	}

	updates := map[string]interface{}{
		"courier":         input.Courier,
		"tracking_number": input.ReceiptNumber, 
		"status":          finalStatus,
	}

	if err := config.DB.Model(&order).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan manifest logistik: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data manifest kurir dan nomor resi pengiriman berhasil diperbarui",
		"status":  finalStatus,
	})
}

// ==========================================
// 6. DELETE: /api/admin/orders/:invoice (Hapus / Batalkan)
// ==========================================
func DeleteOrder(c *gin.Context) {
	var order models.Order
	invoice := c.Param("invoice")

	if err := config.DB.Where("order_invoice = ?", invoice).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pesanan tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menghapus data pesanan dari sistem"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan dengan nomor invoice " + invoice + " berhasil dihapus"})
}

// ==========================================
// 7. GET: /api/v1/user/orders (Riwayat Pesanan User Login)
// ==========================================
func GetUserOrders(c *gin.Context) {
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Sesi tidak valid atau kedaluwarsa"})
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

	var orders []models.Order
	if err := config.DB.Where("customer_id = ?", actualUserID).Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil riwayat pesanan: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}