package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

var jwtSecret = []byte("shamil123")
var ADMINTOKEN string

func AdminLogin(c *gin.Context) {
	// Get the email from the JSON request
	var form struct{ model.AdminLoginRequest }

	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Failed to process the incoming request",
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(form)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "username & password required",
		})
		return
	}

	// Check if email exists in the admin table
	var admin model.Admin
	if tx := database.DB.Where("Username = ?", form.Username).First(&admin); tx.Error != nil {
		fmt.Println(tx)
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Username not present in the admin table",
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
	if err := utils.CheckPassword(admin.Password, form.Password); err == nil {
		token, err := utils.GenerateToken(admin.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token Generation Failed"})
			return
		}
		ADMINTOKEN = token
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Admin signed in successfully",
			"token":   ADMINTOKEN,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Password does not match",
		})
		return
	}
}

func AdminAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ADMINTOKEN

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
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
			c.Set("username", claims["username"])
			c.Set("exp", claims["exp"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
