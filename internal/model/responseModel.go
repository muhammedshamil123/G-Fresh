package model

type UserResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber uint   `json:"phone_number"`
	Picture     string `json:"picture"`
	Blocked     bool   `json:"blocked"`
}
type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type ViewCategoryList struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
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
