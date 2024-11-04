package database

import (
	"fmt"
	"g-fresh/internal/model"
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
	host := os.Getenv("HOST")
	dbname := os.Getenv("DBNAME")
	port := os.Getenv("PORT")
	sslmode := os.Getenv("SSL")
	timezone := os.Getenv("ZONE")
	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=" + sslmode + " TimeZone=" + timezone

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("unable to connect to database, ")
	} else {
		fmt.Println("connection to database :OK")
	}
}
func AutoMigrate() {
	err := DB.AutoMigrate(
		&model.Admin{},
		&model.User{},
		&model.UserReferralHistory{},
		&model.Category{},
		&model.Product{},
		&model.Address{},
		&model.CartItems{},
		&model.Order{},
		&model.OrderItem{},
		&model.ShippingAddress{},
		&model.Rating{},
		&model.Payment{},
		&model.WishlistItems{},
		&model.UserWalletHistory{},
		&model.CouponInventory{},
		&model.CouponUsage{},
		&model.Request{},
	)
	if err != nil {
		log.Fatal("failed to automigrates models")
	}
}
