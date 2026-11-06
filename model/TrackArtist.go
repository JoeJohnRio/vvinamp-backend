package model

type TrackArtist struct {
	ID       string `gorm:"-" json:"id,omitempty"`
	TrackID  string `gorm:"column:track_id;type:uuid;primaryKey"`
	ArtistID string `gorm:"column:artist_id;type:uuid;primaryKey"`
	Role     string `gorm:"column:role"`
}

func (TrackArtist) TableName() string {
	return "track_artist"
}
