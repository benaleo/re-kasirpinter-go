package graph

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
)

// AuthDirective is the implementation of the @auth directive
func AuthDirective(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	// Check if user is authenticated
	userClaims := ForContext(ctx)
	if userClaims == nil {
		return nil, errors.New("unauthorized: authentication required")
	}

	// Continue to the next resolver
	return next(ctx)
}
