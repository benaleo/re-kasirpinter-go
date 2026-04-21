package service

import (
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type IngredientCategoryService struct {
	DB *gorm.DB
}

func NewIngredientCategoryService(db *gorm.DB) (*IngredientCategoryService, error) {
	return &IngredientCategoryService{
		DB: db,
	}, nil
}

func (s *IngredientCategoryService) IngredientCategories(pagination *model.PaginationInput, isOptions *bool) (*model.IngredientCategoriesResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query
	baseQuery := s.DB.Model(&model.IngredientCategoryDB{}).Where("deleted_at IS NULL")

	// Filter by is_active if is_options is true
	getActiveOnly := isOptions != nil && *isOptions
	if getActiveOnly {
		baseQuery = baseQuery.Where("is_active = ?", true)
	}

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return &model.IngredientCategoriesResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to count ingredient categories: %v", countResult.Error),
		}, nil
	}

	// Query ingredient categories with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var categoriesDB []model.IngredientCategoryDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&categoriesDB)
	if result.Error != nil {
		return &model.IngredientCategoriesResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to retrieve ingredient categories: %v", result.Error),
		}, nil
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(categoriesDB))

	// Convert DB models to GraphQL models
	categories := make([]*model.IngredientCategory, len(categoriesDB))
	for i, categoryDB := range categoriesDB {
		categories[i] = helper.ToGraphQLIngredientCategory(categoryDB)
	}

	return &model.IngredientCategoriesResponse{
		Code:       200,
		Success:    true,
		Message:    "ingredient categories retrieved successfully",
		Data:       categories,
		Pagination: paginationResult.PageInfo,
	}, nil
}

func (s *IngredientCategoryService) CreateIngredientCategory(input model.CreateIngredientCategoryInput) (*model.CreateIngredientCategoryResponse, error) {
	// Create ingredient category DB model
	categoryDB := model.IngredientCategoryDB{
		Name:        input.Name,
		Unit:        input.Unit,
		ConvertUnit: input.ConvertUnit,
		IsActive:    input.IsActive,
	}

	// Save to database
	result := s.DB.Create(&categoryDB)
	if result.Error != nil {
		return &model.CreateIngredientCategoryResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create ingredient category: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	category := helper.ToGraphQLIngredientCategory(categoryDB)

	return &model.CreateIngredientCategoryResponse{
		Code:    201,
		Success: true,
		Message: "ingredient category created successfully",
		Data:    category,
	}, nil
}

func (s *IngredientCategoryService) UpdateIngredientCategory(id int64, input model.UpdateIngredientCategoryInput) (*model.UpdateIngredientCategoryResponse, error) {
	// Find ingredient category by ID
	var categoryDB model.IngredientCategoryDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&categoryDB)
	if result.Error != nil {
		return &model.UpdateIngredientCategoryResponse{
			Code:    404,
			Success: false,
			Message: "ingredient category not found",
		}, nil
	}

	// Update fields
	categoryDB.Name = input.Name
	categoryDB.Unit = input.Unit
	categoryDB.ConvertUnit = input.ConvertUnit
	categoryDB.IsActive = input.IsActive

	// Save to database
	result = s.DB.Save(&categoryDB)
	if result.Error != nil {
		return &model.UpdateIngredientCategoryResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update ingredient category: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	category := helper.ToGraphQLIngredientCategory(categoryDB)

	return &model.UpdateIngredientCategoryResponse{
		Code:    200,
		Success: true,
		Message: "ingredient category updated successfully",
		Data:    category,
	}, nil
}

func (s *IngredientCategoryService) DeleteIngredientCategory(id int64) (*model.DeleteIngredientCategoryResponse, error) {
	// Find ingredient category by ID
	var categoryDB model.IngredientCategoryDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&categoryDB)
	if result.Error != nil {
		return &model.DeleteIngredientCategoryResponse{
			Code:    404,
			Success: false,
			Message: "ingredient category not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	categoryDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&categoryDB)
	if result.Error != nil {
		return &model.DeleteIngredientCategoryResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete ingredient category: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	category := helper.ToGraphQLIngredientCategory(categoryDB)

	return &model.DeleteIngredientCategoryResponse{
		Code:    200,
		Success: true,
		Message: "ingredient category deleted successfully",
		Data:    category,
	}, nil
}
