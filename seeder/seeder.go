package seeder

import (
	"context"
	"log"

	"gorm.io/gorm"
)

// Run executes all music-related seeders
func Run(ctx context.Context, db *gorm.DB) error {
	seeder := NewSeeder(db)

	log.Println("Starting music data seeding...")

	if err := seeder.SeedAll(ctx); err != nil {
		return err
	}

	log.Println("Music data seeding completed successfully!")
	return nil
}
