package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ShowWallet(c *gin.Context) {
	var user model.User

	email, exist := c.Get("email")

	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Where("email = ?", email).First(&user); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}

	var wallet []model.UserWalletHistory
	if tx := database.DB.Model(&model.UserWalletHistory{}).Where("user_id=?", user.ID).Find(&wallet); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "wallet empty!",
		})
		return
	}
	var balance float64
	for i, val := range wallet {
		c.JSON(http.StatusOK, gin.H{
			strconv.Itoa(i + 1): val,
		})
		balance = val.CurrentBalance
	}
	c.JSON(http.StatusOK, gin.H{
		"Balance": balance,
	})
}
