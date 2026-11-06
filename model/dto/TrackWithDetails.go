package dto

import "github.com/google/uuid"

type TrackWithDetails struct {
	ID            uuid.UUID
	Title         string
	Duration      int
	AudioFileURL  string
	StreamCount   int
	AlbumID       uuid.UUID
	Explicit      bool
	Lyrics        string
	ReleaseDate   string
	YoutubeID     *string
	// Additional fields from JOINs
	AlbumCoverURL          string  // From albums.cover_art
	ArtistName             string  // From artists.name
	ArtistBio              *string // From artists.bio
	ArtistProfileImageUrl  *string // From artists.profile_image
	ArtistMonthlyListeners *int    // From artists.monthly_listeners
}
