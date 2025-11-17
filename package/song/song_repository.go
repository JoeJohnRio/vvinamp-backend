package song

import (
	"context"
	dbModel "vvinamp/model"
)

type Repository interface {
	GetTrackByID(ctx context.Context, trackId string) (*dbModel.Track, error)

	// ... other methods as needed
}
