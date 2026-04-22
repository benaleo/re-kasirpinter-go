package helper

import (
	"re-kasirpinter-go/graph/model"
)

// Float64FromInt32 converts *int32 to float64, returns 0 if nil
func Float64FromInt32(v *int32) float64 {
	if v == nil {
		return 0
	}
	return float64(*v)
}

// ToGraphQLUser converts UserDB to GraphQL User model
func ToGraphQLUser(userDB model.UserDB, userRoleDB *model.UserRoleDB) *model.User {
	user := &model.User{
		ID:        userDB.ID,
		SecureID:  userDB.SecureID,
		Name:      userDB.Name,
		Email:     userDB.Email,
		Address:   userDB.Address,
		Phone:     userDB.Phone,
		Avatar:    userDB.Avatar,
		IsActive:  userDB.IsActive,
		DeletedAt: userDB.DeletedAt,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
	}

	// Set user role if provided
	if userRoleDB != nil && userRoleDB.ID > 0 {
		user.Role = ToGraphQLUserRole(*userRoleDB)
	}

	return user
}

// ToGraphQLUserRole converts UserRoleDB to GraphQL UserRole model
func ToGraphQLUserRole(userRoleDB model.UserRoleDB) *model.UserRole {
	// Convert permissions
	permissions := make([]*model.UserPermission, len(userRoleDB.Permissions))
	for i, permDB := range userRoleDB.Permissions {
		permissions[i] = ToGraphQLUserPermission(permDB)
	}

	return &model.UserRole{
		ID:          userRoleDB.ID,
		Name:        userRoleDB.Name,
		IsActive:    userRoleDB.IsActive,
		CreatedAt:   userRoleDB.CreatedAt,
		CreatedBy:   userRoleDB.CreatedBy,
		UpdatedAt:   userRoleDB.UpdatedAt,
		UpdatedBy:   userRoleDB.UpdatedBy,
		DeletedAt:   userRoleDB.DeletedAt,
		DeletedBy:   userRoleDB.DeletedBy,
		Permissions: permissions,
	}
}

// ToGraphQLUserPermission converts UserPermissionDB to GraphQL UserPermission model
func ToGraphQLUserPermission(userPermissionDB model.UserPermissionDB) *model.UserPermission {
	return &model.UserPermission{
		ID:   userPermissionDB.ID,
		Name: userPermissionDB.Name,
	}
}

// ToGraphQLIngredientCategory converts IngredientCategoryDB to GraphQL IngredientCategory model
func ToGraphQLIngredientCategory(ingredientCategoryDB model.IngredientCategoryDB) *model.IngredientCategory {
	return &model.IngredientCategory{
		ID:          ingredientCategoryDB.ID,
		Name:        ingredientCategoryDB.Name,
		Unit:        ingredientCategoryDB.Unit,
		ConvertUnit: ingredientCategoryDB.ConvertUnit,
		IsActive:    ingredientCategoryDB.IsActive,
		DeletedAt:   ingredientCategoryDB.DeletedAt,
		CreatedAt:   ingredientCategoryDB.CreatedAt,
		UpdatedAt:   ingredientCategoryDB.UpdatedAt,
	}
}

// ToGraphQLIngredient converts IngredientDB to GraphQL Ingredient model
func ToGraphQLIngredient(ingredientDB model.IngredientDB) *model.Ingredient {
	// Calculate total_stocks: sum of increase - sum of decrease (excluding deleted stocks)
	var totalStocks float64
	for _, stock := range ingredientDB.Stocks {
		// Skip deleted stocks
		if stock.DeletedAt != nil {
			continue
		}
		if stock.Type == model.IngredientStockTypeIncrease {
			totalStocks += stock.Qty
		} else if stock.Type == model.IngredientStockTypeDecrease {
			totalStocks -= stock.Qty
		}
	}

	ingredient := &model.Ingredient{
		ID:          ingredientDB.ID,
		Name:        ingredientDB.Name,
		CategoryID:  ingredientDB.CategoryID,
		IsActive:    ingredientDB.IsActive,
		DeletedAt:   ingredientDB.DeletedAt,
		CreatedAt:   ingredientDB.CreatedAt,
		UpdatedAt:   ingredientDB.UpdatedAt,
		TotalStocks: totalStocks,
	}

	// Set category if provided
	if ingredientDB.Category != nil && ingredientDB.Category.ID > 0 {
		ingredient.Category = ToGraphQLIngredientCategory(*ingredientDB.Category)
	}

	// Set stocks if provided (excluding deleted stocks)
	if len(ingredientDB.Stocks) > 0 {
		var activeStocks []model.IngredientStockDB
		for _, stockDB := range ingredientDB.Stocks {
			if stockDB.DeletedAt == nil {
				activeStocks = append(activeStocks, stockDB)
			}
		}
		if len(activeStocks) > 0 {
			stocks := make([]*model.IngredientStock, len(activeStocks))
			for i, stockDB := range activeStocks {
				stocks[i] = ToGraphQLIngredientStock(stockDB)
			}
			ingredient.Stocks = stocks
		}
	}

	return ingredient
}

// ToGraphQLIngredientStock converts IngredientStockDB to GraphQL IngredientStock model
func ToGraphQLIngredientStock(ingredientStockDB model.IngredientStockDB) *model.IngredientStock {
	stock := &model.IngredientStock{
		ID:           ingredientStockDB.ID,
		Code:         ingredientStockDB.Code,
		Qty:          ingredientStockDB.Qty,
		Type:         model.IngredientStockType(ingredientStockDB.Type),
		Capital:      ingredientStockDB.Capital,
		CapitalItem:  ingredientStockDB.CapitalItem,
		Message:      ingredientStockDB.Message,
		Image:        ingredientStockDB.Image,
		DeletedAt:    ingredientStockDB.DeletedAt,
		CreatedAt:    ingredientStockDB.CreatedAt,
		UpdatedAt:    ingredientStockDB.UpdatedAt,
		IngredientID: ingredientStockDB.IngredientID,
	}

	// Set ingredient if provided
	if ingredientStockDB.Ingredient != nil && ingredientStockDB.Ingredient.ID > 0 {
		stock.Ingredient = ToGraphQLIngredient(*ingredientStockDB.Ingredient)
	}

	return stock
}

// ToGraphQLProductCategory converts ProductCategoryDB to GraphQL ProductCategory model
func ToGraphQLProductCategory(productCategoryDB model.ProductCategoryDB) *model.ProductCategory {
	category := &model.ProductCategory{
		ID:          productCategoryDB.ID,
		Name:        productCategoryDB.Name,
		Description: productCategoryDB.Description,
		ParentID:    productCategoryDB.ParentID,
		IsActive:    productCategoryDB.IsActive,
		DeletedAt:   productCategoryDB.DeletedAt,
		CreatedAt:   productCategoryDB.CreatedAt,
		UpdatedAt:   productCategoryDB.UpdatedAt,
	}

	// Set parent if provided
	if productCategoryDB.Parent != nil && productCategoryDB.Parent.ID > 0 {
		category.Parent = ToGraphQLProductCategory(*productCategoryDB.Parent)
	}

	// Set children if provided (excluding deleted children)
	if len(productCategoryDB.Children) > 0 {
		var activeChildren []model.ProductCategoryDB
		for _, childDB := range productCategoryDB.Children {
			if childDB.DeletedAt == nil {
				activeChildren = append(activeChildren, childDB)
			}
		}
		if len(activeChildren) > 0 {
			children := make([]*model.ProductCategory, len(activeChildren))
			for i, childDB := range activeChildren {
				children[i] = ToGraphQLProductCategory(childDB)
			}
			category.Children = children
		}
	}

	return category
}

// ToGraphQLProduct converts ProductDB to GraphQL Product model
func ToGraphQLProduct(productDB model.ProductDB) *model.Product {
	product := &model.Product{
		ID:          productDB.ID,
		SecureID:    productDB.SecureID,
		Name:        productDB.Name,
		Image:       productDB.Image,
		CategoryID:  productDB.CategoryID,
		Description: productDB.Description,
		IsActive:    productDB.IsActive,
		DeletedAt:   productDB.DeletedAt,
		CreatedAt:   productDB.CreatedAt,
		UpdatedAt:   productDB.UpdatedAt,
	}

	// Set category if provided
	if productDB.Category != nil && productDB.Category.ID > 0 {
		product.Category = ToGraphQLProductCategory(*productDB.Category)
	}

	return product
}

// ToGraphQLDiscount converts DiscountDB to GraphQL Discount model
func ToGraphQLDiscount(discountDB model.DiscountDB) *model.Discount {
	return &model.Discount{
		ID:          discountDB.ID,
		Name:        discountDB.Name,
		Description: discountDB.Description,
		Icon:        discountDB.Icon,
		Code:        discountDB.Code,
		Type:        model.DiscountType(discountDB.Type),
		Value:       discountDB.Value,
		MaxValue:    discountDB.MaxValue,
		MinOrder:    discountDB.MinOrder,
		Quota:       discountDB.Quota,
		StartAt:     discountDB.StartAt,
		EndAt:       discountDB.EndAt,
		IsActive:    discountDB.IsActive,
		DeletedAt:   discountDB.DeletedAt,
		CreatedAt:   discountDB.CreatedAt,
		UpdatedAt:   discountDB.UpdatedAt,
	}
}
