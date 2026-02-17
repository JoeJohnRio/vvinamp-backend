package search

import (
	"context"
)

type key string

const (
	Key key = "search"
)

// Service is interface for search Service
type Service interface {
	Repository
	Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
}

// ForContext is method to get search service from context
func ForContext(ctx context.Context) Service {
	service, ok := ctx.Value(Key).(Service)
	if !ok {
		panic("ctx passing is not contain search service")
	}
	return service
}
