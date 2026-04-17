package graph

import (
	"re-kasirpinter-go/graph/model"
)

func toGraphQLUser(userDB model.UserDB, userRoleDB *model.UserRoleDB) *model.User {
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
		user.Role = toGraphQLUserRole(*userRoleDB)
	}

	return user
}

func toGraphQLUserRole(userRoleDB model.UserRoleDB) *model.UserRole {
	// Convert permissions
	permissions := make([]*model.UserPermission, len(userRoleDB.Permissions))
	for i, permDB := range userRoleDB.Permissions {
		permissions[i] = toGraphQLUserPermission(permDB)
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

func toGraphQLUserPermission(userPermissionDB model.UserPermissionDB) *model.UserPermission {
	return &model.UserPermission{
		ID:   userPermissionDB.ID,
		Name: userPermissionDB.Name,
	}
}

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
