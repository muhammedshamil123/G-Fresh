package model

type AdminLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
type UserEmailLoginRequest struct {
	Email    string `validate:"required,email" json:"email"`
	Password string `json:"password" validate:"required"`
}
type UserEmailSignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	PhoneNumber     string `validate:"required" json:"phonenumber"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}
type AddCategoryList struct {
	Name            string `json:"name"`
	Description     string `validate:"required" json:"description"`
	ImageURL        string `validate:"required" json:"imageurl"`
	OfferPercentage uint   `validate:"required" json:"offerpercentage"`
}
type AddProductsRequest struct {
	CategoryID  uint    `validate:"required" json:"categoryid"`
	Name        string  `validate:"required" json:"name"`
	Description string  `validate:"required" json:"description"`
	ImageURL    string  `gvalidate:"required" json:"imageurl"`
	Price       float64 `validate:"required,number,min=1" json:"price"`
	OfferAmount float64 `json:"offeramount"`
	StockLeft   uint    `validate:"required,number,min=1" json:"stockleft"`
}

type AddAddressRequest struct {
	PhoneNumber  uint   `validate:"number,min=1000000000,max=9999999999" json:"phonenumber"`
	StreetName   string `validate:"required" json:"streetname"`
	StreetNumber string `validate:"required" json:"streetnumber"`
	City         string `validate:"required" json:"city"`
	State        string `validate:"required" json:"state"`
	PostalCode   string `validate:"required" json:"postalcode"`
}
type ProfileEdit struct {
	Name        string `validate:"required" json:"name"`
	Email       string `validate:"required,email" json:"email"`
	PhoneNumber string `validate:"required" json:"phonenumber"`
	Picture     string `validate:"required" json:"picture"`
}

type ChangePasswordRequest struct {
	OldPassword     string `validate:"required" json:"oldpassword"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type InitiatePayment struct {
	OrderID        string `json:"order_id"`
	PaymentGateway string `json:"payment_gateway"`
}

type RazorpayPayment struct {
	PaymentID string `form:"razorpay_payment_id" binding:"required" json:"razorpay_payment_id"`
	OrderID   string `form:"razorpay_order_id" binding:"required" json:"razorpay_order_id"`
	Signature string `form:"razorpay_signature" binding:"required" json:"razorpay_signature"`
}

type PlatformSalesReportInput struct {
	StartDate     string `json:"start_date,omitempty" time_format:"2006-01-02"`
	EndDate       string `json:"end_date,omitempty" time_format:"2006-01-02"`
	Limit         string `json:"limit,omitempty"`
	PaymentStatus string `json:"payment_status"`
}
