package service

import (
	"context"
	"errors"

	"vvinamp/graphql/model"
	"vvinamp/package/user"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

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

func (m *userService) Login(ctx context.Context, email, password string) (*model.User, error) {
	user, err := m.Repository.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials // avoid leaking DB errors
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}

	// at this point credentials are valid
	m.logger.Info("User logged in", zap.String("email", email))

	return user, nil
}

// ComparePassword to compare cryped password
func (m *userService) ComparePassword(ctx context.Context, user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
