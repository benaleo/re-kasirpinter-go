package service

import (
	"context"
	"re-kasirpinter-go/graph/model"

	"gorm.io/gorm"
)

type ProductIngredientService struct {
	DB *gorm.DB
}

func NewProductIngredientService(db *gorm.DB) *ProductIngredientService {
	return &ProductIngredientService{DB: db}
}

func (s *ProductIngredientService) Create(ctx context.Context, input []*model.CreateProductIngredientInput) ([]*model.ProductIngredientDB, error) {
	var productIngredients []*model.ProductIngredientDB

	for _, item := range input {
		// Check if variant exists
		var variant model.ProductVariantDB
		if err := s.DB.Where("id = ? AND deleted_at IS NULL", item.VariantID).First(&variant).Error; err != nil {
			return nil, err
		}

		// Check if ingredient exists
		var ingredient model.IngredientDB
		if err := s.DB.Where("id = ? AND deleted_at IS NULL", item.IngredientID).First(&ingredient).Error; err != nil {
			return nil, err
		}

		// Check if this combination already exists
		var existing model.ProductIngredientDB
		err := s.DB.Where("variant_id = ? AND ingredient_id = ?", item.VariantID, item.IngredientID).First(&existing).Error
		if err == nil {
			return nil, gorm.ErrDuplicatedKey
		}

		productIngredient := &model.ProductIngredientDB{
			VariantID:       item.VariantID,
			IngredientID:    item.IngredientID,
			IngredientValue: item.IngredientValue,
			Unit:            item.Unit,
		}

		if err := s.DB.Create(productIngredient).Error; err != nil {
			return nil, err
		}

		productIngredients = append(productIngredients, productIngredient)
	}

	// Preload variant and ingredient relationships for all created items
	for _, productIngredient := range productIngredients {
		if err := s.DB.Preload("Variant").Preload("Ingredient").First(productIngredient, productIngredient.ID).Error; err != nil {
			return nil, err
		}
	}

	return productIngredients, nil
}

func (s *ProductIngredientService) Update(ctx context.Context, id int64, input model.UpdateProductIngredientInput) (*model.ProductIngredientDB, error) {
	var productIngredient model.ProductIngredientDB
	if err := s.DB.Preload("Variant").Preload("Ingredient").Where("id = ?", id).First(&productIngredient).Error; err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.IngredientValue != nil {
		productIngredient.IngredientValue = *input.IngredientValue
	}
	if input.Unit != nil {
		productIngredient.Unit = *input.Unit
	}

	if err := s.DB.Save(&productIngredient).Error; err != nil {
		return nil, err
	}

	return &productIngredient, nil
}

func (s *ProductIngredientService) Delete(ctx context.Context, variantID int64) ([]*model.ProductIngredientDB, error) {
	var productIngredients []*model.ProductIngredientDB
	if err := s.DB.Where("variant_id = ?", variantID).Find(&productIngredients).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Where("variant_id = ?", variantID).Delete(&model.ProductIngredientDB{}).Error; err != nil {
		return nil, err
	}

	return productIngredients, nil
}

func (s *ProductIngredientService) GetAll(ctx context.Context, pagination *model.PaginationInput, variantID int64) ([]*model.ProductIngredientDB, *model.PageInfo, error) {
	var productIngredients []*model.ProductIngredientDB
	var total int64

	// Build query
	query := s.DB.Model(&model.ProductIngredientDB{})

	if variantID != 0 {
		query = query.Where("variant_id = ?", variantID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Apply pagination
	limit := int(*pagination.Limit)
	offset := (int(*pagination.Page) - 1) * limit

	if err := query.Preload("Variant").Preload("Ingredient").Limit(limit).Offset(offset).Find(&productIngredients).Error; err != nil {
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

	return productIngredients, pageInfo, nil
}

func (s *ProductIngredientService) GetByID(ctx context.Context, id int64) (*model.ProductIngredientDB, error) {
	var productIngredient model.ProductIngredientDB
	if err := s.DB.Where("id = ?", id).First(&productIngredient).Error; err != nil {
		return nil, err
	}

	return &productIngredient, nil
}
