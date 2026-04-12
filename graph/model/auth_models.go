package model

// AuthResponse is the GraphQL model for authentication responses.
// We define it in a non-generated file so it persists across gqlgen runs.
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// LoginInput is the input for login mutation
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
