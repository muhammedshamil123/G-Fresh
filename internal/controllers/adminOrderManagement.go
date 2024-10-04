package controllers

import (
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowOrdersAdmin(c *gin.Context) {

	var orders []model.OrderResponce
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,payment_method,payment_status,ordered_at").Find(&orders); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Orders Empty!",
		})
		return
	}

	if len(orders) < 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Orders Empty!",
		})
		return
	}
	var userid []int
	if tx := database.DB.Model(&model.Order{}).Select("user_id").Find(&userid); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Orders Empty!",
		})
		return
	}
	var add []model.ShippingAddress
	if tx := database.DB.Model(&model.Order{}).Select("phone_number,street_name,street_number,city,state,pin_code").Find(&add); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Orders Empty!",
		})
		return
	}

	for i, val := range orders {
		var items []model.OrderItemResponse
		if tx := database.DB.Model(&model.OrderItem{}).Select("product_id,quantity,amount,order_status").Where("order_id=?", val.OrderID).Find(&items); tx.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Order Item Empty!",
			})
			return
		}
		var pro []any

		for _, value := range items {
			var product model.ViewOrderProductList
			if tx := database.DB.Model(&model.Product{}).Select("name,description,image_url,price,offer_amount").Where("id=?", value.ProductID).First(&product); tx.Error != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Item Empty!",
				})
				return
			}
			item := []any{value, product}
			pro = append(pro, item)
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id":       userid[i],
			"order":         val,
			"order_address": add[i],
			"product":       pro,
		})
	}
	return
}

func CancelOrdersAdmin(c *gin.Context) {

	oid := c.Query("orderid")
	pid := c.Query("pid")
	var order model.Order
	if tx := database.DB.Model(&model.Order{}).Where("order_id=?", oid).First(&order); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order does not exist!",
		})
		return
	}

	var oitem model.OrderItem
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND product_id=?", oid, pid).First(&oitem); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order item does not exist!",
		})
		return
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND product_id=?", oid, pid).Update("order_status", "CANCELED"); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Cancellation failed!",
		})
		return
	}
	order.ItemCount--
	order.TotalAmount -= oitem.Amount

	if tx := database.DB.Model(&model.Order{}).Where("order_id = ?", oid).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount}); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Cancellation failed!",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order Cancelled!",
	})
	var product model.Product
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).First(&product); tx.Error != nil {
		fmt.Println("product does not exist")
	}
	product.StockLeft += oitem.Quantity
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).Update("stock_left", product.StockLeft); tx.Error != nil {
		fmt.Println("stock incerment failed")
	}
	return
}

func ChangeStatus(c *gin.Context) {
	oid := c.Query("orderid")
	pid := c.Query("pid")
	var order model.Order
	if tx := database.DB.Model(&model.Order{}).Where("order_id=?", oid).First(&order); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order does not exist!",
		})
		return
	}

	var oitem model.OrderItem
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND product_id=?", oid, pid).First(&oitem); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order item does not exist!",
		})
		return
	}
	status := ""
	switch oitem.OrderStatus {
	case "PLACED":
		status = "CONFIRMED"
	case "CONFIRMED":
		status = "SHIPPED"
	case "SHIPPED":
		status = "OUT FOR DELIVERY"
	case "OUT FOR DELIVERY":
		status = "DELIVERED"
	case "DELIVERED":
		c.JSON(http.StatusNotFound, gin.H{"message": "Order Delivered!"})
		return
	default:
		c.JSON(http.StatusNotFound, gin.H{"message": "Order Cancelled Aready!"})
		return
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND product_id=?", oid, pid).Update("order_status", status); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "status not changed!",
		})
		return
	}
	if status == "DELIVERED" {
		if tx := database.DB.Model(&model.Order{}).Where("order_id=?", oid).Update("payment_status", "PAID"); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "payement status not changed!",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status Changed!",
		"status":  status,
	})
	return
}
