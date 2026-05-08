package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"

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

type AuthStatusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewAuthStatusResponseWriter(w http.ResponseWriter) *AuthStatusResponseWriter {
	return &AuthStatusResponseWriter{w, http.StatusOK}
}

func (asrw *AuthStatusResponseWriter) WriteHeader(code int) {
	asrw.statusCode = code
	asrw.ResponseWriter.WriteHeader(code)
}

func (asrw *AuthStatusResponseWriter) Write(b []byte) (int, error) {
	// Check if this is a GraphQL response with auth errors
	var response map[string]interface{}
	if err := json.Unmarshal(b, &response); err == nil {
		if errors, ok := response["errors"].([]interface{}); ok {
			for _, err := range errors {
				if errMap, ok := err.(map[string]interface{}); ok {
					if extensions, ok := errMap["extensions"].(map[string]interface{}); ok {
						if code, ok := extensions["code"].(string); ok && code == "UNAUTHENTICATED" {
							asrw.ResponseWriter.WriteHeader(http.StatusUnauthorized)
							break
						}
					}
				}
			}
		}
	}
	return asrw.ResponseWriter.Write(b)
}

func AuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the response writer to capture auth errors
			authWriter := NewAuthStatusResponseWriter(w)

			// Extract client information
			clientInfo := &ClientInfo{
				IP:      getClientIP(r),
				Browser: parseUserAgent(r.UserAgent()),
				OS:      parseOS(r.UserAgent()),
			}

			// Add client info to context
			ctx := context.WithValue(r.Context(), clientInfoCtxKey, clientInfo)
			r = r.WithContext(ctx)

			// Handle GraphQL requests
			if r.URL.Path == "/query" && r.Method == "POST" {
				// Check if there's an Authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					// No auth header, let the resolver handle it (login, createOtp, etc.)
					next.ServeHTTP(authWriter, r)
					return
				}

				// Extract the token from the Authorization header
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				tokenString = strings.TrimSpace(tokenString)

				if tokenString == "" {
					// No token found, let the resolver handle it
					next.ServeHTTP(authWriter, r)
					return
				}

				// Validate the JWT token
				claims, err := validateJWT(tokenString)
				if err != nil {
					// Invalid JWT, let the resolver handle it
					next.ServeHTTP(authWriter, r)
					return
				}

				// Check if token is blacklisted
				if isTokenBlacklisted(db, tokenString) {
					// Token is blacklisted, reject the request
					authWriter.Header().Set("Content-Type", "application/json")
					authWriter.WriteHeader(http.StatusUnauthorized)
					authWriter.Write([]byte(`{"errors":[{"message":"Access denied. Token has been revoked.","extensions":{"code":"UNAUTHENTICATED"}}],"data":null}`))
					return
				}

				// For login tokens, check ActiveTokenDB for expiry
				if claims.Purpose == "login" {
					var activeToken model.ActiveTokenDB
					result := db.Where("token = ? AND expires_at > ?", tokenString, time.Now()).First(&activeToken)
					if result.Error != nil {
						// Token not found or expired, check if this is a logout request
						// For logout, we allow expired tokens to be processed for blacklisting
						// We'll pass a flag to the resolver to handle this case
						ctx = context.WithValue(ctx, "tokenExpired", true)
						ctx = context.WithValue(ctx, "expiredToken", tokenString)
					}
				}

				// Add user info, token, DB, client info, and HTTP writer to context for @auth directive
				ctx = context.WithValue(ctx, userCtxKey, claims)
				ctx = context.WithValue(ctx, tokenCtxKey, tokenString)
				ctx = context.WithValue(ctx, "db", db)
				ctx = context.WithValue(ctx, clientInfoCtxKey, clientInfo)
				ctx = context.WithValue(ctx, "httpWriter", authWriter)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(authWriter, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *helper.Claims {
	raw, _ := ctx.Value(userCtxKey).(*helper.Claims)
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
// Extends token expiry for sliding session (1 hour from now).
// Allows expired tokens for logout mutation.
func AuthDirective(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	user := ForContext(ctx)

	// Check if this is an expired token case (for logout)
	tokenExpired, _ := ctx.Value("tokenExpired").(bool)
	expiredToken, _ := ctx.Value("expiredToken").(string)

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

	// Handle expired tokens - only allow logout
	if tokenExpired {
		if field.Field.Name == "logout" {
			// Allow logout to proceed with expired token
			// Create a minimal user context for logout processing
			expiredUser := &helper.Claims{
				Purpose: "login",
			}
			ctx = context.WithValue(ctx, userCtxKey, expiredUser)
			ctx = context.WithValue(ctx, "expiredTokenForLogout", expiredToken)
			return next(ctx)
		} else {
			// Reject expired tokens for all other methods
			return nil, &gqlerror.Error{
				Message: "Access denied. Token expired or invalid.",
				Extensions: map[string]interface{}{
					"code": "UNAUTHENTICATED",
				},
			}
		}
	}

	// Normal authentication flow for valid tokens
	if user == nil {
		return nil, &gqlerror.Error{
			Message: "Access denied. You must be logged in.",
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
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

	// Implement sliding session for login tokens
	// Extend token expiry to 1 hour from now if token is older than 30 minutes
	if user.Purpose == "login" {
		// Get the current token
		currentToken := GetToken(ctx)
		if currentToken != "" {
			// Check if token should be extended (older than 30 minutes)
			if time.Since(user.IssuedAt.Time) > 30*time.Minute {
				// Get client info for logging
				clientInfo := GetClientInfo(ctx)

				operationContext := graphql.GetOperationContext(ctx)
				if operationContext != nil {
					// Try to get DB from context
					if dbInterface, ok := ctx.Value("db").(*gorm.DB); ok {
						// Update the token expiry in database
						newExpiry := time.Now().Add(1 * time.Hour)
						result := dbInterface.Model(&model.ActiveTokenDB{}).Where("token = ?", currentToken).Update("expires_at", newExpiry)
						if result.Error == nil {
							// Log the extension with device info
							if clientInfo != nil {
								log.Printf("Token extended for user %d from IP %s (%s/%s)", user.UserID, clientInfo.IP, clientInfo.Browser, clientInfo.OS)
							}
							ctx = context.WithValue(ctx, "tokenExtended", true)
							ctx = context.WithValue(ctx, "tokenExtendedMessage", "Token expiry extended to 1 hour from now")
						}
					} else {
						// Fallback: just indicate extension should happen
						ctx = context.WithValue(ctx, "tokenExtended", true)
						ctx = context.WithValue(ctx, "extendToken", currentToken)
					}
				}
			}
		}
	}

	return next(ctx)
}

// AuthStatusInterceptor sets HTTP status codes based on GraphQL errors
type AuthStatusInterceptor struct{}

func (asi *AuthStatusInterceptor) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)

	// Check if response contains authentication errors
	if resp != nil && len(resp.Errors) > 0 {
		for _, err := range resp.Errors {
			if err != nil && err.Extensions != nil {
				if code, ok := err.Extensions["code"].(string); ok && code == "UNAUTHENTICATED" {
					// Set HTTP status code to 401 for authentication errors
					if httpWriter, ok := ctx.Value("httpWriter").(http.ResponseWriter); ok {
						httpWriter.WriteHeader(http.StatusUnauthorized)
					}
					break
				}
			}
		}
	}

	return resp
}

func (asi *AuthStatusInterceptor) ExtensionName() string {
	return "AuthStatusInterceptor"
}

func (asi *AuthStatusInterceptor) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// LoggingInterceptor logs GraphQL request details and handles token extension
type LoggingInterceptor struct{}

func (li *LoggingInterceptor) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	startTime := time.Now()

	// Get client info and user claims from context
	clientInfo := GetClientInfo(ctx)
	userClaims := ForContext(ctx)

	// Get GraphQL operation name
	operationContext := graphql.GetOperationContext(ctx)
	operationName := "unknown"
	if operationContext != nil && operationContext.Operation != nil {
		operationName = operationContext.Operation.Name
	}

	// Execute the response handler
	resp := next(ctx)

	// Check for token extension (sliding session extension)
	if tokenExtended, ok := ctx.Value("tokenExtended").(bool); ok && tokenExtended {
		// Check if extension was already handled in AuthDirective
		if message, ok := ctx.Value("tokenExtendedMessage").(string); ok {
			// Extension was handled in AuthDirective
			if resp != nil && resp.Extensions == nil {
				resp.Extensions = make(map[string]interface{})
			}
			if resp != nil {
				resp.Extensions["token_extended"] = true
				resp.Extensions["message"] = message
			}
		} else if tokenToExtend, ok := ctx.Value("extendToken").(string); ok && tokenToExtend != "" {
			// Fallback: handle extension here (for cases where DB wasn't available in AuthDirective)
			if dbInterface, ok := ctx.Value("db").(*gorm.DB); ok {
				go func(db *gorm.DB, token string) {
					newExpiry := time.Now().Add(1 * time.Hour)
					result := db.Model(&model.ActiveTokenDB{}).Where("token = ?", token).Update("expires_at", newExpiry)
					if result.Error == nil {
						log.Printf("Token extended: %s (expiry reset to 1 hour from now)", token)
					} else {
						log.Printf("Failed to extend token: %v", result.Error)
					}
				}(dbInterface, tokenToExtend)

				// Add extension notification to response
				if resp != nil && resp.Extensions == nil {
					resp.Extensions = make(map[string]interface{})
				}
				if resp != nil {
					resp.Extensions["token_extended"] = true
					resp.Extensions["message"] = "Token expiry extended to 1 hour from now"
				}
			}
		}
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Determine status
	status := "SUCCESS"
	if resp != nil && len(resp.Errors) > 0 {
		status = "FAILED"
	}

	// Extract user ID if available
	userID := "N/A"
	if userClaims != nil {
		userID = fmt.Sprintf("%d", userClaims.UserID)
	}

	// Extract IP
	ip := "N/A"
	if clientInfo != nil {
		ip = clientInfo.IP
	}

	// Log the request
	log.Printf("IP: %s | User ID: %s | Method: %s | Status: %s | Duration: %v",
		ip,
		userID,
		operationName,
		status,
		duration,
	)

	return resp
}

func (li *LoggingInterceptor) ExtensionName() string {
	return "LoggingInterceptor"
}

func (li *LoggingInterceptor) Validate(schema graphql.ExecutableSchema) error {
	return nil
}
