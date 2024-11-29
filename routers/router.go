package routers

import (
	"go_parkir/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin router and sets up the routes
func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/data", controllers.GetDataHandler)
	return router
}
