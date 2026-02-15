package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX interface to allow mocking and easier testing
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

var dbPool DBTX

// InitDB initializes the database connection pool
func InitDB() error {
	// Get database connection string from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to a default connection string if not set
		dbURL = "postgres://username:password@localhost:5432/database_name?sslmode=disable"
		log.Println("Using default database URL. Please set DATABASE_URL environment variable.")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test the connection
	if err = pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	dbPool = pool
	log.Println("Database connection initialized successfully")
	return nil
}

// CloseDB closes the database connection pool
func CloseDB() {
	if pool, ok := dbPool.(*pgxpool.Pool); ok {
		pool.Close()
		log.Println("Database connection closed")
	}
}