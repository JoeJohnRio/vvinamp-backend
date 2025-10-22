package history

import (
	"context"

	"vvinamp/graphql/model"
)

type key string

const (
	Key key = "history"
)

// Service is interface for UserService
type Service interface {
	Repository
	ListRecentHistory(ctx context.Context, limit int) ([]*model.ListeningHistory, error)
}

// ForContext is method to get user service from context
func ForContext(ctx context.Context) Service {
	service, ok := ctx.Value(Key).(Service)
	if !ok {
		panic("ctx passing is not contain history service")
	}
	return service
}
