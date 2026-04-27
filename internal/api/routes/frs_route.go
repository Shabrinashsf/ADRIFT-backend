package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"

	"github.com/gin-gonic/gin"
)

func FRS(route *gin.Engine, frsController controller.FRSController, jwtService service.JWTService) {
	// frs := route.Group("api/frs")
	// {

	// }
}
