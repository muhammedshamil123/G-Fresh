package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	// fmt.Println(os.Getenv("PASSWORD"), "JI")
	// user := os.Getenv("USER")
	// password := os.Getenv("PASSWORD")
	dsn := "host=localhost user=postgres  password=6930 dbname=gfresh port=5432 sslmode=disable TimeZone=Asia/Shanghai"

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
