package input

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	RoleID   *int64 `json:"role_id,omitempty"`
}
