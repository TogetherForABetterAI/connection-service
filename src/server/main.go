package server

import (
	"connection-service/src/config"
	"connection-service/src/db"
	"connection-service/src/middleware"
	"connection-service/src/router"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "connection-service/src/docs"

	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

// Server represents the HTTP server
type Server struct {
	config          *config.GlobalConfig
	database        *db.DB
	http            *http.Server
	shutdownHandler ShutdownHandlerInterface
}

// NewServer creates a new server instance
func NewServer(cfg *config.GlobalConfig) (*Server, error) {
	// Initialize database connection
	database, err := db.NewDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	server := &Server{
		config:   cfg,
		database: database,
	}

	// Create and assign shutdown handler
	server.shutdownHandler = NewShutdownHandler(server)

	return server, nil
}

// Run starts the server with graceful shutdown using ShutdownHandler
func (s *Server) Run() error {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	serverDone := s.startServerGoroutine()

	return s.shutdownHandler.HandleShutdown(serverDone, osSignals)
}

// startServerGoroutine starts the HTTP server in a goroutine and returns a channel for errors
func (s *Server) startServerGoroutine() chan error {
	serverDone := make(chan error, 1)

	go func() {

		middleware, err := middleware.NewMiddleware(s.config)
		s.shutdownHandler.SetMiddleware(middleware)
		r := router.NewRouter(s.config, s.database, middleware)
		// Create HTTP server
		httpServer := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", s.config.GetHost(), s.config.GetPort()),
			Handler: r,
		}
		s.http = httpServer

		slog.Info("Starting connection service",
			"host", s.config.GetHost(),
			"port", s.config.GetPort())

		err = s.startServer()
		serverDone <- err
	}()

	return serverDone
}

// startServer starts the HTTP server and handles errors
func (s *Server) startServer() error {
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
