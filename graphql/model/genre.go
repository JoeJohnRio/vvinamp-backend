package model

type Genre struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Albums      []*Album `json:"albums"`
}
