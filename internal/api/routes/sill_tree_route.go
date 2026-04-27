package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"

	"github.com/gin-gonic/gin"
)

func SkillTree(route *gin.Engine, skillTreeController controller.SkillTreeController, jwtService service.JWTService) {
	// graph := route.Group("api/graph")
	// {

	// }
}
