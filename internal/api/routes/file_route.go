package routes

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func File(route *gin.Engine, fileController controller.FileController, jwtService service.JWTService) {
	routes := route.Group("api/assets").Use(middleware.Authenticate(jwtService))
	{
		routes.GET("/*path", fileController.ServeUpload)
	}

	admin := route.Group("api/admin/assets").Use(middleware.Authenticate(jwtService), middleware.OnlyAllow("ADMIN"))
	{
		admin.POST("/frs/upload", fileController.UploadFRSTempFile)
	}
}
