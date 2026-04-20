package graph

import (
	"re-kasirpinter-go/graph/model"
)

func toGraphQLCreateOtpResponse(code int32, success bool, message string) *model.CreateOtpResponse {
	return &model.CreateOtpResponse{
		Code:    code,
		Success: success,
		Message: message,
	}
}

func toGraphQLVerifyOtpResponse(code int32, success bool, message string, token *string) *model.VerifyOtpResponse {
	return &model.VerifyOtpResponse{
		Code:    code,
		Success: success,
		Message: message,
		Token:   token,
	}
}

func toGraphQLNewPasswordResponse(code int32, success bool, message string) *model.NewPasswordResponse {
	return &model.NewPasswordResponse{
		Code:    code,
		Success: success,
		Message: message,
	}
}

func toGraphQLCreateIngredientCategoryResponse(code int32, success bool, message string, data *model.IngredientCategory) *model.CreateIngredientCategoryResponse {
	return &model.CreateIngredientCategoryResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLUpdateIngredientCategoryResponse(code int32, success bool, message string, data *model.IngredientCategory) *model.UpdateIngredientCategoryResponse {
	return &model.UpdateIngredientCategoryResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLDeleteIngredientCategoryResponse(code int32, success bool, message string, data *model.IngredientCategory) *model.DeleteIngredientCategoryResponse {
	return &model.DeleteIngredientCategoryResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLCreateIngredientResponse(code int32, success bool, message string, data *model.Ingredient) *model.CreateIngredientResponse {
	return &model.CreateIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLUpdateIngredientResponse(code int32, success bool, message string, data *model.Ingredient) *model.UpdateIngredientResponse {
	return &model.UpdateIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLDeleteIngredientResponse(code int32, success bool, message string, data *model.Ingredient) *model.DeleteIngredientResponse {
	return &model.DeleteIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLIngredientCategoriesResponse(code int32, success bool, message string, data []*model.IngredientCategory, pagination *model.PageInfo) *model.IngredientCategoriesResponse {
	return &model.IngredientCategoriesResponse{
		Code:       code,
		Success:    success,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

func toGraphQLIngredientsResponse(code int32, success bool, message string, data []*model.Ingredient, pagination *model.PageInfo) *model.IngredientsResponse {
	return &model.IngredientsResponse{
		Code:       code,
		Success:    success,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}
