package models

// Used by: POST /api/admin/login
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Used by: POST /api/admin/users
// Keep field names aligned with your current swagger (role, confirm_password)
type UserCreateRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Role            int    `json:"role"`
}

// (Optional) if you plan to support password change
type UserPasswordUpdateRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// (Optional) if you plan to support role change
type UserRoleUpdateRequest struct {
	Role int `json:"role"`
}
