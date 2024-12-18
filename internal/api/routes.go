package api

import (
	"g-fresh/internal/controllers"

	"github.com/gin-gonic/gin"
)

func AuthenticationRoutes(router *gin.Engine) {
	//admin
	router.POST("/admin/login", controllers.AdminLogin)

	//user
	router.POST("/user/loginemail", controllers.UserLoginEmail)
	router.POST("/user/signupemail", controllers.UserSignupEmail)
	router.GET("/user/signupemail/otp/:email/:otp", controllers.OtpVerification)
	router.POST("/user/signupemail/resendotp", controllers.ResendOtp)
	router.GET("/auth/google/login", controllers.HandleGoogleLogin)
	router.GET("/auth/google/callback", controllers.HandleGoogleCallback)
	router.GET("/user/logout", controllers.Logout)

}

func AdminRoutes(router *gin.Engine) {
	//user
	router.GET("/admin/users", controllers.AdminAuthorization(), controllers.GetUserList)
	router.PUT("/admin/users/block/:userId", controllers.AdminAuthorization(), controllers.BlockUser)
	router.PUT("/admin/users/unblock/:userId", controllers.AdminAuthorization(), controllers.UnblockUser)

	//categories
	router.GET("/admin/categories", controllers.AdminAuthorization(), controllers.GetCategoryList)
	router.GET("/admin/categories/:id", controllers.AdminAuthorization(), controllers.GetCategory)
	router.POST("/admin/categories/add", controllers.AdminAuthorization(), controllers.AddCategory)
	router.DELETE("/admin/categories/delete/:id", controllers.AdminAuthorization(), controllers.DeleteCategory)
	router.PUT("/admin/categories/edit/:id", controllers.AdminAuthorization(), controllers.EditCategory)

	//products
	router.GET("/admin/products", controllers.AdminAuthorization(), controllers.GetProductList)
	router.POST("/admin/products/add", controllers.AdminAuthorization(), controllers.AddProducts)
	router.DELETE("/admin/products/delete/:id", controllers.AdminAuthorization(), controllers.DeleteProduct)
	router.PUT("/admin/products/edit/:id", controllers.AdminAuthorization(), controllers.EditProduct)

	//orders
	router.GET("/admin/orders", controllers.AdminAuthorization(), controllers.ShowOrdersAdmin)
	router.GET("/admin/orders/return", controllers.AdminAuthorization(), controllers.ReturnRequests)
	router.PATCH("/admin/orders/cancel/:pid/:orderid", controllers.AdminAuthorization(), controllers.CancelOrdersAdmin)
	router.PATCH("/admin/orders/status/:pid/:orderid", controllers.AdminAuthorization(), controllers.ChangeStatus)
	router.GET("/admin/order/invoice/:orderid", controllers.AdminAuthorization(), controllers.AdminOrderInvoice)

	//coupons
	router.GET("/admin/coupon", controllers.AdminAuthorization(), controllers.ShowCoupon)
	router.POST("/admin/coupon/add", controllers.AdminAuthorization(), controllers.AddCoupon)
	router.DELETE("/admin/coupon/delete/:code", controllers.AdminAuthorization(), controllers.DeleteCoupon)
	router.PATCH("/admin/coupon/update/:code", controllers.AdminAuthorization(), controllers.EditCoupon)

	//salesReport
	router.GET("/admin/sales/:download", controllers.AdminAuthorization(), controllers.SalesReport)

	//best selling
	router.GET("/admin/bestselling/products", controllers.AdminAuthorization(), controllers.BestSellingProducts)
	router.GET("/admin/bestselling/category", controllers.AdminAuthorization(), controllers.BestSellingCategory)

	//request
	router.GET("/admin/request/view", controllers.AdminAuthorization(), controllers.ViewRequests)
	router.PATCH("/admin/request/response/:request_id/:count", controllers.AdminAuthorization(), controllers.RequestResponse)
}

func UserRoutes(router *gin.Engine) {
	//home
	router.GET("/user/home", controllers.UserAuthorization(), controllers.GetHome)

	//category
	router.GET("/user/category/:id", controllers.UserAuthorization(), controllers.GetCategory)

	//product
	router.GET("/user/product/:id", controllers.UserAuthorization(), controllers.GetProduct)

	//profile
	router.GET("/user/profile", controllers.UserAuthorization(), controllers.ShowProfile)
	router.PUT("/user/profile/edit", controllers.UserAuthorization(), controllers.EditProfile)
	router.PATCH("/user/profile/password", controllers.UserAuthorization(), controllers.ChangePassword)

	//address
	router.GET("/user/address", controllers.UserAuthorization(), controllers.ShowAddress)
	router.POST("/user/address/add", controllers.UserAuthorization(), controllers.AddAddress)
	router.DELETE("/user/address/delete/:id", controllers.UserAuthorization(), controllers.DeleteAddress)
	router.PUT("/user/address/edit/:id", controllers.UserAuthorization(), controllers.EditAddress)

	//wishlist
	router.GET("/user/wishlist", controllers.UserAuthorization(), controllers.ShowWishlist)
	router.POST("/user/wishlist/add/:pid", controllers.UserAuthorization(), controllers.AddToWishlist)
	router.DELETE("/user/wishlist/delete/:pid", controllers.UserAuthorization(), controllers.DeleteFromWishlist)
	router.DELETE("/user/wishlist/move/:pid/:quantity", controllers.UserAuthorization(), controllers.MoveToCart)

	//cart
	router.GET("/user/cart/:referral/:coupon", controllers.UserAuthorization(), controllers.ShowCart)
	router.POST("/user/cart/add/:pid/:quantity", controllers.UserAuthorization(), controllers.AddToCart)
	router.DELETE("/user/cart/delete/:pid", controllers.UserAuthorization(), controllers.DeleteFromCart)

	//search
	router.GET("/user/products/search/:sort/:category/:available", controllers.UserAuthorization(), controllers.Search)

	//order
	router.GET("/user/order", controllers.UserAuthorization(), controllers.ShowOrders)
	router.POST("/user/order/:aid/:method/:referral/:coupon", controllers.UserAuthorization(), controllers.AddOrder)
	router.PATCH("/user/order/cancel/:pid/:orderid", controllers.UserAuthorization(), controllers.CancelOrders)
	router.PATCH("/user/order/return/:pid/:orderid", controllers.UserAuthorization(), controllers.OrderReturn)
	router.GET("/user/order/invoice/:orderid", controllers.UserAuthorization(), controllers.OrderInvoice)

	//payment
	router.LoadHTMLGlob("templates/*")
	router.GET("/user/order/payement", controllers.UserAuthorization(), controllers.RenderRazorpay)
	router.POST("/user/order/payement/create-order", controllers.UserAuthorization(), controllers.CreateOrder)
	router.POST("/user/order/payement/verify-payment", controllers.UserAuthorization(), controllers.VerifyPayment)
	router.GET("/user/order/payement/failed/:order_id", controllers.UserAuthorization(), controllers.FailedPayements)

	//rating
	router.POST("/user/rating/:pid/:rating", controllers.UserAuthorization(), controllers.AddRating)

	//wallet
	router.GET("/user/wallet", controllers.UserAuthorization(), controllers.ShowWallet)

	//request
	router.GET("/user/request/view", controllers.UserAuthorization(), controllers.ViewRequestsUser)
	router.POST("/user/request/send/:product_id/:count", controllers.UserAuthorization(), controllers.SendRequest)
	router.DELETE("/user/request/delete/:request_id", controllers.UserAuthorization(), controllers.DeleteRequest)
}
