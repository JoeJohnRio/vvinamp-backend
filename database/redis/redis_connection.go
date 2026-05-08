package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Target struct {
	fx.In
	RedisURL string `name:"redis_url"`
	Lc       fx.Lifecycle
	Logger   *zap.Logger
}

type Connection struct {
	client *redis.Client
}

func New(t Target) (*Connection, error) {
	if t.RedisURL == "" {
		return nil, fmt.Errorf("redis_url is empty")
	}

	opt, err := redis.ParseURL(t.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opt)

	t.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("redis ping: %w", err)
			}
			t.Logger.Info("Redis connected", zap.String("url", t.RedisURL))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			t.Logger.Info("Redis disconnect")
			return client.Close()
		},
	})

	return &Connection{client: client}, nil
}

func (c *Connection) Client() (*redis.Client, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("redis connection not initialized")
	}
	return c.client, nil
}
