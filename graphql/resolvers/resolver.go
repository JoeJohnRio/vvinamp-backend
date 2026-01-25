package resolvers

import (
	"context"
	userService "vvinamp/package/user"
	historyService "vvinamp/package/history"
	songService "vvinamp/package/song"
	artistService "vvinamp/package/artist/service"
	"vvinamp/constant"
	"vvinamp/internal/pkg/session"
)

type Resolver struct {
	ArtistService artistService.ArtistService
	UserService   userService.Service
	HistoryService historyService.Service
	SongService   songService.Service
}

func NewResolver(
	artistService artistService.ArtistService,
	userService userService.Service,
	historyService historyService.Service,
	songService songService.Service,
) *Resolver {
	return &Resolver{
		ArtistService:  artistService,
		UserService:    userService,
		HistoryService: historyService,
		SongService:    songService,
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