package controllers

import (
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GetCategoryList(c *gin.Context) {
	var categorylist []model.CategoryResponse
	tx := database.DB.Model(&model.Category{}).Select("id, name, description, image_url").Find(&categorylist)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	if len(categorylist) < 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Category doesn't exist",
		})
		return
	}
	for _, val := range categorylist {
		c.JSON(http.StatusOK, gin.H{
			"category": val,
		})
	}
	return
}

func AddCategory(c *gin.Context) {
	var category model.AddCategoryList
	if err := c.BindJSON(&category); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process the incoming request",
		})
		return
	}
	Validate := validator.New()

	err := Validate.Struct(category)
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

	var existingCat model.Category
	if tx := database.DB.Where("name = ?", category.Name).First(&existingCat); tx.Error == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Category with same name exists!",
		})
		return
	}

	cat := model.Category{
		Name:        category.Name,
		Description: category.Description,
		ImageURL:    category.ImageURL,
		Products:    nil,
	}
	result := database.DB.Create(&cat)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Error creating Category",
			"error":   result.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Category created Created!!",
		"Name":    cat.Name,
		"Id":      cat.ID,
	})
}
func DeleteCategory(c *gin.Context) {

	catid := c.Query("id")
	var category model.Category
	if tx := database.DB.Where("id = ?", catid).First(&category); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Category id does not exists!",
		})
		return
	}
	tx := database.DB.Where("id = ?", catid).Delete(&model.Category{})
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Deletion failed!",
			"error":   tx.Error,
		})
		return
	}
	if tx := database.DB.Model(&model.Product{}).Where("category_id", catid).Update("category_id", nil); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product category updation failed!",
			"error":   tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted",
	})

}

func EditCategory(c *gin.Context) {

	catid := c.Query("id")
	var form model.AddCategoryList
	var category model.Category

	if tx := database.DB.Where("id = ?", catid).First(&category); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Category id does not exists!",
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
	category.Name = form.Name
	category.Description = form.Description
	category.ImageURL = form.ImageURL

	tx := database.DB.Save(&category)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Updation failed!",
			"error":   tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Category updated",
		"category": category.Name,
	})

}
