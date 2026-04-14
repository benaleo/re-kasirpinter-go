package graph

import (
	"context"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

const userCtxKey = "user"
const clientInfoCtxKey = "clientInfo"
const tokenCtxKey = "token"

type ClientInfo struct {
	IP      string
	Browser string
	OS      string
}

func AuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client information
			clientInfo := &ClientInfo{
				IP:      getClientIP(r),
				Browser: parseUserAgent(r.UserAgent()),
				OS:      parseOS(r.UserAgent()),
			}

			// Add client info to context
			ctx := context.WithValue(r.Context(), clientInfoCtxKey, clientInfo)
			r = r.WithContext(ctx)

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
					// Support both "Bearer <token>" and just "<token>" formats
					tokenString := strings.TrimPrefix(authHeader, "Bearer ")
					tokenString = strings.TrimSpace(tokenString)

					if tokenString == "" {
						// No token found
						next.ServeHTTP(w, r)
						return
					}

					// Check if token is blacklisted
					if isTokenBlacklisted(db, tokenString) {
						// Token is blacklisted, reject the request with GraphQL error format
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte(`{"errors":[{"message":"Access denied. You must be logged in.","extensions":{"code":"UNAUTHENTICATED"}}],"data":null}`))
						return
					}

					// Validate the token
					claims, err := validateJWT(tokenString)
					if err != nil {
						// Invalid token, let the resolver handle it
						next.ServeHTTP(w, r)
						return
					}

					// Add user info and token to context
					ctx = context.WithValue(ctx, userCtxKey, claims)
					ctx = context.WithValue(ctx, tokenCtxKey, tokenString)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *Claims {
	raw, _ := ctx.Value(userCtxKey).(*Claims)
	return raw
}

// GetClientInfo finds the client info from the context. REQUIRES Middleware to have run.
func GetClientInfo(ctx context.Context) *ClientInfo {
	raw, _ := ctx.Value(clientInfoCtxKey).(*ClientInfo)
	return raw
}

// GetToken finds the token from the context. REQUIRES Middleware to have run.
func GetToken(ctx context.Context) string {
	raw, _ := ctx.Value(tokenCtxKey).(string)
	return raw
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// parseUserAgent extracts browser information from User-Agent string
func parseUserAgent(userAgent string) string {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		return "Chrome"
	}
	if strings.Contains(ua, "firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Safari"
	}
	if strings.Contains(ua, "edg") {
		return "Edge"
	}
	if strings.Contains(ua, "opera") || strings.Contains(ua, "opr") {
		return "Opera"
	}
	return "Unknown"
}

// parseOS extracts OS information from User-Agent string
func parseOS(userAgent string) string {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "windows") {
		return "Windows"
	}
	if strings.Contains(ua, "mac") || strings.Contains(ua, "darwin") {
		return "macOS"
	}
	if strings.Contains(ua, "linux") {
		return "Linux"
	}
	if strings.Contains(ua, "android") {
		return "Android"
	}
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ios") {
		return "iOS"
	}
	return "Unknown"
}

// AuthDirective adds authentication to a field.
// Only accepts full login tokens (purpose == "login").
// Password-reset tokens are restricted to newPassword mutation only.
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

	// Get the field name from the graphql context
	field := graphql.GetFieldContext(ctx)
	if field == nil {
		return nil, &gqlerror.Error{
			Message: "Access denied. Unable to verify field.",
			Extensions: map[string]interface{}{
				"code": "FORBIDDEN",
			},
		}
	}

	// Allow password_reset purpose only for newPassword mutation
	if user.Purpose == "password_reset" && field.Field.Name != "newPassword" {
		return nil, &gqlerror.Error{
			Message: "Access denied. Password reset token can only be used for password reset.",
			Extensions: map[string]interface{}{
				"code": "FORBIDDEN",
			},
		}
	}

	// Allow login purpose for all authenticated operations
	// Allow password_reset purpose only for newPassword
	if user.Purpose != "login" && user.Purpose != "password_reset" {
		return nil, &gqlerror.Error{
			Message: "Access denied. Token is not authorized for this operation.",
			Extensions: map[string]interface{}{
				"code": "FORBIDDEN",
			},
		}
	}

	return next(ctx)
}
