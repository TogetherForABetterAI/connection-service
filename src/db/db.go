package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"connection-service/src/config"

	_ "github.com/lib/pq"
)

// DB represents the database connection and operations
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(cfg *config.GlobalConfig) (*DB, error) {
	dbConfig := cfg.GetDatabaseConfig()
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.GetHost(),
		dbConfig.GetPort(),
		dbConfig.GetUser(),
		dbConfig.GetPassword(),
		dbConfig.GetDBName(),
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Connected to PostgreSQL database",
		"host", dbConfig.GetHost(),
		"port", dbConfig.GetPort(),
		"database", dbConfig.GetDBName())

	// Execute init.sql to create tables and indexes
	if err := executeInitSQL(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to execute init.sql: %w", err)
	}

	return &DB{conn: conn}, nil
}

// GetConnection returns the underlying sql.DB connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// executeInitSQL reads and executes the init.sql file to create tables and indexes
func executeInitSQL(conn *sql.DB) error {
	// Read init.sql file
	sqlScript, err := os.ReadFile("init.sql")
	if err != nil {
		// Try alternative path for Docker container
		sqlScript, err = os.ReadFile("/app/init.sql")
		if err != nil {
			slog.Warn("init.sql file not found, skipping table creation", "error", err)
			return nil // Don't fail if init.sql doesn't exist
		}
	}

	// Execute the SQL script
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = conn.ExecContext(ctx, string(sqlScript))
	if err != nil {
		return fmt.Errorf("failed to execute SQL script: %w", err)
	}

	slog.Info("Successfully executed init.sql - tables and indexes created/verified")
	return nil
}
