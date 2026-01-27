package model

type Playlist struct {
	ID          string  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title       string  `json:"title" gorm:"column:title;size:100;not null"`
	Description *string `json:"description" gorm:"column:description"`
	CreatorID   *string `json:"creatorID" gorm:"column:creator_id;type:uuid"`
	IsPublic    bool    `json:"isPublic" gorm:"column:is_public;default:true"`
	CoverImage  *string `json:"coverImage" gorm:"column:cover_image"`

	// Relationship
	// Creator *User `json:"-" gorm:"foreignKey:CreatorID;references:ID"`
}

func (Playlist) TableName() string {
	return "playlists"
}