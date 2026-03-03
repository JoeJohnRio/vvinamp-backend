package model

import "time"

type SubscriptionType string

const (
	SubscriptionFree    SubscriptionType = "Free"
	SubscriptionPremium SubscriptionType = "Premium"
)

type User struct {
	ID               string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username         string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email            string     `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash     string     `gorm:"size:255;not null" json:"passwordHash"`
	ProfilePicture   *string    `json:"profilePicture,omitempty"`
	JoinDate         time.Time  `gorm:"not null;autoCreateTime" json:"joinDate"`
	SubscriptionType string     `gorm:"type:varchar(20);default:Free" json:"subscriptionType"`
	LastLogin        *time.Time `gorm:"" json:"lastLogin,omitempty"`
}
