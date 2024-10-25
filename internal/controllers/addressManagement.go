package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func AddAddress(c *gin.Context) {

	var userId model.User
	var address []model.Address
	var newadd model.AddAddressRequest

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
	database.DB.Model(&model.Address{}).Where("user_id = ?", userId.ID).Find(&address)
	if len(address) >= 3 {
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Max address limit reached!",
		})
		return
	}

	if err := c.BindJSON(&newadd); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
			"err":     err,
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(newadd)
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
	a, _ := strconv.Atoi(newadd.PostalCode)
	if a < 100000 || a > 999999 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "postal code invalid",
		})
		return
	}
	add := model.Address{
		UserID:       userId.ID,
		PhoneNumber:  newadd.PhoneNumber,
		StreetName:   newadd.StreetName,
		StreetNumber: newadd.StreetNumber,
		City:         newadd.City,
		State:        newadd.State,
		PostalCode:   newadd.PostalCode,
	}

	result := database.DB.Create(&add)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Error creating Product",
			"error":   result.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Address added succesfully!!",
	})

}

func DeleteAddress(c *gin.Context) {

	addid := c.Query("id")
	var address model.Address
	if tx := database.DB.Where("address_id = ?", addid).First(&address); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Address id does not exists!",
		})
		return
	}
	tx := database.DB.Where("address_id = ?", addid).Delete(&model.Address{})
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Deletion failed!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Address deleted",
	})

}

func EditAddress(c *gin.Context) {

	catid := c.Query("id")
	var form model.AddAddressRequest
	var address model.Address

	if tx := database.DB.Where("address_id = ?", catid).First(&address); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Address id does not exists!",
		})
		return
	}

	if err := c.BindJSON(&form); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
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

	a, _ := strconv.Atoi(form.PostalCode)
	if a < 100000 || a > 999999 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "postal code invalid",
		})
		return
	}
	address.PhoneNumber = form.PhoneNumber
	address.PostalCode = form.PostalCode
	address.State = form.State
	address.City = form.City
	address.StreetName = form.StreetName
	address.StreetNumber = form.StreetNumber

	tx := database.DB.Save(&address)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Updation failed!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Address updated",
		"adderess": address,
	})

}
