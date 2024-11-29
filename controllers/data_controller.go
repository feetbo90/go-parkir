package controllers

import (
	"go_parkir/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDataHandler handles GET requests to fetch the latest serial data
func GetDataHandler(c *gin.Context) {
	data := services.GetLatestData()
	c.JSON(http.StatusOK, gin.H{"data": data})
}
