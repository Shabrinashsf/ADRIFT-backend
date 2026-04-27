package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func User(route *gin.Engine, userController controller.UserController, jwtService service.JWTService) {
	routes := route.Group("api/auth")
	{
		routes.POST("", userController.RegisterUser)
		routes.POST("/login", userController.Login)
		routes.POST("/send-verification-email", userController.SendVerificationEmail)
		routes.GET("/verify-email", userController.VerifyEmail)
		routes.POST("/forgot-password", userController.ForgotPassword)
		routes.POST("/reset-password", userController.ResetPassword)
		routes.GET("/me", middleware.Authenticate(jwtService), userController.MeAuth)
		routes.PATCH("/update", middleware.Authenticate(jwtService), userController.UpdateUser)
	}
}
