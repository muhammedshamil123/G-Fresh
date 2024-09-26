package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserList(c *gin.Context) {

	var users []model.UserResponse

	tx := database.DB.Model(&model.User{}).Select("id, name, email, phone_number, picture, referral_code, wallet_amount, login_method, blocked").Find(&users)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
}
