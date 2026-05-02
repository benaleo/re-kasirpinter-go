package service

import (
	"context"
	"fmt"
	"re-kasirpinter-go/email"
	"re-kasirpinter-go/graph/input"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	DB        *gorm.DB
	R2Service *R2Service
}

func NewUserService(db *gorm.DB) (*UserService, error) {
	r2Service, err := NewR2Service()
	if err != nil {
		// Log the error but don't fail service creation
		// Avatar upload will be optional
		fmt.Printf("Warning: Failed to initialize R2 service: %v\n", err)
	}

	return &UserService{
		DB:        db,
		R2Service: r2Service,
	}, nil
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

		// Handle avatar upload if provided
		var avatarURL *string
		if input.Avatar != nil && *input.Avatar != "" {
			// If avatar is already a URL, use it directly
			if helper.IsImageURL(*input.Avatar) {
				avatarURL = input.Avatar
			} else {
				// Upload to R2 using helper with UUID filename
				avatarURLStr, err := helper.UploadImageToR2(context.Background(), s.R2Service, *input.Avatar, "avatars")
				if err != nil {
					return helper.InternalServerErrorResponse(fmt.Sprintf("failed to upload avatar: %v", err)), nil
				}
				avatarURL = &avatarURLStr
			}
		}

		// Create user DB model
		userDB := model.UserDB{
			SecureID: &secureID,
			Name:     input.Name,
			Email:    input.Email,
			Address:  input.Address, // Optional
			Phone:    input.Phone,
			Password: hashedPassword,
			Avatar:   avatarURL,
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

		// Send welcome email with password
		loginURL := "https://app.kasirpinter.com/login" // Update with actual login URL
		go func() {
			if err := email.SendUserWelcomeEmail(input.Email, input.Name, autoPassword, loginURL); err != nil {
				fmt.Printf("Failed to send welcome email to %s: %v\n", input.Email, err)
			}
		}()

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

		// Handle avatar upload if provided
		var avatarURL *string
		if input.Avatar != nil && *input.Avatar != "" {
			// If avatar is already a URL, use it directly
			if helper.IsImageURL(*input.Avatar) {
				avatarURL = input.Avatar
			} else {
				// Upload to R2 using helper with UUID filename
				avatarURLStr, err := helper.UploadImageToR2(context.Background(), s.R2Service, *input.Avatar, "avatars")
				if err != nil {
					return helper.InternalServerErrorResponse(fmt.Sprintf("failed to upload avatar: %v", err)), nil
				}
				avatarURL = &avatarURLStr
			}
		}

		// Create user DB model
		userDB := model.UserDB{
			SecureID: &secureID,
			Name:     input.Name,
			Email:    input.Email,
			Address:  input.Address,
			Phone:    input.Phone,
			Password: hashedPassword,
			Avatar:   avatarURL,
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

		// Send activation email
		loginURL := "https://app.kasirpinter.com/login" // Update with actual login URL
		roleName := "User"                              // Default role name
		if userRole.ID > 0 {
			roleName = userRole.Name
		}
		go func() {
			if err := email.SendUserActivationEmail(input.Email, input.Name, input.Phone, roleName, loginURL); err != nil {
				fmt.Printf("Failed to send activation email to %s: %v\n", input.Email, err)
			}
		}()

		return helper.SuccessResponse(201, "user created successfully", user), nil
	}
}

func (s *UserService) UpdateUser(ctx context.Context, id string, input input.UpdateUserInput) (*model.UpdateUserResponse, error) {
	// Find user by secure_id
	var userDB model.UserDB
	result := s.DB.Where("secure_id = ?", id).Where("deleted_at IS NULL").First(&userDB)
	if result.Error != nil {
		return &model.UpdateUserResponse{
			Code:    404,
			Success: false,
			Message: "user not found",
		}, nil
	}

	// Handle avatar upload if provided
	var avatarURL *string
	if input.Avatar != nil && *input.Avatar != "" {
		// If avatar is already a URL, use it directly
		if helper.IsImageURL(*input.Avatar) {
			avatarURL = input.Avatar
		} else {
			// Upload to R2 using helper with UUID filename
			avatarURLStr, err := helper.UploadImageToR2(ctx, s.R2Service, *input.Avatar, "avatars")
			if err != nil {
				return &model.UpdateUserResponse{
					Code:    500,
					Success: false,
					Message: fmt.Sprintf("failed to upload avatar: %v", err),
				}, nil
			}
			avatarURL = &avatarURLStr
		}
	}

	// Update fields if provided
	if input.Name != nil {
		userDB.Name = *input.Name
	}
	if input.Email != nil {
		userDB.Email = *input.Email
	}
	if input.Address != nil {
		userDB.Address = *input.Address
	}
	if input.Phone != nil {
		userDB.Phone = *input.Phone
	}
	if input.Password != nil && *input.Password != "" {
		// Decrypt password from frontend using AES
		decryptedPassword, err := helper.Decrypt(*input.Password)
		if err != nil {
			return &model.UpdateUserResponse{
				Code:    400,
				Success: false,
				Message: "invalid password format",
			}, nil
		}

		// Hash the new password
		hashedPassword, err := helper.HashPassword(decryptedPassword)
		if err != nil {
			return &model.UpdateUserResponse{
				Code:    500,
				Success: false,
				Message: fmt.Sprintf("failed to hash password: %v", err),
			}, nil
		}

		userDB.Password = hashedPassword
	}
	if avatarURL != nil {
		userDB.Avatar = avatarURL
	} else if input.Avatar != nil && *input.Avatar == "" {
		// If avatar is explicitly set to empty string, clear it
		userDB.Avatar = nil
	}
	if input.IsActive != nil {
		userDB.IsActive = *input.IsActive
	}
	if input.RoleID != nil {
		userDB.RoleID = input.RoleID
	}

	// Save to database
	result = s.DB.Save(&userDB)
	if result.Error != nil {
		return &model.UpdateUserResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update user: %v", result.Error),
		}, nil
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

	return &model.UpdateUserResponse{
		Code:    200,
		Success: true,
		Message: "user updated successfully",
		Data:    user,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) (*model.DeleteUserResponse, error) {
	// Find user by secure_id
	var userDB model.UserDB
	result := s.DB.Where("secure_id = ?", id).Where("deleted_at IS NULL").First(&userDB)
	if result.Error != nil {
		return &model.DeleteUserResponse{
			Code:    404,
			Success: false,
			Message: "user not found",
		}, nil
	}

	// Soft delete by setting deleted_at and is_active
	now := time.Now()
	userDB.DeletedAt = &now
	userDB.IsActive = false

	// Save to database
	result = s.DB.Save(&userDB)
	if result.Error != nil {
		return &model.DeleteUserResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete user: %v", result.Error),
		}, nil
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

	return &model.DeleteUserResponse{
		Code:    200,
		Success: true,
		Message: "user deleted successfully",
		Data:    user,
	}, nil
}

func (s *UserService) Users(pagination *model.PaginationInput, isUser *bool) (*model.UsersResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query with filters
	baseQuery := s.DB.Model(&model.UserDB{}).Where("deleted_at IS NULL").Where("secure_id IS NOT NULL")

	// Filter by role_id based on is_user parameter (default: true)
	getUserWithRole2 := isUser == nil || *isUser
	if getUserWithRole2 {
		// Get users with role_id = 2
		baseQuery = baseQuery.Where("role_id = ?", 2)
	} else {
		// Get users excluding role_id = 2
		baseQuery = baseQuery.Where("role_id != ?", 2)
	}

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return nil, countResult.Error
	}

	// Query users with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var usersDB []model.UserDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&usersDB)
	if result.Error != nil {
		return nil, result.Error
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(usersDB))

	// Convert DB models to GraphQL models
	users := make([]*model.User, len(usersDB))
	for i, userDB := range usersDB {
		// Get user role if exists
		var userRoleDB *model.UserRoleDB
		if userDB.RoleID != nil {
			var userRole model.UserRoleDB
			s.DB.First(&userRole, *userDB.RoleID)
			if userRole.ID > 0 {
				userRoleDB = &userRole
			}
		}

		users[i] = helper.ToGraphQLUser(userDB, userRoleDB)
	}

	return &model.UsersResponse{
		Code:       200,
		Success:    true,
		Message:    "users retrieved successfully",
		Data:       users,
		Pagination: paginationResult.PageInfo,
	}, nil
}

func (s *UserService) User(id string) (*model.UserResponse, error) {
	// Find user by secure_id
	var userDB model.UserDB
	result := s.DB.Where("secure_id = ?", id).Where("deleted_at IS NULL").First(&userDB)
	if result.Error != nil {
		return &model.UserResponse{
			Code:    404,
			Success: false,
			Message: "user not found",
		}, nil
	}

	// Get user role if exists
	var userRole model.UserRoleDB
	var userRoleDB *model.UserRoleDB
	if userDB.RoleID != nil {
		s.DB.First(&userRole, *userDB.RoleID)
		if userRole.ID > 0 {
			userRoleDB = &userRole
		}
	}

	// Convert DB model to GraphQL model using mapper
	user := helper.ToGraphQLUser(userDB, userRoleDB)

	return &model.UserResponse{
		Code:    200,
		Success: true,
		Message: "user retrieved successfully",
		Data:    user,
	}, nil
}

func (s *UserService) CustomerSearch(keyword string) (*model.CustomerSearchResponse, error) {
	// Search for users by name or phone (case-insensitive) with role_id = 2
	var usersDB []model.UserDB
	result := s.DB.Where("deleted_at IS NULL").
		Where("role_id = ?", 2).
		Where("LOWER(name) LIKE LOWER(?) OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Find(&usersDB)

	if result.Error != nil {
		return &model.CustomerSearchResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("database error: %v", result.Error),
		}, nil
	}

	// Create slice of CustomerSearchData with only the required fields
	var customersData []*model.CustomerSearchData
	for _, userDB := range usersDB {
		customerData := &model.CustomerSearchData{
			SecureID: userDB.SecureID,
			Name:     userDB.Name,
			Phone:    userDB.Phone,
		}
		customersData = append(customersData, customerData)
	}

	return &model.CustomerSearchResponse{
		Code:    200,
		Success: true,
		Message: "customers retrieved successfully",
		Data:    customersData,
	}, nil
}
