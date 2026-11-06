package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"
	"vvinamp/graphql/model"
	dbModel "vvinamp/model"
)

// decodeCursorInt decodes a base64 cursor to an integer offset
func decodeCursorInt(cursor string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	var offset int
	_, err = fmt.Sscanf(string(decoded), "cursor:%d", &offset)
	return offset, err
}

// encodeCursorInt encodes an integer offset to a base64 cursor
func encodeCursorInt(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor:%d", offset)))
}

// convertPlaylistToGraphQL converts a database playlist to a GraphQL model
func convertPlaylistToGraphQL(ctx context.Context, r *Resolver, playlist *dbModel.Playlist) (*model.Playlist, error) {
	// Get creator
	userSvc := r.UserService
	creator, err := userSvc.GetByID(ctx, playlist.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch creator: %w", err)
	}

	return &model.Playlist{
		ID:           playlist.ID,
		Title:        playlist.Title,
		Description:  playlist.Description,
		Creator:      creator,
		CreatedAt:    playlist.CreatedAt.Format(time.RFC3339),
		UpdatedAt: func() *string {
			if playlist.UpdatedAt == nil {
				return nil
			}
			s := playlist.UpdatedAt.Format(time.RFC3339)
			return &s
		}(),
		IsPublic:     playlist.IsPublic,
		CoverImage:   playlist.CoverImage,
		// Tracks, SavedCount, TotalDuration, IsSavedByMe are resolved separately
	}, nil
}

// getPlaylistsByCreator is a helper function to get playlists by creator with pagination
func getPlaylistsByCreator(ctx context.Context, resolver *Resolver, creatorID string, first *int, after *string, last *int, before *string) (*model.PlaylistConnection, error) {
	playlistSvc := resolver.PlaylistService

	// Default limit
	limit := 20
	if first != nil {
		limit = *first
	}
	if last != nil {
		limit = *last
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Handle cursor-based pagination
	offset := 0
	if after != nil {
		cursorOffset, err := decodeCursorInt(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		offset = cursorOffset + 1
	}

	// Get playlists
	playlists, total, err := playlistSvc.GetPlaylistsByCreatorID(ctx, creatorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlists: %w", err)
	}

	// Convert to GraphQL models
	edges := make([]*model.PlaylistEdge, len(playlists))
	for i, playlist := range playlists {
		graphqlPlaylist, err := convertPlaylistToGraphQL(ctx, resolver, playlist)
		if err != nil {
			return nil, fmt.Errorf("failed to convert playlist: %w", err)
		}
		edges[i] = &model.PlaylistEdge{
			Cursor: encodeCursorInt(offset + i),
			Node:   graphqlPlaylist,
		}
	}

	// Build PageInfo
	pageInfo := &model.PageInfo{
		HasNextPage:     offset+len(playlists) < int(total),
		HasPreviousPage: offset > 0,
	}
	if len(edges) > 0 {
		startCursor := edges[0].Cursor
		endCursor := edges[len(edges)-1].Cursor
		pageInfo.StartCursor = &startCursor
		pageInfo.EndCursor = &endCursor
	}

	return &model.PlaylistConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: int(total),
	}, nil
}

// getMyPlaylistFeed is a helper function to get both created and saved playlists with pagination and sorting
func getMyPlaylistFeed(ctx context.Context, resolver *Resolver, userID string, first *int, after *string, last *int, before *string, sortOrder *model.PlaylistSortOrder) (*model.PlaylistConnection, error) {
	playlistSvc := resolver.PlaylistService

	// Default limit
	limit := 20
	if first != nil {
		limit = *first
	}
	if last != nil {
		limit = *last
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Handle cursor-based pagination
	offset := 0
	if after != nil {
		cursorOffset, err := decodeCursorInt(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		offset = cursorOffset + 1
	}

	// Get sort order string
	sortOrderStr := "RECENTS"
	if sortOrder != nil {
		sortOrderStr = string(*sortOrder)
	}

	// Get playlists
	playlists, total, err := playlistSvc.GetMyPlaylistFeed(ctx, userID, limit, offset, sortOrderStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist feed: %w", err)
	}

	// Convert to GraphQL models
	edges := make([]*model.PlaylistEdge, len(playlists))
	for i, playlist := range playlists {
		graphqlPlaylist, err := convertPlaylistToGraphQL(ctx, resolver, playlist)
		if err != nil {
			return nil, fmt.Errorf("failed to convert playlist: %w", err)
		}
		edges[i] = &model.PlaylistEdge{
			Cursor: encodeCursorInt(offset + i),
			Node:   graphqlPlaylist,
		}
	}

	// Build PageInfo
	pageInfo := &model.PageInfo{
		HasNextPage:     offset+len(playlists) < int(total),
		HasPreviousPage: offset > 0,
	}
	if len(edges) > 0 {
		startCursor := edges[0].Cursor
		endCursor := edges[len(edges)-1].Cursor
		pageInfo.StartCursor = &startCursor
		pageInfo.EndCursor = &endCursor
	}

	return &model.PlaylistConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: int(total),
	}, nil
}

// getPlaylistTracks is a helper function to get playlist tracks with pagination
func getPlaylistTracks(ctx context.Context, resolver *Resolver, playlistID string, sortBy *model.PlaylistTrackSortBy, first *int, after *string, last *int, before *string, songSvc interface {
	GetSingleTrack(ctx context.Context, id string) (*dbModel.Track, error)
	GetAlbumByID(ctx context.Context, id string) (*dbModel.Album, error)
}) (*model.PlaylistTrackConnection, error) {
	playlistSvc := resolver.PlaylistService

	// Default limit
	limit := 20
	if first != nil {
		limit = *first
	}
	if last != nil {
		limit = *last
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Handle cursor-based pagination
	offset := 0
	if after != nil {
		cursorOffset, err := decodeCursorInt(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		offset = cursorOffset + 1
	}

	// Get playlist songs
	playlistTracks, err := playlistSvc.GetPlaylistSongs(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}

	sortByValue := "RECENTLY_ADDED"
	if sortBy != nil {
		sortByValue = string(*sortBy)
	}

	switch sortByValue {
	case "TITLE", "ARTIST_NAME", "ALBUM_NAME":
		type sortKey struct {
			track        *dbModel.Track
			album        *dbModel.Album
			artistName   string
			albumName    string
			playlistTrack *dbModel.PlaylistTrack
		}
		keys := make([]sortKey, 0, len(playlistTracks))
		for _, pt := range playlistTracks {
			track, err := songSvc.GetSingleTrack(ctx, pt.TrackID)
			var album *dbModel.Album
			artistName := ""
			albumName := ""
			if err == nil && track != nil {
				album, _ = songSvc.GetAlbumByID(ctx, track.AlbumID.String())
				if album != nil {
					albumName = album.Title
					artistStruct, _ := resolver.ArtistService.GetArtistByID(ctx, album.ArtistID.String())
					if artistStruct != nil {
						artistName = artistStruct.Name
					}
				}
			}
			keys = append(keys, sortKey{track: track, album: album, artistName: artistName, albumName: albumName, playlistTrack: pt})
		}

		sort.SliceStable(keys, func(i, j int) bool {
			switch sortByValue {
			case "TITLE":
				if keys[i].track == nil {
					return false
				}
				if keys[j].track == nil {
					return true
				}
				return strings.ToLower(keys[i].track.Title) < strings.ToLower(keys[j].track.Title)
			case "ARTIST_NAME":
				return strings.ToLower(keys[i].artistName) < strings.ToLower(keys[j].artistName)
			case "ALBUM_NAME":
				return strings.ToLower(keys[i].albumName) < strings.ToLower(keys[j].albumName)
			default:
				return false
			}
		})

		sorted := make([]*dbModel.PlaylistTrack, len(keys))
		for i, k := range keys {
			sorted[i] = k.playlistTrack
		}
		playlistTracks = sorted
	default:
		sort.SliceStable(playlistTracks, func(i, j int) bool {
			return playlistTracks[i].AddedAt.After(playlistTracks[j].AddedAt)
		})
	}

	// Apply pagination manually (since service returns all tracks)
	total := len(playlistTracks)
	if offset >= total {
		playlistTracks = []*dbModel.PlaylistTrack{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		playlistTracks = playlistTracks[offset:end]
	}

	// Convert to GraphQL model with track details
	edges := make([]*model.PlaylistTrackEdge, len(playlistTracks))
	for i, pt := range playlistTracks {
		// Get track details
		track, err := songSvc.GetSingleTrack(ctx, pt.TrackID)
		if err != nil {
			// Skip tracks that don't exist
			continue
		}

		// Get album cover and artist
		album, err := songSvc.GetAlbumByID(ctx, track.AlbumID.String())
		albumCoverURL := ""
		artistName := "Unknown Artist"
		var artistBio *string
		var artistProfileImage *string
		var artistMonthlyListeners *int

		if err == nil && album != nil {
			albumCoverURL = album.CoverArt
			
			// Get artist name from album's artist ID
			artist, artistErr := resolver.ArtistService.GetArtistByID(ctx, album.ArtistID.String())
			if artistErr == nil && artist != nil {
				artistName = artist.Name
				artistBio = artist.Bio
				artistProfileImage = artist.ProfileImage
				monthly := int(artist.MonthlyListeners)
				artistMonthlyListeners = &monthly
			}
		}

		graphqlTrack := &model.PlaylistTrack{
			Track: &model.Track{
				ID:                     track.ID.String(),
				Title:                  track.Title,
				Duration:               track.Duration,
				StreamCount:            track.StreamCount,
				AudioFileURL:           track.AudioFileURL,
				AlbumCoverURL:          albumCoverURL,
				ArtistName:             artistName,
				Explicit:               track.Explicit,
				Lyrics:                 &track.Lyrics,
				ReleaseDate:            track.ReleaseDate,
				ArtistBio:              artistBio,
				ArtistProfileImageURL:  artistProfileImage,
				ArtistMonthlyListeners: artistMonthlyListeners,
				YoutubeID:              track.YoutubeID,
			},
			Position: pt.Position,
			AddedAt:  pt.AddedAt.Format(time.RFC3339),
		}

		edges[i] = &model.PlaylistTrackEdge{
			Cursor: encodeCursorInt(offset + i),
			Node:   graphqlTrack,
		}
	}

	// Build PageInfo
	pageInfo := &model.PageInfo{
		HasNextPage:     offset+len(edges) < total,
		HasPreviousPage: offset > 0,
	}
	if len(edges) > 0 {
		startCursor := edges[0].Cursor
		endCursor := edges[len(edges)-1].Cursor
		pageInfo.StartCursor = &startCursor
		pageInfo.EndCursor = &endCursor
	}

	return &model.PlaylistTrackConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: total,
	}, nil
}
