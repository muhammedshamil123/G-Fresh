package controllers

import (
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func BestSellingProducts(c *gin.Context) {
	var products []model.ViewProductList

	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name, COUNT(order_items.product_id) AS order_count").Joins("JOIN categories ON categories.id=products.category_id").Joins("JOIN order_items ON order_items.product_id = products.id").Group("products.id, categories.name").Order("order_count DESC").Limit(10).Find(&products)
	if ty.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
			"error":   ty.Error,
		})
		return
	}
	if len(products) < 1 {
		c.JSON(http.StatusOK, gin.H{
			"products": "empty",
		})
	}
	for i, val := range products {
		if val.StockLeft < 1 {
			c.JSON(http.StatusOK, gin.H{
				strconv.Itoa(i + 1): val,
				"status":            "out of stock",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				strconv.Itoa(i + 1): val,
			})
		}
	}

}

func BestSellingCategory(c *gin.Context) {
	type count struct {
		Name       string `json:"name"`
		Id         uint   `json:"id"`
		CategoryID uint   `json:"category_id"`
	}
	var counts []count
	ty := database.DB.Model(&model.Product{}).Select("products.name, products.id, products.category_id").Joins("JOIN order_items ON order_items.product_id = products.id").Find(&counts)
	if ty.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
			"error":   ty.Error,
		})
		return
	}
	mapp := make(map[uint]uint)
	var cat []uint

	for _, val := range counts {
		mapp[val.CategoryID]++
		i := 0
		for i < len(cat) {
			if cat[i] == val.CategoryID {
				break
			}
			i++
		}
		if i == len(cat) {
			cat = append(cat, val.CategoryID)
		}
	}
	for i := 0; i < len(cat)-1; i++ {
		for j := i + 1; j < len(cat); j++ {
			if mapp[cat[i]] < mapp[cat[j]] {
				cat[i], cat[j] = cat[j], cat[i]
			}
		}
	}
	fmt.Println(mapp)
	for i, val := range cat {
		if i+1 == 10 {
			break
		}
		var category model.ViewCategoryList
		database.DB.Model(&model.Category{}).Where("id=?", val).First(&category)
		c.JSON(http.StatusOK, gin.H{
			strconv.Itoa(i + 1): category,
		})

	}
}
