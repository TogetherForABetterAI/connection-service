package main

import (
	"auth-gateway/src/config"
	"auth-gateway/src/router"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "auth-gateway/src/docs"

	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

// @title Auth Gateway API
// @version 1.0
// @description API Gateway for authentication and authorization

// @contact.name   Auth Gateway Team
// @contact.url    https://github.com/your-org/auth-gateway
// @contact.email  auth-gateway@example.com

func main() {
	config := loadConfig()
	setupLogging()
	server := createServer(config)
	startServerWithGracefulShutdown(server, config)
}

func loadConfig() config.GlobalConfig {
	config, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	return config
}

func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func createServer(config config.GlobalConfig) *http.Server {
	r := router.NewRouter(config)
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler: r,
	}
}

func startServerWithGracefulShutdown(server *http.Server, config config.GlobalConfig) {
	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server", "host", config.Host, "port", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	slog.Info("Shutting down server...")

	// Attempt graceful shutdown without timeout
	if err := server.Shutdown(context.Background()); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		return
	}

	slog.Info("Server exited gracefully")
}
