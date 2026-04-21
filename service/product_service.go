package service

import (
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductService struct {
	DB *gorm.DB
}

func NewProductService(db *gorm.DB) (*ProductService, error) {
	return &ProductService{
		DB: db,
	}, nil
}

func (s *ProductService) Products(pagination *model.PaginationInput) (*model.ProductsResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query with category preload
	baseQuery := s.DB.Model(&model.ProductDB{}).Preload("Category").Where("deleted_at IS NULL")

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return &model.ProductsResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to count products: %v", countResult.Error),
		}, nil
	}

	// Query products with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var productsDB []model.ProductDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&productsDB)
	if result.Error != nil {
		return &model.ProductsResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to retrieve products: %v", result.Error),
		}, nil
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(productsDB))

	// Convert DB models to GraphQL models
	products := make([]*model.Product, len(productsDB))
	for i, productDB := range productsDB {
		products[i] = helper.ToGraphQLProduct(productDB)
	}

	return &model.ProductsResponse{
		Code:       200,
		Success:    true,
		Message:    "products retrieved successfully",
		Data:       products,
		Pagination: paginationResult.PageInfo,
	}, nil
}

func (s *ProductService) CreateProduct(input model.CreateProductInput) (*model.CreateProductResponse, error) {
	// Generate UUID v4 for secure_id
	secureID := uuid.New().String()

	// Create product DB model
	productDB := model.ProductDB{
		SecureID:    &secureID,
		Name:        input.Name,
		Image:       input.Image,
		CategoryID:  input.CategoryID,
		Description: input.Description,
		IsActive:    input.IsActive,
	}

	// Save to database
	result := s.DB.Create(&productDB)
	if result.Error != nil {
		return &model.CreateProductResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create product: %v", result.Error),
		}, nil
	}

	// Reload with category
	s.DB.Preload("Category").First(&productDB, productDB.ID)

	// Convert DB model to GraphQL model
	product := helper.ToGraphQLProduct(productDB)

	return &model.CreateProductResponse{
		Code:    201,
		Success: true,
		Message: "product created successfully",
		Data:    product,
	}, nil
}

func (s *ProductService) UpdateProduct(id int64, input model.UpdateProductInput) (*model.UpdateProductResponse, error) {
	// Find product by ID
	var productDB model.ProductDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&productDB)
	if result.Error != nil {
		return &model.UpdateProductResponse{
			Code:    404,
			Success: false,
			Message: "product not found",
		}, nil
	}

	// Update fields if provided
	if input.Name != nil {
		productDB.Name = *input.Name
	}
	if input.Image != nil {
		productDB.Image = input.Image
	}
	if input.CategoryID != nil {
		productDB.CategoryID = input.CategoryID
	}
	if input.Description != nil {
		productDB.Description = input.Description
	}
	if input.IsActive != nil {
		productDB.IsActive = *input.IsActive
	}

	// Save to database
	result = s.DB.Save(&productDB)
	if result.Error != nil {
		return &model.UpdateProductResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update product: %v", result.Error),
		}, nil
	}

	// Reload with category
	s.DB.Preload("Category").First(&productDB, productDB.ID)

	// Convert DB model to GraphQL model
	product := helper.ToGraphQLProduct(productDB)

	return &model.UpdateProductResponse{
		Code:    200,
		Success: true,
		Message: "product updated successfully",
		Data:    product,
	}, nil
}

func (s *ProductService) DeleteProduct(id int64) (*model.DeleteProductResponse, error) {
	// Find product by ID
	var productDB model.ProductDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&productDB)
	if result.Error != nil {
		return &model.DeleteProductResponse{
			Code:    404,
			Success: false,
			Message: "product not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	productDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&productDB)
	if result.Error != nil {
		return &model.DeleteProductResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete product: %v", result.Error),
		}, nil
	}

	// Reload with category
	s.DB.Preload("Category").First(&productDB, productDB.ID)

	// Convert DB model to GraphQL model
	product := helper.ToGraphQLProduct(productDB)

	return &model.DeleteProductResponse{
		Code:    200,
		Success: true,
		Message: "product deleted successfully",
		Data:    product,
	}, nil
}
