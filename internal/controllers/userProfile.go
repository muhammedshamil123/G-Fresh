package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ShowProfile(c *gin.Context) {

	var userDetails model.UserResponse

	user, exist := c.Get("email")

	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	if tx := database.DB.Model(&model.User{}).Select("id, name, email, phone_number,picture,blocked").Where("email = ?", user).First(&userDetails); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": userDetails,
	})
}

func ShowAddress(c *gin.Context) {

	var userId model.User
	var address []model.Address

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
	if tx := database.DB.Model(&model.Address{}).Where("user_id = ?", userId.ID).Find(&address); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "address does not exists!",
		})
		return
	}
	for _, val := range address {
		c.JSON(http.StatusOK, gin.H{
			"user": val,
		})
	}
}
func EditProfile(c *gin.Context) {

	var userDetails model.ProfileEdit
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

	if err := c.BindJSON(&userDetails); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
			"err":     err,
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(userDetails)
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
			"error": errs,
		})
		return
	}

	user.Name = userDetails.Name
	if user.Email != userDetails.Email {
		user.Email = userDetails.Email
		token, err := utils.GenerateToken(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token Generation Failed"})
			return
		}
		USERTOKEN = token
	}
	user.PhoneNumber = userDetails.PhoneNumber
	user.Picture = userDetails.Picture
	tx := database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Updation failed!",
			"error":   tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated",
		"Name":    user.Name,
	})
}

func ChangePassword(c *gin.Context) {
	var form model.ChangePasswordRequest
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

	if err := c.BindJSON(&form); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
			"err":     err,
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
			"error": errs,
		})
		return
	}
	if err := utils.CheckPassword(user.HashedPassword, form.OldPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Password Incorrect",
		})
		return
	}
	if form.Password != form.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Password does not match",
		})
		return
	}
	user.HashedPassword, _ = utils.HashPassword(form.Password)
	tx := database.DB.Model(&user).Update("hashed_password", user.HashedPassword)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Updation failed!",
			"error":   tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated",
		"Name":    user.Name,
	})
}
func ShowCart(c *gin.Context) {
	var cartItems []model.CartItems

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

	if tx := database.DB.Model(&model.CartItems{}).Where("user_id=?", user.ID).Find(&cartItems); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cart items does not exists!",
		})
		return
	}
	var total int
	for _, val := range cartItems {
		var products model.Product

		if tx := database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).First(&products); tx.Error != nil {
			c.JSON(http.StatusOK, gin.H{
				"Availabitity": "Product unavailable",
			})
			database.DB.Where("user_id=? AND product_id=?", user.ID, val.ProductID).Delete(&model.CartItems{})
			continue
		}
		product := model.ViewCartList{
			Name:        products.Name,
			Description: products.Description,
			Price:       products.Price,
			OfferAmount: products.OfferAmount,
			StockLeft:   products.StockLeft,
		}
		if product.StockLeft < val.Quantity {
			if product.StockLeft <= 0 {
				c.JSON(http.StatusNotModified, gin.H{
					"item":         product,
					"Availabitity": "Product unavailable",
				})
				// database.DB.Where("user_id=? AND product_id=?", user.ID, val.ProductID).Delete(&model.CartItems{})
				continue
			}
			val.Quantity = product.StockLeft
			database.DB.Model(&model.CartItems{}).Where("user_id=? AND product_id=?", user.ID, val.ProductID).Update("quantity", val.Quantity)

			c.JSON(http.StatusOK, gin.H{
				"item":     product,
				"quantity": val.Quantity,
				"updation": "Quantity Decreased to Availability",
			})
			total += int(product.Price) * int(val.Quantity)
			continue
		}
		c.JSON(http.StatusOK, gin.H{
			"item":     product,
			"quantity": val.Quantity,
		})
		total += int(product.Price) * int(val.Quantity)
	}
	if len(cartItems) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": "Cart Empty",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"Total Amount": total,
	})
	return
}
