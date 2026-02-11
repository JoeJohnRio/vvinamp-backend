package dto

import "github.com/google/uuid"

type TrackWithDetails struct {
	ID           uuid.UUID
	Title        string
	Duration     int
	AudioFileURL string
	StreamCount  int
	AlbumID      uuid.UUID
	Explicit     bool
	Lyrics       string
	ReleaseDate  string
	// Additional fields from JOINs
	AlbumCoverURL string // From albums.cover_art
}
