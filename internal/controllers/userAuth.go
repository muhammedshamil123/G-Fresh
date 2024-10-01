package controllers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     "561809634466-7789k623dstmaotg90edbil5r07iscl4.apps.googleusercontent.com",
	ClientSecret: "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",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

var oauthStateString = "randomstate"
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
		if user.Blocked {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
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
			"status":  true,
			"message": "User signed in successfully",
			"welcome": user.Name,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP", "err": err})
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
			c.Set("username", claims["name"])
			c.Set("exp", claims["exp"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
func HandleGoogleLoginsbacks(c *gin.Context) {
	http.HandleFunc("/auth/google/login", HandleGoogleLogin)
}
func HandleGoogleCallbacks(c *gin.Context) {
	http.HandleFunc("/auth/google/callback", HandleGoogleCallback)
}

// Step 1: Redirect to Google for authentication
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Step 2: Handle callback from Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		log.Printf("Invalid OAuth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Code exchange failed: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Use the token to get user info
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error getting user info: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer response.Body.Close()

	// Read the user's information
	userInfo, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading user info: %s", err.Error())
		return
	}

	// Display or store the user information (e.g., email, name)
	fmt.Fprintf(w, "User Info: %s\n", userInfo)
}
