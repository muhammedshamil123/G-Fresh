package model

type UserResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
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
