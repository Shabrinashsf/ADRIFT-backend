package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"

	"github.com/gin-gonic/gin"
)

func Admin(route *gin.Engine, adminController controller.AdminController, jwtService service.JWTService) {
	// admin := route.Group("api/admin")
	// {

	// }
}
