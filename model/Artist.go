package model

import (
	"gorm.io/gorm"
	"time"
)

type Artist struct {
	ID                      string         `gorm:"type:uuid;primaryKey" json:"id"`
	Name                    string         `gorm:"not null" json:"name"`
	Bio                     *string        `gorm:"type:text" json:"bio,omitempty"`
	ProfileImage            *string        `gorm:"type:text" json:"profile_image,omitempty"`
	VerticalVideoHeaderUrl  *string        `gorm:"type:text" json:"vertical_video_header_url,omitempty"`
	Verified                bool           `gorm:"default:false" json:"verified"`
	MonthlyListeners        int32          `gorm:"default:0" json:"monthly_listeners"`
	IsFollowed              bool           `gorm:"-" json:"is_followed"` // This field is computed at runtime based on user relationship
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Note: We're not including relationships to avoid migration issues
	// Relationships would be handled in queries when needed
}

// TableName specifies the table name for the Artist model
func (Artist) TableName() string {
	return "artists"
}