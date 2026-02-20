package repository

import (
	"context"
	"strings"
	"vvinamp/database/postgre"
	dbModel "vvinamp/model"
	"vvinamp/package/search"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type sqlRepository struct {
	logger *zap.Logger
}

type SQLRepositoryTarget struct {
	fx.In
	Connection *postgre.Connection
	Logger     *zap.Logger
}

func NewSQLRepository(target SQLRepositoryTarget) (search.Repository, error) {
	db, err := target.Connection.Client()
	if err != nil {
		return nil, err
	}

	// AutoMigrate if needed (optional for search since it doesn't own tables)
	_ = db

	return &sqlRepository{
		logger: target.Logger,
	}, nil
}

func (s *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return postgre.ForContext(ctx)
}

// Search implements search.Repository.
func (s *sqlRepository) Search(ctx context.Context, query search.SearchQuery) (*search.SearchResult, error) {
	result := &search.SearchResult{}

	// Build search query pattern
	searchPattern := ""
	if query.Query != nil {
		searchPattern = "%" + strings.TrimSpace(*query.Query) + "%"
	}

	// Search tracks
	if query.ContentType == nil || *query.ContentType == search.ContentTypeTrack {
		var tracks []*dbModel.Track
		if searchPattern != "" {
			s.DB(ctx).Where("title ILIKE ?", searchPattern).Limit(query.Limit).Find(&tracks)
		}
		
		// Convert to TrackWithDetails and fetch album cover and artist name
		tracksWithDetails := make([]*search.TrackWithDetails, len(tracks))
		for i, track := range tracks {
			tracksWithDetails[i] = &search.TrackWithDetails{
				Track: *track,
			}
			
			// Fetch album cover if track has an album
			if track.AlbumID != uuid.Nil {
				var album dbModel.Album
				if err := s.DB(ctx).Where("id = ?", track.AlbumID).First(&album).Error; err == nil {
					tracksWithDetails[i].AlbumCoverURL = album.CoverArt
					
					// Fetch artist name for this album
					var artist dbModel.Artist
					if err := s.DB(ctx).Where("id = ?", album.ArtistID).First(&artist).Error; err == nil {
						tracksWithDetails[i].ArtistName = artist.Name
					}
				}
			}
			
			// If no album cover found, try to get from track_artist relationship
			if tracksWithDetails[i].AlbumCoverURL == "" {
				var trackArtist dbModel.TrackArtist
				if err := s.DB(ctx).Where("track_id = ?", track.ID.String()).First(&trackArtist).Error; err == nil {
					var artist dbModel.Artist
					if err := s.DB(ctx).Where("id = ?", trackArtist.ArtistID).First(&artist).Error; err == nil {
						if tracksWithDetails[i].ArtistName == "" {
							tracksWithDetails[i].ArtistName = artist.Name
						}
					}
				}
			}
		}
		result.Tracks = tracksWithDetails
		result.TotalCount += len(tracksWithDetails)
	}

	// Search albums
	if query.ContentType == nil || *query.ContentType == search.ContentTypeAlbum {
		var albums []*dbModel.Album
		if searchPattern != "" {
			s.DB(ctx).Where("title ILIKE ?", searchPattern).Limit(query.Limit).Find(&albums)
		}
		
		// Convert to AlbumWithArtist and fetch artist data
		albumsWithArtist := make([]*search.AlbumWithArtist, len(albums))
		for i, album := range albums {
			albumsWithArtist[i] = &search.AlbumWithArtist{
				Album: *album,
			}
			// Fetch artist for this album
			var artist dbModel.Artist
			if err := s.DB(ctx).Where("id = ?", album.ArtistID).First(&artist).Error; err == nil {
				albumsWithArtist[i].Artist = &artist
			}
		}
		result.Albums = albumsWithArtist
		result.TotalCount += len(albumsWithArtist)
	}

	// Search artists
	if query.ContentType == nil || *query.ContentType == search.ContentTypeArtist {
		var artists []*dbModel.Artist
		if searchPattern != "" {
			s.DB(ctx).Where("name ILIKE ?", searchPattern).Limit(query.Limit).Find(&artists)
		}
		result.Artists = artists
		result.TotalCount += len(artists)
	}

	// Search playlists (only public)
	if query.ContentType == nil || *query.ContentType == search.ContentTypePlaylist {
		var playlists []*dbModel.Playlist
		if searchPattern != "" {
			s.DB(ctx).Where("title ILIKE ? AND is_public = ?", searchPattern, true).Limit(query.Limit).Find(&playlists)
		}
		result.Playlists = playlists
		result.TotalCount += len(playlists)
	}

	// Search podcast episodes
	if query.ContentType == nil || *query.ContentType == search.ContentTypePodcastEpisode {
		var episodes []*dbModel.PodcastEpisode
		if searchPattern != "" {
			s.DB(ctx).Where("title ILIKE ?", searchPattern).Limit(query.Limit).Find(&episodes)
		}
		result.PodcastEpisodes = episodes
		result.TotalCount += len(episodes)
	}

	// Search podcast series
	if query.ContentType == nil || *query.ContentType == search.ContentTypePodcastSeries {
		var series []*dbModel.PodcastSeries
		if searchPattern != "" {
			s.DB(ctx).Where("title ILIKE ?", searchPattern).Limit(query.Limit).Find(&series)
		}
		result.PodcastSeries = series
		result.TotalCount += len(series)
	}

	return result, nil
}
