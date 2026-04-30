package service

import (
	"context"
	"fmt"
	"re-kasirpinter-go/graph/model"

	"gorm.io/gorm"
)

type ProductVariantService struct {
	DB        *gorm.DB
	R2Service *R2Service
}

func NewProductVariantService(db *gorm.DB) (*ProductVariantService, error) {
	r2Service, err := NewR2Service()
	if err != nil {
		// Log the error but don't fail service creation
		// Image upload will be optional
		fmt.Printf("Warning: Failed to initialize R2 service: %v\n", err)
	}

	return &ProductVariantService{
		DB:        db,
		R2Service: r2Service,
	}, nil
}

func (s *ProductVariantService) Create(ctx context.Context, input model.CreateProductVariantInput) (*model.ProductVariantDB, error) {
	// Check if product exists
	var product model.ProductDB
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", input.ProductID).First(&product).Error; err != nil {
		return nil, err
	}

	// Handle image upload if provided
	var imageURL *string
	if input.Image != nil && *input.Image != "" && s.R2Service != nil {
		imageURLStr, err := s.R2Service.UploadFromBase64(
			context.Background(),
			*input.Image,
			"product-variants",
			fmt.Sprintf("%d", input.ProductID),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %v", err)
		}
		imageURL = &imageURLStr
	}

	variant := &model.ProductVariantDB{
		Image:         imageURL,
		ProductID:     input.ProductID,
		Name:          input.Name,
		Price:         input.Price,
		PriceOriginal: input.PriceOriginal,
		IsActive:      input.IsActive,
	}

	if err := s.DB.Create(variant).Error; err != nil {
		return nil, err
	}

	return variant, nil
}

func (s *ProductVariantService) Update(ctx context.Context, id int64, input model.UpdateProductVariantInput) (*model.ProductVariantDB, error) {
	var variant model.ProductVariantDB
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&variant).Error; err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Image != nil {
		// Handle image upload if provided
		if *input.Image != "" && s.R2Service != nil {
			imageURLStr, err := s.R2Service.UploadFromBase64(
				context.Background(),
				*input.Image,
				"product-variants",
				fmt.Sprintf("%d", variant.ProductID),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to upload image: %v", err)
			}
			variant.Image = &imageURLStr
		} else if *input.Image == "" {
			// If empty string, set to nil (remove image)
			variant.Image = nil
		}
	}
	if input.Name != nil {
		variant.Name = *input.Name
	}
	if input.Price != nil {
		variant.Price = *input.Price
	}
	if input.PriceOriginal != nil {
		variant.PriceOriginal = input.PriceOriginal
	}
	if input.IsActive != nil {
		variant.IsActive = *input.IsActive
	}

	if err := s.DB.Save(&variant).Error; err != nil {
		return nil, err
	}

	return &variant, nil
}

func (s *ProductVariantService) Delete(ctx context.Context, id int64) (*model.ProductVariantDB, error) {
	var variant model.ProductVariantDB
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&variant).Error; err != nil {
		return nil, err
	}

	// Soft delete
	if err := s.DB.Delete(&variant).Error; err != nil {
		return nil, err
	}

	return &variant, nil
}

func (s *ProductVariantService) GetAll(ctx context.Context, pagination *model.PaginationInput, productID int64, isActive *bool) ([]*model.ProductVariantDB, *model.PageInfo, error) {
	var variants []*model.ProductVariantDB
	var total int64

	// Build query
	query := s.DB.Where("deleted_at IS NULL")

	if productID != 0 {
		query = query.Where("product_id = ?", productID)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Get total count
	if err := query.Model(&model.ProductVariantDB{}).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Apply pagination
	limit := int(*pagination.Limit)
	offset := (int(*pagination.Page) - 1) * limit

	if err := query.Limit(limit).Offset(offset).Find(&variants).Error; err != nil {
		return nil, nil, err
	}

	// Build pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	pageInfo := &model.PageInfo{
		CurrentPage:     *pagination.Page,
		PerPage:         *pagination.Limit,
		TotalItems:      int32(total),
		TotalPages:      int32(totalPages),
		HasNextPage:     *pagination.Page < int32(totalPages),
		HasPreviousPage: *pagination.Page > 1,
	}

	return variants, pageInfo, nil
}

func (s *ProductVariantService) GetByID(ctx context.Context, id int64) (*model.ProductVariantDB, error) {
	var variant model.ProductVariantDB
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&variant).Error; err != nil {
		return nil, err
	}

	return &variant, nil
}
