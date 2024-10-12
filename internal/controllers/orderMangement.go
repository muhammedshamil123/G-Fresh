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

	"github.com/gin-gonic/gin"
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
	order.PaymentMethod = methods[method]
	order.PaymentStatus = "PENDING"

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
		}
		o_id = 0
		c.JSON(http.StatusOK, gin.H{
			"message": "Order Created!",
			"order":   order,
		})
		return

	}
	database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
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
	// Capture the Razorpay Payment ID and other details from the frontend
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
	// Get the Razorpay secret key from environment variables
	secret := "XEPMrjfiphZjlQHlmlxmgWy6"

	// Verify payment signature using HMAC-SHA256
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
	// Payment verified
	c.JSON(http.StatusOK, gin.H{"status": "Payment verified successfully"})
}

func verifySignature(orderID, paymentID, signature, secret string) bool {
	// Concatenate the Razorpay Order ID and Payment ID
	data := orderID + "|" + paymentID

	// Create a new HMAC by defining the hash type and the secret key
	h := hmac.New(sha256.New, []byte(secret))

	// Write the data to the HMAC
	h.Write([]byte(data))

	// Get the computed HMAC in hex format
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare the computed HMAC with the provided Razorpay signature
	return expectedSignature == signature
}
