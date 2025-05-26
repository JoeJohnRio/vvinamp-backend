package album

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type AlbumRepository interface {
	GetByID(ctx context.Context, id string) (*Album, error)
	Create(ctx context.Context, album *Album) error
	GetByArtist(ctx context.Context, artistID string) ([]*Album, error)
	// Add other methods...
}

// NewAlbumRepository returns a new instance of the albumRepository (constructor)
func NewAlbumRepository(db *sql.DB) AlbumRepository {
	return &albumRepository{db: db}
}

type NullString struct {
	String string
	Valid  bool
}
type albumRepository struct {
	db *sql.DB
}

// Implementation
func (r *albumRepository) GetByID(ctx context.Context, id string) (*Album, error) {
	query := `
    SELECT 
        a.id, a.title, a.artist_id, 
        DATE_FORMAT(a.release_date, '%Y-%m-%d') as release_date,
        a.cover_art, a.genre_id,
        ar.name as artist_name,
        g.name as genre_name
    FROM albums a
    JOIN artists ar ON a.artist_id = ar.id
    LEFT JOIN genres g ON a.genre_id = g.id
    WHERE a.id = ?
`

	var a Album
	var releaseDateStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID,
		&a.Title,
		&a.ArtistID,
		&releaseDateStr,
		&a.CoverArt,
		&a.GenreID,
		&a.ArtistName,
		&a.GenreName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Album not found
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Parse the date (adjust layout based on your DB format)
	if releaseDateStr != "" {
		a.ReleaseDate, err = time.Parse("2006-01-02", releaseDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	return &a, nil
}

func (r *albumRepository) Create(ctx context.Context, album *Album) error {
	// TODO: implement actual insert logic
	return nil
}

func (r *albumRepository) GetByArtist(ctx context.Context, artistID string) ([]*Album, error) {
	// TODO: implement actual query logic
	return nil, nil
}
