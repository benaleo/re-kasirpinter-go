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

	// Set variants if provided (excluding deleted variants)
	if len(productDB.Variants) > 0 {
		var activeVariants []model.ProductVariantDB
		for _, variantDB := range productDB.Variants {
			if variantDB.DeletedAt == nil {
				activeVariants = append(activeVariants, variantDB)
			}
		}
		if len(activeVariants) > 0 {
			variants := make([]*model.ProductVariant, len(activeVariants))
			for i, variantDB := range activeVariants {
				variants[i] = ToGraphQLProductVariant(variantDB)
			}
			product.Variants = variants

			// Calculate is_available: true if any variant has available_stock > 0 or is_unlimited is true and is_active is true
			isAvailable := false
			for _, variant := range variants {
				if variant.IsActive && (variant.IsUnlimited || variant.AvailableStock > 0) {
					isAvailable = true
					break
				}
			}
			product.IsAvailable = isAvailable
		} else {
			product.IsAvailable = false
		}
	} else {
		product.IsAvailable = false
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

// ToGraphQLProductVariant converts ProductVariantDB to GraphQL ProductVariant model
func ToGraphQLProductVariant(productVariantDB model.ProductVariantDB) *model.ProductVariant {
	variant := &model.ProductVariant{
		ID:            productVariantDB.ID,
		Image:         productVariantDB.Image,
		ProductID:     productVariantDB.ProductID,
		Name:          productVariantDB.Name,
		Price:         productVariantDB.Price,
		PriceOriginal: productVariantDB.PriceOriginal,
		IsUnlimited:   productVariantDB.IsUnlimited,
		IsActive:      productVariantDB.IsActive,
		DeletedAt:     productVariantDB.DeletedAt,
		CreatedAt:     productVariantDB.CreatedAt,
		UpdatedAt:     productVariantDB.UpdatedAt,
	}

	// Set product if available
	if productVariantDB.Product != nil {
		variant.Product = ToGraphQLProduct(*productVariantDB.Product)
	}

	// Set ingredients if available and calculate available stock
	if len(productVariantDB.Ingredients) > 0 {
		ingredients := make([]*model.ProductIngredient, len(productVariantDB.Ingredients))
		for i, ingredientDB := range productVariantDB.Ingredients {
			ingredients[i] = ToGraphQLProductIngredient(ingredientDB)
		}
		variant.Ingredients = ingredients

		// If is_unlimited is true, set available_stock to -1 to indicate unlimited
		if productVariantDB.IsUnlimited {
			variant.AvailableStock = -1
		} else {
			// Calculate available stock based on ingredient availability
			// For each ingredient, calculate how many units can be made
			// The variant's available stock is the minimum of all these calculations
			var availableStock float64 = -1 // -1 means no ingredients
			for _, ingredientDB := range productVariantDB.Ingredients {
				if ingredientDB.Ingredient != nil {
					// Calculate total stock for this ingredient
					var totalStock float64
					for _, stock := range ingredientDB.Ingredient.Stocks {
						if stock.DeletedAt != nil {
							continue
						}
						if stock.Type == model.IngredientStockTypeIncrease {
							totalStock += stock.Qty
						} else if stock.Type == model.IngredientStockTypeDecrease {
							totalStock -= stock.Qty
						}
					}

					// Calculate how many units can be made from this ingredient
					if ingredientDB.IngredientValue > 0 {
						unitsFromIngredient := totalStock / ingredientDB.IngredientValue
						if availableStock == -1 || unitsFromIngredient < availableStock {
							availableStock = unitsFromIngredient
						}
					}
				}
			}

			// If availableStock is still -1, set it to 0 (no ingredients defined)
			if availableStock == -1 {
				availableStock = 0
			}

			// Ensure stock is not negative
			if availableStock < 0 {
				availableStock = 0
			}

			// Floor to integer (no decimals)
			availableStock = float64(int(availableStock))

			variant.AvailableStock = availableStock
		}
	} else {
		variant.AvailableStock = 0
	}

	return variant
}

// ToGraphQLProductIngredient converts ProductIngredientDB to GraphQL ProductIngredient model
func ToGraphQLProductIngredient(productIngredientDB model.ProductIngredientDB) *model.ProductIngredient {
	ingredient := &model.ProductIngredient{
		ID:              productIngredientDB.ID,
		VariantID:       productIngredientDB.VariantID,
		IngredientID:    productIngredientDB.IngredientID,
		IngredientValue: productIngredientDB.IngredientValue,
		Unit:            productIngredientDB.Unit,
		CreatedAt:       productIngredientDB.CreatedAt,
		UpdatedAt:       productIngredientDB.UpdatedAt,
	}

	// Set variant if available
	if productIngredientDB.Variant != nil {
		ingredient.Variant = ToGraphQLProductVariant(*productIngredientDB.Variant)
	}

	// Set ingredient if available
	if productIngredientDB.Ingredient != nil {
		ingredient.Ingredient = ToGraphQLIngredient(*productIngredientDB.Ingredient)
	}

	return ingredient
}

// ToGraphQLProductIngredients converts []ProductIngredientDB to []*model.ProductIngredient
func ToGraphQLProductIngredients(productIngredientsDB []model.ProductIngredientDB) []*model.ProductIngredient {
	ingredients := make([]*model.ProductIngredient, len(productIngredientsDB))
	for i, ingredientDB := range productIngredientsDB {
		ingredients[i] = ToGraphQLProductIngredient(ingredientDB)
	}
	return ingredients
}

// ToGraphQLProductIngredientSlice converts []*ProductIngredientDB to []*model.ProductIngredient
func ToGraphQLProductIngredientSlice(productIngredientsDB []*model.ProductIngredientDB) []*model.ProductIngredient {
	ingredients := make([]*model.ProductIngredient, len(productIngredientsDB))
	for i, ingredientDB := range productIngredientsDB {
		ingredients[i] = ToGraphQLProductIngredient(*ingredientDB)
	}
	return ingredients
}
