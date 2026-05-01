package service

import (
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type ProductExtraService struct {
	DB *gorm.DB
}

func NewProductExtraService(db *gorm.DB) (*ProductExtraService, error) {
	return &ProductExtraService{
		DB: db,
	}, nil
}

func (s *ProductExtraService) CreateProductExtra(input model.CreateProductExtraInput) (*model.CreateProductExtraResponse, error) {
	// Create product extra DB model
	productExtraDB := model.ProductExtraDB{
		Name:     input.Name,
		Price:    input.Price,
		IsActive: input.IsActive,
	}

	// Save to database
	result := s.DB.Create(&productExtraDB)
	if result.Error != nil {
		return &model.CreateProductExtraResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create product extra: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	productExtra := helper.ToGraphQLProductExtra(productExtraDB)

	return &model.CreateProductExtraResponse{
		Code:    201,
		Success: true,
		Message: "product extra created successfully",
		Data:    productExtra,
	}, nil
}

func (s *ProductExtraService) UpdateProductExtra(id int64, input model.UpdateProductExtraInput) (*model.UpdateProductExtraResponse, error) {
	// Find product extra by ID
	var productExtraDB model.ProductExtraDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&productExtraDB)
	if result.Error != nil {
		return &model.UpdateProductExtraResponse{
			Code:    404,
			Success: false,
			Message: "product extra not found",
		}, nil
	}

	// Update fields if provided
	if input.Name != nil {
		productExtraDB.Name = *input.Name
	}
	if input.Price != nil {
		productExtraDB.Price = *input.Price
	}
	if input.IsActive != nil {
		productExtraDB.IsActive = *input.IsActive
	}

	// Save to database
	result = s.DB.Save(&productExtraDB)
	if result.Error != nil {
		return &model.UpdateProductExtraResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update product extra: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	productExtra := helper.ToGraphQLProductExtra(productExtraDB)

	return &model.UpdateProductExtraResponse{
		Code:    200,
		Success: true,
		Message: "product extra updated successfully",
		Data:    productExtra,
	}, nil
}

func (s *ProductExtraService) DeleteProductExtra(id int64) (*model.DeleteProductExtraResponse, error) {
	// Find product extra by ID
	var productExtraDB model.ProductExtraDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&productExtraDB)
	if result.Error != nil {
		return &model.DeleteProductExtraResponse{
			Code:    404,
			Success: false,
			Message: "product extra not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	productExtraDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&productExtraDB)
	if result.Error != nil {
		return &model.DeleteProductExtraResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete product extra: %v", result.Error),
		}, nil
	}

	// Convert DB model to GraphQL model
	productExtra := helper.ToGraphQLProductExtra(productExtraDB)

	return &model.DeleteProductExtraResponse{
		Code:    200,
		Success: true,
		Message: "product extra deleted successfully",
		Data:    productExtra,
	}, nil
}

func (s *ProductExtraService) ProductExtras(pagination *model.PaginationInput, isActive *bool) (*model.ProductExtrasResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query
	baseQuery := s.DB.Model(&model.ProductExtraDB{}).Where("deleted_at IS NULL")

	// Filter by is_active if provided
	if isActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *isActive)
	}

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return &model.ProductExtrasResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to count product extras: %v", countResult.Error),
		}, nil
	}

	// Query product extras with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var productExtrasDB []model.ProductExtraDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&productExtrasDB)
	if result.Error != nil {
		return &model.ProductExtrasResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to retrieve product extras: %v", result.Error),
		}, nil
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(productExtrasDB))

	// Convert DB models to GraphQL models
	productExtras := make([]*model.ProductExtra, len(productExtrasDB))
	for i, productExtraDB := range productExtrasDB {
		productExtras[i] = helper.ToGraphQLProductExtra(productExtraDB)
	}

	return &model.ProductExtrasResponse{
		Code:       200,
		Success:    true,
		Message:    "product extras retrieved successfully",
		Data:       productExtras,
		Pagination: paginationResult.PageInfo,
	}, nil
}
