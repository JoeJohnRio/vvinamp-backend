package repository

import (
	"time"

	"vvinamp/database/mysql"
	"vvinamp/graphql/model"
	"vvinamp/package/user"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

// sqlRepository contains all the interactions
// with the user collection stored in sql.
type sqlRepository struct {
	logger *zap.Logger
}

// SQLRepositoryTarget is `fx.In` struct for `fx` to get all dependency to create `UserSQLRepository`
type SQLRepositoryTarget struct {
	fx.In
	Connection *mysql.Connection
	Logger     *zap.Logger
}

// NewSQLRepository is UserRepository's constructor
func NewSQLRepository(target SQLRepositoryTarget) (user.Repository, error) {
	err := target.Connection.Client().Debug().AutoMigrate(&model.User{})
	if err != nil {
		return nil, err
	}
	return &sqlRepository{
		logger: target.Logger,
	}, err
}

// DB method extract database client from context
func (m *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return mysql.ForContext(ctx)
}

// GetAll returns all the Users stored in the database.
func (m *sqlRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	err := m.DB(ctx).Find(&users).Error
	return users, err
}

// GetByID returns one User which is matched by input ID from the database.
func (m *sqlRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := m.DB(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail returns one user which is matched by email
func (m *sqlRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := m.DB(ctx).Where("email = ?", email).First(&user).Error

	return &user, err
}

// Create will insert new user into database
func (m *sqlRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	hashedPassword, err := hashPassword(user.PasswordHash)
	free := string(model.SubscriptionFree)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = hashedPassword
	user.SubscriptionType = &free
	user.JoinDate = time.Now().String()
	err = m.DB(ctx).Debug().Create(user).Error
	return user, err
}

// Update will update user by id
func (m *sqlRepository) UpdateSubscription(ctx context.Context, id string, update *model.User) (*model.User, error) {
	// Get existing user
	user, err := m.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only update subscription type if it's provided
	if update.SubscriptionType != nil {
		user.SubscriptionType = update.SubscriptionType
	}

	// Save updated user to DB
	err = m.DB(ctx).Save(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Delete will remove all users
func (m *sqlRepository) Delete(ctx context.Context) error {
	return m.DB(ctx).Delete(&model.User{}).Error
}

// DeleteByID will remove user by id from database
func (m *sqlRepository) DeleteByID(ctx context.Context, id string) error {
	return m.DB(ctx).Delete(ctx, model.User{ID: id}).Error
}
