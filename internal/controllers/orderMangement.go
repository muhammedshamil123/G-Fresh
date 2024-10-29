package controllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"g-fresh/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

var o_id uint

var FAILED map[uint]int

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
	delivery_charge := utils.CalculateDistance(address.PostalCode)
	fmt.Println(delivery_charge)
	order.UserID = userId
	order.ShippingAddress = model.ShippingAddress{
		PhoneNumber:  address.PhoneNumber,
		StreetName:   address.StreetName,
		StreetNumber: address.StreetNumber,
		City:         address.City,
		State:        address.State,
		PinCode:      address.PostalCode,
	}
	var refferaloffer int
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
	var couponoffer, couponmin, couponmax int
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
		couponmin = int(existCoupon.MinimumAmount)
		couponmax = int(existCoupon.MaximumAmount)
		couponoffer = int(existCoupon.Percentage)
	}
	println("hi", couponoffer, couponmin, couponmax)
	order_total := 0
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
		coupon_amount := (price * float64(couponoffer)) / 100
		order_total += int(price * float64(val.Quantity))
		order.ProductOfferAmount += (cat_amount + (price - offer)) * float64(val.Quantity)
		order.CouponDiscountAmount += (ref_amount + coupon_amount) * float64(val.Quantity)
		order.FinalAmount += price * float64(val.Quantity)
		order.TotalAmount += float64(int((offer - cat_amount - ref_amount - coupon_amount) * float64(val.Quantity)))
		fmt.Println(val.ProductID, " offer amount:", offer, " category_offer:", cat_amount, "refferal_offer:", ref_amount, "order_total:", order.TotalAmount)
	}
	if order_total < 500 && referral != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Referral code cannot use order below 500!",
		})
		return
	}
	if order_total < couponmin && coupon != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Coupon code cannot use order below limit!",
			"limit":   couponmin,
		})
		return
	}
	if order_total > couponmax && coupon != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Coupon code cannot use order above limit!",
			"limit":   couponmax,
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
	order.DeliveryCharge = uint(RoundDecimalValue(float64(delivery_charge)))
	order.CouponDiscountAmount = RoundDecimalValue(order.CouponDiscountAmount)
	order.FinalAmount = RoundDecimalValue(order.FinalAmount)
	order.TotalAmount = RoundDecimalValue(order.TotalAmount)
	order.ProductOfferAmount = RoundDecimalValue(order.ProductOfferAmount)
	if tx := database.DB.Model(&model.Order{}).Create(&order); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating order!",
		})
		return
	}
	var orders model.OrderResponce
	database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).First(&orders)
	order.TotalAmount = order.TotalAmount + float64(order.DeliveryCharge)
	orders.TotalAmount = orders.TotalAmount + float64(orders.DeliveryCharge)
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
				"order":   orders,
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
				Amount:            order.TotalAmount,
			}
			database.DB.Model(&model.Payment{}).Create(&payement)
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "PAID")
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).First(&order)

			c.JSON(http.StatusOK, gin.H{
				"message":        "Order Created!",
				"order":          orders,
				"wallet balance": newWallet.CurrentBalance,
			})
			database.DB.Model(&model.CartItems{}).Where("user_id=?", order.UserID).Delete(&model.CartItems{})
			return
		} else if order.PaymentMethod == "COD" {
			o_id = 0
			c.JSON(http.StatusOK, gin.H{
				"message": "Order Created!",
				"order":   orders,
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

func placeOrder(order model.Order, cart []model.CartItems, referraloffer, couponoffer int) bool {
	var orderitems []model.OrderItem
	for _, val := range cart {
		var offer, price float64
		database.DB.Model(&model.Product{}).Select("offer_amount").Where("id=?", val.ProductID).First(&offer)

		database.DB.Model(&model.Product{}).Select("price").Where("id=?", val.ProductID).First(&price)
		cid, cat_offer := 0, 0
		database.DB.Model(&model.Product{}).Select("category_id").Where("id=?", val.ProductID).First(&cid)
		database.DB.Model(&model.Category{}).Select("offer_percentage").Where("id=?", cid).First(&cat_offer)
		cat_amount := (price * float64(cat_offer)) / 100
		ref_amount := (price * float64(referraloffer)) / 100
		coupon_amount := (price * float64(couponoffer)) / 100
		orderitem := model.OrderItem{
			OrderID:     order.OrderID,
			UserID:      order.UserID,
			ProductID:   val.ProductID,
			Quantity:    val.Quantity,
			Amount:      float64(int((offer - cat_amount - ref_amount - coupon_amount) * float64(val.Quantity))),
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
	if tx := database.DB.Model(&model.Order{}).Select("order_id,item_count,total_amount,final_amount,payment_method,payment_status,ordered_at,order_status,coupon_discount_amount,product_offer_amount,delivery_charge").Where("user_id=?", userId).Order("ordered_at DESC").Find(&orders); tx.Error != nil {
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
		val.TotalAmount += float64(val.DeliveryCharge)
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
				Amount:            order.TotalAmount,
			}
			database.DB.Model(&model.Payment{}).Create(&payement)
			order.ItemCount = 0
			order.TotalAmount = 0
			if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", order.OrderID, order.UserID).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount}); tx.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Cancel failed!",
				})
				return
			}

			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "REFUND")
			c.JSON(http.StatusOK, gin.H{
				"message":        "Order Cancelled!",
				"Refund":         "Success",
				"wallet balance": newWallet.CurrentBalance,
			})
		} else {
			order.ItemCount = 0
			order.TotalAmount = 0
			if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", order.OrderID, order.UserID).Updates(map[string]interface{}{"item_count": order.ItemCount, "total_amount": order.TotalAmount}); tx.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Cancel failed!",
				})
				return
			}
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
			Amount:            oitem.Amount,
		}
		database.DB.Model(&model.Payment{}).Create(&payement)
		var orders []model.OrderItem
		database.DB.Model(&model.OrderItem{}).Where("order_id=?", order.OrderID).Find(&orders)
		i := 0
		for i = 0; i < len(orders); i++ {
			if (orders[i].OrderStatus != "CANCELED") && (orders[i].OrderStatus != "RETURNED") {
				break
			}
		}
		if i == len(orders) {
			database.DB.Model(&model.Order{}).Where("order_id=?", order.OrderID).Update("payment_status", "REFUND")
		}
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

var orderID uint

func RenderRazorpay(c *gin.Context) {
	orderID = o_id
	c.HTML(http.StatusOK, "razorpay.html", nil)
}
func CreateOrder(c *gin.Context) {
	client := razorpay.NewClient("rzp_test_Mg8qA7Z2ycbKOB", "XEPMrjfiphZjlQHlmlxmgWy6")

	var order model.Order
	database.DB.Model(&model.Order{}).Where("order_id=?", orderID).First(&order)
	order.TotalAmount = order.TotalAmount + float64(order.DeliveryCharge)
	amount := RoundDecimalValue(order.TotalAmount * 100)
	println(amount, order.TotalAmount)
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
}

func VerifyPayment(c *gin.Context) {
	orderid := strconv.Itoa(int(orderID))
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
	amount := 0.0
	database.DB.Model(&model.Order{}).Where("order_id =?", orderID).Select("total_amount").First(&amount)
	payement := model.Payment{
		OrderID:           orderid,
		WalletPaymentID:   "",
		RazorpayOrderID:   paymentInfo.OrderID,
		RazorpayPaymentID: paymentInfo.PaymentID,
		RazorpaySignature: paymentInfo.Signature,
		PaymentGateway:    "Razorpay",
		PaymentStatus:     "PAID",
		Amount:            amount,
	}
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

func FailedPayements(c *gin.Context) {
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
	id, _ := strconv.Atoi(c.Query("order_id"))
	orderID = uint(id)
	var order model.Order
	if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", orderID, userId).First(&order); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order not found!",
		})
		return
	}
	if order.PaymentMethod == "Razorpay" && order.PaymentStatus == "PENDING" {
		if FAILED == nil {
			FAILED = make(map[uint]int)
		}
		if FAILED[order.OrderID] >= 3 {
			deleteOrder(int(order.OrderID))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximm limit for payment reached reached!",
			})
			return
		}
		FAILED[order.OrderID]++
		c.HTML(http.StatusOK, "razorpay.html", nil)
	}
	if order.PaymentMethod != "Razorpay" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Payment method!",
		})
		return
	}

	if order.PaymentStatus != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Payment already done!",
		})
		return
	}
}
func OrderInvoice(c *gin.Context) {
	user, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}
	var User model.User
	if tx := database.DB.Model(&model.User{}).Where("email = ?", user).First(&User); tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exists!",
		})
		return
	}
	id, _ := strconv.Atoi(c.Query("orderid"))
	orderID = uint(id)
	var order model.Order
	if tx := database.DB.Model(&model.Order{}).Where("order_id = ? AND user_id = ?", orderID, User.ID).First(&order); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order not found!",
		})
		return
	}
	var orders []model.OrderItem
	if tx := database.DB.Model(&model.OrderItem{}).Where("order_id = ? AND user_id = ?", orderID, User.ID).Find(&orders); tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order not found!",
		})
		return
	}
	pdfBytes, err := GeneratePDFInvoice(order, orders, User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate PDF",
			"error":   err.Error(),
		})
		return
	}
	c.Writer.Header().Set("Content-type", "application/pdf")
	c.Writer.Header().Set("Content-Disposition", "inline; filename=salesreport.pdf")
	c.Writer.Write(pdfBytes)
}
func GeneratePDFInvoice(order model.Order, orderItems []model.OrderItem, user model.User) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Tabloid", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(260, 10, "Order Invoice", "0", 0, "C", false, 0, "")
	pdf.Ln(30)

	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(40, 10, "G-FRESH")
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 13)
	pdf.CellFormat(50, 8, "Maradu,452", "0", 0, "L", false, 0, "")

	pdf.CellFormat(130, 8, "", "0", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, "Date", "1", 0, "C", false, 0, "")
	year, month, date := order.OrderedAt.Date()

	pdf.CellFormat(50, 8, fmt.Sprintf("%v-%v-%v", date, month, year), "1", 1, "C", false, 0, "")

	pdf.CellFormat(50, 8, "Kochi, 678496, Eranakulam", "0", 0, "L", false, 0, "")

	pdf.CellFormat(130, 8, "", "0", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, "Order ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%v", order.OrderID), "1", 1, "C", false, 0, "")

	pdf.CellFormat(50, 8, "Phone: 7845126903", "0", 0, "L", false, 0, "")

	pdf.CellFormat(130, 8, "", "0", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, "User ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%v", order.UserID), "1", 1, "C", false, 0, "")

	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(30, 8, "Customer", "0", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 13)
	pdf.CellFormat(30, 8, fmt.Sprintf("%v", user.Name), "0", 1, "L", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%v, %v", order.ShippingAddress.StreetName, order.ShippingAddress.StreetNumber), "0", 1, "L", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%v, %v, %v", order.ShippingAddress.City, order.ShippingAddress.PinCode, order.ShippingAddress.State), "0", 1, "L", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%v", order.ShippingAddress.PhoneNumber), "0", 1, "L", false, 0, "")
	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(20, 10, "S.NO", "1", 0, "C", false, 0, "")
	pdf.CellFormat(110, 10, "Name", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Price", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Amount", "1", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 13)
	for i, val := range orderItems {
		pdf.CellFormat(20, 10, fmt.Sprintf("%v", i+1), "1", 0, "C", false, 0, "")
		var product model.Product
		database.DB.Model(&model.Product{}).Where("id=?", val.ProductID).First(&product)
		pdf.CellFormat(110, 10, fmt.Sprintf("%v", product.Name), "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 10, fmt.Sprintf("%v", product.Price), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%v", val.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 10, fmt.Sprintf("%v", product.Price*float64(val.Quantity)), "1", 1, "R", false, 0, "")
	}
	for i := len(orderItems); i < 10; i++ {
		pdf.CellFormat(20, 10, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(110, 10, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 10, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 10, "", "1", 1, "R", false, 0, "")
	}
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(210, 10, "Total", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%v", order.FinalAmount), "1", 1, "R", false, 0, "")

	pdf.Ln(10)
	pdf.SetFont("Arial", "", 13)
	pdf.CellFormat(130, 10, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(80, 10, "Total offer", "0", 0, "R", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("- %v", order.ProductOfferAmount), "0", 1, "R", false, 0, "")

	pdf.CellFormat(130, 10, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(80, 10, "Total Discount", "0", 0, "R", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("- %v", order.CouponDiscountAmount), "0", 1, "R", false, 0, "")

	pdf.CellFormat(130, 10, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(80, 10, "Delivery Charge", "0", 0, "R", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("+ %v", order.DeliveryCharge), "0", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(130, 10, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(80, 10, "Final Amount", "0", 0, "R", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%v", order.TotalAmount+float64(order.DeliveryCharge)), "0", 1, "R", false, 0, "")

	pdf.SetY(-50)
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 10, "If you have any questionsabout this price quote, please contact", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 10, "G-FRESH, 7845126903, gfreshproject@gmail.com", "", 0, "C", false, 0, "")
	var pdfBytes bytes.Buffer
	err := pdf.Output(&pdfBytes)
	if err != nil {
		return nil, err
	}

	return pdfBytes.Bytes(), nil
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
