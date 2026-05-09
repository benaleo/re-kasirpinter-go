package main

import (
	"fmt"
	"log"
	"os"

	"re-kasirpinter-go/config"
	"re-kasirpinter-go/service"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		fmt.Printf("Warning: Failed to load environment variables: %v\n", err)
	}

	// Initialize database
	db, err := config.InitDb()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	// Initialize cleanup service
	cleanupService := service.NewCleanupService(db)

	// Run cleanup
	log.Println("[Manual] Running cleanup of expired active_tokens")
	if err := cleanupService.CleanupExpiredTokens(); err != nil {
		fmt.Printf("Error cleaning up expired tokens: %v\n", err)
		os.Exit(1)
	}

	log.Println("[Manual] Cleanup completed successfully")
}
