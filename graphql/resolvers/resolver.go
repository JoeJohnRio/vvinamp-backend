package resolvers

import (
	"context"
	"vvinamp/constant"
	"vvinamp/internal/pkg/session"
	youtubeService "vvinamp/internal/services/youtube"
	artistService "vvinamp/package/artist/service"
	historyService "vvinamp/package/history"
	playlistService "vvinamp/package/playlist"
	songService "vvinamp/package/song"
	streamingService "vvinamp/package/streaming"
	userService "vvinamp/package/user"
)

type Resolver struct {
	ArtistService          artistService.ArtistService
	UserService            userService.Service
	HistoryService         historyService.Service
	SongService            songService.Service
	StreamingService       *streamingService.Service
	PlaylistService        playlistService.Service
	YoutubeSearchService   *youtubeService.SearchService
	YoutubeDownloadService *youtubeService.DownloadService
}

func NewResolver(
	artistService artistService.ArtistService,
	userService userService.Service,
	historyService historyService.Service,
	songService songService.Service,
	streamingService *streamingService.Service,
	playlistService playlistService.Service,
	youtubeSearchService *youtubeService.SearchService,
	youtubeDownloadService *youtubeService.DownloadService,
) *Resolver {
	return &Resolver{
		ArtistService:          artistService,
		UserService:            userService,
		HistoryService:         historyService,
		SongService:            songService,
		StreamingService:       streamingService,
		PlaylistService:        playlistService,
		YoutubeSearchService:   youtubeSearchService,
		YoutubeDownloadService: youtubeDownloadService,
	}
}

// Helper function to get user ID from context if needed
func getUserIDFromContext(ctx context.Context) *string {
	// Extract the user ID from the context using the constant.Session key
	// This key is set by the authentication middleware
	userIDInterface := ctx.Value(constant.Session)
	if userIDInterface != nil {
		if userID, ok := userIDInterface.(string); ok {
			return &userID
		}
	}

	// Return nil if no authenticated user is found in the context
	return nil
}

// Helper function to validate session and get user ID with proper error handling
func requireValidSession(ctx context.Context) (string, error) {
	sessionHelper := session.NewGraphQLSessionHelper()
	return sessionHelper.RequireSessionForResolver(ctx)
}
