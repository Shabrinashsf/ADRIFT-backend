package config

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"ADRIFT-backend/internal/api/routes"
	"ADRIFT-backend/internal/pkg/logger"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RestConfig struct {
	server     *gin.Engine
	httpServer *http.Server
	port       string
	env        string
	provider   *Provider
}

func NewRestConfig(db *gorm.DB) *RestConfig {
	app := gin.Default()
	server := NewRouter(app)

	// Initialize dependency injection provider
	provider := NewProvider(db)

	// =========== (INJECT DEPENDENCIES) ===========
	// Controllers
	userController := provider.InvokeUserController()
	fileController := provider.InvokeFileController()
	frsController := provider.InvokeFRSController()
	skillTreeController := provider.InvokeSkillTreeController()
	adminController := provider.InvokeAdminController()

	// Services needed for routes
	jwtService := provider.InvokeJWTService()

	// =========== (ROUTES) ===========
	routes.User(server, userController, jwtService)
	routes.File(server, fileController, jwtService)
	routes.FRS(server, frsController, jwtService)
	routes.SkillTree(server, skillTreeController, jwtService)
	routes.Admin(server, adminController, jwtService)

	// Get configuration
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	mode := os.Getenv("APP_ENV")
	if mode == "" {
		mode = "localhost"
	}

	return &RestConfig{
		server:     server,
		httpServer: nil,
		port:       port,
		env:        mode,
		provider:   provider,
	}
}

func (rc *RestConfig) Start() error {
	var addr string
	if rc.env == "localhost" {
		addr = "127.0.0.1:" + rc.port
	} else {
		addr = ":" + rc.port
	}

	rc.httpServer = &http.Server{
		Addr:    addr,
		Handler: rc.server,
	}

	go func() {
		myFigure := figure.NewColorFigure("ADRIFT Backend", "", "blue", true)
		myFigure.Print()
		fmt.Printf("Starting server on %s\n", addr)

		if err := rc.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Server error: %v", err)
		}
	}()

	return nil
}

func (rc *RestConfig) GetEngine() *gin.Engine {
	return rc.server
}

func (rc *RestConfig) GetPort() string {
	return rc.port
}

func (rc *RestConfig) GetEnv() string {
	return rc.env
}

func (rc *RestConfig) Shutdown(ctx context.Context) error {
	logger.Infof("Starting graceful shutdown...")

	if rc.httpServer != nil {
		logger.Infof("Shutting down HTTP server...")
		if err := rc.httpServer.Shutdown(ctx); err != nil {
			logger.Errorf("HTTP server shutdown error: %v", err)
			return err
		}
		logger.Infof("HTTP server stopped")
	}

	// Clean up DI provider
	if rc.provider != nil {
		rc.provider.Shutdown()
	}

	logger.Infof("Graceful shutdown completed")
	return nil
}
