package model

import "github.com/google/uuid"

type Album struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title    string
	ArtistID uuid.UUID
	// Artist      Artist `gorm:"foreignKey:ArtistID;references:ID"`
	ReleaseDate string
	CoverArt    string
	GenreID     uuid.UUID
	// Genre       Genre `gorm:"foreignKey:GenreID;references:ID"`
}
