package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "https://www.shaamil.in.net/auth/google/callback",
	ClientID:     "",
	ClientSecret: "",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

var oauthStateString = os.Getenv("oauthStateString")
var USERTOKEN, storedOTP string
var otpTimer time.Time
var USER model.User

func UserLoginEmail(c *gin.Context) {
	var form struct{ model.UserEmailLoginRequest }
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
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
			"err": errs,
		})
		return
	}

	var user model.User
	if tx := database.DB.Where("email = ?", form.Email).First(&user); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Email not present in the user table",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error",
			})
			return
		}
	}

	if err := utils.CheckPassword(user.HashedPassword, form.Password); err == nil {
		if user.Blocked {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User is blocked",
			})
			return
		}
		token, err := utils.GenerateToken(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token Generation Failed"})
			return
		}
		USERTOKEN = token
		c.JSON(http.StatusOK, gin.H{
			"message": "User signed in successfully",
			"welcome": user.Name,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Password does not match",
		})
		return
	}
}

func UserSignupEmail(c *gin.Context) {
	var form model.UserEmailSignupRequest
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
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
			"error": errs,
		})
		return
	}
	var existingUser model.User
	if tx, ty := database.DB.Where("email = ?", form.Email).First(&existingUser), database.DB.Where("phone_number = ?", form.PhoneNumber).First(&existingUser); tx.Error == nil || ty.Error == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Account With this email/phone number already exist",
		})
		return
	}
	if len(form.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "password min length is 8",
		})
		return
	}
	if form.Password != form.ConfirmPassword {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Password Does not match!!",
		})
		return
	}
	hPassword, err := utils.HashPassword(form.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err,
		})
		return
	}
	p, _ := strconv.Atoi(form.PhoneNumber)
	fmt.Println(p)
	if p < 1000000000 || p > 9999999999 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "phone number not valid",
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
		ReferralCode:   utils.GenerateReferralCode(),
	}

	otp, err := utils.GenerateOTP(6, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	err = utils.SendEmailOTP(user.Email, otp.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP", "err": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
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
			"message": "OTP has expired",
		})
		return
	}
	if otpParam == storedOTP && email == user.Email {

		result := database.DB.Create(&user)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Error creating user",
				"error":   result.Error,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "OTP verified successfully and user Created!!",
			"welcome": user.Name,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
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
		"message": "OTP sent to email",
	})
	storedOTP = otp.OTP
	otpTimer = otp.ExpiryTime
}
func UserAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := USERTOKEN

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
			c.Abort()
			return
		}
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("email", claims["username"])
			c.Set("exp", claims["exp"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}

func HandleGoogleLogin(c *gin.Context) {
	googleOauthConfig.ClientID = os.Getenv("ClientID")
	googleOauthConfig.ClientSecret = os.Getenv("ClientSecret")
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {

	code := strings.TrimSpace(c.Query("code"))

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "missing code parameter",
		})
		return
	}
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Token Exchange Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to exchange token",
		})
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get user information",
		})
		return
	}

	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to read user information",
		})
		return
	}
	var googleUser model.GoogleResponse
	err = json.Unmarshal(content, &googleUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to parse user information",
		})
		return
	}

	var existingUser model.User
	if err := database.DB.Where("email = ?", googleUser.Email).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			newUser := model.User{
				Email:   googleUser.Email,
				Name:    googleUser.Name,
				Picture: googleUser.Picture,
			}
			if err := database.DB.Create(&newUser).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to create new user",
				})
				return
			}
			existingUser = newUser
		} else {

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to fetch user from database",
			})
			return
		}
	}

	if existingUser.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "user is unauthorized to access",
		})
		return
	}

	tokenstring, err := utils.GenerateToken(existingUser.Email)
	if tokenstring == "" || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "failed to create authorization token",
		})
		return
	}
	USERTOKEN = tokenstring

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "login successful",
	})
}
