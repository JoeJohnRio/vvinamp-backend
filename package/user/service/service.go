package service

import (
	"context"

	"vvinamp/graphql/model"
	"vvinamp/package/user"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	user.Repository
	logger *zap.Logger
}

type Target struct {
	fx.In
	Repository user.Repository
	Logger     *zap.Logger
}

func NewService(target Target) user.Service {
	return &userService{
		Repository: target.Repository,
		logger:     target.Logger,
	}
}

// Register new user and stored it to database with hashed password
func (m *userService) Register(ctx context.Context, user *model.User) (*model.User, error) {
	return m.Repository.Create(ctx, user)
}

// ComparePassword to compare cryped password
func (m *userService) ComparePassword(ctx context.Context, user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
