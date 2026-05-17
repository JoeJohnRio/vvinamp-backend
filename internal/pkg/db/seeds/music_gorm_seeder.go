package seeds

import (
	"context"
	"log"
	"vvinamp/seeder"

	"gorm.io/gorm"
)

// RunMusicSeeder runs the GORM-based music seeder
func RunMusicSeeder(db *gorm.DB) {
	log.Println("Running GORM-based music seeder...")

	ctx := context.Background()
	seederInstance := seeder.NewSeeder(db)

	if err := seederInstance.SeedAll(ctx); err != nil {
		log.Fatalf("Failed to run music seeder: %v", err)
	}

	log.Println("GORM-based music seeder completed successfully!")
}
