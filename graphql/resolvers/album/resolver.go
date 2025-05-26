package album

import (
	"context"

	"github.com/JoeJohnRio/youtube-music/graphql/model"
	"github.com/JoeJohnRio/youtube-music/internal/repository/album"
)

type Resolver struct {
	repo album.AlbumRepository
}

func (r *Resolver) Album(ctx context.Context, id string) (*model.Album, error) {
	dbAlbum, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToGraphQL(dbAlbum), nil
}
