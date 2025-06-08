package user

import (
	"context"

	"vvinamp/graphql/model"
)

type key string

const (
	Key key = "user"
)

// Service is interface for UserService
type Service interface {
	Repository
	Register(ctx context.Context, user *model.User) (*model.User, error)
	ComparePassword(ctx context.Context, user *model.User, password string) bool
}

// ForContext is method to get user service from context
func ForContext(ctx context.Context) Service {
	service, ok := ctx.Value(Key).(Service)
	if !ok {
		panic("ctx passing is not contain user service")
	}
	return service
}
