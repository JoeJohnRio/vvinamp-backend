package genre

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type GenreRepository interface {
	GetAllGenres(ctx context.Context) ([]Genre, error)
}

// NewAlbumRepository returns a new instance of the albumRepository (constructor)
func NewGenreRepository(db *sql.DB) GenreRepository {
	return &genreRepository{db: db}
}

type NullString struct {
	String string
	Valid  bool
}
type genreRepository struct {
	db *sql.DB
}

// Implementation
func (r *genreRepository) GetAllGenres(ctx context.Context) ([]Genre, error) {
	query := `
		SELECT id, name, description
		FROM genres
	`

	log.Println("joel123", "jalan 1")

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query genres: %w", err)
	}
	defer rows.Close()

	log.Println("joel123", "jalan 2")
	var genres []Genre
	for rows.Next() {
		var g Genre
		err := rows.Scan(&g.ID, &g.Name, &g.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan genre row: %w", err)
		}
		genres = append(genres, g)
	}

	log.Println("joel123", "jalan 3")
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return genres, nil
}
