package resolvers

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"fmt"

	"spotify-clone/graphql"
	"spotify-clone/graphql/model"
	"spotify-clone/graphql/resolvers/album"
	"spotify-clone/graphql/resolvers/genre"
	"spotify-clone/graphql/resolvers/user"
	"spotify-clone/internal/repository"
	"spotify-clone/pkg/jwt"
)

type Resolver struct {
	Repo *repository.Repository
}

// Tracks is the resolver for the tracks field.
func (r *albumResolver) Tracks(ctx context.Context, obj *model.Album) ([]*model.Track, error) {
	panic("not implemented")
}

// Tracks is the resolver for the tracks field.
func (r *artistResolver) Tracks(ctx context.Context, obj *model.Artist) ([]*model.Track, error) {
	panic("not implemented")
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	username, err := jwt.ParseToken(input.Token)
	if err != nil {
		return "", fmt.Errorf("access denied")
	}
	token, err := jwt.GenerateToken(username)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetAlbum is the resolver for the getAlbum field.
func (r *queryResolver) GetAlbum(ctx context.Context, id string) (*model.Album, error) {
	// 1. Call repository
	dbAlbum, err := r.Repo.Album.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("album not found: %w", err)
	}
	// 2. Convert to GraphQL model
	return album.ToGraphQL(dbAlbum), nil
}

// GetAllGenres is the resolver for the getAllGenres field.
func (r *queryResolver) GetAllGenres(ctx context.Context) ([]*model.Genre, error) {
	dbGenre, err := r.Repo.Genre.GetAllGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("genre is empty: %w", err)
	}
	// 2. Convert to GraphQL model
	return genre.ToGraphQL(dbGenre), nil
}

// GetQuickPicks is the resolver for the getQuickPicks field.
func (r *queryResolver) GetQuickPicks(ctx context.Context, userID string) ([]*model.QuickPick, error) {
	picks, err := r.Repo.User.GetUserQuickPicks(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToGraphQLQuickPicks(picks), nil
}

// Album returns graphql1.AlbumResolver implementation.
func (r *Resolver) Album() graphql.AlbumResolver { return &albumResolver{r} }

// Artist returns graphql1.ArtistResolver implementation.
func (r *Resolver) Artist() graphql.ArtistResolver { return &artistResolver{r} }

// Mutation returns graphql1.MutationResolver implementation.
func (r *Resolver) Mutation() graphql.MutationResolver { return &mutationResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql.QueryResolver { return &queryResolver{r} }

type albumResolver struct{ *Resolver }
type artistResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
