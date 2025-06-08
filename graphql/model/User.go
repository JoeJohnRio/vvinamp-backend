package model

type SubscriptionType string

const (
	SubscriptionFree    SubscriptionType = "Free"
	SubscriptionPremium SubscriptionType = "Premium"
)

type User struct {
	ID               string  `gorm:"primaryKey" json:"id"`
	Username         string  `gorm:"type:varchar(255);unique" json:"username"`
	Email            string  `gorm:"type:varchar(255);unique" json:"email"`
	PasswordHash     string  `gorm:"not null" json:"passwordHash"`
	ProfilePicture   *string `json:"profilePicture,omitempty"`
	JoinDate         string  `gorm:"not null" json:"joinDate"`
	SubscriptionType *string `json:"subscriptionType,omitempty"`
	LastLogin        *string `json:"lastLogin,omitempty"`
	// Playlists        []*Playlist         `json:"playlists"`
	// ListeningHistory []*ListeningHistory `json:"listeningHistory"`
	// LikedTracks      []*UserLike         `json:"likedTracks"`
	// FollowedArtists  []*UserFollow       `json:"followedArtists"`
}

// type User struct {
// 	ID               string  `gorm:"primaryKey" json:"id"`
// 	Username         string  `gorm:"type:varchar(255);unique" json:"username"`
// 	Email            string  `gorm:"type:varchar(255);unique" json:"email"`
// 	PasswordHash     string  `gorm:"not null" json:"passwordHash"`
// 	ProfilePicture   *string `json:"profilePicture,omitempty"`
// 	JoinDate         string  `gorm:"not null" json:"joinDate"`
// 	SubscriptionType *string `json:"subscriptionType,omitempty"`
// 	LastLogin        *string `json:"lastLogin,omitempty"`
// }
