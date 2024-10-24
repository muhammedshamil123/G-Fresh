package controllers

import (
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ShowOrdersAdmin(c *gin.Context) {

	var orders []model.OrderResponce
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,final_amount,payment_method,payment_status,ordered_at,order_status,coupon_discount_amount,product_offer_amount").Order("ordered_at DESC").Find(&orders); tx.Error != nil {
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
	if tx := database.DB.Model(&model.Order{}).Select("user_id").Order("ordered_at DESC").Find(&userid); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Orders Empty!",
		})
		return
	}
	var add []model.ShippingAddress
	if tx := database.DB.Model(&model.Order{}).Select("phone_number,street_name,street_number,city,state,pin_code").Order("ordered_at DESC").Find(&add); tx.Error != nil {
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
	if pid == "" {
		if order.OrderStatus == "CANCELED" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Order cancelled already!",
			})
			return
		}
		database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", oid, order.UserID).Update("order_status", "CANCELED")
		if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, order.UserID).Update("order_status", "CANCELED"); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Cancellation failed!",
			})
			return
		}
		if order.PaymentStatus == "PAID" {
			var wallet model.UserWalletHistory
			var cBallance float64
			if tx := database.DB.Model(&model.UserWalletHistory{}).Where("user_id=?", order.UserID).Last(&wallet); tx.Error == nil {
				cBallance = wallet.CurrentBalance
				fmt.Println("no error", cBallance)
			} else {
				fmt.Println(tx.Error)
			}
			newBalance := cBallance + order.TotalAmount
			fmt.Println(newBalance, cBallance, order)
			newWallet := model.UserWalletHistory{
				TransactionTime: time.Now(),
				WalletPaymentID: uuid.New().String(),
				UserID:          order.UserID,
				Type:            "Incoming",
				Amount:          order.TotalAmount,
				CurrentBalance:  newBalance,
				OrderID:         strconv.Itoa(int(order.OrderID)),
				Reason:          "Order Cancel",
			}
			database.DB.Model(&model.UserWalletHistory{}).Create(&newWallet)
			payement := model.Payment{
				OrderID:           strconv.Itoa(int(order.OrderID)),
				WalletPaymentID:   newWallet.WalletPaymentID,
				RazorpayOrderID:   "",
				RazorpayPaymentID: "",
				RazorpaySignature: "",
				PaymentGateway:    "Wallet",
				PaymentStatus:     "REFUND",
				Amount:            order.TotalAmount,
			}
			database.DB.Model(&model.Payment{}).Create(&payement)
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "REFUND")
			c.JSON(http.StatusOK, gin.H{
				"message":        "Order Cancelled!",
				"Refund":         "Success",
				"wallet balance": newWallet.CurrentBalance,
			})
		} else {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "CANCELED")
			c.JSON(http.StatusOK, gin.H{
				"message": "Order Cancelled!",
			})
		}
		order.ItemCount = 0
		order.TotalAmount = 0
		if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", order.OrderID, order.UserID).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount}); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Cancel failed!",
			})
			return
		}
		var oitem []model.OrderItem
		if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, order.UserID).Find(&oitem); tx.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Order item does not exist!",
			})
			return
		}
		for _, val := range oitem {
			var product model.Product
			if tx := database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).First(&product); tx.Error != nil {
				fmt.Println("product does not exist")
			}
			product.StockLeft += val.Quantity
			if tx := database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).Update("stock_left", product.StockLeft); tx.Error != nil {
				fmt.Println("stock incerment failed")
			}
		}
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

	var orderitems []model.OrderItem
	flag := true
	database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, order.UserID).Find(&orderitems)
	for _, val := range orderitems {
		if val.OrderStatus != "CANCELED" {
			flag = false
		}
	}
	if flag {
		order.OrderStatus = "CANCELED"
	}

	var product model.Product
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).First(&product); tx.Error != nil {
		fmt.Println("product does not exist")
	}
	product.StockLeft += oitem.Quantity
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).Update("stock_left", product.StockLeft); tx.Error != nil {
		fmt.Println("stock incerment failed")
	}
	if order.PaymentStatus == "PAID" {
		var wallet model.UserWalletHistory
		var cBallance float64
		if tx := database.DB.Model(&model.UserWalletHistory{}).Where("user_id=?", order.UserID).Last(&wallet); tx.Error == nil {
			cBallance = wallet.CurrentBalance
			fmt.Println("no error", cBallance)
		} else {
			fmt.Println(tx.Error)
		}
		newBalance := cBallance + oitem.Amount
		fmt.Println(newBalance, cBallance, order)
		newWallet := model.UserWalletHistory{
			TransactionTime: time.Now(),
			WalletPaymentID: uuid.New().String(),
			UserID:          order.UserID,
			Type:            "Incoming",
			Amount:          oitem.Amount,
			CurrentBalance:  newBalance,
			OrderID:         strconv.Itoa(int(order.OrderID)),
			Reason:          "Order Cancel",
		}

		database.DB.Model(&model.UserWalletHistory{}).Create(&newWallet)
		database.DB.Model(&model.User{}).Where("id=?", order.UserID).Update("wallet_amount", newWallet.CurrentBalance)
		payement := model.Payment{
			OrderID:           strconv.Itoa(int(order.OrderID)),
			WalletPaymentID:   newWallet.WalletPaymentID,
			RazorpayOrderID:   "",
			RazorpayPaymentID: "",
			RazorpaySignature: "",
			PaymentGateway:    "Wallet",
			PaymentStatus:     "REFUND",
			Amount:            oitem.Amount,
		}
		database.DB.Model(&model.Payment{}).Create(&payement)
		var orders []model.OrderItem
		database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Find(&orders)
		i := 0
		for i = 0; i < len(orders); i++ {
			if orders[i].OrderStatus != "CANCELED" && orders[i].OrderStatus != "RETURNED" {
				break
			}
		}
		if i == len(orders) {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "REFUND")
		}

		order.ItemCount--
		order.TotalAmount -= oitem.Amount

		if tx := database.DB.Model(&model.Order{}).Where("order_id = ?", oid).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount, "order_status": order.OrderStatus}); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Cancellation failed!",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":        "Order Cancelled!",
			"Refund":         "Success",
			"wallet balance": newWallet.CurrentBalance,
		})
		return
	}

	database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "CANCELED")
	c.JSON(http.StatusOK, gin.H{
		"message": "Order Cancelled!",
	})
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
	case "CANCELED":
		c.JSON(http.StatusNotFound, gin.H{"message": "Order Cancelled Aready!"})
		return
	case "RETURN REQUEST":
		status = "RETURNED"
	case "RETURNED":
		c.JSON(http.StatusNotFound, gin.H{"message": "Order Returned Aready!"})
		return
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND product_id=?", oid, pid).Update("order_status", status); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "status not changed!",
		})
		return
	}

	if status == "DELIVERED" && order.PaymentStatus != "PAID" {

		payement := model.Payment{
			OrderID:           strconv.Itoa(int(order.OrderID)),
			WalletPaymentID:   "",
			RazorpayOrderID:   "",
			RazorpayPaymentID: "",
			RazorpaySignature: "",
			PaymentGateway:    "COD",
			PaymentStatus:     "PAID",
			Amount:            order.TotalAmount,
		}
		database.DB.Model(&model.Payment{}).Create(&payement)
		database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "PAID")
		c.JSON(http.StatusOK, gin.H{
			"message":         "Order Delivered!",
			"Payement Status": "PAID",
		})
		return
	}
	if status == "DELIVERED" {
		var orderitems []model.OrderItem
		flag := true
		database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, order.UserID).Find(&orderitems)
		for _, val := range orderitems {
			if val.OrderStatus != "DELIVERED" {
				flag = false
			}
		}
		if flag {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("order_status", "DELIVERED")
			if tx := database.DB.Model(&model.Order{}).Where("order_id=?", oid).Update("payment_status", "PAID"); tx.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "payement status not changed!",
				})
				return
			}
		}
	}

	if status == "RETURNED" && order.PaymentStatus == "PAID" {
		order.ItemCount -= 1
		order.TotalAmount -= oitem.Amount

		if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", order.OrderID, order.UserID).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount}); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Return failed!",
			})
			return
		}
		var wallet model.UserWalletHistory
		var cBallance float64
		if tx := database.DB.Model(&model.UserWalletHistory{}).Where("user_id=?", order.UserID).Last(&wallet); tx.Error == nil {
			cBallance = wallet.CurrentBalance
			fmt.Println("no error", cBallance)
		} else {
			fmt.Println(tx.Error)
		}
		newBalance := cBallance + oitem.Amount
		fmt.Println(newBalance, cBallance, order)
		newWallet := model.UserWalletHistory{
			TransactionTime: time.Now(),
			WalletPaymentID: uuid.New().String(),
			UserID:          order.UserID,
			Type:            "Incoming",
			Amount:          oitem.Amount,
			CurrentBalance:  newBalance,
			OrderID:         strconv.Itoa(int(order.OrderID)),
			Reason:          "Order Return",
		}
		database.DB.Model(&model.UserWalletHistory{}).Create(&newWallet)
		database.DB.Model(&model.User{}).Where("id=?", order.UserID).Update("wallet_amount", newWallet.CurrentBalance)
		payement := model.Payment{
			OrderID:           strconv.Itoa(int(order.OrderID)),
			WalletPaymentID:   newWallet.WalletPaymentID,
			RazorpayOrderID:   "",
			RazorpayPaymentID: "",
			RazorpaySignature: "",
			PaymentGateway:    "Wallet",
			PaymentStatus:     "REFUND",
			Amount:            oitem.Amount,
		}
		database.DB.Model(&model.Payment{}).Create(&payement)
		var orders []model.OrderItem
		database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Find(&orders)
		i := 0
		for i = 0; i < len(orders); i++ {
			if orders[i].OrderStatus != "CANCELED" && orders[i].OrderStatus != "RETURNED" {
				break
			}
		}
		if i == len(orders) {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "REFUND")
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Order Returned!",
			"Refund":  "Success",
		})
		var product model.Product
		if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).First(&product); tx.Error != nil {
			fmt.Println("product does not exist")
		}
		product.StockLeft += oitem.Quantity
		if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).Update("stock_left", product.StockLeft); tx.Error != nil {
			fmt.Println("stock incerment failed")
		}

	}
	if status == "RETURNED" {
		var orderitems []model.OrderItem
		flag := true
		database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, order.UserID).Find(&orderitems)
		for _, val := range orderitems {
			if val.OrderStatus != "RETURNED" {
				flag = false
			}
		}
		if flag {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("order_status", "RETURNED")
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Status Changed!",
		"status":  status,
	})

}

func ReturnRequests(c *gin.Context) {
	var orders []model.OrderResponce
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,final_amount,payment_method,payment_status,ordered_at,coupon_discount_amount,product_offer_amount").Find(&orders); tx.Error != nil {
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
		if tx := database.DB.Model(&model.OrderItem{}).Where("order_status=?", "RETURN REQUEST").Select("product_id,quantity,amount,order_status").Where("order_id=?", val.OrderID).Find(&items); tx.Error != nil {
			continue
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
		if len(pro) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"user_id":       userid[i],
				"order":         val,
				"order_address": add[i],
				"products":      pro,
			})
		}
	}
}
