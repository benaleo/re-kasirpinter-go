package helper

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateRandomString generates a random string using UUID
func GenerateRandomString(n int) (string, error) {
	return uuid.New().String(), nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if a password matches a hashed password
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
