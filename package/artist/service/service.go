package service

import (
	"context"
	dbModel "vvinamp/model"
	"vvinamp/package/artist/repository"
)

type ArtistService interface {
	GetArtistHeaderByID(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, bool, error)
	GetArtistScreenData(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, []*dbModel.Track, []*dbModel.Album, error)
}

type artistService struct {
	repo repository.ArtistRepository
}

func NewService(repo repository.ArtistRepository) ArtistService {
	return &artistService{repo: repo}
}

func (s *artistService) GetArtistHeaderByID(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, bool, error) {
	artist, err := s.repo.GetArtistHeaderByID(ctx, artistID)
	if err != nil {
		return nil, false, err
	}

	isFollowed := false
	if userID != nil {
		isFollowed, err = s.repo.IsUserFollowingArtist(ctx, *userID, artistID)
		if err != nil {
			// Log the error but continue with isFollowed as false
			// In a real application, you might want to handle this differently
		}
	}

	return artist, isFollowed, nil
}

func (s *artistService) GetArtistScreenData(ctx context.Context, artistID string, userID *string) (*dbModel.Artist, []*dbModel.Track, []*dbModel.Album, error) {
	// Get artist profile - this is critical and if it fails, we return an error
	artist, err := s.repo.GetArtistHeaderByID(ctx, artistID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get top songs (tracks) by stream count
	// If this fails, we still continue with empty slice
	topSongs, err := s.repo.GetTopSongsByArtist(ctx, artistID)
	if err != nil {
		// Log the error if needed, but continue execution
		topSongs = []*dbModel.Track{}
	}

	// Get top releases (albums) by popularity metric
	// If this fails, we still continue with empty slice
	topReleases, err := s.repo.GetTopReleasesByArtist(ctx, artistID)
	if err != nil {
		// Log the error if needed, but continue execution
		topReleases = []*dbModel.Album{}
	}

	return artist, topSongs, topReleases, nil
}