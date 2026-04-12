package graph

import (
	"context"

	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DB *gorm.DB
}

// ForContext returns the Claims associated with the context
func ForContext(ctx context.Context) *Claims {
	raw, _ := ctx.Value("claims").(*Claims)
	return raw
}
