package persistdb

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	RoleID   int    `json:"role_id"`
}

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permission struct {
	ID       int    `json:"id"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

type RolePermission struct {
	RoleID       int `json:"role_id"`
	PermissionID int `json:"permission_id"`
}
