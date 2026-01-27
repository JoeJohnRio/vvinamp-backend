package model

type PodcastSeries struct {
	ID          string  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title       string  `json:"title" gorm:"column:title;size:150"`
	Description *string `json:"description" gorm:"column:description"`
	CoverArt    *string `json:"coverArt" gorm:"column:cover_art"`
	GenreID     *string `json:"genreID" gorm:"column:genre_id;type:uuid"`
	Publisher   *string `json:"publisher" gorm:"column:publisher;size:100"`
}

func (PodcastSeries) TableName() string {
	return "podcast_series"
}