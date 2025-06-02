package repository

import (
	"spotify-clone/internal/repository/album"
	"spotify-clone/internal/repository/genre"
	"spotify-clone/internal/repository/user"
)

type Repository struct {
	Album album.AlbumRepository
	Genre genre.GenreRepository
	User  user.UserRepository
}
