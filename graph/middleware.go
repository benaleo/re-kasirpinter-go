package graph

import (
	"context"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const userCtxKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login and register mutations
		if r.URL.Path == "/query" {
			// This is a GraphQL request, check the operation name
			if r.Method == "POST" {
				// In a real app, you'd parse the request body to check the operation name
				// For simplicity, we'll just check the auth header for now
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					// No auth header, let the resolver handle it
					next.ServeHTTP(w, r)
					return
				}

				// Extract the token from the Authorization header
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString == authHeader {
					// No Bearer token found
					next.ServeHTTP(w, r)
					return
				}

				// Validate the token
				claims, err := validateJWT(tokenString)
				if err != nil {
					// Invalid token, let the resolver handle it
					next.ServeHTTP(w, r)
					return
				}

				// Add user info to context
				ctx := context.WithValue(r.Context(), userCtxKey, claims)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *Claims {
	raw, _ := ctx.Value(userCtxKey).(*Claims)
	return raw
}

// AuthDirective adds authentication to a field.
// Only accepts full login tokens (purpose == "login").
// Password-reset tokens are restricted to userUpdatePassword only.
func AuthDirective(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	user := ForContext(ctx)
	if user == nil {
		return nil, &gqlerror.Error{
			Message: "Access denied. You must be logged in.",
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}
	if user.Purpose != "login" && user.Purpose != "" {
		return nil, &gqlerror.Error{
			Message: "Access denied. Token is not authorized for this operation.",
			Extensions: map[string]interface{}{
				"code": "FORBIDDEN",
			},
		}
	}

	return next(ctx)
}
