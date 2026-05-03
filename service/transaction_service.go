package service

import (
	"context"
	"fmt"
	"re-kasirpinter-go/graph/model"
	"time"

	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{
		DB: db,
	}
}

// generateInvoice generates invoice number with format: sequence-date-month-year/KASIRPINTER
func (s *TransactionService) generateInvoice(date time.Time, sequence int32) string {
	sequenceStr := fmt.Sprintf("%04d", sequence)
	dateStr := fmt.Sprintf("%02d%02d%04d", date.Day(), date.Month(), date.Year())
	return fmt.Sprintf("%s-%s/KASIRPINTER", sequenceStr, dateStr)
}

// getNextSequence gets the next sequence number for the given date
func (s *TransactionService) getNextSequence(date time.Time) (int32, error) {
	var maxSequence int32 = 0

	// Find the maximum sequence for the given date
	err := s.DB.Model(&model.TransactionDB{}).
		Select("COALESCE(MAX(sequence), 0)").
		Where("DATE(date) = ?", date.Format("2006-01-02")).
		Scan(&maxSequence).Error

	if err != nil {
		return 1, err
	}

	return maxSequence + 1, nil
}

// CreateTransaction creates a new transaction with its products and extras
func (s *TransactionService) CreateTransaction(ctx context.Context, input model.CreateTransactionInput) (*model.CreateTransactionResponse, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current date and sequence
	now := time.Now()
	sequence, err := s.getNextSequence(now)
	if err != nil {
		tx.Rollback()
		return &model.CreateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to generate sequence number: " + err.Error(),
		}, err
	}

	// Generate invoice number
	invoice := s.generateInvoice(now, sequence)

	// Create transaction
	transaction := model.TransactionDB{
		Date:          now,
		Sequence:      sequence,
		Invoice:       invoice,
		PaymentMethod: input.PaymentMethod,
		TotalAmount:   input.TotalAmount,
		TotalBilled:   input.TotalBilled,
		Tax:           input.Tax,
		Subtotal:      input.Subtotal,
		Discount:      input.Discount,
		CustomerID:    input.CustomerID,
		CreatedBy:     input.CreatedBy,
		IsCompleted:   false,
		IsCanceled:    false,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return &model.CreateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to create transaction: " + err.Error(),
		}, err
	}

	// Create transaction products and extras
	for _, productInput := range input.Products {
		// Create transaction product
		var attribute string
		if productInput.Attribute != nil {
			attribute = *productInput.Attribute
		}

		product := model.TransactionProductDB{
			TransactionID: transaction.ID,
			ProductID:     productInput.ProductID,
			AvailableType: productInput.AvailableType,
			VariantType:   productInput.VariantType,
			Attribute:     attribute,
			VariantName:   productInput.VariantName,
			Quantity:      productInput.Quantity,
			ProductPrice:  productInput.ProductPrice,
			TotalExtras:   productInput.TotalExtras,
			TotalPrice:    productInput.TotalPrice,
			Notes:         productInput.Notes,
		}

		if err := tx.Create(&product).Error; err != nil {
			tx.Rollback()
			return &model.CreateTransactionResponse{
				Code:    500,
				Success: false,
				Message: "Failed to create transaction product: " + err.Error(),
			}, err
		}

		// Create transaction extras
		for _, extraInput := range productInput.Extras {
			extra := model.TransactionExtraDB{
				TransactionProductID: product.ID,
				ExtraName:            extraInput.ExtraName,
				Quantity:             extraInput.Quantity,
				ExtraPrice:           extraInput.ExtraPrice,
			}

			if err := tx.Create(&extra).Error; err != nil {
				tx.Rollback()
				return &model.CreateTransactionResponse{
					Code:    500,
					Success: false,
					Message: "Failed to create transaction extra: " + err.Error(),
				}, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &model.CreateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to commit transaction: " + err.Error(),
		}, err
	}

	// Get the complete transaction with relations
	var result model.TransactionDB
	if err := s.DB.Preload("Customer").Preload("Products.Product").Preload("Products.Extras").First(&result, transaction.ID).Error; err != nil {
		return &model.CreateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve created transaction: " + err.Error(),
		}, err
	}

	return &model.CreateTransactionResponse{
		Code:    201,
		Success: true,
		Message: "Transaction created successfully",
		Data:    s.convertTransactionDBToGraphQL(&result),
	}, nil
}

// GetTransactions retrieves paginated transactions
func (s *TransactionService) GetTransactions(ctx context.Context, pagination model.PaginationInput) (*model.TransactionsResponse, error) {
	var transactions []model.TransactionDB
	var total int64

	// Parse sort by
	sortBy := "created_at desc"
	if pagination.SortBy != nil && *pagination.SortBy != "" {
		sortBy = *pagination.SortBy
	}

	// Get total count
	if err := s.DB.Model(&model.TransactionDB{}).Count(&total).Error; err != nil {
		return &model.TransactionsResponse{
			Code:    500,
			Success: false,
			Message: "Failed to count transactions: " + err.Error(),
		}, err
	}

	// Calculate pagination
	limit := int(*pagination.Limit)
	page := int(*pagination.Page)
	offset := (page - 1) * limit

	// Get transactions with relations
	if err := s.DB.Preload("Customer").
		Preload("Products.Product").
		Preload("Products.Extras").
		Order(sortBy).
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return &model.TransactionsResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve transactions: " + err.Error(),
		}, err
	}

	// Convert to GraphQL
	var result []*model.Transaction
	for _, tx := range transactions {
		result = append(result, s.convertTransactionDBToGraphQL(&tx))
	}

	// Create pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	paginationInfo := model.PageInfo{
		CurrentPage:     int32(page),
		PerPage:         int32(limit),
		TotalItems:      int32(total),
		TotalPages:      int32(totalPages),
		HasNextPage:     page < totalPages,
		HasPreviousPage: page > 1,
	}

	return &model.TransactionsResponse{
		Code:       200,
		Success:    true,
		Message:    "Transactions retrieved successfully",
		Data:       result,
		Pagination: &paginationInfo,
	}, nil
}

// GetTransactionByID retrieves a single transaction by ID
func (s *TransactionService) GetTransactionByID(ctx context.Context, id int64) (*model.TransactionResponse, error) {
	var transaction model.TransactionDB

	if err := s.DB.Preload("Customer").
		Preload("Products.Product").
		Preload("Products.Extras").
		First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.TransactionResponse{
				Code:    404,
				Success: false,
				Message: "Transaction not found",
			}, nil
		}
		return &model.TransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve transaction: " + err.Error(),
		}, err
	}

	return &model.TransactionResponse{
		Code:    200,
		Success: true,
		Message: "Transaction retrieved successfully",
		Data:    s.convertTransactionDBToGraphQL(&transaction),
	}, nil
}

// UpdateTransaction updates an existing transaction
func (s *TransactionService) UpdateTransaction(ctx context.Context, id int64, input model.UpdateTransactionInput) (*model.UpdateTransactionResponse, error) {
	var transaction model.TransactionDB

	// Check if transaction exists
	if err := s.DB.First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.UpdateTransactionResponse{
				Code:    404,
				Success: false,
				Message: "Transaction not found",
			}, nil
		}
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve transaction: " + err.Error(),
		}, err
	}

	// Update fields
	if input.PaymentMethod != nil {
		transaction.PaymentMethod = *input.PaymentMethod
	}
	if input.TotalAmount != nil {
		transaction.TotalAmount = *input.TotalAmount
	}
	if input.TotalBilled != nil {
		transaction.TotalBilled = *input.TotalBilled
	}
	if input.Tax != nil {
		transaction.Tax = *input.Tax
	}
	if input.Subtotal != nil {
		transaction.Subtotal = *input.Subtotal
	}
	if input.Discount != nil {
		transaction.Discount = *input.Discount
	}
	if input.CustomerID != nil {
		transaction.CustomerID = input.CustomerID
	}
	if input.IsCompleted != nil {
		transaction.IsCompleted = *input.IsCompleted
	}
	if input.IsCanceled != nil {
		transaction.IsCanceled = *input.IsCanceled
	}
	if input.UpdatedBy != nil {
		transaction.UpdatedBy = input.UpdatedBy
	}

	if err := s.DB.Save(&transaction).Error; err != nil {
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to update transaction: " + err.Error(),
		}, err
	}

	// Get updated transaction with relations
	var result model.TransactionDB
	if err := s.DB.Preload("Customer").Preload("Products.Product").Preload("Products.Extras").First(&result, id).Error; err != nil {
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve updated transaction: " + err.Error(),
		}, err
	}

	return &model.UpdateTransactionResponse{
		Code:    200,
		Success: true,
		Message: "Transaction updated successfully",
		Data:    s.convertTransactionDBToGraphQL(&result),
	}, nil
}

// DeleteTransaction soft deletes a transaction
func (s *TransactionService) DeleteTransaction(ctx context.Context, id int64) (*model.DeleteTransactionResponse, error) {
	var transaction model.TransactionDB

	// Check if transaction exists
	if err := s.DB.First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.DeleteTransactionResponse{
				Code:    404,
				Success: false,
				Message: "Transaction not found",
			}, nil
		}
		return &model.DeleteTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve transaction: " + err.Error(),
		}, err
	}

	// Soft delete transaction (this will cascade delete related products and extras)
	if err := s.DB.Delete(&transaction).Error; err != nil {
		return &model.DeleteTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to delete transaction: " + err.Error(),
		}, err
	}

	return &model.DeleteTransactionResponse{
		Code:    200,
		Success: true,
		Message: "Transaction deleted successfully",
		Data:    s.convertTransactionDBToGraphQL(&transaction),
	}, nil
}

// Helper function to convert TransactionDB to GraphQL Transaction
func (s *TransactionService) convertTransactionDBToGraphQL(tx *model.TransactionDB) *model.Transaction {
	var customer *model.CustomerSearchData
	if tx.Customer != nil {
		customer = &model.CustomerSearchData{
			SecureID: tx.Customer.SecureID,
			Name:     tx.Customer.Name,
			Phone:    tx.Customer.Phone,
		}
	}

	var products []*model.TransactionProduct
	for _, product := range tx.Products {
		var extras []*model.TransactionExtra
		for _, extra := range product.Extras {
			extras = append(extras, &model.TransactionExtra{
				ID:                   extra.ID,
				TransactionProductID: extra.TransactionProductID,
				ExtraName:            extra.ExtraName,
				Quantity:             extra.Quantity,
				ExtraPrice:           extra.ExtraPrice,
				CreatedAt:            extra.CreatedAt,
			})
		}

		products = append(products, &model.TransactionProduct{
			ID:            product.ID,
			TransactionID: product.TransactionID,
			ProductID:     product.ProductID,
			AvailableType: product.AvailableType,
			VariantType:   product.VariantType,
			Attribute:     &product.Attribute,
			VariantName:   product.VariantName,
			Quantity:      product.Quantity,
			ProductPrice:  product.ProductPrice,
			TotalExtras:   product.TotalExtras,
			TotalPrice:    product.TotalPrice,
			Notes:         product.Notes,
			CreatedAt:     product.CreatedAt,
			Extras:        extras,
		})
	}

	return &model.Transaction{
		ID:            tx.ID,
		Date:          tx.Date,
		Sequence:      tx.Sequence,
		Invoice:       tx.Invoice,
		PaymentMethod: tx.PaymentMethod,
		TotalAmount:   tx.TotalAmount,
		TotalBilled:   tx.TotalBilled,
		Tax:           tx.Tax,
		Subtotal:      tx.Subtotal,
		Discount:      tx.Discount,
		CustomerID:    tx.CustomerID,
		Customer:      customer,
		IsCompleted:   tx.IsCompleted,
		IsCanceled:    tx.IsCanceled,
		CreatedAt:     tx.CreatedAt,
		CreatedBy:     tx.CreatedBy,
		UpdatedAt:     tx.UpdatedAt,
		UpdatedBy:     tx.UpdatedBy,
		Products:      products,
	}
}
