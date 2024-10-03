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
	Name        string `json:"name"`
	Description string `validate:"required" json:"description"`
	ImageURL    string `validate:"required" json:"imageurl"`
}
type AddProductsRequest struct {
	CategoryID  uint    `validate:"required" json:"categoryid"`
	Name        string  `validate:"required" json:"name"`
	Description string  `validate:"required" json:"description"`
	ImageURL    string  `gvalidate:"required" json:"imageurl"`
	Price       float64 `validate:"required,number" json:"price"`
	OfferAmount float64 `json:"offeramount"`
	StockLeft   uint    `validate:"required,number" json:"stockleft"`
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
	Email       string `validate:"required" json:"email"`
	PhoneNumber string `validate:"required" json:"phonenumber"`
	Picture     string `validate:"required" json:"picture"`
}

type ChangePasswordRequest struct {
	OldPassword     string `validate:"required" json:"oldpassword"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}
