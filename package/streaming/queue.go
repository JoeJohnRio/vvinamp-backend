package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"vvinamp/graphql/model"

	"github.com/redis/go-redis/v9"
)

func getQueueKey(userId string) string {
	return fmt.Sprintf("player_queue:%s", userId)
}
func (s *Service) GetPlayerQueue(ctx context.Context, userId string) (*model.PlayerQueue, error) {
	key := getQueueKey(userId)
	val, err := s.Redis.Get(ctx, key).Result()

	// If no queue exists in Redis, return a fresh one
	if err == redis.Nil {
		return &model.PlayerQueue{
			UpcomingTrack: []*model.TrackMinimal{},
			PreviousTrack: []*model.TrackMinimal{}, // Initialize this too
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get queue from redis: %w", err)
	}

	var queue model.PlayerQueue
	if err := json.Unmarshal([]byte(val), &queue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal queue: %w", err)
	}

	// Ensure slices are initialized if they unmarshalled as nil
	if queue.UpcomingTrack == nil {
		queue.UpcomingTrack = []*model.TrackMinimal{}
	}
	if queue.PreviousTrack == nil {
		queue.PreviousTrack = []*model.TrackMinimal{}
	}

	return &queue, nil
}

func (s *Service) savePlayerQueue(ctx context.Context, userId string, queue *model.PlayerQueue) error {
	key := getQueueKey(userId)
	data, err := json.Marshal(queue)
	if err != nil {
		return fmt.Errorf("failed to marshal queue: %w", err)
	}

	return s.Redis.Set(ctx, key, string(data), 0).Err()
}

func (s *Service) SetPlayerQueue(ctx context.Context, userId string, queue *model.PlayerQueue) error {
	return s.savePlayerQueue(ctx, userId, queue)
}

func (s *Service) AddToQueue(ctx context.Context, userId string, trackIds []string) (*model.PlayerQueue, error) {
	queue, err := s.GetPlayerQueue(ctx, userId)
	if err != nil {
		return nil, err
	}

	tracksWithDetails, err := s.Repo.GetTracksWithDetails(ctx, trackIds)
	if err != nil {
		return nil, err
	}

	for _, t := range tracksWithDetails {
		title := t.Title
		artistName := t.ArtistName
		coverUrl := t.AlbumCoverURL

		minimalTrack := &model.TrackMinimal{
			TrackID:                t.ID.String(),
			Title:                  &title,
			Artist:                 &artistName,
			CoverImageURL:          &coverUrl,
			ArtistBio:              t.ArtistBio,
			ArtistProfileImageURL:  t.ArtistProfileImageUrl,
			ArtistMonthlyListeners: t.ArtistMonthlyListeners,
		}

		queue.UpcomingTrack = append(queue.UpcomingTrack, minimalTrack)
	}
	// 3. Save to Redis
	if err := s.savePlayerQueue(ctx, userId, queue); err != nil {
		return nil, err
	}

	return queue, nil
}

func (s *Service) ClearQueue(ctx context.Context, userId string) (*model.PlayerQueue, error) {
	queue := &model.PlayerQueue{
		UpcomingTrack: []*model.TrackMinimal{},
	}

	if err := s.savePlayerQueue(ctx, userId, queue); err != nil {
		return nil, err
	}

	return queue, nil
}

func (s *Service) SkipToPrevious(ctx context.Context, userId string) (*model.PlayerQueue, error) {
	queue, err := s.GetPlayerQueue(ctx, userId)
	if err != nil {
		return nil, err
	}

	if len(queue.PreviousTrack) > 0 {
		// 1. If there's a current track, push it back to the front of upcoming
		if queue.CurrentTrack != nil {
			queue.UpcomingTrack = append(
				[]*model.TrackMinimal{queue.CurrentTrack},
				queue.UpcomingTrack...,
			)
		}

		// 2. Pop the last item from previous to be the new current
		lastIndex := len(queue.PreviousTrack) - 1
		prevTrack := queue.PreviousTrack[lastIndex]
		queue.CurrentTrack = prevTrack
		queue.PreviousTrack = queue.PreviousTrack[:lastIndex]
	}

	if err := s.savePlayerQueue(ctx, userId, queue); err != nil {
		return nil, err
	}
	return queue, nil
}

func (s *Service) SkipToNext(ctx context.Context, userId string) (*model.PlayerQueue, error) {
	queue, err := s.GetPlayerQueue(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 1. Add current track to the end of the previous track history
	if queue.CurrentTrack != nil {
		queue.PreviousTrack = append(queue.PreviousTrack, queue.CurrentTrack)
	}

	// 2. Shift the first upcoming track to be the new current track
	if len(queue.UpcomingTrack) > 0 {
		nextTrack := queue.UpcomingTrack[0]
		queue.CurrentTrack = nextTrack // nextTrack is already a pointer, no '&' needed
		queue.UpcomingTrack = queue.UpcomingTrack[1:]
	} else {
		queue.CurrentTrack = nil // Optional: clears the player if nothing is next
		queue.UpcomingTrack = nil
	}

	// Save to Redis
	if err := s.savePlayerQueue(ctx, userId, queue); err != nil {
		return nil, err
	}

	return queue, nil
}
