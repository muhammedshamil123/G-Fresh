package controllers

import (
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddToCart(c *gin.Context) {

	var cart model.CartItems
	var userId model.User

	user, exist := c.Get("email")

	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Where("email = ?", user).First(&userId); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}

	prodid, _ := strconv.Atoi(c.Query("pid"))
	qty, _ := strconv.Atoi(c.Query("quantity"))
	fmt.Println(prodid, qty)
	var product model.Product

	if tx := database.DB.Where("id = ?", prodid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product id does not exists!",
		})
		return
	}

	if product.StockLeft < uint(qty) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Insufficient Quantity!",
		})
		return
	}
	if 1 > uint(qty) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Min Quantity Required!",
		})
		return
	}

	if 10 < uint(qty) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Max quantity exceeded!",
		})
		return
	}

	if tx := database.DB.Model(&model.CartItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).First(&cart); tx.Error == nil {
		cart.Quantity += uint(qty)
		if 10 < cart.Quantity {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Max quantity exceeded!",
			})
			return
		}
		database.DB.Model(&model.CartItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).Update("quantity", cart.Quantity)
		// database.DB.Model(&model.Product{}).Where("id=?", prodid).Update("stock_left", product.StockLeft-uint(qty))
		c.JSON(http.StatusOK, gin.H{
			"message": "Added to cart succesfully!!",
		})
		return
	}
	newcart := model.CartItems{
		ProductID: uint(prodid),
		Quantity:  uint(qty),
		UserID:    userId.ID,
	}

	result := database.DB.Create(&newcart)
	// database.DB.Model(&model.Product{}).Where("id=?", prodid).Update("stock_left", product.StockLeft-uint(qty))
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Error adding to cart",
			"error":   result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Add to cart succesfully!!",
	})

}

func DeleteFromCart(c *gin.Context) {

	var cart model.CartItems
	var userId model.User

	user, exist := c.Get("email")

	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Where("email = ?", user).First(&userId); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}

	prodid, _ := strconv.Atoi(c.Query("pid"))

	var product model.Product

	if tx := database.DB.Where("id = ?", prodid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product id does not exists!",
		})
		return
	}
	if tx := database.DB.Model(&model.CartItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).First(&cart); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cart item does not exists!",
		})
		return
	}
	if tx := database.DB.Where("user_id=? AND product_id=?", cart.UserID, cart.ProductID).Delete(&model.CartItems{}); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Deletion Failed!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Item deleted succesfully!!",
	})
}
