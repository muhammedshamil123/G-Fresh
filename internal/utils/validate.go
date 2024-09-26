package utils

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func Validate(value interface{}) error {
	var customValidationMessages = map[string]string{
		"username_required":        "Please enter a username",
		"Name_required":            "Please enter a name.",
		"Email_required":           "Please enter an email address.",
		"Email_email":              "Please enter a valid email address.",
		"PhoneNumber_required":     "Please enter a phone number.",
		"PhoneNumber_number":       "Please enter a numerical value for phone number.",
		"PhoneNumber_min":          "Phone number should be at least 10 digits.",
		"PhoneNumber_max":          "Phone number should not exceed 10 digits.",
		"password_required":        "Please enter a password.",
		"ConfirmPassword_required": "Please confirm your password.",
		"Description_required":     "Please enter a description.",
		"Address_required":         "Please enter an address.",
		"ImageURL_required":        "Please enter a URL for the image.",
		"CertificateURL_required":  "Please enter a URL for the certificate.",
		"ProductID_required":       "Please enter a product ID.",
		"ProductID_number":         "Please enter a numerical value for product ID.",
		"Quantity_required":        "Please enter a quantity.",
		"Quantity_number":          "Please enter a numerical value for quantity.",
		"CategoryID_required":      "Please enter a category ID.",
		"CategoryID_number":        "Please enter a numerical value for category ID.",
		"Price_required":           "Please enter a price.",
		"Price_number":             "Please enter a numerical value for price.",
		"OfferAmount_number":       "Please enter a numerical value for offer amount.",
		"MaxStock_required":        "Please enter a maximum stock count.",
		"MaxStock_number":          "Please enter a numerical value for maximum stock count.",
		"StockLeft_required":       "Please enter a stock count.",
		"StockLeft_number":         "Please enter a numerical value for stock count.",
		"Expiry_required":          "Please enter an expiry date.",
		"Expiry_number":            "Please enter a numerical value for expiry date.",
		"Percentage_required":      "Please enter a percentage.",
		"Percentage_number":        "Please enter a numerical value for percentage.",
		"MaximumUsage_required":    "Please enter a maximum usage count.",
		"MaximumUsage_number":      "Please enter a numerical value for maximum usage count.",
		"MinimumAmount_required":   "Please enter a minimum amount.",
		"MinimumAmount_number":     "Please enter a numerical value for minimum amount.",
		"UserID_required":          "Please enter a user ID.",
		"UserID_number":            "Please enter a numerical value for user ID.",
		"AddressID_required":       "Please enter an address ID.",
		"AddressID_number":         "Please enter a numerical value for address ID.",
		"PaymentMethod_required":   "Please enter a payment method.",
		"OrderID_required":         "Please enter an order ID.",
		"OrderID_number":           "Please enter a numerical value for order ID.",
		"UserRating_required":      "Please provide a rating.",
		"UserRating_number":        "Please enter a numerical value for rating.",
		"RestaurantID_required":    "Please provide the restaurant_id for placing the orders of that particular cart",
	}

	// validate the struct body
	validate := validator.New()
	err := validate.Struct(value)
	if err != nil {
		var errs []string
		for _, e := range err.(validator.ValidationErrors) {
			translationKey := e.Field() + "_" + e.Tag()
			errMsg := customValidationMessages[translationKey]
			if errMsg == "" {
				errMsg = e.Error()
			}
			errs = append(errs, errMsg)
		}
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}
