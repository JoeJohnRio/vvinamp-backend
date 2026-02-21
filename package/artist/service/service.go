package service

import (
	"context"
	dbModel "vvinamp/model"
	"vvinamp/model/dto"
	"vvinamp/package/artist/repository"
)

type key string

const (
	Key key = "artist"
)

type ArtistService interface {
	GetArtistHeaderByID(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, bool, error)
	GetArtistScreenData(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, []*dto.TrackWithDetails, []*dbModel.Album, error)
	GetArtistByID(ctx context.Context, artistID string) (*dbModel.Artist, error)
}

type artistService struct {
	repo repository.ArtistRepository
}

func NewService(repo repository.ArtistRepository) ArtistService {
	return &artistService{repo: repo}
}

func (s *artistService) GetArtistByID(ctx context.Context, artistID string) (*dbModel.Artist, error) {
	return s.repo.GetArtistHeaderByID(ctx, artistID, nil)
}

func (s *artistService) GetArtistHeaderByID(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, bool, error) {
	artist, err := s.repo.GetArtistHeaderByID(ctx, artistID, userID)
	if err != nil {
		return nil, false, err
	}

	isFollowed := false
	if userID != nil {
		isFollowed, err = s.repo.IsUserFollowingArtist(ctx, *userID, artistID)
		if err != nil {
		}
	}

	return artist, isFollowed, nil
}

func (s *artistService) GetArtistScreenData(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, []*dto.TrackWithDetails, []*dbModel.Album, error) {
	artist, err := s.repo.GetArtistHeaderByID(ctx, artistID, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	topSongs, err := s.repo.GetTopSongsByArtist(ctx, artistID, userID)
	if err != nil {
		topSongs = []*dto.TrackWithDetails{}
	}

	topReleases, err := s.repo.GetTopReleasesByArtist(ctx, artistID, userID)
	if err != nil {
		topReleases = []*dbModel.Album{}
	}

	return artist, topSongs, topReleases, nil
}
