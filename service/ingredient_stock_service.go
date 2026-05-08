package service

import (
	"fmt"
	"math"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"gorm.io/gorm"
)

type IngredientStockService struct {
	DB *gorm.DB
}

func NewIngredientStockService(db *gorm.DB) (*IngredientStockService, error) {
	return &IngredientStockService{
		DB: db,
	}, nil
}

func (s *IngredientStockService) IngredientStocks(pagination *model.PaginationInput, ingredientID *int64) (*model.IngredientStocksResponse, error) {
	// Parse pagination parameters
	params := helper.ParsePagination(pagination)

	// Build base query with ingredient preload
	baseQuery := s.DB.Model(&model.IngredientStockDB{}).Preload("Ingredient").Preload("Ingredient.Category").Preload("Ingredient.Stocks", "deleted_at IS NULL").Where("deleted_at IS NULL")

	// Filter by ingredient_id if provided
	if ingredientID != nil {
		baseQuery = baseQuery.Where("ingredient_id = ?", *ingredientID)
	}

	// Get total count
	var total int64
	countResult := baseQuery.Count(&total)
	if countResult.Error != nil {
		return &model.IngredientStocksResponse{
			Code:           500,
			Success:        false,
			Message:        fmt.Sprintf("failed to count ingredient stocks: %v", countResult.Error),
			IngredientName: nil,
			TotalStocks:    nil,
			Unit:           nil,
			ConvertUnit:    nil,
		}, nil
	}

	// Query ingredient stocks with pagination
	paginationResult := helper.BuildPaginationResult(params, total, 0)
	var stocksDB []model.IngredientStockDB
	result := baseQuery.Order(paginationResult.SortBy).Limit(int(paginationResult.Limit)).Offset(paginationResult.Offset).Find(&stocksDB)
	if result.Error != nil {
		return &model.IngredientStocksResponse{
			Code:           500,
			Success:        false,
			Message:        fmt.Sprintf("failed to retrieve ingredient stocks: %v", result.Error),
			IngredientName: nil,
			TotalStocks:    nil,
			Unit:           nil,
			ConvertUnit:    nil,
		}, nil
	}

	// Rebuild pagination result with actual item count
	paginationResult = helper.BuildPaginationResult(params, total, len(stocksDB))

	// Convert DB models to GraphQL models
	stocks := make([]*model.IngredientStock, len(stocksDB))
	for i, stockDB := range stocksDB {
		stocks[i] = helper.ToGraphQLIngredientStock(stockDB)
	}

	// Calculate summary fields if filtered by ingredient
	var ingredientName *string
	var totalStocks *float64
	var unit *string
	var convertUnit *string
	var convertCalc *int32

	if ingredientID != nil {
		// Query ingredient data separately to ensure it's available even when no stocks exist
		var ingredientDB model.IngredientDB
		ingredientResult := s.DB.Preload("Category").Where("id = ? AND deleted_at IS NULL", *ingredientID).First(&ingredientDB)

		if ingredientResult.Error == nil {
			name := ingredientDB.Name
			ingredientName = &name

			// Get unit, convert_unit, and convert_calc from ingredient's category
			if ingredientDB.Category != nil {
				u := ingredientDB.Category.Unit
				unit = &u
				convertUnit = ingredientDB.Category.ConvertUnit
				calc := ingredientDB.Category.ConvertCalc
				convertCalc = &calc
			}
		}

		// Calculate total stocks from the stocks data
		if len(stocksDB) > 0 {
			total := 0.0
			for _, stock := range stocksDB {
				total += stock.Qty
			}
			totalStocks = &total
		}
	}

	return &model.IngredientStocksResponse{
		Code:           200,
		Success:        true,
		Message:        "ingredient stocks retrieved successfully",
		IngredientName: ingredientName,
		TotalStocks:    totalStocks,
		Unit:           unit,
		ConvertUnit:    convertUnit,
		ConvertCalc:    convertCalc,
		Data:           stocks,
		Pagination:     paginationResult.PageInfo,
	}, nil
}

func (s *IngredientStockService) CreateIngredientStock(input model.CreateIngredientStockInput) (*model.CreateIngredientStockResponse, error) {
	// Calculate capital_item as capital divided by qty, rounded to 2 decimal places
	capitalItem := math.Round(input.Capital/input.Qty*100) / 100

	// Create ingredient stock DB model
	stockDB := model.IngredientStockDB{
		Code:         input.Code,
		Qty:          input.Qty,
		Type:         model.IngredientStockType(input.Type),
		Capital:      input.Capital,
		CapitalItem:  capitalItem,
		Message:      input.Message,
		Image:        input.Image,
		IngredientID: input.IngredientID,
	}

	// Save to database
	result := s.DB.Create(&stockDB)
	if result.Error != nil {
		return &model.CreateIngredientStockResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create ingredient stock: %v", result.Error),
		}, nil
	}

	// Reload with ingredient and category
	s.DB.Preload("Ingredient").Preload("Ingredient.Category").First(&stockDB, stockDB.ID)

	// Convert DB model to GraphQL model
	stock := helper.ToGraphQLIngredientStock(stockDB)

	return &model.CreateIngredientStockResponse{
		Code:    201,
		Success: true,
		Message: "ingredient stock created successfully",
		Data:    stock,
	}, nil
}

func (s *IngredientStockService) UpdateIngredientStock(id int64, input model.UpdateIngredientStockInput) (*model.UpdateIngredientStockResponse, error) {
	// Find ingredient stock by ID
	var stockDB model.IngredientStockDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&stockDB)
	if result.Error != nil {
		return &model.UpdateIngredientStockResponse{
			Code:    404,
			Success: false,
			Message: "ingredient stock not found",
		}, nil
	}

	// Update fields if provided
	needsRecalculation := false
	if input.Code != nil {
		stockDB.Code = input.Code
	}
	if input.Qty != nil {
		stockDB.Qty = *input.Qty
		needsRecalculation = true
	}
	if input.Type != nil {
		stockDB.Type = model.IngredientStockType(*input.Type)
	}
	if input.Capital != nil {
		stockDB.Capital = *input.Capital
		needsRecalculation = true
	}
	if input.Message != nil {
		stockDB.Message = input.Message
	}
	if input.Image != nil {
		stockDB.Image = input.Image
	}
	if input.IngredientID != nil {
		stockDB.IngredientID = *input.IngredientID
	}

	// Recalculate capital_item if capital or qty changed, rounded to 2 decimal places
	if needsRecalculation && stockDB.Qty > 0 {
		stockDB.CapitalItem = math.Round(stockDB.Capital/stockDB.Qty*100) / 100
	}

	// Save to database
	result = s.DB.Save(&stockDB)
	if result.Error != nil {
		return &model.UpdateIngredientStockResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update ingredient stock: %v", result.Error),
		}, nil
	}

	// Reload with ingredient and category
	s.DB.Preload("Ingredient").Preload("Ingredient.Category").First(&stockDB, stockDB.ID)

	// Convert DB model to GraphQL model
	stock := helper.ToGraphQLIngredientStock(stockDB)

	return &model.UpdateIngredientStockResponse{
		Code:    200,
		Success: true,
		Message: "ingredient stock updated successfully",
		Data:    stock,
	}, nil
}

func (s *IngredientStockService) DeleteIngredientStock(id int64) (*model.DeleteIngredientStockResponse, error) {
	// Find ingredient stock by ID
	var stockDB model.IngredientStockDB
	result := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&stockDB)
	if result.Error != nil {
		return &model.DeleteIngredientStockResponse{
			Code:    404,
			Success: false,
			Message: "ingredient stock not found",
		}, nil
	}

	// Soft delete by setting deleted_at
	now := time.Now()
	stockDB.DeletedAt = &now

	// Save to database
	result = s.DB.Save(&stockDB)
	if result.Error != nil {
		return &model.DeleteIngredientStockResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to delete ingredient stock: %v", result.Error),
		}, nil
	}

	// Reload with ingredient and category
	s.DB.Preload("Ingredient").Preload("Ingredient.Category").First(&stockDB, stockDB.ID)

	// Convert DB model to GraphQL model
	stock := helper.ToGraphQLIngredientStock(stockDB)

	return &model.DeleteIngredientStockResponse{
		Code:    200,
		Success: true,
		Message: "ingredient stock deleted successfully",
		Data:    stock,
	}, nil
}
