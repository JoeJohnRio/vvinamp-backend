package album

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Album, error)
	Create(ctx context.Context, album *Album) error
	GetByArtist(ctx context.Context, artistID string) ([]*Album, error)
	// Add other methods...
}

type albumRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &albumRepository{db: db}
}

// Implementation
func (r *albumRepository) GetByID(ctx context.Context, id string) (*Album, error) {
	query := `
		SELECT 
			a.id, a.title, a.artist_id, a.release_date, 
			a.cover_art, a.genre_id,
			ar.name as artist_name,
			g.name as genre_name
		FROM albums a
		JOIN artists ar ON a.artist_id = ar.id
		LEFT JOIN genres g ON a.genre_id = g.id
		WHERE a.id = ?
	`

	var a Album
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Title, &a.ArtistID, &a.ReleaseDate,
		&a.CoverArt, &a.GenreID,
		&a.ArtistName, &a.GenreName,
	)

	if err != nil {
		return nil, err
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
