package graphql // must match the package at the top of graphql/generated.go

import (
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalDateTimeISO(t time.Time) graphql.Marshaler      { return graphql.MarshalTime(t) }
func UnmarshalDateTimeISO(v interface{}) (time.Time, error) { return graphql.UnmarshalTime(v) }
