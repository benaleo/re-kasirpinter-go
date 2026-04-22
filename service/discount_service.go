package service

import (
	"context"
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountService struct {
	DB        *gorm.DB
	R2Service *R2Service
}

func NewDiscountService(db *gorm.DB) (*DiscountService, error) {
	r2Service, err := NewR2Service()
	if err != nil {
		// Log the error but don't fail service creation
		// Icon upload will be optional
		fmt.Printf("Warning: Failed to initialize R2 service: %v\n", err)
	}

	return &DiscountService{
		DB:        db,
		R2Service: r2Service,
	}, nil
}

func (s *DiscountService) Discounts(pagination *model.PaginationInput, isActive *bool, isPeriod *bool, isQuota *bool) (*model.DiscountsResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query
	baseQuery := s.DB.Model(&model.DiscountDB{}).Where("deleted_at IS NULL")

	// Apply is_active filter
	if isActive != nil && *isActive {
		baseQuery = baseQuery.Where("is_active = ?", true)
	}

	// Apply is_period filter: current time between start_at and end_at
	if isPeriod != nil && *isPeriod {
		now := time.Now()
		baseQuery = baseQuery.Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", now, now)
	}

	// Apply is_quota filter: quota > 0
	if isQuota != nil && *isQuota {
		baseQuery = baseQuery.Where("quota > 0")
	}

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return &model.DiscountsResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to count discounts: %v", countResult.Error),
		}, nil
	}

	// Query discounts with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var discountsDB []model.DiscountDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&discountsDB)
	if result.Error != nil {
		return &model.DiscountsResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to retrieve discounts: %v", result.Error),
		}, nil
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(discountsDB))

	// Convert DB models to GraphQL models
	discounts := make([]*model.Discount, len(discountsDB))
	for i, discountDB := range discountsDB {
		discounts[i] = helper.ToGraphQLDiscount(discountDB)
	}

	return &model.DiscountsResponse{
		Code:       200,
		Success:    true,
		Message:    "discounts retrieved successfully",
		Data:       discounts,
		Pagination: paginationResult.PageInfo,
	}, nil
}

func (s *DiscountService) CreateDiscount(input model.CreateDiscountInput) (*model.CreateDiscountResponse, error) {
	// Generate UUID v4 for icon upload
	secureID := uuid.New().String()

	// Handle icon upload if provided
	var iconURL *string
	if input.Icon != nil && *input.Icon != "" && s.R2Service != nil {
		iconURLStr, err := s.R2Service.UploadFromBase64(
			context.Background(),
			*input.Icon,
			"discounts",
			secureID,
		)
		if err != nil {
			return &model.CreateDiscountResponse{
				Code:    500,
				Success: false,
				Message: fmt.Sprintf("failed to upload icon: %v", err),
			}, nil
		}
		iconURL = &iconURLStr
	}

	// Create discount DB model
	discountDB := model.DiscountDB{
		Name:        input.Name,
		Description: input.Description,
		Icon:        iconURL,
		Code:        input.Code,
		Type:        model.DiscountType(input.Type),
		Value:       input.Value,
		MaxValue:    input.MaxValue,
		MinOrder:    input.MinOrder,
		Quota:       input.Quota,
		StartAt:     input.StartAt,
		EndAt:       input.EndAt,
		IsActive:    input.IsActive,
	}

	// Save to database
	result := s.DB.Create(&discountDB)
	if result.Error != nil {
		if helper.IsDuplicateCodeError(result.Error) {
			return &model.CreateDiscountResponse{
				Code:    400,
				Success: false,
				Message: "ups code already created",
			}, nil
		}
		return &model.CreateDiscountResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create discount: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	discount := helper.ToGraphQLDiscount(discountDB)

	return &model.CreateDiscountResponse{
		Code:    201,
		Success: true,
		Message: "discount created successfully",
		Data:    discount,
	}, nil
}

func (s *DiscountService) UpdateDiscount(ctx context.Context, id int64, input model.UpdateDiscountInput) (*model.UpdateDiscountResponse, error) {
	// Find discount by ID
	var discountDB model.DiscountDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&discountDB)
	if result.Error != nil {
		return &model.UpdateDiscountResponse{
			Code:    404,
			Success: false,
			Message: "discount not found",
		}, nil
	}

	// Handle icon upload if provided
	var iconURL *string
	if input.Icon != nil && *input.Icon != "" && s.R2Service != nil {
		// Use ID for upload folder naming
		iconURLStr, err := s.R2Service.UploadFromBase64(
			ctx,
			*input.Icon,
			"discounts",
			fmt.Sprintf("%d", id),
		)
		if err != nil {
			return &model.UpdateDiscountResponse{
				Code:    500,
				Success: false,
				Message: fmt.Sprintf("failed to upload icon: %v", err),
			}, nil
		}
		iconURL = &iconURLStr
	}

	// Update fields if provided
	if input.Name != nil {
		discountDB.Name = *input.Name
	}
	if input.Description != nil {
		discountDB.Description = input.Description
	}
	if iconURL != nil {
		discountDB.Icon = iconURL
	} else if input.Icon != nil && *input.Icon == "" {
		// If icon is explicitly set to empty string, clear it
		discountDB.Icon = nil
	}
	if input.Code != nil {
		discountDB.Code = input.Code
	}
	if input.Type != nil {
		discountDB.Type = model.DiscountType(*input.Type)
	}
	if input.Value != nil {
		discountDB.Value = *input.Value
	}
	if input.MaxValue != nil {
		discountDB.MaxValue = input.MaxValue
	}
	if input.MinOrder != nil {
		discountDB.MinOrder = input.MinOrder
	}
	if input.Quota != nil {
		discountDB.Quota = input.Quota
	}
	if input.StartAt != nil {
		discountDB.StartAt = input.StartAt
	}
	if input.EndAt != nil {
		discountDB.EndAt = input.EndAt
	}
	if input.IsActive != nil {
		discountDB.IsActive = *input.IsActive
	}

	// Save to database
	result = s.DB.Save(&discountDB)
	if result.Error != nil {
		if helper.IsDuplicateCodeError(result.Error) {
			return &model.UpdateDiscountResponse{
				Code:    400,
				Success: false,
				Message: "ups code already created",
			}, nil
		}
		return &model.UpdateDiscountResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update discount: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	discount := helper.ToGraphQLDiscount(discountDB)

	return &model.UpdateDiscountResponse{
		Code:    200,
		Success: true,
		Message: "discount updated successfully",
		Data:    discount,
	}, nil
}

func (s *DiscountService) DeleteDiscount(id int64) (*model.DeleteDiscountResponse, error) {
	// Find discount by ID
	var discountDB model.DiscountDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&discountDB)
	if result.Error != nil {
		return &model.DeleteDiscountResponse{
			Code:    404,
			Success: false,
			Message: "discount not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	discountDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&discountDB)
	if result.Error != nil {
		return &model.DeleteDiscountResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete discount: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	discount := helper.ToGraphQLDiscount(discountDB)

	return &model.DeleteDiscountResponse{
		Code:    200,
		Success: true,
		Message: "discount deleted successfully",
		Data:    discount,
	}, nil
}

func (s *DiscountService) CheckDiscount(code string) (*model.CheckDiscountResponse, error) {
	// Find discount by code with validation
	var discountDB model.DiscountDB
	now := time.Now()

	result := s.DB.Where("code = ? AND deleted_at IS NULL AND is_active = ?", code, true).First(&discountDB)

	// Check if discount is valid (found, quota available, and within date range)
	isValid := true
	if result.Error != nil {
		isValid = false
	}
	if discountDB.Quota != nil && *discountDB.Quota <= 0 {
		isValid = false
	}
	if discountDB.StartAt != nil && now.Before(*discountDB.StartAt) {
		isValid = false
	}
	if discountDB.EndAt != nil && now.After(*discountDB.EndAt) {
		isValid = false
	}

	if !isValid {
		return &model.CheckDiscountResponse{
			Code:    404,
			Success: false,
			Message: "invalid discount code",
			Data:    nil,
		}, nil
	}

	// Return only type and value
	return &model.CheckDiscountResponse{
		Code:    200,
		Success: true,
		Message: "discount valid",
		Data: &model.CheckDiscountData{
			Type:  model.DiscountType(discountDB.Type),
			Value: discountDB.Value,
		},
	}, nil
}
