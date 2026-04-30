package routes

import (
	"ADRIFT-backend/constants"
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Admin(route *gin.Engine, adminController controller.AdminController, jwtService service.JWTService) {
	admin := route.Group("api/admin", middleware.Authenticate(jwtService), middleware.OnlyAllow(constants.ENUM_ROLE_ADMIN))
	{
		// Lab Paths
		admin.GET("/lab-paths", adminController.GetAllLabPaths)
		admin.POST("/lab-paths", adminController.CreateLabPath)
		admin.PATCH("/lab-paths/:labPathId", adminController.UpdateLabPath)
		admin.DELETE("/lab-paths/:labPathId", adminController.DeleteLabPath)

		// Prerequisites
		admin.POST("/prerequisites", adminController.CreatePrerequisite)
		admin.DELETE("/prerequisites/:courseId/:requireId", adminController.DeletePrerequisite)

		// Path Edges
		admin.POST("/path-edges", adminController.CreatePathEdge)
		admin.DELETE("/path-edges/:pathEdgeId", adminController.DeletePathEdge)

		// Lectures
		admin.GET("/lectures", adminController.GetAllLectures)
		admin.POST("/lectures", adminController.CreateLecture)
		admin.PATCH("/lectures/:lectureId", adminController.UpdateLecture)
	}
}
