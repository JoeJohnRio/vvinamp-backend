package model

type Album struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      *Artist `json:"artist"`
	ReleaseDate string  `json:"releaseDate"`
	CoverArt    *string `json:"coverArt,omitempty"`
	Genre       *Genre  `json:"genre,omitempty"`
}
