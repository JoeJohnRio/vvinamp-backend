package resolvers

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"log"
	"vvinamp/constant"
	graphql1 "vvinamp/graphql"
	"vvinamp/graphql/model"
	"vvinamp/package/user"
)

type Resolver struct{}

// Tracks is the resolver for the tracks field.
func (r *albumResolver) Tracks(ctx context.Context, obj *model.Album) ([]*model.Track, error) {
	panic("not implemented")
}

// Tracks is the resolver for the tracks field.
func (r *artistResolver) Tracks(ctx context.Context, obj *model.Artist) ([]*model.Track, error) {
	panic("not implemented")
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*model.LoginPayload, error) {
	panic("not implemented")
}

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.RegisterPayload, error) {
	log.Println("Joel123 Register Context", user.ForContext(ctx))
	a, err := user.ForContext(ctx).Create(ctx, &model.User{
		Email:        input.Email,
		PasswordHash: input.Password,
		Username:     input.Name,
	})
	if err != nil {
		return nil, err
	}

	session := ctx.Value(constant.Session).(*string)
	*session = string(a.ID)

	return &model.RegisterPayload{a}, err
}

// GetAlbum is the resolver for the getAlbum field.
func (r *queryResolver) GetAlbum(ctx context.Context, id string) (*model.Album, error) {
	// // 1. Call repository
	// dbAlbum, err := r.Repo.Album.GetByID(ctx, id)
	// if err != nil {
	// 	return nil, fmt.Errorf("album not found: %w", err)
	// }
	// // 2. Convert to GraphQL model
	// return album.ToGraphQL(dbAlbum), nil

	panic("not implemented")
}

// GetAllGenres is the resolver for the getAllGenres field.
func (r *queryResolver) GetAllGenres(ctx context.Context) ([]*model.Genre, error) {
	// dbGenre, err := r.Repo.Genre.GetAllGenres(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("genre is empty: %w", err)
	// }
	// // 2. Convert to GraphQL model
	// return genre.ToGraphQL(dbGenre), nil

	panic("not implemented")
}

// GetQuickPicks is the resolver for the getQuickPicks field.
func (r *queryResolver) GetQuickPicks(ctx context.Context, userID string) ([]*model.QuickPick, error) {
	// picks, err := r.Repo.User.GetUserQuickPicks(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }
	// return user.ToGraphQLQuickPicks(picks), nil

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

// Mutation returns graphql1.MutationResolver implementation.
func (r *Resolver) Mutation() graphql1.MutationResolver { return &mutationResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql1.QueryResolver { return &queryResolver{r} }

// User returns graphql1.UserResolver implementation.
func (r *Resolver) User() graphql1.UserResolver { return &userResolver{r} }

type albumResolver struct{ *Resolver }
type artistResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
