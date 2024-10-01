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
	router.GET("/admin/users", controllers.AdminAuthorization(), controllers.GetUserList)
	router.PUT("/admin/users/block/:userId", controllers.AdminAuthorization(), controllers.BlockUser)
	router.PUT("/admin/users/unblock/:userId", controllers.AdminAuthorization(), controllers.UnblockUser)

	router.GET("/admin/categories", controllers.AdminAuthorization(), controllers.GetCategoryList)
	router.GET("/admin/categories/:id", controllers.AdminAuthorization(), controllers.GetCategory)
	router.POST("/admin/categories/add", controllers.AdminAuthorization(), controllers.AddCategory)
	router.DELETE("/admin/categories/delete/:id", controllers.AdminAuthorization(), controllers.DeleteCategory)
	router.PUT("/admin/categories/edit/:id", controllers.AdminAuthorization(), controllers.EditCategory)

	router.GET("/admin/products", controllers.AdminAuthorization(), controllers.GetProductList)
	router.POST("/admin/products/add", controllers.AdminAuthorization(), controllers.AddProducts)
	router.DELETE("/admin/products/delete/:id", controllers.AdminAuthorization(), controllers.DeleteProduct)
	router.PUT("/admin/products/edit/:id", controllers.AdminAuthorization(), controllers.EditProduct)
}

func UserRoutes(router *gin.Engine) {
	router.GET("/user/home", controllers.UserAuthorization(), controllers.GetHome)
	router.GET("/user/category/:id", controllers.UserAuthorization(), controllers.GetCategory)
	router.GET("/user/product/:id", controllers.UserAuthorization(), controllers.GetProduct)
}
