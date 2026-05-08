package service

import (
	"context"
	"encoding/json"
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProductService struct {
	DB        *gorm.DB
	Redis     *redis.Client
	R2Service *R2Service
}

func NewProductService(db *gorm.DB, redisClient *redis.Client) (*ProductService, error) {
	r2Service, err := NewR2Service()
	if err != nil {
		// Log the error but don't fail service creation
		// Image upload will be optional
		fmt.Printf("Warning: Failed to initialize R2 service: %v\n", err)
	}

	return &ProductService{
		DB:        db,
		Redis:     redisClient,
		R2Service: r2Service,
	}, nil
}

// generateProductCacheKey generates a cache key for product list queries
func (s *ProductService) generateProductCacheKey(pagination *model.PaginationInput, isActive *bool, productExtraIds *bool) string {
	limitVal := int32(0)
	if pagination != nil && pagination.Limit != nil {
		limitVal = *pagination.Limit
	}
	pageVal := int32(0)
	if pagination != nil && pagination.Page != nil {
		pageVal = *pagination.Page
	}
	isActiveStr := ""
	if isActive != nil {
		isActiveStr = fmt.Sprintf("%v", *isActive)
	}
	productExtraIdsStr := ""
	if productExtraIds != nil {
		productExtraIdsStr = fmt.Sprintf("%v", *productExtraIds)
	}
	return fmt.Sprintf("products:%v:%v:%s:%s", limitVal, pageVal, isActiveStr, productExtraIdsStr)
}

func (s *ProductService) Products(ctx context.Context, pagination *model.PaginationInput, isActive *bool, productExtraIds *bool) (*model.ProductsResponse, error) {
	// Generate cache key
	cacheKey := s.generateProductCacheKey(pagination, isActive, productExtraIds)

	// Try to get from cache first
	if s.Redis != nil {
		cachedData, err := s.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var response model.ProductsResponse
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				fmt.Printf("[Redis] Loading products from cache: %s\n", cacheKey)
				return &response, nil
			}
		}
	}

	fmt.Printf("[DB] Loading products from database\n")

	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query with category preload
	baseQuery := s.DB.Model(&model.ProductDB{}).Preload("Category").Where("deleted_at IS NULL")

	// Preload ProductHasExtras if productExtraIds flag is true
	if productExtraIds != nil && *productExtraIds {
		baseQuery = baseQuery.Preload("ProductHasExtras")
	}

	// Filter by is_active if provided
	if isActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *isActive)
	}

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

	// Query products with pagination and preload variants with ingredients and stocks
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var productsDB []model.ProductDB
	result := baseQuery.
		Preload("Variants", "deleted_at IS NULL").
		Preload("Variants.Ingredients").
		Preload("Variants.Ingredients.Ingredient").
		Preload("Variants.Ingredients.Ingredient.Stocks", "deleted_at IS NULL").
		Order(paginationResult.SortBy).
		Limit(int(paginationResult.Limit)).
		Offset(paginationResult.Offset).
		Find(&productsDB)
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

	response := &model.ProductsResponse{
		Code:       200,
		Success:    true,
		Message:    "products retrieved successfully",
		Data:       products,
		Pagination: paginationResult.PageInfo,
	}

	// Cache the result
	if s.Redis != nil {
		cachedData, err := json.Marshal(response)
		if err == nil {
			s.Redis.Set(ctx, cacheKey, cachedData, 5*time.Minute)
			fmt.Printf("[Redis] Cached products list: %s\n", cacheKey)
		}
	}

	return response, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, input model.CreateProductInput) (*model.CreateProductResponse, error) {
	// Generate UUID v4 for secure_id
	secureID := uuid.New().String()

	// Handle image upload if provided
	var imageURL *string
	if input.Image != nil && *input.Image != "" {
		// If image is already a URL, use it directly
		if helper.IsImageURL(*input.Image) {
			imageURL = input.Image
		} else {
			// Upload to R2 using helper with UUID filename
			imageURLStr, err := helper.UploadImageToR2(context.Background(), s.R2Service, *input.Image, "products")
			if err != nil {
				return &model.CreateProductResponse{
					Code:    500,
					Success: false,
					Message: fmt.Sprintf("failed to upload image: %v", err),
				}, nil
			}
			imageURL = &imageURLStr
		}
	}

	// Create product DB model
	productDB := model.ProductDB{
		SecureID:      &secureID,
		Name:          input.Name,
		Image:         imageURL,
		CategoryID:    input.CategoryID,
		Description:   input.Description,
		AvailableType: input.AvailableType,
		VariantType:   input.VariantType,
		IsActive:      input.IsActive,
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

	// Handle product_extra_ids if provided
	if input.ProductExtraIds != nil && len(input.ProductExtraIds) > 0 {
		// Create ProductHasExtra relationships
		var productHasExtras []model.ProductHasExtraDB
		for _, extraID := range input.ProductExtraIds {
			if extraID != nil {
				productHasExtras = append(productHasExtras, model.ProductHasExtraDB{
					ProductID:      productDB.ID,
					ProductExtraID: *extraID,
				})
			}
		}

		// Save all ProductHasExtra relationships
		if len(productHasExtras) > 0 {
			result := s.DB.Create(&productHasExtras)
			if result.Error != nil {
				return &model.CreateProductResponse{
					Code:    500,
					Success: false,
					Message: fmt.Sprintf("failed to create product extras: %v", result.Error),
				}, nil
			}
		}
	}

	// Reload with category and product extras
	s.DB.Preload("Category").Preload("ProductHasExtras").First(&productDB, productDB.ID)

	// Convert DB model to GraphQL model
	product := helper.ToGraphQLProduct(productDB)

	// Evict cache
	s.evictProductCache(ctx)

	return &model.CreateProductResponse{
		Code:    201,
		Success: true,
		Message: "product created successfully",
		Data:    product,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id int64, input model.UpdateProductInput) (*model.UpdateProductResponse, error) {
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

	// Handle image upload if provided
	var imageURL *string
	if input.Image != nil && *input.Image != "" {
		// If image is already a URL, use it directly
		if helper.IsImageURL(*input.Image) {
			imageURL = input.Image
		} else {
			// Upload to R2 using helper with UUID filename
			imageURLStr, err := helper.UploadImageToR2(ctx, s.R2Service, *input.Image, "products")
			if err != nil {
				return &model.UpdateProductResponse{
					Code:    500,
					Success: false,
					Message: fmt.Sprintf("failed to upload image: %v", err),
				}, nil
			}
			imageURL = &imageURLStr
		}
	}

	// Update fields if provided
	if input.Name != nil {
		productDB.Name = *input.Name
	}
	if imageURL != nil {
		productDB.Image = imageURL
	} else if input.Image != nil && *input.Image == "" {
		// If image is explicitly set to empty string, clear it
		productDB.Image = nil
	}
	if input.CategoryID != nil {
		productDB.CategoryID = input.CategoryID
	}
	if input.Description != nil {
		productDB.Description = input.Description
	}
	productDB.AvailableType = input.AvailableType
	productDB.VariantType = input.VariantType
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

	// Handle product_extra_ids if provided (replace all existing relationships)
	if input.ProductExtraIds != nil {
		// Delete existing ProductHasExtra relationships for this product
		s.DB.Where("product_id = ?", productDB.ID).Delete(&model.ProductHasExtraDB{})

		// Create new ProductHasExtra relationships if any are provided
		if len(input.ProductExtraIds) > 0 {
			var productHasExtras []model.ProductHasExtraDB
			for _, extraID := range input.ProductExtraIds {
				if extraID != nil {
					productHasExtras = append(productHasExtras, model.ProductHasExtraDB{
						ProductID:      productDB.ID,
						ProductExtraID: *extraID,
					})
				}
			}

			// Save all new ProductHasExtra relationships
			if len(productHasExtras) > 0 {
				result := s.DB.Create(&productHasExtras)
				if result.Error != nil {
					return &model.UpdateProductResponse{
						Code:    500,
						Success: false,
						Message: fmt.Sprintf("failed to update product extras: %v", result.Error),
					}, nil
				}
			}
		}
	}

	// Reload with category and product extras
	s.DB.Preload("Category").Preload("ProductHasExtras").First(&productDB, productDB.ID)

	// Convert DB model to GraphQL model
	product := helper.ToGraphQLProduct(productDB)

	// Evict cache
	s.evictProductCache(ctx)

	return &model.UpdateProductResponse{
		Code:    200,
		Success: true,
		Message: "product updated successfully",
		Data:    product,
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int64) (*model.DeleteProductResponse, error) {
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

	// Evict cache
	s.evictProductCache(ctx)

	return &model.DeleteProductResponse{
		Code:    200,
		Success: true,
		Message: "product deleted successfully",
		Data:    product,
	}, nil
}

// evictProductCache evicts product-related cache entries
func (s *ProductService) evictProductCache(ctx context.Context) {
	if s.Redis == nil {
		return
	}

	// Evict all product list caches (pattern-based deletion)
	iter := s.Redis.Scan(ctx, 0, "products:*", 0).Iterator()
	count := 0
	for iter.Next(ctx) {
		s.Redis.Del(ctx, iter.Val())
		count++
	}
	fmt.Printf("[Redis] Evicted %d product cache entries\n", count)
}
