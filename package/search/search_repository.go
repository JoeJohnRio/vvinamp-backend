package search

import (
	"context"
	dbModel "vvinamp/model"
	"vvinamp/graphql/model"
)

// ContentType alias for convenience
type ContentType = model.ContentType

type SearchQuery struct {
	Query       *string
	ContentType *ContentType
	Genre       *string
	Tags        []string
	Explicit    *bool
	Limit       int
	Offset      int
}

type SearchResult struct {
	Tracks          []*dbModel.Track
	Albums          []*dbModel.Album
	Artists         []*dbModel.Artist
	Playlists       []*dbModel.Playlist
	PodcastEpisodes []*dbModel.PodcastEpisode
	PodcastSeries   []*dbModel.PodcastSeries
	TotalCount      int
}

type Repository interface {
	Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
}
