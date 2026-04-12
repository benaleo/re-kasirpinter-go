package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Load the appropriate .env file
	err := godotenv.Load(".env." + env)
	if err != nil {
		// Fallback to default .env file
		err = godotenv.Load()
		if err != nil {
			log.Printf("Warning: No .env file found. Using system environment variables.")
		}
	}

	// Set default values for required environment variables
	setDefaultEnv("JWT_SECRET", os.Getenv("JWT_SECRET"))
	setDefaultEnv("ENV", os.Getenv("ENV"))
	setDefaultEnv("PORT", os.Getenv("PORT"))
	setDefaultEnv("ALLOWED_ORIGINS", os.Getenv("ALLOWED_ORIGINS"))

	return nil
}

// setDefaultEnv sets an environment variable to a default value if it's not already set
func setDefaultEnv(key, defaultValue string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, defaultValue)
	}
}
