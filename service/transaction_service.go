package service

import (
	"context"
	"fmt"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"strings"
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

	// Generate secure_id
	secureID, err := helper.GenerateRandomString(32)
	if err != nil {
		return &model.CreateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to generate secure ID: " + err.Error(),
		}, err
	}

	// Get current date and sequence
	now := time.Now()
	sequence, err := s.getNextSequence(now)
	if err != nil {
		return &model.CreateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to get next sequence: " + err.Error(),
		}, err
	}

	// Generate invoice
	invoice := s.generateInvoice(now, sequence)

	// Create transaction
	transaction := model.TransactionDB{
		SecureID:      &secureID,
		Date:          now.Format("2006-01-02"),
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
		IsCompleted:   input.IsCompleted != nil && *input.IsCompleted,
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
func (s *TransactionService) GetTransactions(ctx context.Context, pagination model.PaginationInput, date *string, isCompleted *bool, isCanceled *bool) (*model.TransactionsResponse, error) {
	var transactions []model.TransactionDB
	var total int64

	// Parse sort by
	sortBy := "created_at desc"
	if pagination.SortBy != nil && *pagination.SortBy != "" {
		// Replace comma with space for proper SQL ORDER BY syntax
		sortBy = strings.ReplaceAll(*pagination.SortBy, ",", " ")
	}

	// Build query
	query := s.DB.Model(&model.TransactionDB{})

	// Filter by date if provided
	if date != nil && *date != "" {
		query = query.Where("date = ?", *date)
	}

	// Filter by is_completed if provided
	if isCompleted != nil {
		query = query.Where("is_completed = ?", *isCompleted)
	}

	// Filter by is_canceled if provided
	if isCanceled != nil {
		query = query.Where("is_canceled = ?", *isCanceled)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
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
	if err := query.Preload("Customer").
		Preload("Products").
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

	// Manually load products if preload didn't work
	for i := range transactions {
		if len(transactions[i].Products) > 0 {
			for j := range transactions[i].Products {
				var product model.ProductDB
				if err := s.DB.Where("id = ?", transactions[i].Products[j].ProductID).First(&product).Error; err == nil {
					transactions[i].Products[j].Product = &product
				}
			}
		}
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

// GetTransactionByID retrieves a single transaction by secure_id
func (s *TransactionService) GetTransactionByID(ctx context.Context, secureID string) (*model.TransactionResponse, error) {
	var transaction model.TransactionDB

	if err := s.DB.Preload("Customer").
		Preload("Products.Product").
		Preload("Products.Extras").
		Where("secure_id = ?", secureID).
		First(&transaction).Error; err != nil {
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
func (s *TransactionService) UpdateTransaction(ctx context.Context, secureID string, input model.UpdateTransactionInput) (*model.UpdateTransactionResponse, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var transaction model.TransactionDB

	// Check if transaction exists
	if err := tx.Where("secure_id = ?", secureID).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return &model.UpdateTransactionResponse{
				Code:    404,
				Success: false,
				Message: "Transaction not found",
			}, nil
		}
		tx.Rollback()
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

	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to update transaction: " + err.Error(),
		}, err
	}

	// Remove all existing transaction products and their extras for this transaction
	// First, get all transaction product IDs for this transaction
	var transactionProductIDs []int64
	if err := tx.Model(&model.TransactionProductDB{}).Where("transaction_id = ?", transaction.ID).Pluck("id", &transactionProductIDs).Error; err != nil {
		tx.Rollback()
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to get transaction product IDs: " + err.Error(),
		}, err
	}

	// Delete transaction extras using the correct column (transaction_product_id)
	if len(transactionProductIDs) > 0 {
		if err := tx.Where("transaction_product_id IN ?", transactionProductIDs).Delete(&model.TransactionExtraDB{}).Error; err != nil {
			tx.Rollback()
			return &model.UpdateTransactionResponse{
				Code:    500,
				Success: false,
				Message: "Failed to delete transaction extras: " + err.Error(),
			}, err
		}
	}

	// Delete transaction products
	if err := tx.Where("transaction_id = ?", transaction.ID).Delete(&model.TransactionProductDB{}).Error; err != nil {
		tx.Rollback()
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to delete transaction products: " + err.Error(),
		}, err
	}

	// Create new transaction products and extras from input
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
			return &model.UpdateTransactionResponse{
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
				return &model.UpdateTransactionResponse{
					Code:    500,
					Success: false,
					Message: "Failed to create transaction extra: " + err.Error(),
				}, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to commit transaction: " + err.Error(),
		}, err
	}

	// Get updated transaction with relations
	var result model.TransactionDB
	if err := s.DB.Preload("Customer").Preload("Products.Product").Preload("Products.Extras").Where("secure_id = ?", secureID).First(&result).Error; err != nil {
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

// CancelTransaction cancels a transaction by setting is_canceled to true
func (s *TransactionService) CancelTransaction(ctx context.Context, secureID string) (*model.UpdateTransactionResponse, error) {
	var transaction model.TransactionDB

	// Check if transaction exists
	if err := s.DB.Where("secure_id = ?", secureID).First(&transaction).Error; err != nil {
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

	// Check if transaction is already completed
	if transaction.IsCompleted {
		return &model.UpdateTransactionResponse{
			Code:    400,
			Success: false,
			Message: "Cannot cancel completed transaction",
		}, nil
	}

	// Check if transaction is already canceled
	if transaction.IsCanceled {
		return &model.UpdateTransactionResponse{
			Code:    400,
			Success: false,
			Message: "Transaction is already canceled",
		}, nil
	}

	// Cancel transaction
	transaction.IsCanceled = true
	if err := s.DB.Save(&transaction).Error; err != nil {
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to cancel transaction: " + err.Error(),
		}, err
	}

	// Get updated transaction with relations
	var result model.TransactionDB
	if err := s.DB.Preload("Customer").Preload("Products.Product").Preload("Products.Extras").Where("secure_id = ?", secureID).First(&result).Error; err != nil {
		return &model.UpdateTransactionResponse{
			Code:    500,
			Success: false,
			Message: "Failed to retrieve updated transaction: " + err.Error(),
		}, err
	}

	return &model.UpdateTransactionResponse{
		Code:    200,
		Success: true,
		Message: "Transaction canceled successfully",
		Data:    s.convertTransactionDBToGraphQL(&result),
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

		// Convert ProductDB to GraphQL Product
		var productData *model.Product
		if product.Product != nil {
			productData = &model.Product{
				ID:            product.Product.ID,
				SecureID:      product.Product.SecureID,
				Name:          product.Product.Name,
				Image:         product.Product.Image,
				CategoryID:    product.Product.CategoryID,
				Description:   product.Product.Description,
				AvailableType: product.Product.AvailableType,
				VariantType:   product.Product.VariantType,
				IsActive:      product.Product.IsActive,
				DeletedAt:     product.Product.DeletedAt,
				CreatedAt:     product.Product.CreatedAt,
				UpdatedAt:     product.Product.UpdatedAt,
			}
		}

		products = append(products, &model.TransactionProduct{
			ID:            product.ID,
			TransactionID: product.TransactionID,
			ProductID:     product.ProductID,
			Product:       productData,
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

	var id string
	if tx.SecureID != nil {
		id = *tx.SecureID
	}

	return &model.Transaction{
		ID:            id,
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
