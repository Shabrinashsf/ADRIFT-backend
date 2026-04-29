package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func FRS(route *gin.Engine, frsController controller.FRSController, jwtService service.JWTService) {
	admin := route.Group("api/admin").Use(middleware.Authenticate(jwtService), middleware.OnlyAllow("ADMIN"))
	{
		admin.POST("/schedule/upload", frsController.UploadScheduleFile)
		admin.POST("/schedule/revise", frsController.DeleteScheduleArtifacts)
		admin.POST("/schedule/submit", frsController.SubmitSchedule)
	}

	student := route.Group("api/frs").Use(middleware.Authenticate(jwtService))
	{
		student.GET("/schedules", frsController.ListSchedules)
		student.POST("", frsController.CreateFRSPlan)
		student.POST("/alternative", frsController.FindAlternatives)
		student.GET("", frsController.ListFRSPlans)
		student.GET("/:planId", frsController.GetFRSPlanDetail)
		student.DELETE("/:planId", frsController.DeleteFRSPlan)
	}
}
