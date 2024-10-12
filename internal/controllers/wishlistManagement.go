package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ShowWishlist(c *gin.Context) {
	database.DB.AutoMigrate(&model.WishlistItems{})
	var wishlistItems []model.WishlistItems

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

	if tx := database.DB.Model(&model.WishlistItems{}).Where("user_id=?", user.ID).Find(&wishlistItems); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Wishlist items does not exists!",
		})
		return
	}
	for _, val := range wishlistItems {
		var products model.Product

		if tx := database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).First(&products); tx.Error != nil {
			c.JSON(http.StatusOK, gin.H{
				"Availabitity": "Product unavailable",
			})
			database.DB.Where("user_id=? AND product_id=?", user.ID, val.ProductID).Delete(&model.WishlistItems{})
			continue
		}
		product := model.ViewCartList{
			Name:        products.Name,
			Description: products.Description,
			Price:       products.Price,
			OfferAmount: products.OfferAmount,
			StockLeft:   products.StockLeft,
		}

		c.JSON(http.StatusOK, gin.H{
			"item": product,
		})
	}
	if len(wishlistItems) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": "Wislist Empty",
		})
	}
}
func AddToWishlist(c *gin.Context) {

	var wishlist model.WishlistItems
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

	if tx := database.DB.Model(&model.WishlistItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).First(&wishlist); tx.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Item already in wishlist",
		})
		return
	}
	newwishlist := model.WishlistItems{
		ProductID: uint(prodid),
		UserID:    userId.ID,
	}

	result := database.DB.Create(&newwishlist)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Error adding to wishlist",
			"error":   result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Add to wislist succesfully!!",
	})
}
func DeleteFromWishlist(c *gin.Context) {
	var wishlist model.WishlistItems
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
	if tx := database.DB.Model(&model.WishlistItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).First(&wishlist); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Wishlist item does not exists!",
		})
		return
	}
	if tx := database.DB.Where("user_id=? AND product_id=?", wishlist.UserID, wishlist.ProductID).Delete(&model.WishlistItems{}); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Removing Failed!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed succesfully!!",
	})
}

func MoveToCart(c *gin.Context) {
	var wishlist model.WishlistItems
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

	var product model.Product

	if tx := database.DB.Where("id = ?", prodid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product id does not exists!",
		})
		return
	}
	if tx := database.DB.Model(&model.WishlistItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).First(&wishlist); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Wishlist item does not exists!",
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
	var cart model.CartItems
	if tx := database.DB.Model(&model.CartItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).First(&cart); tx.Error == nil {
		cart.Quantity += uint(qty)
		if 10 < cart.Quantity {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Max quantity exceeded!",
			})
			return
		}
		if product.StockLeft < cart.Quantity {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Insufficient Quantity!",
			})
			return
		}
		database.DB.Model(&model.CartItems{}).Where("product_id = ? AND user_id = ?", prodid, userId.ID).Update("quantity", cart.Quantity)
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
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Error adding to cart",
			"error":   result.Error,
		})
		return
	}
	if tx := database.DB.Where("user_id=? AND product_id=?", wishlist.UserID, wishlist.ProductID).Delete(&model.WishlistItems{}); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Removing Failed!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Item moved succesfully!!",
	})
}
