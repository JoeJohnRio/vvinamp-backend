package repository

import (
	"github.com/JoeJohnRio/youtube-music/internal/repository/album"
	"github.com/JoeJohnRio/youtube-music/internal/repository/genre"
)

type Repository struct {
	Album album.AlbumRepository
	Genre genre.GenreRepository
}
