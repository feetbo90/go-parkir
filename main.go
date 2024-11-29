package main

import (
	"go_parkir/routers"
)

func main() {
	// Initialize and start the router
	router := routers.SetupRouter()
	router.Run(":8081")
}
