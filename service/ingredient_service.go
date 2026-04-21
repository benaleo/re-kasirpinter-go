package service

import (
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type IngredientService struct {
	DB *gorm.DB
}

func NewIngredientService(db *gorm.DB) (*IngredientService, error) {
	return &IngredientService{
		DB: db,
	}, nil
}

func (s *IngredientService) CreateIngredient(input model.CreateIngredientInput) (*model.CreateIngredientResponse, error) {
	// Create ingredient DB model
	ingredientDB := model.IngredientDB{
		Name:       input.Name,
		CategoryID: input.CategoryID,
		IsActive:   input.IsActive,
	}

	// Save to database
	result := s.DB.Create(&ingredientDB)
	if result.Error != nil {
		return &model.CreateIngredientResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create ingredient: %v", result.Error),
		}, nil
	}

	// Reload with category and stocks
	s.DB.Preload("Category").Preload("Stocks", "deleted_at IS NULL").First(&ingredientDB, ingredientDB.ID)

	// Convert DB model to GraphQL model
	ingredient := helper.ToGraphQLIngredient(ingredientDB)

	return &model.CreateIngredientResponse{
		Code:    201,
		Success: true,
		Message: "ingredient created successfully",
		Data:    ingredient,
	}, nil
}

func (s *IngredientService) UpdateIngredient(id int64, input model.UpdateIngredientInput) (*model.UpdateIngredientResponse, error) {
	// Find ingredient by ID
	var ingredientDB model.IngredientDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&ingredientDB)
	if result.Error != nil {
		return &model.UpdateIngredientResponse{
			Code:    404,
			Success: false,
			Message: "ingredient not found",
		}, nil
	}

	// Update fields if provided
	if input.Name != nil {
		ingredientDB.Name = *input.Name
	}
	if input.CategoryID != nil {
		ingredientDB.CategoryID = input.CategoryID
	}
	if input.IsActive != nil {
		ingredientDB.IsActive = *input.IsActive
	}

	// Save to database
	result = s.DB.Save(&ingredientDB)
	if result.Error != nil {
		return &model.UpdateIngredientResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update ingredient: %v", result.Error),
		}, nil
	}

	// Reload with category and stocks
	s.DB.Preload("Category").Preload("Stocks", "deleted_at IS NULL").First(&ingredientDB, ingredientDB.ID)

	// Convert DB model to GraphQL model
	ingredient := helper.ToGraphQLIngredient(ingredientDB)

	return &model.UpdateIngredientResponse{
		Code:    200,
		Success: true,
		Message: "ingredient updated successfully",
		Data:    ingredient,
	}, nil
}

func (s *IngredientService) DeleteIngredient(id int64) (*model.DeleteIngredientResponse, error) {
	// Find ingredient by ID
	var ingredientDB model.IngredientDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&ingredientDB)
	if result.Error != nil {
		return &model.DeleteIngredientResponse{
			Code:    404,
			Success: false,
			Message: "ingredient not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	ingredientDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&ingredientDB)
	if result.Error != nil {
		return &model.DeleteIngredientResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete ingredient: %v", result.Error),
		}, nil
	}

	// Reload with category and stocks
	s.DB.Preload("Category").Preload("Stocks", "deleted_at IS NULL").First(&ingredientDB, ingredientDB.ID)

	// Convert DB model to GraphQL model
	ingredient := helper.ToGraphQLIngredient(ingredientDB)

	return &model.DeleteIngredientResponse{
		Code:    200,
		Success: true,
		Message: "ingredient deleted successfully",
		Data:    ingredient,
	}, nil
}
