package model

type TrackArtist struct {
	ID       string `gorm:"primaryKey"`
	TrackID  string // Foreign key, references Track.ID
	ArtistID string // Foreign key, references Artist.ID
	Role     string
}
