package history

import (
	"context"
	dbModel "vvinamp/model"
)

type Repository interface {
	GetRecentHistoryByUser(ctx context.Context, limit int) ([]*dbModel.ListeningHistory, error)

	// ... other methods as needed
}
