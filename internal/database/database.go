package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func ConnectToDB() {
	var err error

	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")
	dsn := "host=localhost user=" + user + " password=" + password + " dbname=gfresh port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("unable to connect to database, ")
	} else {
		fmt.Println("connection to database :OK")
	}
}
func AutoMigrate() {
	err := DB.AutoMigrate()
	if err != nil {
		log.Fatal("failed to automigrates models")
	}
}
