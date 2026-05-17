package model

import "github.com/google/uuid"

type Track struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string
	Duration     int
	AudioFileURL string
	StreamCount  int
	AlbumID      uuid.UUID
	// Album        Album `gorm:"foreignKey:AlbumID;references:ID"`
	Explicit    bool
	Lyrics      string
	ReleaseDate string
}
