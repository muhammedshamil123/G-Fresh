package controllers

import (
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func ShowCoupon(c *gin.Context) {
	var coupons []model.CouponInventory
	database.DB.Model(&model.CouponInventory{}).Find(&coupons)
	if len(coupons) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Coupons Empty",
		})
		return
	}
	for i, val := range coupons {
		c.JSON(http.StatusOK, gin.H{
			strconv.Itoa(i + 1): val,
		})
	}
}
func AddCoupon(c *gin.Context) {

	// database.DB.AutoMigrate(&model.CouponInventory{})
	var coupon, exist model.CouponInventory

	if err := c.BindJSON(&coupon); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
		})
		return
	}
	fmt.Println("coupon: ", coupon)
	Validate := validator.New()

	err := Validate.Struct(coupon)
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
	if coupon.Percentage <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "percentage should be greater than 1!!",
		})
		return
	}
	if coupon.MaximumUsage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "max usage must be greater than 1!!",
		})
		return
	}
	if coupon.MinimumAmount <= 1 || coupon.MinimumAmount > coupon.MaximumAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "min amount must be between 1 and maximum amount!!",
		})
		return
	}
	if coupon.MaximumAmount <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "max amount must be greater than 1!!",
		})
		return
	}
	if tx := database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", coupon.CouponCode).First(&exist); tx.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Coupon code": "same coupon code exist!!",
		})
		return
	}
	database.DB.Model(&model.CouponInventory{}).Create(&coupon)
	c.JSON(http.StatusOK, gin.H{
		"message": "Created",
		"coupon":  coupon,
	})
}
func DeleteCoupon(c *gin.Context) {
	code := c.Query("code")
	// database.DB.AutoMigrate(&model.CouponUsage{})
	if tx := database.DB.Model(&model.CouponUsage{}).Where("coupon_code=?", code).Delete(&model.CouponUsage{}); tx.Error != nil {
		if tx.Error != gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "coupon usage deletion failed!",
			})
			return
		}
	}
	if tx := database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Delete(&model.CouponInventory{}); tx.Error != nil {
		if tx.Error != gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "coupon deletion failed!",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "coupon not found!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "coupon deleted!",
	})
}
func EditCoupon(c *gin.Context) {
	code := c.Query("code")
	var coupon, exist model.CouponInventory

	if err := c.BindJSON(&coupon); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
		})
		return
	}
	fmt.Println("coupon: ", coupon)
	Validate := validator.New()

	err := Validate.Struct(coupon)
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
	if tx := database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).First(&exist); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Coupon code": "coupon not found!!",
		})
		return
	}

	if coupon.Percentage <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "percentage should be greater than 1!!",
		})
		return
	}
	if coupon.MaximumUsage < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "max usage must be greater than 1!!",
		})
		return
	}
	if coupon.MinimumAmount <= 1 || coupon.MinimumAmount > coupon.MaximumAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "min amount must be between 1 and maximum amount!!",
		})
		return
	}
	if coupon.MaximumAmount <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "max amount must be greater than 1!!",
		})
		return
	}
	if coupon.Expiry != exist.Expiry {
		database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Update("expiry", coupon.Expiry)
	}
	if coupon.MaximumUsage != exist.MaximumUsage {
		database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Update("maximum_usage", coupon.MaximumUsage)
	}
	if coupon.Percentage != exist.Percentage {
		database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Update("percentage", coupon.Percentage)
	}
	if coupon.MinimumAmount != exist.MinimumAmount {
		database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Update("minimum_amount", coupon.MinimumAmount)
	}
	if coupon.CouponCode != exist.CouponCode {
		database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Update("coupon_code", coupon.CouponCode)
	}
	if coupon.MaximumAmount != exist.MaximumAmount {
		database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", code).Update("maximum_amount", coupon.MaximumAmount)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "updated!!",
		"coupon":  coupon,
	})
}
