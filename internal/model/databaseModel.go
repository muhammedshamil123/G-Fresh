package model

import (
	"gorm.io/gorm"
)

type Admin struct {
	ID        uint `gorm:"primaryKey"`
	Username  string
	Password  string
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	gorm.Model
	Name           string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string `gorm:"column:email;type:varchar(255);unique" validate:"required,email" json:"email"`
	PhoneNumber    string `gorm:"column:phone_number;type:varchar(255);unique" validate:"required,e164" json:"phone_number"`
	Picture        string `gorm:"column:picture;type:text" json:"picture"`
	Blocked        bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255)" validate:"required" json:"hashed_password"`
}

type Category struct {
	gorm.Model
	ID          uint      `gorm:"column:id" json:"id"`
	Name        string    `validate:"required" json:"name"`
	Description string    `gorm:"column:description" validate:"required" json:"description"`
	ImageURL    string    `gorm:"column:image_url" validate:"required" json:"image_url"`
	Products    []Product `gorm:"foreignKey:CategoryID"`
}
type Product struct {
	gorm.Model
	ID            uint
	CategoryID    uint    `gorm:"foreignKey:CategoryID" validate:"required" json:"category_id"`
	Name          string  `validate:"required" json:"name"`
	Description   string  `gorm:"column:description" validate:"required" json:"description"`
	ImageURL      string  `gorm:"column:image_url" validate:"required" json:"image_url"`
	Price         float64 `validate:"required,number" json:"price"`
	OfferAmount   float64 `gorm:"column:offer_amount" json:"offer_amount"`
	StockLeft     uint    `validate:"required,number" json:"stock_left"`
	RatingSum     float64 `gorm:"column:rating_sum" json:"rating_sum"`
	RatingCount   uint    `gorm:"column:rating_count" json:"rating_count"`
	AverageRating float64 `gorm:"column:average_rating" json:"average_rating"`
}
