package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var USERTOKEN, storedOTP string
var otpTimer time.Time
var USER model.User

func UserLoginEmail(c *gin.Context) {
	// Get the email from the JSON request
	var form struct{ model.UserEmailLoginRequest }
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Failed to process the incoming request",
			"error":   err,
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(form)
	if err != nil {
		var errs []string
		for _, e := range err.(validator.ValidationErrors) {
			errMsg := e.Field() + "_" + e.Tag()
			if errMsg == "" {
				errMsg = e.Error()
			}
			errs = append(errs, errMsg)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"err":    errs,
		})
		return
	}

	// Check if email exists in the admin table
	var user model.User
	if tx := database.DB.Where("email = ?", form.Email).First(&user); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Email not present in the user table",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "Database error",
			})
			return
		}
	}

	// Check if password matches the username
	if err := utils.CheckPassword(user.HashedPassword, form.Password); err == nil {
		token, err := utils.GenerateToken(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token Generation Failed"})
			return
		}
		USERTOKEN = token
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "User signed in successfully",
			"token":   USERTOKEN,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Password does not match",
		})
		return
	}
}

func UserSignupEmail(c *gin.Context) {
	var form model.UserEmailSignupRequest
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Failed to process the incoming request",
			"error":   err,
		})
		return
	}
	Validate := validator.New()
	fmt.Println(form)
	err := Validate.Struct(form)
	if err != nil {

		var errs []string
		for _, e := range err.(validator.ValidationErrors) {
			errMsg := e.Field() + "_" + e.Tag()
			if errMsg == "" {
				errMsg = e.Error()
			}
			errs = append(errs, errMsg)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  errs,
		})
		return
	}
	var existingUser model.User
	if tx, ty := database.DB.Where("email = ?", form.Email).First(&existingUser), database.DB.Where("phone_number = ?", form.PhoneNumber).First(&existingUser); tx.Error == nil || ty.Error == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Account With this email/phone number already exist",
		})
		return
	}
	if form.Password != form.ConfirmPassword {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Password Does not match!!",
		})
		return
	}
	hPassword, err := utils.HashPassword(form.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": err,
		})
		return
	}
	user := model.User{
		Name:           form.Name,
		Email:          form.Email,
		PhoneNumber:    form.PhoneNumber,
		Picture:        "",
		Blocked:        false,
		HashedPassword: hPassword,
	}

	otp, err := utils.GenerateOTP(6, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	err = utils.SendEmailOTP(user.Email, otp.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "OTP sent to email",
	})
	storedOTP = otp.OTP
	otpTimer = otp.ExpiryTime
	USER = user
}
func OtpVerification(c *gin.Context) {

	email := c.Query("email")
	otpParam := c.Query("otp")
	user := USER

	if time.Now().After(otpTimer) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "OTP has expired",
		})
		return
	}
	if otpParam == storedOTP && email == user.Email {

		result := database.DB.Create(&user)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Error creating user",
				"error":   result.Error,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "OTP verified successfully and user Created!!",
			"welcome": user.Name,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Invalid OTP",
		})
	}
}
func ResendOtp(c *gin.Context) {
	user := USER
	otp, err := utils.GenerateOTP(6, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	err = utils.SendEmailOTP(user.Email, otp.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "OTP sent to email",
	})
	storedOTP = otp.OTP
	otpTimer = otp.ExpiryTime
}
