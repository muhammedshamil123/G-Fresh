package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SendRequest(c *gin.Context) {
	database.DB.AutoMigrate(&model.Request{})
	var user model.User

	pid, _ := strconv.Atoi(c.Query("product_id"))
	count, _ := strconv.Atoi(c.Query("count"))
	useremail, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Where("email = ?", useremail).First(&user); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	var product model.Product

	if tx := database.DB.Where("id = ?", pid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product id does not exists!",
		})
		return
	}

	if count > 10 || count < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Count should be less than 10 and greater than 0!",
		})
		return
	}
	if product.StockLeft > uint(count) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Sufficient Quantity Available!",
		})
		return
	}
	var request model.Request
	if tx := database.DB.Model(&model.Request{}).Where("user_id=? AND product_id=?", user.ID, product.ID).First(&request); tx.Error == nil {

		if request.Response == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Request already submitted!",
			})
			return
		}
	}
	request = model.Request{
		UserID:    user.ID,
		ProductID: product.ID,
		Count:     uint(count),
	}
	if tx := database.DB.Model(&model.Request{}).Create(&request); tx.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Request submitted succesfully!",
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to send request!",
	})
}
func DeleteRequest(c *gin.Context) {
	var user model.User

	request_id, _ := strconv.Atoi(c.Query("request_id"))
	useremail, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Where("email = ?", useremail).First(&user); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}

	var request model.Request
	if tx := database.DB.Model(&model.Request{}).Where("user_id=? AND request_id=?", user.ID, request_id).First(&request); tx.Error != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "No request found!",
		})
		return
	}
	if request.Response != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Request already resonded!",
		})
		return
	}
	if tx := database.DB.Where("user_id=? AND request_id=?", user.ID, request_id).Delete(&model.Request{}); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete request",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Request deleted successfully!",
	})
}

func ViewRequestsUser(c *gin.Context) {
	var user model.User
	useremail, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Where("email = ?", useremail).First(&user); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	var requests []model.Request
	if tx := database.DB.Model(&model.Request{}).Where("user_id", user.ID).Find(&requests); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching request details",
		})
		return
	}
	if len(requests) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Requests empty!",
		})
		return
	}
	for i, val := range requests {
		c.JSON(http.StatusOK, gin.H{
			strconv.Itoa(i + 1): val,
		})
	}
}
func ViewRequests(c *gin.Context) {
	var requests []model.Request
	if tx := database.DB.Model(&model.Request{}).Where("response=?", "").Find(&requests); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching request details",
		})
		return
	}
	if len(requests) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Requests empty!",
		})
		return
	}
	for i, val := range requests {
		c.JSON(http.StatusOK, gin.H{
			strconv.Itoa(i + 1): val,
		})
	}
}

func RequestResponse(c *gin.Context) {

	request_id, _ := strconv.Atoi(c.Query("request_id"))
	count, _ := strconv.Atoi(c.Query("count"))
	var request model.Request
	if tx := database.DB.Model(&model.Request{}).Where("request_id=?", request_id).First(&request); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Request not found!",
		})
		return
	}
	var product model.Product
	if tx := database.DB.Model(&model.Product{}).Where("id=?", request.ProductID).First(&product); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Product not found!",
		})
		return
	}
	if count+int(product.StockLeft) < int(request.Count) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "count less than required count!",
		})
		return
	}
	product.StockLeft += uint(count)
	if tx := database.DB.Model(&model.Product{}).Where("id=?", request.ProductID).Update("stock_left", product.StockLeft); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Product updation failed",
		})
		return
	}
	if tx := database.DB.Model(&model.Request{}).Where("request_id=?", request.RequestID).Update("response", "Stock Updated"); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Product updation failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Responded successfully!",
	})

}
