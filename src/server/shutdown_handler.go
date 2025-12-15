package server

import (
	"connection-service/src/middleware"
	"context"
	"log/slog"
	"os"
)

// ShutdownHandlerInterface defines the interface for handling graceful shutdown
type ShutdownHandlerInterface interface {
	// HandleShutdown orchestrates the shutdown process
	// Returns an error if shutdown encounters an issue
	HandleShutdown(serverDone chan error, osSignals chan os.Signal) error

	// ShutdownServer initiates server shutdown
	ShutdownServer()

	// SetMiddleware sets the middleware for the shutdown handler
	SetMiddleware(mw *middleware.Middleware)
}

// ShutdownHandler implements the ShutdownHandlerInterface
type ShutdownHandler struct {
	server *Server
	middleware *middleware.Middleware
}

// NewShutdownHandler creates a new shutdown handler
func NewShutdownHandler(server *Server) ShutdownHandlerInterface {
	return &ShutdownHandler{
		server: server,
	}
}

// HandleShutdown orchestrates graceful shutdown based on shutdown sources
func (h *ShutdownHandler) HandleShutdown(serverDone chan error, osSignals chan os.Signal) error {
	// Wait for one of two shutdown triggers:
	// 1. Server error/completion (serverDone)
	// 2. OS signal (SIGTERM/SIGINT from Kubernetes or user)
	select {
	case err := <-serverDone:
		// Server stopped (error or normal completion)
		slog.Info("Server stopped, initiating shutdown")
		close(osSignals) // Signal OS goroutine to stop if it's listening
		h.ShutdownServer()
		return h.handleServerError(err)

	case sig, ok := <-osSignals:
		// OS signal received (SIGTERM from Kubernetes or user)
		if !ok {
			return nil
		}
		slog.Info("Received OS signal, initiating shutdown", "signal", sig)
		h.ShutdownServer()

		// Wait for server to finish
		err := <-serverDone
		return h.handleServerError(err)
	}
}

// handleServerError handles shutdown when server stops
func (h *ShutdownHandler) handleServerError(err error) error {
	if err != nil {
		slog.Error("Service stopped with an error", "error", err)
		return err
	}
	slog.Info("Service stopped cleanly")
	return nil
}

// ShutdownServer initiates the shutdown of all server components
func (h *ShutdownHandler) ShutdownServer() {
	slog.Info("Shutting down server components...")

	// Attempt graceful shutdown of HTTP server
	if err := h.server.http.Shutdown(context.Background()); err != nil {
		slog.Error("Error during HTTP server shutdown", "error", err)
	}

	if h.middleware != nil {
		h.middleware.HandleSigterm()
	}

	// Close database connection
	if h.server.database != nil {
		h.server.database.Close()
		slog.Info("Database connection closed")
	}

	slog.Info("Server shutdown complete")
}


func (h *ShutdownHandler) SetMiddleware(mw *middleware.Middleware) {
	h.middleware = mw
}