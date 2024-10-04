package controllers

import (
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddOrder(c *gin.Context) {

	// database.DB.AutoMigrate(&model.Order{})

	var order model.Order
	var userId uint
	var carts []model.CartItems
	var address model.Address

	addressid := c.Query("aid")
	user, exist := c.Get("email")

	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	if tx := database.DB.Model(&model.User{}).Select("id").Where("email = ?", user).First(&userId); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}

	if tx := database.DB.Model(&model.CartItems{}).Where("user_id=?", userId).Find(&carts); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cart items does not exists!",
		})
		return
	}
	if len(carts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cart empty!",
		})
		return
	}

	if tx := database.DB.Model(&model.Address{}).Where("address_id=? AND user_id=?", addressid, userId).First(&address); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Address does not exists!",
		})
		return
	}

	order.UserID = userId
	order.ShippingAddress = model.ShippingAddress{
		PhoneNumber:  address.PhoneNumber,
		StreetName:   address.StreetName,
		StreetNumber: address.StreetNumber,
		City:         address.City,
		State:        address.State,
		PinCode:      address.PostalCode,
	}
	price := 0
	for _, val := range carts {
		order.ItemCount++
		if tx := database.DB.Model(&model.Product{}).Select("price").Where("id=?", val.ProductID).First(&price); tx.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Address does not exists!",
			})
			return
		}
		order.TotalAmount += float64(price * int(val.Quantity))
	}
	order.PaymentMethod = "COD"
	order.PaymentStatus = "PENDING"

	if tx := database.DB.Model(&model.Order{}).Create(&order); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating order!",
		})
		return
	}
	if placeOrder(order, carts) {
		database.DB.Model(&model.CartItems{}).Where("user_id=?", order.UserID).Delete(&model.CartItems{})
		c.JSON(http.StatusOK, gin.H{
			"message": "Order Created!",
			"order":   order,
		})
		return
	}
	database.DB.Model(&model.Order{}).Delete(&order)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Error placing order order!",
	})
	return
}

func placeOrder(order model.Order, cart []model.CartItems) bool {
	var orderitems []model.OrderItem
	for _, val := range cart {
		price := 0
		database.DB.Model(&model.Product{}).Select("price").Where("id=?", val.ProductID).First(&price)
		orderitem := model.OrderItem{
			OrderID:     order.OrderID,
			UserID:      order.UserID,
			ProductID:   val.ProductID,
			Quantity:    val.Quantity,
			Amount:      float64(price * int(val.Quantity)),
			OrderStatus: "PLACED",
		}
		if tx := database.DB.Model(&model.OrderItem{}).Create(&orderitem); tx.Error != nil {
			database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
			return false
		}
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("user_id=?", order.UserID).Find(&orderitems); tx.Error != nil {
		database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
		return false
	}
	for _, val := range orderitems {
		stock := 0
		if tx := database.DB.Model(&model.Product{}).Select("stock_left").Where("id=?", val.ProductID).First(&stock); tx.Error != nil {
			database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
			return false
		}
		stock -= int(val.Quantity)
		if tx := database.DB.Model(&model.Product{}).Where("id = ?", val.ProductID).Update("stock_left", stock); tx.Error != nil {
			database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
			return false
		}
	}
	return true
}

func ShowOrders(c *gin.Context) {
	user, exist := c.Get("email")
	userId := 0
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	if tx := database.DB.Model(&model.User{}).Select("id").Where("email = ?", user).First(&userId); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	var orders []model.OrderResponce
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,payment_method,payment_status,ordered_at").Where("user_id=?", userId).Find(&orders); tx.Error != nil {
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
	var add []model.ShippingAddress
	if tx := database.DB.Model(&model.Order{}).Select("phone_number,street_name,street_number,city,state,pin_code").Where("user_id=?", userId).Find(&add); tx.Error != nil {
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
			"order":         val,
			"order_address": add[i],
			"product":       pro,
		})
	}
	return
}

func CancelOrders(c *gin.Context) {
	user, exist := c.Get("email")
	userId := 0
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	if tx := database.DB.Model(&model.User{}).Select("id").Where("email = ?", user).First(&userId); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	oid := c.Query("orderid")
	pid := c.Query("pid")
	var order model.Order
	if tx := database.DB.Model(&model.Order{}).Where("order_id=? AND user_id=?", oid, userId).First(&order); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order does not exist!",
		})
		return
	}

	var oitem model.OrderItem
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=? AND product_id=?", oid, userId, pid).First(&oitem); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order item does not exist!",
		})
		return
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=? AND product_id=?", oid, userId, pid).Update("order_status", "CANCELED"); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Cancellation failed!",
		})
		return
	}
	order.ItemCount--
	order.TotalAmount -= oitem.Amount

	if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", oid, userId).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount}); tx.Error != nil {
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
