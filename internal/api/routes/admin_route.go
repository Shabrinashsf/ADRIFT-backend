package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Admin(route *gin.Engine, adminController controller.AdminController, frsController controller.FRSController, fileController controller.FileController, jwtService service.JWTService) {
	admin := route.Group("api/admin")
	admin.Use(middleware.Authenticate(jwtService), middleware.OnlyAllow("ADMIN"))
	{
		// Course
		admin.GET("/courses", adminController.ListCourseGroups)
		admin.GET("/courses/:semester", adminController.ListCoursesBySemester)
		admin.POST("/courses", adminController.CreateCourse)
		admin.PATCH("/courses/:courseId", adminController.UpdateCourse)
		admin.DELETE("/courses/:courseId", adminController.DeleteCourse)

		// Schedule
		admin.GET("/schedules", adminController.ListSchedules)
		admin.POST("/schedules", adminController.CreateSchedule)
		admin.PATCH("/schedules/:scheduleId", adminController.UpdateSchedule)
		admin.DELETE("/schedules/:scheduleId", adminController.DeleteSchedule)

		// FRS schedule
		admin.POST("/schedule/upload", frsController.UploadScheduleFile)
		admin.POST("/schedule/revise", frsController.DeleteScheduleArtifacts)
		admin.POST("/schedule/submit", frsController.SubmitSchedule)

		// FRS file upload
		admin.POST("/assets/frs/upload", fileController.UploadFRSTempFile)

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

	authenticated := route.Group("api/assets").Use(middleware.Authenticate(jwtService))
	{
		authenticated.GET("/*path", fileController.ServeUpload)
	}

	FRS(route, frsController, jwtService)
}
