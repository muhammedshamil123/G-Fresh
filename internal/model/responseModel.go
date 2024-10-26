package model

import "time"

type UserResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	Picture      string `json:"picture"`
	Blocked      bool   `json:"blocked"`
	ReferralCode string `json:"referral_code"`
}
type CategoryResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ImageURL        string `json:"image_url"`
	OfferPercentage uint   `json:"offer_percentage"`
}

type ViewCategoryList struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	ImageURL        string `json:"image_url"`
	OfferPercentage uint   `json:"offer_percentage"`
}

type ProductResponse struct {
	ID            uint
	CategoryID    uint    `json:"category_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ImageURL      string  `json:"image_url"`
	Price         float64 `json:"price"`
	OfferAmount   float64 `json:"offer_amount"`
	StockLeft     uint    `json:"stock_left"`
	RatingCount   uint    `json:"rating_count"`
	AverageRating float64 `json:"average_rating"`
}

type ViewProductList struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ImageURL      string  `json:"image_url"`
	Price         float64 `json:"price"`
	OfferAmount   float64 `json:"offer_amount"`
	StockLeft     uint    `json:"stock_left"`
	RatingCount   uint    `json:"rating_count"`
	AverageRating float64 `json:"average_rating"`
	CategoryName  string  `json:"category_name"`
}

type ViewCartList struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	OfferAmount float64 `json:"offer_amount"`
	StockLeft   uint    `json:"stock_left"`
}

type GoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type OrderResponce struct {
	OrderID              uint      `json:"order_id"`
	ItemCount            uint      `json:"item_count"`
	TotalAmount          float64   `json:"final_amount"`
	FinalAmount          float64   `json:"total_amount"`
	PaymentMethod        string    `json:"payment_method"`
	PaymentStatus        string    `json:"payment_status"`
	OrderedAt            time.Time `json:"ordered_at"`
	OrderStatus          string    `json:"order_status"`
	CouponDiscountAmount float64   `json:"coupon_discount_amount"`
	ProductOfferAmount   float64   `json:"product_offer_amount"`
	DeliveryCharge       uint      `json:"delivery_charge"`
}

type OrderItemResponse struct {
	ProductID   uint    `json:"product_id"`
	Quantity    uint    `json:"quantity"`
	Amount      float64 `json:"amount"`
	OrderStatus string  `json:"order_status"`
}

type ShippingAddressResponse struct {
	PhoneNumber  uint   `json:"phone_number"`
	StreetName   string `json:"street_name"`
	StreetNumber string `json:"street_number"`
	City         string `json:"city"`
	State        string `json:"state"`
	PinCode      string `json:"pincode"`
}

type ViewOrderProductList struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	OfferAmount float64 `json:"offer_amount"`
}

type OrderCount struct {
	TotalOrder          uint `json:"total_order"`
	TotalPLACED         uint `json:"total_placed"`
	TotalCONFIRMED      uint `json:"total_confirmed"`
	TotalSHIPPED        uint `json:"total_shipped"`
	TotalOUTFORDELIVERY uint `json:"total_out_for_delivery"`
	TotalDELIVERED      uint `json:"total_delivered"`
	TotalCANCELED       uint `json:"total_cancelled"`
	TotalRETURNREQUEST  uint `json:"total_return_request"`
	TotalRETURNED       uint `json:"total_returned"`
}

type AmountInformation struct {
	TotalCouponDeduction       float64 `json:"total_coupon_deduction"`
	TotalProductOfferDeduction float64 `json:"total_product_offer_deduction"`
	TotalAmountBeforeDeduction float64 `json:"total_amount_before_deduction"`
	TotalAmountAfterDeduction  float64 `json:"total_amount_after_deduction"`
	TotalSalesRevenue          float64 `json:"total_sales_revenue"`
	TotalRefundAmount          float64 `json:"total_refund_amount"`
	AverageOrderValue          float64 `json:"average_order_value"`
	TotalProductSold           uint    `json:"total_products_sold"`
	TotalProductReturned       uint    `json:"total_products_returned"`
	TotalCustomers             uint    `json:"total_customers"`
}
type GeoCodeResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
		Components struct {
			Postcode string `json:"postcode"`
			Country  string `json:"country"`
			State    string `json:"state"`
			City     string `json:"city"`
		} `json:"components"`
	} `json:"results"`
}
