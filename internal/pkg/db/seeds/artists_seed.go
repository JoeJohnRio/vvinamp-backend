package seeds

import (
	"log"

	"github.com/bxcodec/faker/v3"
)

func (s Seed) ArtistsSeed() {
	// This is the original seeder - keeping it for reference
	// Delete it first
	_, erro := s.db.Exec("DELETE FROM artists")
	if erro != nil {
		log.Fatalf("Failed to delete artists: %v", erro)
	}

	for i := 0; i < 100; i++ {
		stmt, err := s.db.Prepare(`INSERT INTO artists(name, bio, profile_image, verified, monthly_listeners) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			log.Fatalf("Failed to prepare statement: %v", err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(faker.Name(), faker.Sentence(), faker.URL(), false, 100)
		if err != nil {
			log.Fatalf("Failed to execute statement: %v", err)
		}
	}
}

// IndonesianArtistsSeed seeds the database with real Indonesian artists
func (s Seed) IndonesianArtistsSeed() {
	// This function would be for the SQL-based approach
	// For now, we'll use our GORM-based seeder in the seeder package
	log.Println("Indonesian artists seeding should be done via GORM seeder in seeder package")
}