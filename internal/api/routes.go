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
}

func AdminRoutes(router *gin.Engine) {
	router.GET("/admin/users", controllers.AdminAuthorization(), controllers.GetUserList)
}
