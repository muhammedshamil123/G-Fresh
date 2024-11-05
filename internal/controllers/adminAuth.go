package controllers

import (
	"errors"
	"net/http"
	"os"
	"log"

	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

var jwtSecret = []byte(os.Getenv("JWTSECRET"))
var ADMINTOKEN string

func AdminLogin(c *gin.Context) {

	var form struct{ model.AdminLoginRequest }
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(form)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "username & password required",
		})
		return
	}
	var admin model.Admin
	if tx := database.DB.Where("Username = ?", form.Username).First(&admin); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Username not present in the admin table",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error",
			})
			return
		}
	}
	log.Println(admin.Password,form.Password)
	if admin.Password==form.Password {
		token, err := utils.GenerateToken(admin.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token Generation Failed"})
			return
		}
		ADMINTOKEN = token
		c.JSON(http.StatusOK, gin.H{
			"message": "Admin signed in successfully",
			"welcome": admin.Username,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Password does not match",
		})
		return
	}
}

func AdminAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ADMINTOKEN

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin not logged in"})
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
