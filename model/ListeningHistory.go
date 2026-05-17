package model

import (
	"time"
	"vvinamp/graphql/model"
)

type ListeningHistory struct {
	ID          string      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      string      `json:"userID" gorm:"column:user_id;type:uuid"`
	ContentID   string      `json:"contentID" gorm:"column:content_id"`
	ContentType string      `json:"contentType" gorm:"column:content_type"`
	PlayedAt    time.Time   `json:"playedAt" gorm:"column:played_at"`
	Progress    int         `json:"progress" gorm:"column:progress"`
	User        *model.User `json:"user,omitempty" gorm:"-"`
}

func (ListeningHistory) TableName() string {
	return "listening_history"
}
