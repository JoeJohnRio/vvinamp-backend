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
	if err == redis.Nil {
		return &model.PlayerQueue{
			UpcomingTrackIds: []string{},
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get queue from redis: %w", err)
	}

	var queue model.PlayerQueue
	if err := json.Unmarshal([]byte(val), &queue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal queue: %w", err)
	}

	if queue.UpcomingTrackIds == nil {
		queue.UpcomingTrackIds = []string{}
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

func (s *Service) AddToQueue(ctx context.Context, userId string, trackIds []string) (*model.PlayerQueue, error) {
	queue, err := s.GetPlayerQueue(ctx, userId)
	if err != nil {
		return nil, err
	}

	if queue.CurrentTrackID == nil && len(trackIds) > 0 {
		currentTrack := trackIds[0]
		queue.CurrentTrackID = &currentTrack
		if len(trackIds) > 1 {
			queue.UpcomingTrackIds = append(queue.UpcomingTrackIds, trackIds[1:]...)
		}
	} else {
		queue.UpcomingTrackIds = append(queue.UpcomingTrackIds, trackIds...)
	}

	if err := s.savePlayerQueue(ctx, userId, queue); err != nil {
		return nil, err
	}

	return queue, nil
}

func (s *Service) ClearQueue(ctx context.Context, userId string) (*model.PlayerQueue, error) {
	queue := &model.PlayerQueue{
		UpcomingTrackIds: []string{},
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

	if len(queue.UpcomingTrackIds) > 0 {
		nextTrack := queue.UpcomingTrackIds[0]
		queue.CurrentTrackID = &nextTrack
		queue.UpcomingTrackIds = queue.UpcomingTrackIds[1:]
	} else {
		queue.CurrentTrackID = nil
	}

	if err := s.savePlayerQueue(ctx, userId, queue); err != nil {
		return nil, err
	}

	return queue, nil
}
