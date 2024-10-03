package model

import (
	"time"

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
	PhoneNumber    string `gorm:"column:phone_number;type:varchar(255);unique" validate:"number,min=1000000000,max=9999999999" json:"phone_number"`
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

type Address struct {
	UserID       uint   `json:"user_id" gorm:"column:user_id"`
	AddressID    uint   `gorm:"primaryKey;autoIncrement;column:address_id" json:"address_id"`
	PhoneNumber  uint   `gorm:"column:phone_number" validate:"number,min=1000000000,max=9999999999" json:"phone_number"`
	StreetName   string `validate:"required" json:"street_name" gorm:"column:street_name"`
	StreetNumber string `validate:"required" json:"street_number" gorm:"column:street_number"`
	City         string `validate:"required" json:"city" gorm:"column:city"`
	State        string `validate:"required" json:"state" gorm:"column:state"`
	PostalCode   string `validate:"required" json:"postal_code" gorm:"column:postal_code"`
}

type CartItems struct {
	UserID    uint `gorm:"column:user_id" validate:"required,number" json:"user_id"`
	ProductID uint `validate:"required,number" json:"product_id"`
	Quantity  uint ` validate:"required,number" json:"quantity"`
}

type Order struct {
	OrderID       string    `validate:"required" json:"order_id"`
	UserID        uint      `validate:"required,number" json:"user_id"`
	AddressID     uint      `validate:"required,number" json:"address_id"`
	ItemCount     uint      `json:"item_count"`
	TotalAmount   float64   `validate:"required,number" json:"total_amount"`
	PaymentMethod string    `validate:"required" json:"payment_method" gorm:"column:payment_method"`
	PaymentStatus string    `validate:"required" json:"payment_status" gorm:"column:payment_status"`
	OrderedAt     time.Time `gorm:"autoCreateTime" json:"ordered_at"`
}

type OrderItem struct {
	OrderID            string  `validate:"required" csv:"OrderID" json:"order_id"`
	UserID             uint    `validate:"required,number" json:"user_id" csv:"UserID"`
	ProductID          uint    `validate:"required,number" json:"product_id" csv:"ProductID"`
	Quantity           uint    `validate:"required,number" json:"quantity" csv:"Quantity"`
	Amount             float64 `validate:"required,number" json:"amount" csv:"Amount"`
	ProductOfferAmount float64 `json:"product_offer_amount" csv:"ProductOfferAmount"`
	OrderStatus        string  `json:"order_status" gorm:"column:order_status" csv:"OrderStatus"`
}
