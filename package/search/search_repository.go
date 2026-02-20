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

// AlbumWithArtist extends dbModel.Album to include artist data for search results
type AlbumWithArtist struct {
	dbModel.Album
	Artist *dbModel.Artist
}

// TrackWithDetails extends dbModel.Track to include album cover and artist name for search results
type TrackWithDetails struct {
	dbModel.Track
	AlbumCoverURL string
	ArtistName    string
}

type SearchResult struct {
	Tracks          []*TrackWithDetails
	Albums          []*AlbumWithArtist
	Artists         []*dbModel.Artist
	Playlists       []*dbModel.Playlist
	PodcastEpisodes []*dbModel.PodcastEpisode
	PodcastSeries   []*dbModel.PodcastSeries
	TotalCount      int
}

type Repository interface {
	Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
}
