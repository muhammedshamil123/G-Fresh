package controllers

import (
	"errors"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserList(c *gin.Context) {

	var users []model.UserResponse

	tx := database.DB.Model(&model.User{}).Select("id, name, email, phone_number, picture, referral_code, wallet_amount, login_method, blocked").Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	for _, val := range users {
		c.JSON(http.StatusOK, gin.H{
			"users": val,
		})
	}

}

func BlockUser(c *gin.Context) {
	userid := c.Query("userId")

	if tx := database.DB.Model(&model.User{}).Where("id = ?", userid).Update("blocked", true); tx.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "User is Blocked blocked",
		})
		return
	} else {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User Id not present in the database",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error",
			})
			return
		}
	}

}
func UnblockUser(c *gin.Context) {
	userid := c.Query("userId")
	if tx := database.DB.Model(&model.User{}).Where("id = ?", userid).Update("blocked", false); tx.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "User is Unblocked blocked",
		})
		return
	} else {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User Id not present in the database",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error",
			})
			return
		}
	}

}
