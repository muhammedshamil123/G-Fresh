package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GetProductList(c *gin.Context) {
	var productlist []model.ProductResponse

	tx := database.DB.Model(&model.Product{}).Select("id, category_id, name, description, image_url,price,offer_amount,stock_left,rating_count,average_rating").Find(&productlist)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	if len(productlist) < 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Products empty",
		})
		return
	}
	for _, val := range productlist {
		c.JSON(http.StatusOK, gin.H{
			"product": val,
		})
	}
}

func AddProducts(c *gin.Context) {
	// database.DB.AutoMigrate(&model.Product{})

	var product model.AddProductsRequest
	if err := c.BindJSON(&product); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(product)
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

	var existingProd model.Product
	if tx := database.DB.Where("name = ?", product.Name).First(&existingProd); tx.Error == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Product with same name exists!",
		})
		return
	}
	var cat model.Category
	if tx := database.DB.Model(&model.Category{}).Where("id", product.CategoryID).First(&cat); tx.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Category does not exist!",
		})
		return
	}
	prod := model.Product{
		CategoryID:    product.CategoryID,
		Name:          product.Name,
		Description:   product.Description,
		ImageURL:      product.ImageURL,
		Price:         product.Price,
		OfferAmount:   product.OfferAmount,
		StockLeft:     product.StockLeft,
		RatingSum:     0,
		RatingCount:   0,
		AverageRating: 0,
	}

	result := database.DB.Create(&prod)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Error creating Product",
			"error":   result.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Product created Created!!",
		"Name":    prod.Name,
		"Id":      prod.ID,
	})
}

func DeleteProduct(c *gin.Context) {

	proid := c.Query("id")
	var product model.Product
	if tx := database.DB.Where("id = ?", proid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product id does not exists!",
		})
		return
	}
	tx := database.DB.Where("id = ?", proid).Delete(&model.Product{})
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Deletion failed!",
			"error":   tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted",
	})

}

func EditProduct(c *gin.Context) {

	prodid := c.Query("id")
	var form model.AddProductsRequest
	var product model.Product

	if tx := database.DB.Where("id = ?", prodid).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product id does not exists!",
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
		var errMsg string
		for _, e := range err.(validator.ValidationErrors) {
			errMsg = e.Field() + "_" + e.Tag()
			if errMsg == "" {
				errMsg = e.Error()
			}
			errs = append(errs, errMsg)
			if errMsg == "StockLeft_required" {
				break
			}
		}
		if errMsg != "StockLeft_required" {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": errs,
			})
			return
		}
	}
	product.CategoryID = form.CategoryID
	product.Name = form.Name
	product.Description = form.Description
	product.ImageURL = form.ImageURL
	product.Price = form.Price
	product.OfferAmount = form.OfferAmount
	product.StockLeft = form.StockLeft

	tx := database.DB.Save(&product)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Updation failed!",
			"error":   tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated",
		"product": product.Name,
	})

}
