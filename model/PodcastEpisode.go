package model

import (
	"time"
)

type PodcastEpisode struct {
	ID            string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SeriesID      string     `json:"seriesID" gorm:"column:series_id;type:uuid;not null"`
	Title         string     `json:"title" gorm:"column:title;size:200;not null"`
	EpisodeNumber *int       `json:"episodeNumber" gorm:"column:episode_number"`
	SeasonNumber  *int       `json:"seasonNumber" gorm:"column:season_number"`
	Description   *string    `json:"description" gorm:"column:description"`
	ReleaseDate   *time.Time `json:"releaseDate" gorm:"column:release_date"`
	Duration      int        `json:"duration" gorm:"column:duration;check:duration > 0"`
	AudioFileURL  string     `json:"audioFileUrl" gorm:"column:audio_file_url;not null"`
	CoverArt      *string    `json:"coverArt" gorm:"column:cover_art"`

	// Relationship
	Series *PodcastSeries `json:"-" gorm:"foreignKey:SeriesID;references:ID"`
}

func (PodcastEpisode) TableName() string {
	return "podcast_episodes"
}