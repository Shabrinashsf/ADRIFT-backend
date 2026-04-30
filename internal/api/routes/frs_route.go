package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func FRS(route *gin.Engine, frsController controller.FRSController, jwtService service.JWTService) {
	authenticated := route.Group("api/frs").Use(middleware.Authenticate(jwtService))
	{
		authenticated.GET("/schedules", frsController.ListSchedules)
	}

	student := route.Group("api/frs").Use(middleware.Authenticate(jwtService), middleware.OnlyAllow("STUDENT"))
	{
		student.POST("", frsController.CreateFRSPlan)
		student.POST("/alternative", frsController.FindAlternatives)
		student.GET("", frsController.ListFRSPlans)
		student.GET("/:planId", frsController.GetFRSPlanDetail)
		student.DELETE("/:planId", frsController.DeleteFRSPlan)
	}
}
