package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
)

var o_id uint

func AddOrder(c *gin.Context) {

	// database.DB.AutoMigrate(&model.Order{})

	var order model.Order
	var userId uint
	var carts []model.CartItems
	var address model.Address
	methods := map[string]string{
		"1": "COD",
		"2": "Razorpay",
		"3": "Wallet",
	}

	addressid := c.Query("aid")
	method := c.Query("method")
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

	for _, val := range carts {
		order.ItemCount++
		var offer float64
		if tx := database.DB.Model(&model.Product{}).Select("offer_amount").Where("id=?", val.ProductID).First(&offer); tx.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product does not exists!",
			})
			return
		}
		cid, cat_offer := 0, 0
		database.DB.Model(&model.Product{}).Select("category_id").Where("id=?", val.ProductID).First(&cid)
		database.DB.Model(&model.Category{}).Select("offer_percentage").Where("id=?", cid).First(&cat_offer)
		cat_amount := (offer * float64(cat_offer)) / 100
		order.TotalAmount += float64(int(offer-cat_amount) * int(val.Quantity))
	}
	order.PaymentMethod = methods[method]
	order.PaymentStatus = "PENDING"
	order.OrderStatus = "PLACED"
	if order.PaymentMethod == "COD" && order.TotalAmount > 1000 {
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Cash on delivery not available for order greater than 1000/-",
		})
		return
	}
	if tx := database.DB.Model(&model.Order{}).Create(&order); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating order!",
		})
		return
	}
	if placeOrder(order, carts) {

		database.DB.Model(&model.CartItems{}).Where("user_id=?", order.UserID).Delete(&model.CartItems{})
		if order.PaymentMethod == "Razorpay" {
			o_id = order.OrderID
			c.JSON(http.StatusOK, gin.H{
				"message": "Payement pending!",
			})
			return
		} else if order.PaymentMethod == "Wallet" {
			var wallet model.UserWalletHistory
			if tx := database.DB.Model(&model.UserWalletHistory{}).Where("user_id=?", order.UserID).Last(&wallet); tx.Error != nil {
				database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Insufficient wallet balance!",
				})
				return
			}
			if wallet.CurrentBalance < order.TotalAmount {
				database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Insufficient wallet balance!",
				})
				return
			}
			newBalance := wallet.CurrentBalance - order.TotalAmount
			newWallet := model.UserWalletHistory{
				TransactionTime: time.Now(),
				WalletPaymentID: uuid.New().String(),
				UserID:          order.UserID,
				Type:            "Outgoing",
				Amount:          order.TotalAmount,
				CurrentBalance:  newBalance,
				OrderID:         strconv.Itoa(int(order.OrderID)),
				Reason:          "Order Payement",
			}
			database.DB.Model(&model.UserWalletHistory{}).Create(&newWallet)
			o_id = 0
			payement := model.Payment{
				OrderID:           strconv.Itoa(int(order.OrderID)),
				WalletPaymentID:   newWallet.WalletPaymentID,
				RazorpayOrderID:   "",
				RazorpayPaymentID: "",
				RazorpaySignature: "",
				PaymentGateway:    "Wallet",
				PaymentStatus:     "PAID",
			}
			database.DB.Model(&model.Payment{}).Create(&payement)
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "PAID")
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).First(&order)
			c.JSON(http.StatusOK, gin.H{
				"message":        "Order Created!",
				"order":          order,
				"wallet balance": newWallet.CurrentBalance,
			})
			return
		} else if order.PaymentMethod == "COD" {
			o_id = 0
			c.JSON(http.StatusOK, gin.H{
				"message": "Order Created!",
				"order":   order,
			})
			return
		} else {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Invalid payement method!",
			})
			return
		}
	}
	database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Error placing order!",
	})
}

func placeOrder(order model.Order, cart []model.CartItems) bool {
	var orderitems []model.OrderItem
	for _, val := range cart {
		var offer float64
		database.DB.Model(&model.Product{}).Select("offer_amount").Where("id=?", val.ProductID).First(&offer)
		cid, cat_offer := 0, 0
		database.DB.Model(&model.Product{}).Select("category_id").Where("id=?", val.ProductID).First(&cid)
		database.DB.Model(&model.Category{}).Select("offer_percentage").Where("id=?", cid).First(&cat_offer)
		cat_amount := (offer * float64(cat_offer)) / 100
		orderitem := model.OrderItem{
			OrderID:     order.OrderID,
			UserID:      order.UserID,
			ProductID:   val.ProductID,
			Quantity:    val.Quantity,
			Amount:      float64((int(offer-cat_amount) * int(val.Quantity))),
			OrderStatus: "PLACED",
		}
		if tx := database.DB.Model(&model.OrderItem{}).Create(&orderitem); tx.Error != nil {
			database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
			return false
		}
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("user_id=? AND order_id=?", order.UserID, order.OrderID).Find(&orderitems); tx.Error != nil {
		database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
		return false
	}
	fmt.Println(orderitems)
	for _, val := range orderitems {
		var stock model.Product
		if tx := database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).First(&stock); tx.Error != nil {
			database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
			return false
		}
		if stock.StockLeft < val.Quantity {
			fmt.Println(stock.StockLeft)
			database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Delete(&model.OrderItem{})
			return false
		}
		stock.StockLeft = stock.StockLeft - val.Quantity
		if tx := database.DB.Model(&model.Product{}).Where("id = ?", val.ProductID).Update("stock_left", stock.StockLeft); tx.Error != nil {
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
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,payment_method,payment_status,ordered_at,order_status").Where("user_id=?", userId).Find(&orders); tx.Error != nil {
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
	if pid == "" {
		if order.OrderStatus == "CANCELED" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Order cancelled already!",
			})
			return
		}
		database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", oid, userId).Update("order_status", "CANCELED")
		if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, userId).Update("order_status", "CANCELED"); tx.Error != nil {
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
			}
			database.DB.Model(&model.Payment{}).Create(&payement)
			c.JSON(http.StatusOK, gin.H{
				"message":        "Order Cancelled!",
				"Refund":         "Success",
				"wallet balance": newWallet.CurrentBalance,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Order Cancelled!",
			})
		}
		var oitem []model.OrderItem
		if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, userId).Find(&oitem); tx.Error != nil {
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
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=? AND product_id=?", oid, userId, pid).First(&oitem); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Order item does not exist!",
		})
		return
	}
	if oitem.OrderStatus == "CANCELED" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Item cancelled already!",
		})
		return
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=? AND product_id=?", oid, userId, pid).Update("order_status", "CANCELED"); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Cancellation failed!",
		})
		return
	}
	order.ItemCount -= 1
	order.TotalAmount -= oitem.Amount
	var orderitems []model.OrderItem
	flag := true
	database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, userId).Find(&orderitems)
	for _, val := range orderitems {
		if val.OrderStatus != "CANCELED" {
			flag = false
		}
	}
	if flag {
		order.OrderStatus = "CANCELED"
	}
	if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", oid, userId).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount, "order_status": order.OrderStatus}); tx.Error != nil {
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
		payement := model.Payment{
			OrderID:           strconv.Itoa(int(order.OrderID)),
			WalletPaymentID:   newWallet.WalletPaymentID,
			RazorpayOrderID:   "",
			RazorpayPaymentID: "",
			RazorpaySignature: "",
			PaymentGateway:    "Wallet",
			PaymentStatus:     "REFUND",
		}
		database.DB.Model(&model.Payment{}).Create(&payement)
		c.JSON(http.StatusOK, gin.H{
			"message":        "Order Cancelled!",
			"Refund":         "Success",
			"wallet balance": newWallet.CurrentBalance,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Order Cancelled!",
		})
	}
	var product model.Product
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).First(&product); tx.Error != nil {
		fmt.Println("product does not exist")
	}
	product.StockLeft += oitem.Quantity
	if tx := database.DB.Model(&model.Product{}).Where("id=?", pid).Update("stock_left", product.StockLeft); tx.Error != nil {
		fmt.Println("stock incerment failed")
	}

}

func RenderRazorpay(c *gin.Context) {
	c.HTML(http.StatusOK, "razorpay.html", nil)
}

func CreateOrder(c *gin.Context) {
	client := razorpay.NewClient("rzp_test_Mg8qA7Z2ycbKOB", "XEPMrjfiphZjlQHlmlxmgWy6")

	var order model.Order
	database.DB.Model(&model.Order{}).Where("order_id=?", o_id).First(&order)
	amount := order.TotalAmount * 100
	razorpayOrder, err := client.Order.Create(map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  "order_rcptid_11",
	}, nil)

	if err != nil {
		fmt.Println("Error creating Razorpay order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating order"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"order_id": razorpayOrder["id"],
		"amount":   amount,
		"currency": "INR",
	})
	fmt.Println(razorpayOrder)
}

func VerifyPayment(c *gin.Context) {
	orderid := strconv.Itoa(int(o_id))
	fmt.Println(orderid)
	var paymentInfo struct {
		PaymentID string `json:"razorpay_payment_id"`
		OrderID   string `json:"razorpay_order_id"`
		Signature string `json:"razorpay_signature"`
	}

	if err := c.BindJSON(&paymentInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment information"})
		return
	}
	fmt.Println(paymentInfo)
	secret := "XEPMrjfiphZjlQHlmlxmgWy6"
	if !verifySignature(paymentInfo.OrderID, paymentInfo.PaymentID, paymentInfo.Signature, secret) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment signature"})
		return
	}
	// database.DB.AutoMigrate(&model.Payment{})
	payement := model.Payment{
		OrderID:           orderid,
		WalletPaymentID:   "",
		RazorpayOrderID:   paymentInfo.OrderID,
		RazorpayPaymentID: paymentInfo.PaymentID,
		RazorpaySignature: paymentInfo.Signature,
		PaymentGateway:    "Razorpay",
		PaymentStatus:     "PAID",
	}
	fmt.Println(payement)
	database.DB.Model(&model.Payment{}).Create(&payement)
	database.DB.Model(&model.Order{}).Where("order_id=?", orderid).Update("payment_status", payement.PaymentStatus)
	o_id = 0
	c.JSON(http.StatusOK, gin.H{"status": "Payment verified successfully"})
}

func verifySignature(orderID, paymentID, signature, secret string) bool {

	data := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return expectedSignature == signature
}

func OrderReturn(c *gin.Context) {
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
	if pid == "" {
		if order.OrderStatus != "DELIVERED" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Order not delivered!",
			})
			return
		}
		if order.OrderStatus == "RETURNED" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Order returned already!",
			})
			return
		}
		if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=?", oid, userId).Update("order_status", "RETURN REQUEST"); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Return failed!",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Return Requested!",
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
	if oitem.OrderStatus == "RETURNED" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Item returned already!",
		})
		return
	}
	if oitem.OrderStatus != "DELIVERED" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Item not delivered!",
		})
		return
	}
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id=? AND user_id=? AND product_id=?", oid, userId, pid).Update("order_status", "RETURN REQUEST"); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Return failed!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Return Requested!",
	})
}
