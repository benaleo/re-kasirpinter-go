package service

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type CleanupService struct {
	DB *gorm.DB
}

func NewCleanupService(db *gorm.DB) *CleanupService {
	return &CleanupService{
		DB: db,
	}
}

// CleanupExpiredTokens removes expired active_tokens from the database
func (s *CleanupService) CleanupExpiredTokens() error {
	result := s.DB.Where("expires_at < ?", time.Now()).Delete(&struct{}{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("[Cleanup] Removed %d expired active_tokens", result.RowsAffected)
	} else {
		log.Printf("[Cleanup] No expired active_tokens found")
	}

	return nil
}
