package main

import (
	"g-fresh/internal/api"
	"g-fresh/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {

	database.ConnectToDB()
	database.AutoMigrate()

	router := gin.Default()

	api.AuthenticationRoutes(router)
	api.AdminRoutes(router)

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
