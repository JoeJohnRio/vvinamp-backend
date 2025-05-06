package graph

import (
	"github.com/JoeJohnRio/youtube-music/internal/repository/album"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AlbumRepo album.Repository
}

// Add this to ensure we implement all required interfaces
// var _ generated.ResolverRoot = (*Resolver)(nil)
