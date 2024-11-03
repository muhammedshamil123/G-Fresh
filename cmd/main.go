package main

import (
	"g-fresh/internal/api"
	"g-fresh/internal/database"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	database.ConnectToDB()
	database.AutoMigrate()

	router := gin.Default()

	api.AuthenticationRoutes(router)
	api.AdminRoutes(router)
	api.UserRoutes(router)
	log.Println("main client id:=", os.Getenv("ClientID"))
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}

// func init() {
// 	godotenv.Load()
// }
