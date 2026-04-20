package graph

import (
	"re-kasirpinter-go/service"

	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DB          *gorm.DB
	R2Service   *service.R2Service
	UserService *service.UserService
}
