package helper

import (
	"re-kasirpinter-go/graph/model"
)

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
		Status:      ingredientCategoryDB.Status,
		DeletedAt:   ingredientCategoryDB.DeletedAt,
		CreatedAt:   ingredientCategoryDB.CreatedAt,
		UpdatedAt:   ingredientCategoryDB.UpdatedAt,
	}
}

// ToGraphQLIngredient converts IngredientDB to GraphQL Ingredient model
func ToGraphQLIngredient(ingredientDB model.IngredientDB) *model.Ingredient {
	ingredient := &model.Ingredient{
		ID:         ingredientDB.ID,
		Name:       ingredientDB.Name,
		CategoryID: ingredientDB.CategoryID,
		IsActive:   ingredientDB.IsActive,
		DeletedAt:  ingredientDB.DeletedAt,
		CreatedAt:  ingredientDB.CreatedAt,
		UpdatedAt:  ingredientDB.UpdatedAt,
	}

	// Set category if provided
	if ingredientDB.Category != nil && ingredientDB.Category.ID > 0 {
		ingredient.Category = ToGraphQLIngredientCategory(*ingredientDB.Category)
	}

	return ingredient
}
