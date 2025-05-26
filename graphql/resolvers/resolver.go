package resolvers

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"fmt"

	"github.com/JoeJohnRio/youtube-music/graphql"
	"github.com/JoeJohnRio/youtube-music/graphql/model"
	"github.com/JoeJohnRio/youtube-music/graphql/resolvers/album"
	"github.com/JoeJohnRio/youtube-music/internal/repository"
	"github.com/JoeJohnRio/youtube-music/pkg/jwt"
)

type Resolver struct {
	Repo *repository.Repository
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

// Mutation returns graphql1.MutationResolver implementation.
func (r *Resolver) Mutation() graphql.MutationResolver { return &mutationResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
