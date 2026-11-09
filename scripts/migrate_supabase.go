package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("POSTGRE_URL")
	if dsn == "" {
		log.Fatal("POSTGRE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping failed: %v", err)
	}
	fmt.Println("Connected to Supabase PostgreSQL.")

	sqlBytes, err := os.ReadFile("supabase_ready.sql")
	if err != nil {
		log.Fatalf("Failed to read supabase_ready.sql: %v", err)
	}

	fmt.Println("Executing database schema and seed data import...")
	start := time.Now()
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Printf("Migration completed successfully in %v!\n", time.Since(start).Round(time.Millisecond))

	// Verify imported tables
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name;
	`)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nCreated Tables in Supabase:")
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			fmt.Printf(" - %s\n", name)
		}
	}
}
