package service

import (
	"fmt"
	"re-kasirpinter-go/graph/input"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

func (s *UserService) CreateUser(input input.CreateUserInput, isUser *bool) (*model.CreateUserResponse, error) {
	// Determine is_user value (default true if not provided)
	isUserValue := isUser == nil || *isUser

	// Validate based on is_user flag
	if isUserValue {
		// For is_user true: mandatory fields are name, email, phone
		if input.Name == "" {
			return helper.BadRequestResponse("name is required"), nil
		}
		if input.Email == "" {
			return helper.BadRequestResponse("email is required"), nil
		}
		if input.Phone == "" {
			return helper.BadRequestResponse("phone is required"), nil
		}

		// Auto-generate password
		autoPassword := "Kasirpinter2026!"
		hashedPassword, err := helper.HashPassword(autoPassword)
		if err != nil {
			return helper.InternalServerErrorResponse(fmt.Sprintf("failed to hash password: %v", err)), nil
		}

		// Set role_id to 2 automatically
		roleID := int64(2)

		// Generate secure_id (UUID)
		secureID, err := helper.GenerateRandomString(16)
		if err != nil {
			return helper.InternalServerErrorResponse(fmt.Sprintf("failed to generate secure_id: %v", err)), nil
		}

		// Create user DB model
		userDB := model.UserDB{
			SecureID: &secureID,
			Name:     input.Name,
			Email:    input.Email,
			Address:  input.Address, // Optional
			Phone:    input.Phone,
			Password: hashedPassword,
			IsActive: true,
			RoleID:   &roleID,
		}

		// Save to database
		result := s.DB.Create(&userDB)
		if result.Error != nil {
			return helper.InternalServerErrorResponse(fmt.Sprintf("failed to create user: %v", result.Error)), nil
		}

		// Get user role
		var userRole model.UserRoleDB
		s.DB.First(&userRole, *userDB.RoleID)

		// Convert DB model to GraphQL model using mapper
		var userRoleDB *model.UserRoleDB
		if userRole.ID > 0 {
			userRoleDB = &userRole
		}
		user := helper.ToGraphQLUser(userDB, userRoleDB)

		return helper.SuccessResponse(201, "user created successfully with auto-generated password", user), nil
	} else {
		// For is_user false: mandatory fields are name, email, address, phone, password, role_id
		if input.Name == "" {
			return helper.BadRequestResponse("name is required"), nil
		}
		if input.Email == "" {
			return helper.BadRequestResponse("email is required"), nil
		}
		if input.Address == "" {
			return helper.BadRequestResponse("address is required"), nil
		}
		if input.Phone == "" {
			return helper.BadRequestResponse("phone is required"), nil
		}
		if input.Password == "" {
			return helper.BadRequestResponse("password is required"), nil
		}
		if input.RoleID == nil {
			return helper.BadRequestResponse("role_id is required"), nil
		}

		// Validate that role_id is not 2
		if *input.RoleID == 2 {
			return helper.BadRequestResponse("role_id cannot be 2 when is_user is false"), nil
		}

		// Hash password
		hashedPassword, err := helper.HashPassword(input.Password)
		if err != nil {
			return helper.InternalServerErrorResponse(fmt.Sprintf("failed to hash password: %v", err)), nil
		}

		// Generate secure_id (UUID)
		secureID, err := helper.GenerateRandomString(16)
		if err != nil {
			return helper.InternalServerErrorResponse(fmt.Sprintf("failed to generate secure_id: %v", err)), nil
		}

		// Create user DB model
		userDB := model.UserDB{
			SecureID: &secureID,
			Name:     input.Name,
			Email:    input.Email,
			Address:  input.Address,
			Phone:    input.Phone,
			Password: hashedPassword,
			IsActive: true,
			RoleID:   input.RoleID,
		}

		// Save to database
		result := s.DB.Create(&userDB)
		if result.Error != nil {
			return helper.InternalServerErrorResponse(fmt.Sprintf("failed to create user: %v", result.Error)), nil
		}

		// Get user role if exists
		var userRole model.UserRoleDB
		if userDB.RoleID != nil {
			s.DB.First(&userRole, *userDB.RoleID)
		}

		// Convert DB model to GraphQL model using mapper
		var userRoleDB *model.UserRoleDB
		if userRole.ID > 0 {
			userRoleDB = &userRole
		}
		user := helper.ToGraphQLUser(userDB, userRoleDB)

		return helper.SuccessResponse(201, "user created successfully", user), nil
	}
}
