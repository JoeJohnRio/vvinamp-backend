package repository

import (
	"github.com/JoeJohnRio/youtube-music/internal/repository/album"
)

type Repository struct {
	Album album.AlbumRepository
}
