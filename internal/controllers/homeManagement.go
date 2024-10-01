package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHome(c *gin.Context) {

	var categories []model.ViewCategoryList
	var products []model.ViewProductList
	tx := database.DB.Model(&model.Category{}).Select("id, name, description, image_url").Find(&categories)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the category database, or the data doesn't exist",
		})
		return
	}
	for _, val := range categories {
		c.JSON(http.StatusOK, gin.H{
			"category": val,
		})
	}
	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Find(&products)
	if ty.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
			"error":   ty.Error,
		})
		return
	}
	for _, val := range products {
		c.JSON(http.StatusOK, gin.H{
			"products": val,
		})
	}
}
func GetCategory(c *gin.Context) {
	catid := c.Query("id")
	var products []model.ViewProductList
	var category model.Category

	if tx := database.DB.Where("id = ?", catid).First(&category); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Category id does not exists!",
		})
		return
	}
	tx := database.DB.Model(&model.Product{}).Select("name, description, image_url,price,offer_amount,stock_left,rating_count,average_rating").Where("category_id = ?", catid).Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}
	if len(products) == 0 {
		c.JSON(http.StatusNoContent, gin.H{
			"products": "No Products",
		})
		return
	}
	for _, val := range products {
		c.JSON(http.StatusOK, gin.H{
			"products": val,
		})
	}
}

func GetProduct(c *gin.Context) {
	prodid := c.Query("id")
	var products model.ViewProductList

	tx := database.DB.Model(&model.Product{}).Select("name, description, image_url,price,offer_amount,stock_left,rating_count,average_rating").Where("id = ?", prodid).Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Product with this id does not exist!!",
		})
		return
	}

	if products.StockLeft < 1 {
		c.JSON(http.StatusOK, gin.H{
			"products": products,
			"stock":    "Out of stock",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})

}
