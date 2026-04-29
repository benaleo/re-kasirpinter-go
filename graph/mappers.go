package graph

import (
	"re-kasirpinter-go/graph/model"
)

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

func toGraphQLCreateIngredientStockResponse(code int32, success bool, message string, data *model.IngredientStock) *model.CreateIngredientStockResponse {
	return &model.CreateIngredientStockResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLUpdateIngredientStockResponse(code int32, success bool, message string, data *model.IngredientStock) *model.UpdateIngredientStockResponse {
	return &model.UpdateIngredientStockResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLDeleteIngredientStockResponse(code int32, success bool, message string, data *model.IngredientStock) *model.DeleteIngredientStockResponse {
	return &model.DeleteIngredientStockResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLIngredientStocksResponse(code int32, success bool, message string, data []*model.IngredientStock, pagination *model.PageInfo) *model.IngredientStocksResponse {
	return &model.IngredientStocksResponse{
		Code:       code,
		Success:    success,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

// Response mappers for ProductVariant
func toGraphQLCreateProductVariantResponse(code int32, success bool, message string, data *model.ProductVariant) *model.CreateProductVariantResponse {
	return &model.CreateProductVariantResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLUpdateProductVariantResponse(code int32, success bool, message string, data *model.ProductVariant) *model.UpdateProductVariantResponse {
	return &model.UpdateProductVariantResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLDeleteProductVariantResponse(code int32, success bool, message string, data *model.ProductVariant) *model.DeleteProductVariantResponse {
	return &model.DeleteProductVariantResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLProductVariantsResponse(code int32, success bool, message string, data []*model.ProductVariant, pagination *model.PageInfo) *model.ProductVariantsResponse {
	return &model.ProductVariantsResponse{
		Code:       code,
		Success:    success,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

func toGraphQLProductVariantResponse(code int32, success bool, message string, data *model.ProductVariant) *model.ProductVariantResponse {
	return &model.ProductVariantResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

// Response mappers for ProductIngredient
func toGraphQLCreateProductIngredientResponse(code int32, success bool, message string, data *model.ProductIngredient) *model.CreateProductIngredientResponse {
	return &model.CreateProductIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLUpdateProductIngredientResponse(code int32, success bool, message string, data *model.ProductIngredient) *model.UpdateProductIngredientResponse {
	return &model.UpdateProductIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLDeleteProductIngredientResponse(code int32, success bool, message string, data *model.ProductIngredient) *model.DeleteProductIngredientResponse {
	return &model.DeleteProductIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}

func toGraphQLProductIngredientsResponse(code int32, success bool, message string, data []*model.ProductIngredient, pagination *model.PageInfo) *model.ProductIngredientsResponse {
	return &model.ProductIngredientsResponse{
		Code:       code,
		Success:    success,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

func toGraphQLProductIngredientResponse(code int32, success bool, message string, data *model.ProductIngredient) *model.ProductIngredientResponse {
	return &model.ProductIngredientResponse{
		Code:    code,
		Success: success,
		Message: message,
		Data:    data,
	}
}
