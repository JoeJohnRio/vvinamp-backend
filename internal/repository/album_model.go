package album

import (
	"database/sql"
	"time"
)

type Album struct {
	ID          string
	Title       string
	ArtistID    string
	ReleaseDate time.Time
	CoverArt    sql.NullString
	GenreID     sql.NullString

	// Joined fields (not in DB table)
	ArtistName string         `db:"-"`
	GenreName  sql.NullString `db:"-"`
}
