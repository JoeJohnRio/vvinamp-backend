package artist

import (
	"context"
	"vvinamp/package/artist/service"
)

// ForContext is method to get artist service from context
func ForContext(ctx context.Context) service.ArtistService {
	s, ok := ctx.Value(service.Key).(service.ArtistService)
	if !ok {
		panic("ctx passing is not contain artist service")
	}
	return s
}
