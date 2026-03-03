package model

import (
	"time"
)

type Playlist struct {
	ID           string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string     `json:"title" gorm:"column:title;size:100;not null"`
	Description  *string    `json:"description" gorm:"column:description"`
	CreatorID    string     `json:"creatorId" gorm:"column:creator_id;type:uuid;not null"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt    *time.Time `json:"updatedAt" gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	IsPublic     bool       `json:"isPublic" gorm:"column:is_public;default:true"`
	CoverImage   *string    `json:"coverImage" gorm:"column:cover_image"`
	BannerImage  *string    `json:"bannerImage" gorm:"column:banner_image"`

	// Relationships (not mapped in GORM, handled in queries)
	// Creator *User `json:"-" gorm:"foreignKey:CreatorID;references:ID"`
}

func (Playlist) TableName() string {
	return "playlists"
}

// PlaylistTrack represents the join table between playlists and tracks
type PlaylistTrack struct {
	PlaylistID string    `json:"playlistId" gorm:"column:playlist_id;type:uuid;primaryKey"`
	TrackID    string    `json:"trackId" gorm:"column:track_id;type:uuid;primaryKey"`
	Position   int       `json:"position" gorm:"column:position;not null;primaryKey"`
	AddedAt    time.Time `json:"addedAt" gorm:"column:added_at;default:CURRENT_TIMESTAMP"`
}

func (PlaylistTrack) TableName() string {
	return "playlist_track"
}

// UserSavedPlaylist represents the join table for users saving playlists
type UserSavedPlaylist struct {
	UserID     string    `json:"userId" gorm:"column:user_id;type:uuid;primaryKey"`
	PlaylistID string    `json:"playlistId" gorm:"column:playlist_id;type:uuid;primaryKey"`
	SavedAt    time.Time `json:"savedAt" gorm:"column:saved_at;default:CURRENT_TIMESTAMP"`
}

func (UserSavedPlaylist) TableName() string {
	return "user_saved_playlists"
}