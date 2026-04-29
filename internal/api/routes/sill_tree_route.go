package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SkillTree(route *gin.Engine, skillTreeController controller.SkillTreeController, jwtService service.JWTService) {
	graph := route.Group("api/graph")
	{
		// Public endpoints
		graph.GET("", skillTreeController.GetGraph)
		graph.GET("/nodes/:courseId", skillTreeController.GetNodeDetail)
		graph.GET("/nodes/:courseId/chain", skillTreeController.GetNodeChain)

		// Authenticated (student only)
		authGraph := graph.Group("", middleware.Authenticate(jwtService))
		{
			authGraph.GET("/progress", skillTreeController.GetProgressGraph)
			authGraph.GET("/progress/summary", skillTreeController.GetProgressSummary)
			authGraph.POST("/progress/claim/:courseId", skillTreeController.ClaimCourse)
			authGraph.DELETE("/progress/claim/:courseId", skillTreeController.UnclaimCourse)
		}
	}
}
