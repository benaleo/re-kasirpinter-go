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
