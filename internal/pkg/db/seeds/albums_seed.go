package seeds

import (
	"log"
	"math/rand"

	"github.com/bxcodec/faker/v3"
)

func (s Seed) AlbumsSeed() {
	// Step 1: Delete existing albums
	_, err := s.db.Exec("DELETE FROM albums")
	if err != nil {
		log.Fatalf("Failed to delete albums: %v", err)
	}

	// Step 2: Fetch artist IDs
	artistRows, err := s.db.Query("SELECT id FROM artists")
	if err != nil {
		log.Fatalf("Failed to fetch artist IDs: %v", err)
	}
	defer artistRows.Close()

	var artistIDs []string
	for artistRows.Next() {
		var id string
		if err := artistRows.Scan(&id); err != nil {
			log.Fatalf("Failed to scan artist ID: %v", err)
		}
		artistIDs = append(artistIDs, id)
	}

	// Step 3: Fetch genre IDs
	genreRows, err := s.db.Query("SELECT id FROM genres")
	if err != nil {
		log.Fatalf("Failed to fetch genre IDs: %v", err)
	}
	defer genreRows.Close()

	var genreIDs []string
	for genreRows.Next() {
		var id string
		if err := genreRows.Scan(&id); err != nil {
			log.Fatalf("Failed to scan genre ID: %v", err)
		}
		genreIDs = append(genreIDs, id)
	}

	if len(artistIDs) == 0 || len(genreIDs) == 0 {
		log.Fatal("Artist or genre IDs are empty. Make sure to seed them first.")
	}

	// Step 4: Insert new albums
	for i := 0; i < 100; i++ {
		stmt, err := s.db.Prepare(`
			INSERT INTO albums (title, artist_id, release_date, cover_art, genre_id)
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			log.Fatalf("Failed to prepare album insert: %v", err)
		}
		defer stmt.Close()

		title := faker.Word() + " " + faker.Word()
		artistID := artistIDs[rand.Intn(len(artistIDs))]
		genreID := genreIDs[rand.Intn(len(genreIDs))]
		releaseDate := faker.Date() // returns time.Time
		coverArt := faker.URL()

		_, err = stmt.Exec(title, artistID, releaseDate, coverArt, genreID)
		if err != nil {
			log.Fatalf("Failed to insert album: %v", err)
		}
	}
}
