package controllers

import (
	"log"
	"net/http"
	"reptil-adventure-api/config"
	"reptil-adventure-api/models"

	// Kita siapkan jika perlu konversi string ke int
	"github.com/gin-gonic/gin"
)

func MidtransNotificationHandler(c *gin.Context) {
	var notificationData map[string]interface{}

	// 1. Ambil data JSON dari Midtrans
	if err := c.ShouldBindJSON(&notificationData); err != nil {
		log.Println("❌ Gagal membaca payload Midtrans:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// 2. Ekstrak data
	orderIDStr, _ := notificationData["order_id"].(string)
	transactionStatus, _ := notificationData["transaction_status"].(string)
	fraudStatus, _ := notificationData["fraud_status"].(string)

	log.Printf("📥 Notifikasi Masuk - OrderID: %s, Status: %s\n", orderIDStr, transactionStatus)

	var newStatus string
	if transactionStatus == "settlement" || transactionStatus == "capture" {
		if transactionStatus == "capture" && fraudStatus != "accept" {
			newStatus = "fraud"
		} else {
			newStatus = "success"
		}
	} else if transactionStatus == "pending" {
		newStatus = "pending"
	} else if transactionStatus == "cancel" || transactionStatus == "expire" {
		newStatus = "expired"
	}

	// 4. Proses Update ke Database dengan Proteksi Tipe Data
// 4. Proses Update ke Database berdasarkan Nomor Invoice
	if newStatus != "" {
		// Menggunakan target kolom "order_invoice" sesuai struktur model Go Anda
		queryErr := config.DB.Model(&models.Order{}).
			Where("order_invoice = ?", orderIDStr). 
			Update("status", newStatus).Error

		if queryErr != nil {
			log.Println("❌ EROR DATABASE GORM:", queryErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui database: " + queryErr.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Status diperbarui menjadi " + newStatus,
	})
}