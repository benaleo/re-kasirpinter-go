package helper

import (
	"re-kasirpinter-go/graph/model"
	"strings"
)

// BadRequestResponse creates a standard bad request response
func BadRequestResponse(message string) *model.CreateUserResponse {
	return &model.CreateUserResponse{
		Code:    400,
		Success: false,
		Message: message,
	}
}

// InternalServerErrorResponse creates a standard internal server error response
func InternalServerErrorResponse(message string) *model.CreateUserResponse {
	return &model.CreateUserResponse{
		Code:    500,
		Success: false,
		Message: message,
	}
}

// SuccessResponse creates a standard success response
func SuccessResponse(code int32, message string, data *model.User) *model.CreateUserResponse {
	return &model.CreateUserResponse{
		Code:    code,
		Success: true,
		Message: message,
		Data:    data,
	}
}

// IsDuplicateCodeError checks if the error is a duplicate code constraint violation
func IsDuplicateCodeError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "idx_discounts_code")
}
