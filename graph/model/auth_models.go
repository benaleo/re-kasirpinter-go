package model

// AuthResponse is the GraphQL model for authentication responses.
// We define it in a non-generated file so it persists across gqlgen runs.
type AuthResponse struct {
	Code    int32     `json:"code"`
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *AuthData `json:"data"`
}
