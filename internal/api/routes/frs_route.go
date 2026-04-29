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
}
