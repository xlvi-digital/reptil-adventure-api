package controllers

import (
	"net/http"
	"reptil-adventure-api/config"
	"reptil-adventure-api/models"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/provinces
func GetProvinces(c *gin.Context) {
	var provinces []models.Province
	if err := config.DB.Order("name asc").Find(&provinces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, provinces)
}

// GET /api/v1/regencies?province_id=32
func GetRegencies(c *gin.Context) {
	provinceID := c.Query("province_id")
	if provinceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Parameter province_id wajib diisi"})
		return
	}

	var regencies []models.Regency
	if err := config.DB.Where("province_id = ?", provinceID).Order("name asc").Find(&regencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, regencies)
}

// GET /api/v1/districts?regency_id=3203
func GetDistricts(c *gin.Context) {
	regencyID := c.Query("regency_id")
	if regencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Parameter regency_id wajib diisi"})
		return
	}

	var districts []models.District
	if err := config.DB.Where("regency_id = ?", regencyID).Order("name asc").Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, districts)
}