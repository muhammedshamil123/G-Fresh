package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetHome(c *gin.Context) {

	var categories []model.ViewCategoryList
	var products []model.ViewProductList
	tx := database.DB.Model(&model.Category{}).Select("id, name, description, image_url").Find(&categories)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the category database, or the data doesn't exist",
		})
		return
	}
	if len(categories) < 1 {
		c.JSON(http.StatusOK, gin.H{
			"category": "empty",
		})
	}
	for _, val := range categories {
		c.JSON(http.StatusOK, gin.H{
			"category": val,
		})
	}
	ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Find(&products)
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
	for _, val := range products {
		if val.StockLeft < 1 {
			c.JSON(http.StatusOK, gin.H{
				"products": val,
				"status":   "out of stock",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"products": val,
			})
		}
	}
}
func GetCategory(c *gin.Context) {
	catid := c.Query("id")
	var products []model.ViewProductList
	var category model.Category

	if tx := database.DB.Where("id = ?", catid).First(&category); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Category id does not exists!",
		})
		return
	}
	tx := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Where("category_id = ?", catid).Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the products database, or the data doesn't exist",
		})
		return
	}
	if len(products) < 1 {
		c.JSON(http.StatusNotFound, gin.H{
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

	tx := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Where("products.id = ?", prodid).Find(&products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product with this id does not exist!!",
		})
		return
	}
	if products.Name == "" {
		c.JSON(http.StatusNotFound, gin.H{
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

func Search_P_LtoH(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("price ASC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("price ASC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func Search_P_HtoL(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("price DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("price DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func SearchNew(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("products.created_at DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("products.created_at DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}

}

func SearchAtoZ(c *gin.Context) {

	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("LOWER(products.name) ASC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("LOWER(products.name) ASC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func SearchZtoA(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("LOWER(products.name) DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("LOWER(products.name) DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func SearchAverageRating(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("average_rating DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).Select("products.name, products.description, products.image_url,price,offer_amount,stock_left,rating_count,average_rating,categories.name AS category_name").Joins("JOIN categories ON categories.id=products.category_id").Order("average_rating DESC").Find(&products)
		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func SearchPopular(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).
			Select("products.name, products.description, products.image_url, price, offer_amount, stock_left, rating_count, average_rating, categories.name AS category_name, COUNT(order_items.product_id) AS order_count").
			Joins("JOIN categories ON categories.id = products.category_id").
			Joins("LEFT JOIN order_items ON order_items.product_id = products.id").
			Group("products.id, categories.name").
			Order("order_count DESC").
			Find(&products)

		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).
			Select("products.name, products.description, products.image_url, price, offer_amount, stock_left, rating_count, average_rating, categories.name AS category_name, COUNT(order_items.product_id) AS order_count").
			Joins("JOIN categories ON categories.id = products.category_id").
			Joins("LEFT JOIN order_items ON order_items.product_id = products.id").
			Group("products.id, categories.name").
			Order("order_count DESC").
			Find(&products)

		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func SearchFeatured(c *gin.Context) {
	cat, _ := strconv.Atoi(c.Query("category"))
	if cat == 0 {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).
			Select("products.name, products.description, products.image_url, price, offer_amount, stock_left, rating_count, average_rating, categories.name AS category_name, COUNT(cart_items.product_id) AS cart_count").
			Joins("JOIN categories ON categories.id = products.category_id").
			Joins("LEFT JOIN cart_items ON cart_items.product_id = products.id").
			Group("products.id, categories.name").
			Order("cart_count DESC").
			Find(&products)

		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else if categoryExist(cat) {
		var products []model.ViewProductList
		ty := database.DB.Model(&model.Product{}).Where("category_id=?", cat).
			Select("products.name, products.description, products.image_url, price, offer_amount, stock_left, rating_count, average_rating, categories.name AS category_name, COUNT(cart_items.product_id) AS cart_count").
			Joins("JOIN categories ON categories.id = products.category_id").
			Joins("LEFT JOIN cart_items ON cart_items.product_id = products.id").
			Group("products.id, categories.name").
			Order("cart_count DESC").
			Find(&products)

		if ty.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Category": "Does not exist",
		})
	}
}

func categoryExist(cat int) bool {
	var category model.Category
	if tx := database.DB.Model(&model.Category{}).Where("id=?", uint(cat)).First(&category); tx.Error != nil {
		return false
	}
	return true
}
