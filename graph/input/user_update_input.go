package input

type UpdateUserInput struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Address  *string `json:"address,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	RoleID   *int64  `json:"role_id,omitempty"`
}
