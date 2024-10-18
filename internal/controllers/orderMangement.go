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
	"gorm.io/gorm"
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
	referral := c.Query("referral")
	coupon := c.Query("coupon")
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
	var refferaloffer float64
	var referredby uint
	if referral != "" {
		result := database.DB.Model(&model.User{}).Where("referral_code = ? AND id <> ?", referral, order.UserID).Select("id").First(&referredby)
		fmt.Println("reffered by:", referredby)

		if result.Error != nil {
			fmt.Println("Referral code not found or query error:", result.Error)
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Referral code not found!",
			})
			return
		}
		var referralhistory model.UserReferralHistory
		// database.DB.AutoMigrate(&model.UserReferralHistory{})
		if tx := database.DB.Model(&model.UserReferralHistory{}).Where("user_id=?", order.UserID).First(&referralhistory); tx.Error != nil {
			newreferralhistory := model.UserReferralHistory{
				UserID:       order.UserID,
				ReferralCode: referral,
				ReferredBy:   referredby,
				ReferClaimed: false,
			}
			database.DB.Model(&model.UserReferralHistory{}).Create(&newreferralhistory)
		}
		if referralhistory.ReferClaimed {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Referral already applied",
			})
			return
		}
		refferaloffer = 2

	}
	var couponoffer, couponmin float64
	if coupon != "" {
		var existCoupon model.CouponInventory
		if tx := database.DB.Model(&model.CouponInventory{}).Where("coupon_code=?", coupon).First(&existCoupon); tx.Error != nil {
			if tx.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Coupon not found!",
				})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "error fetching coupon!",
			})
			return
		}
		if existCoupon.Expiry.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Coupon expired!",
			})
			return
		}
		var usage model.CouponUsage
		if tx := database.DB.Model(&model.CouponUsage{}).Where("coupon_code=? AND user_id=?", coupon, order.UserID).First(&usage); tx.Error != nil {
			if tx.Error != gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "error fetching coupon usage!",
				})
				return
			}
		} else {
			if usage.UsageCount >= existCoupon.MaximumUsage {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Coupon usage limit reached!",
				})
				return
			}
		}
		couponmin = existCoupon.MinimumAmount
		couponoffer = float64(existCoupon.Percentage)
	}
	println(couponoffer, couponmin)
	order_total := 0.00
	for _, val := range carts {
		order.ItemCount++
		var offer, price float64
		if tx := database.DB.Model(&model.Product{}).Select("offer_amount").Where("id=?", val.ProductID).First(&offer); tx.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product does not exists!",
			})
			return
		}
		if tx := database.DB.Model(&model.Product{}).Select("price").Where("id=?", val.ProductID).First(&price); tx.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product does not exists!",
			})
			return
		}
		cid, cat_offer := 0, 0
		database.DB.Model(&model.Product{}).Select("category_id").Where("id=?", val.ProductID).First(&cid)
		database.DB.Model(&model.Category{}).Select("offer_percentage").Where("id=?", cid).First(&cat_offer)

		cat_amount := (price * float64(cat_offer)) / 100
		ref_amount := (price * float64(refferaloffer)) / 100
		coupon_amount := (price * couponoffer) / 100
		order_total += price * float64(val.Quantity)
		order.ProductOfferAmount += (cat_amount + (price - offer)) * float64(val.Quantity)
		order.CouponDiscountAmount += (ref_amount + coupon_amount) * float64(val.Quantity)
		order.FinalAmount += price * float64(val.Quantity)
		order.TotalAmount += (offer - cat_amount - ref_amount - coupon_amount) * float64(val.Quantity)
		fmt.Println(val.ProductID, " offer amount:", offer, " category_offer:", cat_amount, "refferal_offer:", ref_amount, "order_total:", order.TotalAmount)
	}
	if order_total < 500 && referral != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Referral code cannot use order below 500!",
		})
		return
	}
	if order_total < float64(couponmin) && coupon != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Coupon code cannot use order below limit!",
			"limit":   couponmin,
		})
		return
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
	if placeOrder(order, carts, refferaloffer, couponoffer) {
		if referral != "" {
			database.DB.Model(&model.UserReferralHistory{}).Where("user_id=?", order.UserID).Update("refer_claimed", true)
		}
		if coupon != "" {
			var usage model.CouponUsage
			if tx := database.DB.Model(&model.CouponUsage{}).Where("coupon_code=? AND user_id=?", coupon, order.UserID).First(&usage); tx.Error != nil {
				if tx.Error == gorm.ErrRecordNotFound {
					usage := model.CouponUsage{
						UserID:     order.UserID,
						CouponCode: coupon,
						UsageCount: 1,
					}
					database.DB.Model(&model.CouponUsage{}).Create(&usage)
				}
			} else {
				usage.UsageCount++
				database.DB.Model(&model.CouponUsage{}).Where("coupon_code=? AND user_id=?", coupon, order.UserID).Update("usage_count", usage.UsageCount)
			}
		}

		if order.PaymentMethod == "Razorpay" {
			o_id = order.OrderID
			c.JSON(http.StatusOK, gin.H{
				"message": "Payement pending!",
				"order":   order,
			})
			database.DB.Model(&model.CartItems{}).Where("user_id=?", order.UserID).Delete(&model.CartItems{})
			return
		} else if order.PaymentMethod == "Wallet" {
			var user model.User
			if tx := database.DB.Model(&model.User{}).Where("id = ?", order.UserID).First(&user); tx.Error != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "User does not exists!",
				})
				return
			}
			// var wallet model.UserWalletHistory
			// if tx := database.DB.Model(&model.UserWalletHistory{}).Where("user_id=?", order.UserID).Last(&wallet); tx.Error != nil {
			// 	database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
			// 	c.JSON(http.StatusInternalServerError, gin.H{
			// 		"message": "Insufficient wallet balance!",
			// 	})
			// 	return
			// }
			if user.WalletAmount < order.TotalAmount {
				database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Delete(&model.Order{})
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Insufficient wallet balance!",
				})
				return
			}
			newBalance := user.WalletAmount - order.TotalAmount
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
			newWallet.Amount = RoundDecimalValue(newWallet.Amount)
			newWallet.CurrentBalance = RoundDecimalValue(newWallet.CurrentBalance)
			database.DB.Model(&model.UserWalletHistory{}).Create(&newWallet)
			database.DB.Model(&model.User{}).Where("id=?", order.UserID).Update("wallet_amount", newWallet.CurrentBalance)
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
			order.CouponDiscountAmount = RoundDecimalValue(order.CouponDiscountAmount)
			order.FinalAmount = RoundDecimalValue(order.FinalAmount)
			order.TotalAmount = RoundDecimalValue(order.TotalAmount)
			order.ProductOfferAmount = RoundDecimalValue(order.ProductOfferAmount)
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).First(&order)

			c.JSON(http.StatusOK, gin.H{
				"message":        "Order Created!",
				"order":          order,
				"wallet balance": newWallet.CurrentBalance,
			})
			database.DB.Model(&model.CartItems{}).Where("user_id=?", order.UserID).Delete(&model.CartItems{})
			return
		} else if order.PaymentMethod == "COD" {
			o_id = 0
			c.JSON(http.StatusOK, gin.H{
				"message": "Order Created!",
				"order":   order,
			})
			database.DB.Model(&model.CartItems{}).Where("user_id=?", order.UserID).Delete(&model.CartItems{})
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

func placeOrder(order model.Order, cart []model.CartItems, referraloffer, couponoffer float64) bool {
	var orderitems []model.OrderItem
	for _, val := range cart {
		var offer, price float64
		database.DB.Model(&model.Product{}).Select("offer_amount").Where("id=?", val.ProductID).First(&offer)

		database.DB.Model(&model.Product{}).Select("price").Where("id=?", val.ProductID).First(&price)
		cid, cat_offer := 0, 0.0
		database.DB.Model(&model.Product{}).Select("category_id").Where("id=?", val.ProductID).First(&cid)
		database.DB.Model(&model.Category{}).Select("offer_percentage").Where("id=?", cid).First(&cat_offer)
		cat_amount := (price * cat_offer) / 100
		ref_amount := (price * referraloffer) / 100
		coupon_amount := (price * float64(couponoffer)) / 100
		orderitem := model.OrderItem{
			OrderID:     order.OrderID,
			UserID:      order.UserID,
			ProductID:   val.ProductID,
			Quantity:    val.Quantity,
			Amount:      (offer - cat_amount - ref_amount - coupon_amount) * float64(val.Quantity),
			OrderStatus: "PLACED",
		}
		fmt.Println(orderitem.ProductID, " offer amount:", offer, " category_offer:", cat_amount, "refferal_offer:", ref_amount, "order_total:", orderitem.Amount)
		orderitem.Amount = RoundDecimalValue(orderitem.Amount)
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
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,final_amount,payment_method,payment_status,ordered_at,order_status,coupon_discount_amount,product_offer_amount").Where("user_id=?", userId).Order("ordered_at DESC").Find(&orders); tx.Error != nil {
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
			database.DB.Model(&model.User{}).Where("id=?", order.UserID).Update("wallet_amount", newWallet.CurrentBalance)
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
		database.DB.Model(&model.User{}).Where("id=?", order.UserID).Update("wallet_amount", newWallet.CurrentBalance)
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
		deleteOrder(int(o_id))
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
		deleteOrder(int(o_id))
		return
	}
	fmt.Println(paymentInfo)
	secret := "XEPMrjfiphZjlQHlmlxmgWy6"
	if !verifySignature(paymentInfo.OrderID, paymentInfo.PaymentID, paymentInfo.Signature, secret) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment signature"})
		deleteOrder(int(o_id))
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

func deleteOrder(order_id int) {
	var order []model.OrderItem
	database.DB.Model(&model.OrderItem{}).Where("order_id=?", order_id).Find(&order)
	for _, val := range order {
		var product model.Product
		database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).First(&product)
		product.StockLeft += val.Quantity
		database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).Update("stock_left", product.StockLeft)
	}
	database.DB.Model(&model.Order{}).Where("order_id=?", order_id).Delete(&model.Order{})
	database.DB.Model(&model.OrderItem{}).Where("order_id=?", order_id).Delete(&model.OrderItem{})
}
