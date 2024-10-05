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
	router.PATCH("/admin/orders/cancel/:pid/:orderid", controllers.AdminAuthorization(), controllers.CancelOrdersAdmin)
	router.PATCH("/admin/orders/status/:pid/:orderid", controllers.AdminAuthorization(), controllers.ChangeStatus)
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

	//cart
	router.GET("/user/cart", controllers.UserAuthorization(), controllers.ShowCart)
	router.POST("/user/cart/add/:pid/:quantity", controllers.UserAuthorization(), controllers.AddToCart)
	router.POST("/user/cart/delete/:pid", controllers.UserAuthorization(), controllers.DeleteFromCart)

	//search
	router.GET("/user/products/lowtohigh", controllers.UserAuthorization(), controllers.Search_P_LtoH)
	router.GET("/user/products/hightolow", controllers.UserAuthorization(), controllers.Search_P_HtoL)
	router.GET("/user/products/new", controllers.UserAuthorization(), controllers.SearchNew)
	router.GET("/user/products/AtoZ", controllers.UserAuthorization(), controllers.SearchAtoZ)
	router.GET("/user/products/ZtoA", controllers.UserAuthorization(), controllers.SearchZtoA)
	router.GET("/user/products/rating", controllers.UserAuthorization(), controllers.SearchAverageRating)
	router.GET("/user/products/popular", controllers.UserAuthorization(), controllers.SearchPopular)
	router.GET("/user/products/featured", controllers.UserAuthorization(), controllers.SearchFeatured)

	//order
	router.GET("/user/order", controllers.UserAuthorization(), controllers.ShowOrders)
	router.POST("/user/order/:aid", controllers.UserAuthorization(), controllers.AddOrder)
	router.PATCH("/user/order/cancel/:pid/:orderid", controllers.UserAuthorization(), controllers.CancelOrders)

	//rating
	router.POST("/user/rating/:pid/:rating", controllers.UserAuthorization(), controllers.AddRating)
}
