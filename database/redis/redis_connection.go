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
		t.Logger.Warn("REDIS_URL is empty; caching disabled")
		return &Connection{client: nil}, nil
	}

	opt, err := redis.ParseURL(t.RedisURL)
	if err != nil {
		t.Logger.Warn("invalid redis url; caching disabled", zap.Error(err))
		return &Connection{client: nil}, nil
	}

	conn := &Connection{client: nil}
	client := redis.NewClient(opt)

	t.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				t.Logger.Warn("redis ping failed; caching disabled", zap.Error(err))
				return nil
			}
			conn.client = client
			t.Logger.Info("Redis connected")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if conn.client != nil {
				t.Logger.Info("Redis disconnect")
				return conn.client.Close()
			}
			return nil
		},
	})

	return conn, nil
}

func (c *Connection) Client() (*redis.Client, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("redis connection not initialized")
	}
	return c.client, nil
}
