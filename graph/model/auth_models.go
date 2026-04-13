package model

// AuthResponse is the GraphQL model for authentication responses.
// We define it in a non-generated file so it persists across gqlgen runs.
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
