package model

type Artist struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Bio              *string  `json:"bio,omitempty"`
	ProfileImage     *string  `json:"profileImage,omitempty"`
	Verified         bool     `json:"verified"`
	MonthlyListeners int32    `json:"monthlyListeners"`
	Albums           []*Album `json:"albums"`
}
