package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
