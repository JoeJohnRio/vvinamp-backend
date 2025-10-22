package resolvers

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"vvinamp/constant"
	"vvinamp/controller/auth"
	graphql1 "vvinamp/graphql"
	"vvinamp/graphql/model"
	"vvinamp/package/history"
	"vvinamp/package/user"
)

type Resolver struct{}

// Artist is the resolver for the artist field.
func (r *albumResolver) Artist(ctx context.Context, obj *model.Album) (*model.Artist, error) {
	panic("not implemented")
}

// Tracks is the resolver for the tracks field.
func (r *albumResolver) Tracks(ctx context.Context, obj *model.Album) ([]*model.Track, error) {
	panic("not implemented")
}

// Albums is the resolver for the albums field.
func (r *artistResolver) Albums(ctx context.Context, obj *model.Artist) ([]*model.Album, error) {
	panic("not implemented")
}

// Tracks is the resolver for the tracks field.
func (r *artistResolver) Tracks(ctx context.Context, obj *model.Artist) ([]*model.Track, error) {
	panic("not implemented")
}

// Albums is the resolver for the albums field.
func (r *genreResolver) Albums(ctx context.Context, obj *model.Genre) ([]*model.Album, error) {
	panic("not implemented")
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginPayload, error) {
	// This should be handled by the authentication middleware
	// Return appropriate response or implement proper token refresh logic
	return nil, errors.New("refresh token handling should be implemented in middleware")
}

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.RegisterPayload, error) {
	a, err := user.ForContext(ctx).Register(ctx, &model.User{
		Email:        input.Email,
		PasswordHash: input.Password,
		Username:     input.Username,
	})
	if err != nil {
		return nil, err
	}

	session := ctx.Value(constant.Session).(*string)
	*session = string(a.ID)

	return &model.RegisterPayload{a}, err
}

// Creator is the resolver for the creator field.
func (r *playlistResolver) Creator(ctx context.Context, obj *model.Playlist) (*model.User, error) {
	panic("not implemented")
}

// Tracks is the resolver for the tracks field.
func (r *playlistResolver) Tracks(ctx context.Context, obj *model.Playlist) ([]*model.PlaylistTrack, error) {
	panic("not implemented")
}

// Login is the resolver for the login field.
func (r *queryResolver) Login(ctx context.Context, email string, password string) (*model.LoginPayload, error) {
	// Assume user.ForContext(ctx) returns a service to authenticate
	userSvc := user.ForContext(ctx)
	u, err := userSvc.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}
	var accessToken, refreshToken string
	log.Println("joel123 Login")
	// If you're using context to pass auth writer for middleware/token things:
	if arw, ok := ctx.Value("authResponseWriter").(AuthResponseWriter); ok {
		arw.SetSessionToResolver(u.ID)
		// At this point, the authentication.go middleware will set cookies

		if arw.GetAuth() != nil {
			// Generate the tokens (logic centralized)
			accessToken, refreshToken, err = arw.GetAuth().GenerateTokens(u.ID)
			if err != nil {
				return nil, fmt.Errorf("internal: failed to generate tokens: %w", err)
			}
		}
	}

	// Typically, JWT + Cookie logic is handled by Gin or your HTTP middleware.
	// In GraphQL response, deliver basic session data:
	return &model.LoginPayload{
		Session: &model.AuthSession{
			IsLoggedIn:   true,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

// ListRecentHistory is the resolver for the listRecentHistory field.
func (r *queryResolver) ListRecentHistory(ctx context.Context, limit int) ([]*model.ListeningHistory, error) {
	historySvc := history.ForContext(ctx)
	a, err := historySvc.ListRecentHistory(ctx, limit)

	if err != nil {
		return nil, err
	}

	return a, err
}

// Logout is the resolver for the logout field.
func (r *queryResolver) Logout(ctx context.Context) (bool, error) {
	if arw, ok := ctx.Value("authResponseWriter").(AuthResponseWriter); ok {
		arw.SetSessionToResolver("") // Empty string signals logout
	}
	return true, nil
}

// Album is the resolver for the album field.
func (r *trackResolver) Album(ctx context.Context, obj *model.Track) (*model.Album, error) {
	panic("not implemented")
}

// Artists is the resolver for the artists field.
func (r *trackResolver) Artists(ctx context.Context, obj *model.Track) ([]*model.TrackArtist, error) {
	panic("not implemented")
}

// Playlists is the resolver for the playlists field.
func (r *userResolver) Playlists(ctx context.Context, obj *model.User) ([]*model.Playlist, error) {
	panic("not implemented")
}

// ListeningHistory is the resolver for the listeningHistory field.
func (r *userResolver) ListeningHistory(ctx context.Context, obj *model.User) ([]*model.ListeningHistory, error) {
	panic("not implemented")
}

// LikedTracks is the resolver for the likedTracks field.
func (r *userResolver) LikedTracks(ctx context.Context, obj *model.User) ([]*model.UserLike, error) {
	panic("not implemented")
}

// FollowedArtists is the resolver for the followedArtists field.
func (r *userResolver) FollowedArtists(ctx context.Context, obj *model.User) ([]*model.UserFollow, error) {
	panic("not implemented")
}

// Album returns graphql1.AlbumResolver implementation.
func (r *Resolver) Album() graphql1.AlbumResolver { return &albumResolver{r} }

// Artist returns graphql1.ArtistResolver implementation.
func (r *Resolver) Artist() graphql1.ArtistResolver { return &artistResolver{r} }

// Genre returns graphql1.GenreResolver implementation.
func (r *Resolver) Genre() graphql1.GenreResolver { return &genreResolver{r} }

// Mutation returns graphql1.MutationResolver implementation.
func (r *Resolver) Mutation() graphql1.MutationResolver { return &mutationResolver{r} }

// Playlist returns graphql1.PlaylistResolver implementation.
func (r *Resolver) Playlist() graphql1.PlaylistResolver { return &playlistResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql1.QueryResolver { return &queryResolver{r} }

// Track returns graphql1.TrackResolver implementation.
func (r *Resolver) Track() graphql1.TrackResolver { return &trackResolver{r} }

// User returns graphql1.UserResolver implementation.
func (r *Resolver) User() graphql1.UserResolver { return &userResolver{r} }

type albumResolver struct{ *Resolver }
type artistResolver struct{ *Resolver }
type genreResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type playlistResolver struct{ *Resolver }
type queryResolver struct {
	*Resolver
}
type trackResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

type AuthResponseWriter interface {
	SetSessionToResolver(userID string)
	GetAuth() *auth.Auth
}

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
