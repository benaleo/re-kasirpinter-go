package service

import (
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type ProductCategoryService struct {
	DB *gorm.DB
}

func NewProductCategoryService(db *gorm.DB) (*ProductCategoryService, error) {
	return &ProductCategoryService{
		DB: db,
	}, nil
}

func (s *ProductCategoryService) ProductCategories(pagination *model.PaginationInput) (*model.ProductCategoriesResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query with parent and children preload
	baseQuery := s.DB.Model(&model.ProductCategoryDB{}).Preload("Parent").Preload("Children", "deleted_at IS NULL").Where("deleted_at IS NULL")

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return &model.ProductCategoriesResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to count product categories: %v", countResult.Error),
		}, nil
	}

	// Query product categories with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var categoriesDB []model.ProductCategoryDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&categoriesDB)
	if result.Error != nil {
		return &model.ProductCategoriesResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to retrieve product categories: %v", result.Error),
		}, nil
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(categoriesDB))

	// Convert DB models to GraphQL models
	categories := make([]*model.ProductCategory, len(categoriesDB))
	for i, categoryDB := range categoriesDB {
		categories[i] = helper.ToGraphQLProductCategory(categoryDB)
	}

	return &model.ProductCategoriesResponse{
		Code:       200,
		Success:    true,
		Message:    "product categories retrieved successfully",
		Data:       categories,
		Pagination: paginationResult.PageInfo,
	}, nil
}

func (s *ProductCategoryService) CreateProductCategory(input model.CreateProductCategoryInput) (*model.CreateProductCategoryResponse, error) {
	// Create product category DB model
	categoryDB := model.ProductCategoryDB{
		Name:        input.Name,
		Description: input.Description,
		ParentID:    input.ParentID,
		IsActive:    input.IsActive,
	}

	// Save to database
	result := s.DB.Create(&categoryDB)
	if result.Error != nil {
		return &model.CreateProductCategoryResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create product category: %v", result.Error),
		}, nil
	}

	// Reload with parent and children
	s.DB.Preload("Parent").Preload("Children", "deleted_at IS NULL").First(&categoryDB, categoryDB.ID)

	// Convert DB model to GraphQL model
	category := helper.ToGraphQLProductCategory(categoryDB)

	return &model.CreateProductCategoryResponse{
		Code:    201,
		Success: true,
		Message: "product category created successfully",
		Data:    category,
	}, nil
}

func (s *ProductCategoryService) UpdateProductCategory(id int64, input model.UpdateProductCategoryInput) (*model.UpdateProductCategoryResponse, error) {
	// Find product category by ID
	var categoryDB model.ProductCategoryDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&categoryDB)
	if result.Error != nil {
		return &model.UpdateProductCategoryResponse{
			Code:    404,
			Success: false,
			Message: "product category not found",
		}, nil
	}

	// Update fields if provided
	if input.Name != nil {
		categoryDB.Name = *input.Name
	}
	if input.Description != nil {
		categoryDB.Description = input.Description
	}
	if input.ParentID != nil {
		categoryDB.ParentID = input.ParentID
	}
	if input.IsActive != nil {
		categoryDB.IsActive = *input.IsActive
	}

	// Save to database
	result = s.DB.Save(&categoryDB)
	if result.Error != nil {
		return &model.UpdateProductCategoryResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update product category: %v", result.Error),
		}, nil
	}

	// Reload with parent and children
	s.DB.Preload("Parent").Preload("Children", "deleted_at IS NULL").First(&categoryDB, categoryDB.ID)

	// Convert DB model to GraphQL model
	category := helper.ToGraphQLProductCategory(categoryDB)

	return &model.UpdateProductCategoryResponse{
		Code:    200,
		Success: true,
		Message: "product category updated successfully",
		Data:    category,
	}, nil
}

func (s *ProductCategoryService) DeleteProductCategory(id int64) (*model.DeleteProductCategoryResponse, error) {
	// Find product category by ID
	var categoryDB model.ProductCategoryDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&categoryDB)
	if result.Error != nil {
		return &model.DeleteProductCategoryResponse{
			Code:    404,
			Success: false,
			Message: "product category not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	categoryDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&categoryDB)
	if result.Error != nil {
		return &model.DeleteProductCategoryResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete product category: %v", result.Error),
		}, nil
	}

	// Reload with parent and children
	s.DB.Preload("Parent").Preload("Children", "deleted_at IS NULL").First(&categoryDB, categoryDB.ID)

	// Convert DB model to GraphQL model
	category := helper.ToGraphQLProductCategory(categoryDB)

	return &model.DeleteProductCategoryResponse{
		Code:    200,
		Success: true,
		Message: "product category deleted successfully",
		Data:    category,
	}, nil
}
