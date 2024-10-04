package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddRating(c *gin.Context) {

	database.DB.AutoMigrate(&model.Rating{})
	var userId int

	pid := c.Query("pid")
	rat, _ := strconv.Atoi(c.Query("rating"))
	user, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	if rat < 1 || rat > 5 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Rating should be between 1-5",
		})
		return
	}
	if tx := database.DB.Model(&model.User{}).Select("id").Where("email = ?", user).First(&userId); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	var product model.Product
	if tx := database.DB.Model(&model.Product{}).Where("id = ?", pid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product does not exists!",
		})
		return
	}
	rating := model.Rating{
		UserID:    uint(userId),
		ProductID: product.ID,
		Rating:    uint(rat),
	}
	if tx := database.DB.Model(&model.Rating{}).Create(&rating); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product does not exists!",
		})
		return
	}
	product.RatingSum += float64(rat)
	product.RatingCount++
	product.AverageRating = product.RatingSum / float64(product.RatingCount)
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).Updates(map[string]interface{}{"rating_sum": product.RatingSum, "rating_count": product.RatingCount, "average_rating": product.AverageRating}); tx.Error != nil {
		database.DB.Model(&model.Rating{}).Where("user_id=? AND product_id=?", userId, pid).Delete(&model.Rating{})
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Updation failed!",
		})
		return
	}
	// database.DB.Model(&model.Product{}).Where("id=?", pid).Updates(map[string]interface{}{"rating_sum": 0, "rating_count": 0, "average_rating": 0})
	// database.DB.Model(&model.Rating{}).Where("user_id=? AND product_id=?", userId, pid).Delete(&model.Rating{})
	c.JSON(http.StatusOK, gin.H{
		"message": "Rating added",
	})
}
