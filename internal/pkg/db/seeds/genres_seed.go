package seeds

import "log"

func (s Seed) GenresSeed() {
	// Step 1: Delete existing genres
	_, err := s.db.Exec("DELETE FROM genres")
	if err != nil {
		log.Fatalf("Failed to delete genres: %v", err)
	}

	// Step 2: Define genre data
	genres := []struct {
		Name        string
		Description string
	}{
		{"Pop", "Popular and catchy music."},
		{"Rock", "Guitar-driven music with strong rhythms."},
		{"Hip-Hop", "Rhythmic music with rapping vocals."},
		{"Jazz", "Improvisational music with swing and blue notes."},
		{"Classical", "Orchestral music with rich harmonies."},
		{"EDM", "Electronic dance music, club-oriented."},
		{"R&B", "Smooth rhythms and soulful vocals."},
		{"Metal", "Loud, aggressive guitar-driven music."},
		{"Reggae", "Caribbean music with offbeat rhythms."},
		{"Country", "Southern U.S. music with storytelling."},
	}

	// Step 3: Insert genres
	for _, genre := range genres {
		stmt, err := s.db.Prepare(`INSERT INTO genres (id, name, description) VALUES (UUID(), ?, ?)`)
		if err != nil {
			log.Fatalf("Failed to prepare genre insert: %v", err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(genre.Name, genre.Description)
		if err != nil {
			log.Fatalf("Failed to insert genre: %v", err)
		}
	}
}
