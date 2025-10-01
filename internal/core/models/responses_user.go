package models

// Returned by: GET /api/admin/users
// Mirrors persistdb.UserListDTO (but decoupled from persistdb)
type UserSummary struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	HasPassword bool   `json:"has_password"`
	Role        string `json:"role"`
}

type UserListResponse struct {
	Users []UserSummary `json:"users"`
}

// Returned by: POST /api/admin/login
// Keep it minimal now (token only). You can add "user" later if you want.
type AuthResponse struct {
	Token string `json:"token"`
}

// (Optional) If you ever need to return 1 user (e.g., create/read)
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
