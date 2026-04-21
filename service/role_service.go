package service

import (
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type RoleService struct {
	DB *gorm.DB
}

func NewRoleService(db *gorm.DB) (*RoleService, error) {
	return &RoleService{
		DB: db,
	}, nil
}

func (s *RoleService) Roles() (*model.RolesResponse, error) {
	// Query all roles with permissions
	var rolesDB []model.UserRoleDB
	result := s.DB.Preload("Permissions").Where("deleted_at IS NULL").Find(&rolesDB)
	if result.Error != nil {
		return &model.RolesResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to retrieve roles: %v", result.Error),
		}, nil
	}

	// Convert DB models to GraphQL models
	roles := make([]*model.UserRole, len(rolesDB))
	for i, roleDB := range rolesDB {
		roles[i] = helper.ToGraphQLUserRole(roleDB)
	}

	return &model.RolesResponse{
		Code:    200,
		Success: true,
		Message: "roles retrieved successfully",
		Data:    roles,
	}, nil
}

func (s *RoleService) Role(id int64) (*model.RoleResponse, error) {
	// Find role by ID with permissions
	var roleDB model.UserRoleDB
	result := s.DB.Preload("Permissions").Where("id = ? AND deleted_at IS NULL", id).First(&roleDB)
	if result.Error != nil {
		return &model.RoleResponse{
			Code:    404,
			Success: false,
			Message: "role not found",
		}, nil
	}

	// Convert DB model to GraphQL model
	role := helper.ToGraphQLUserRole(roleDB)

	return &model.RoleResponse{
		Code:    200,
		Success: true,
		Message: "role retrieved successfully",
		Data:    role,
	}, nil
}

func (s *RoleService) CreateRole(input model.CreateRoleInput) (*model.CreateRoleResponse, error) {
	// Create role DB model
	roleDB := model.UserRoleDB{
		Name:     input.Name,
		IsActive: true,
	}

	// Save to database
	result := s.DB.Create(&roleDB)
	if result.Error != nil {
		return &model.CreateRoleResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create role: %v", result.Error),
		}, nil
	}

	// Associate permissions if provided
	if len(input.PermissionIds) > 0 {
		var permissions []model.UserPermissionDB
		s.DB.Where("id IN ?", input.PermissionIds).Find(&permissions)
		if len(permissions) > 0 {
			s.DB.Model(&roleDB).Association("Permissions").Append(&permissions)
		}
	}

	// Reload role with permissions
	s.DB.Preload("Permissions").First(&roleDB, roleDB.ID)

	// Convert DB model to GraphQL model
	role := helper.ToGraphQLUserRole(roleDB)

	return &model.CreateRoleResponse{
		Code:    201,
		Success: true,
		Message: "role created successfully",
		Data:    role,
	}, nil
}

func (s *RoleService) UpdateRole(id int64, input model.UpdateRoleInput) (*model.UpdateRoleResponse, error) {
	// Find role by ID
	var roleDB model.UserRoleDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&roleDB)
	if result.Error != nil {
		return &model.UpdateRoleResponse{
			Code:    404,
			Success: false,
			Message: "role not found",
		}, nil
	}

	// Update fields
	roleDB.Name = input.Name
	roleDB.IsActive = input.Status

	// Update permissions if provided
	if input.PermissionIds != nil {
		// Clear existing permissions
		s.DB.Model(&roleDB).Association("Permissions").Clear()

		// Add new permissions
		if len(input.PermissionIds) > 0 {
			var permissions []model.UserPermissionDB
			s.DB.Where("id IN ?", input.PermissionIds).Find(&permissions)
			if len(permissions) > 0 {
				s.DB.Model(&roleDB).Association("Permissions").Append(&permissions)
			}
		}
	}

	// Save to database
	result = s.DB.Save(&roleDB)
	if result.Error != nil {
		return &model.UpdateRoleResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update role: %v", result.Error),
		}, nil
	}

	// Reload role with permissions
	s.DB.Preload("Permissions").First(&roleDB, roleDB.ID)

	// Convert DB model to GraphQL model
	role := helper.ToGraphQLUserRole(roleDB)

	return &model.UpdateRoleResponse{
		Code:    200,
		Success: true,
		Message: "role updated successfully",
		Data:    role,
	}, nil
}

func (s *RoleService) DeleteRole(id int64) (*model.DeleteRoleResponse, error) {
	// Find role by ID
	var roleDB model.UserRoleDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&roleDB)
	if result.Error != nil {
		return &model.DeleteRoleResponse{
			Code:    404,
			Success: false,
			Message: "role not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	roleDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&roleDB)
	if result.Error != nil {
		return &model.DeleteRoleResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete role: %v", result.Error),
		}, nil
	}

	// Reload role with permissions
	s.DB.Preload("Permissions").First(&roleDB, roleDB.ID)

	// Convert DB model to GraphQL model
	role := helper.ToGraphQLUserRole(roleDB)

	return &model.DeleteRoleResponse{
		Code:    200,
		Success: true,
		Message: "role deleted successfully",
		Data:    role,
	}, nil
}
