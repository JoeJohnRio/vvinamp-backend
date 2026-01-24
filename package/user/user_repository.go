package user

import (
	"context"

	"vvinamp/graphql/model"
)

// Repository is interface for UserRepository
type Repository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}
