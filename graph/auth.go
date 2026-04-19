package graph

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"re-kasirpinter-go/graph/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	jwtSecret = []byte(getEnv("JWT_SECRET", "your-256-bit-secret"))
)

type Claims struct {
	UserID   int32  `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	SecureID string `json:"secure_id"`
	Purpose  string `json:"purpose"` // "login" or "password_reset"
	jwt.RegisteredClaims
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func generateOTPCode() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < 6; i++ {
		code += string(rune('0' + rand.Intn(10)))
	}
	return code
}

func generateJWT(userID int32, email string, role string, secureID string, purpose string) (string, error) {
	expiry := 24 * time.Hour
	if purpose == "password_reset" {
		expiry = 15 * time.Minute
	}
	expirationTime := time.Now().Add(expiry)

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		SecureID: secureID,
		Purpose:  purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kasirpinter",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// isTokenBlacklisted checks if a token is in the blacklist
func isTokenBlacklisted(db *gorm.DB, tokenString string) bool {
	var count int64
	db.Model(&model.BlacklistedTokenDB{}).Where("token = ? AND expires_at > ?", tokenString, time.Now()).Count(&count)
	return count > 0
}

// blacklistToken adds a token to the blacklist
func blacklistToken(db *gorm.DB, tokenString string, userID int32, expiresAt time.Time) error {
	blacklistedToken := model.BlacklistedTokenDB{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return db.Create(&blacklistedToken).Error
}

// cleanupExpiredBlacklistedTokens removes expired tokens from the blacklist
func cleanupExpiredBlacklistedTokens(db *gorm.DB) error {
	return db.Where("expires_at <= ?", time.Now()).Delete(&model.BlacklistedTokenDB{}).Error
}

// RequireSuperAdmin checks if the user from context has superadmin role
func RequireSuperAdmin(ctx context.Context) error {
	userClaims := ForContext(ctx)
	if userClaims == nil {
		return errors.New("unauthorized: no user context found")
	}

	if userClaims.Role != "superadmin" {
		return errors.New("unauthorized: superadmin access required")
	}

	return nil
}

// ValidateUserAuthorization checks if the authenticated user is authorized to access a resource
func ValidateUserAuthorization(ctx context.Context, requestedUserID int32) error {
	userClaims := ForContext(ctx)
	if userClaims == nil {
		return errors.New("unauthorized: no user context found")
	}

	// Superadmin can access any resource
	if userClaims.Role == "superadmin" {
		return nil
	}

	// Check if the requested userID matches the authenticated user's ID
	if userClaims.UserID != requestedUserID {
		return errors.New("unauthorized: you can only access resources for yourself")
	}

	return nil
}
