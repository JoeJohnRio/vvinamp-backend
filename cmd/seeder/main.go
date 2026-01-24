package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"vvinamp/seeder"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	// Parse command line flags
	flag.Parse()
	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "seed-music":
			runMusicSeeder()
		case "reset-and-seed-music":
			resetAndSeedMusic()
		default:
			fmt.Println("Usage:")
			fmt.Println("  go run cmd/seeder/main.go seed-music          # Seed music data (Indonesian artists)")
			fmt.Println("  go run cmd/seeder/main.go reset-and-seed-music # Reset and seed music data (Indonesian artists)")
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/seeder/main.go seed-music          # Seed music data (Indonesian artists)")
		fmt.Println("  go run cmd/seeder/main.go reset-and-seed-music # Reset and seed music data (Indonesian artists)")
		os.Exit(1)
	}
}

func runMusicSeeder() {
	// Initialize database connection directly
	db := initializeDatabase()

	// Run the music seeder
	ctx := context.Background()
	seederInstance := seeder.NewSeeder(db)

	log.Println("Starting to seed Indonesian music data...")

	if err := seederInstance.SeedAll(ctx); err != nil {
		log.Fatalf("Failed to run music seeder: %v", err)
	}

	log.Println("Music data seeding completed successfully!")
}

func resetAndSeedMusic() {
	// Initialize database connection directly
	db := initializeDatabase()

	// Run the music seeder (this will clear existing music data first)
	ctx := context.Background()
	seederInstance := seeder.NewSeeder(db)

	log.Println("Starting to reset and seed Indonesian music data...")

	if err := seederInstance.SeedAll(ctx); err != nil {
		log.Fatalf("Failed to run reset and music seeder: %v", err)
	}

	log.Println("Reset and music data seeding completed successfully!")
}

// initializeDatabase creates a direct database connection using environment variables
func initializeDatabase() *gorm.DB {
	// Try to use the POSTGRE_URL from environment first (as used in the main app)
	postgreURL := os.Getenv("POSTGRE_URL")
	if postgreURL != "" {
		// Remove quotes if present
		if len(postgreURL) > 1 && postgreURL[0] == '"' && postgreURL[len(postgreURL)-1] == '"' {
			postgreURL = postgreURL[1 : len(postgreURL)-1]
		}

		// Connect using the full connection string
		db, err := gorm.Open(postgres.Open(postgreURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database using POSTGRE_URL: %v", err)
		}

		// Test the connection
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("Failed to get database instance: %v", err)
		}

		if err := sqlDB.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		log.Println("Database connection established successfully using POSTGRE_URL")
		return db
	}

	// Check if POSTGRE_URL is commented in the .env file and extract the connection string
	envContent, err := os.ReadFile("../../.env") // Read the .env file
	if err == nil {
		lines := strings.Split(string(envContent), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# POSTGRE_URL=") {
				// Extract the connection string from the commented line
				commentedURL := strings.TrimPrefix(line, "# ")
				parts := strings.SplitN(commentedURL, "=", 2)
				if len(parts) == 2 {
					extractedURL := strings.Trim(parts[1], "\"'")
					db, err := gorm.Open(postgres.Open(extractedURL), &gorm.Config{})
					if err == nil {
						// Test the connection
						sqlDB, err := db.DB()
						if err == nil {
							if sqlDB.Ping() == nil {
								log.Println("Database connection established successfully using extracted POSTGRE_URL from .env")
								return db
							}
						}
					}
				}
			}
		}
	}

	// Fallback to individual environment variables
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASS", "password")
	dbname := getEnvOrDefault("DB_NAME", "vvinamp")
	sslMode := getEnvOrDefault("DB_SSL_MODE", "disable") // Allow SSL mode to be configured

	// Construct connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslMode)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connection established successfully")
	return db
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
