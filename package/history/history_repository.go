package history

import (
	"context"
	"vvinamp/graphql/model"
)

type Repository interface {
	GetRecentHistoryByUser(ctx context.Context, limit int) ([]*model.ListeningHistory, error)

	// ... other methods as needed
}
