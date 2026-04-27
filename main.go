package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ADRIFT-backend/cmd"
	"ADRIFT-backend/database"
	"ADRIFT-backend/internal/config"
	"ADRIFT-backend/internal/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize environment variables
	if err := godotenv.Load(".env"); err != nil {
		logger.Errorf("cannot load .env: %v", err)
	} else {
		logger.Infof(".env loaded successfully")
	}

	// Initialized database
	logger.Infof("Setting up database connection...")
	db := database.SetUpDatabaseConnection()
	defer database.CloseDatabaseConnection(db)
	logger.Infof("Database connection established.")

	// Handle CLI command
	if len(os.Args) > 1 {
		logger.Infof("Running commands...")
		cmd.Command(db)
		return
	}

	// Create REST config and start server
	restConfig := config.NewRestConfig(db)
	logger.Infof("Services initialized")
	logger.Infof("Setting up server...")

	if err := restConfig.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-quit
	logger.Infof("Received signal: %v", sig)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := restConfig.Shutdown(ctx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}

	logger.Infof("Application exited")
}
