package model

import (
	"encoding/json"
	"fmt"
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
	Name           string  `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string  `gorm:"column:email;type:varchar(255);unique" validate:"required,email" json:"email"`
	PhoneNumber    string  `gorm:"column:phone_number;type:varchar(255);unique" validate:"number,min=1000000000,max=9999999999" json:"phone_number"`
	Picture        string  `gorm:"column:picture;type:text" json:"picture"`
	Blocked        bool    `gorm:"column:blocked;type:bool" json:"blocked"`
	HashedPassword string  `gorm:"column:hashed_password;type:varchar(255)" validate:"required" json:"hashed_password"`
	ReferralCode   string  `gorm:"column:referral_code" json:"referral_code"`
	WalletAmount   float64 `gorm:"column:wallet_amount;" json:"wallet_amount"`
}

type UserReferralHistory struct {
	UserID       uint   `gorm:"column:user_id" json:"user_id"`
	ReferralCode string `gorm:"column:referral_code" json:"referral_code"`
	ReferredBy   uint   `gorm:"column:referred_by" json:"referred_by"`
	ReferClaimed bool   `gorm:"column:refer_claimed" json:"refer_claimed"`
}

type Category struct {
	gorm.Model
	ID              uint   `gorm:"column:id" json:"id"`
	Name            string `validate:"required" json:"name"`
	Description     string `gorm:"column:description" validate:"required" json:"description"`
	ImageURL        string `gorm:"column:image_url" validate:"required" json:"image_url"`
	OfferPercentage uint   `gorm:"column:offer_percentage" json:"offer_percentage"`
}
type Product struct {
	gorm.Model
	ID            uint
	CategoryID    uint    `validate:"required" json:"category_id"`
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
	OrderID              uint            `json:"order_id" gorm:"autoCreateTime"`
	UserID               uint            `validate:"required,number" json:"user_id"`
	ShippingAddress      ShippingAddress `gorm:"embedded" json:"shippingAddress"`
	ItemCount            uint            `json:"item_count"`
	TotalAmount          float64         `validate:"required,number" json:"total_amount"`
	FinalAmount          float64         `validate:"required,number" json:"final_amount"`
	PaymentMethod        string          `validate:"required" json:"payment_method" gorm:"column:payment_method"`
	PaymentStatus        string          `validate:"required" json:"payment_status" gorm:"column:payment_status"`
	OrderedAt            time.Time       `gorm:"autoCreateTime" json:"ordered_at"`
	OrderStatus          string          `validate:"required" json:"order_status" gorm:"column:order_status"`
	CouponDiscountAmount float64         `validate:"required,number" json:"coupon_discount_amount"`
	ProductOfferAmount   float64         `validate:"required,number" json:"product_offer_amount"`
	DeliveryCharge       uint            `json:"delivery_charge"`
}

type OrderItem struct {
	OrderID     uint    `validate:"required" json:"order_id"`
	UserID      uint    `validate:"required,number" json:"user_id" `
	ProductID   uint    `validate:"required,number" json:"product_id"`
	Quantity    uint    `validate:"required,number" json:"quantity"`
	Amount      float64 `validate:"required,number" json:"amount"`
	OrderStatus string  `json:"order_status" gorm:"column:order_status"`
}

type ShippingAddress struct {
	PhoneNumber  uint   `gorm:"column:phone_number" validate:"number,min=1000000000,max=9999999999" json:"phone_number"`
	StreetName   string `gorm:"type:varchar(255)" json:"street_name"`
	StreetNumber string `gorm:"type:varchar(255)" json:"street_number"`
	City         string `gorm:"type:varchar(255)" json:"city"`
	State        string `gorm:"type:varchar(255)" json:"state"`
	PinCode      string `gorm:"type:varchar(20)" json:"pincode"`
}

type Rating struct {
	UserID    uint `gorm:"column:user_id"  validate:"required" json:"user_id"`
	ProductID uint `gorm:"column:product_id"  validate:"required,number" json:"product_id"`
	Rating    uint `gorm:"column:rating" validate:"number,min=1,max=5" json:"rating"`
}

type Payment struct {
	OrderID           string  `validate:"required" json:"order_id"`
	WalletPaymentID   string  `json:"wallet_payment_id" gorm:"column:wallet_payment_id"`
	RazorpayOrderID   string  `validate:"required" json:"razorpay_order_id" gorm:"column:razorpay_order_id"`
	RazorpayPaymentID string  `validate:"required" json:"razorpay_payment_id" gorm:"column:razorpay_payment_id"`
	RazorpaySignature string  `validate:"required" json:"razorpay_signature" gorm:"column:razorpay_signature"`
	PaymentGateway    string  `json:"payment_gateway" gorm:"payment_gateway"`
	PaymentStatus     string  `validate:"required" json:"payment_status" gorm:"column:payment_status"`
	Amount            float64 `validate:"required" json:"amount" gorm:"column:amount"`
}

type WishlistItems struct {
	UserID    uint `gorm:"column:user_id" validate:"required,number" json:"user_id"`
	ProductID uint `validate:"required,number" json:"product_id"`
}

type UserWalletHistory struct {
	TransactionTime time.Time `gorm:"autoCreateTime" json:"transaction_time"`
	WalletPaymentID string    `gorm:"column:wallet_payment_id" json:"wallet_payment_id"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	Type            string    `gorm:"column:type" json:"type"`
	OrderID         string    `gorm:"column:order_id" json:"order_id"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	CurrentBalance  float64   `gorm:"column:current_balance" json:"current_balance"`
	Reason          string    `gorm:"column:reason" json:"reason"`
}

type CouponInventory struct {
	CouponCode    string    `validate:"required" json:"coupon_code" gorm:"primary_key"`
	Expiry        time.Time `validate:"required" json:"expiry"`
	Percentage    uint      `validate:"required" json:"percentage"`
	MaximumUsage  uint      `validate:"required" json:"maximum_usage"`
	MinimumAmount float64   `validate:"required" json:"minimum_amount"`
	MaximumAmount float64   `validate:"required" json:"maximum_amount"`
}

func (c *CouponInventory) UnmarshalJSON(data []byte) error {

	type Alias CouponInventory
	aux := &struct {
		Expiry string `json:"expiry"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	layout := "02-01-2006"
	parsedTime, err := time.Parse(layout, aux.Expiry)
	if err != nil {
		return fmt.Errorf("invalid date format for expiry, use dd-mm-yyyy")
	}
	c.Expiry = parsedTime
	return nil
}

type CouponUsage struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	CouponCode string `json:"coupon_code"`
	UsageCount uint   `json:"usage_count"`
	OrderID    uint   `json:"order_id"`
}

type Request struct {
	RequestID uint   `json:"request_id" gorm:"autoCreateTime"`
	ProductID uint   `validate:"required,number" json:"product_id"`
	UserID    uint   `gorm:"column:user_id" validate:"required,number" json:"user_id"`
	Count     uint   `validate:"number" json:"count"`
	Response  string `json:"response"`
}
