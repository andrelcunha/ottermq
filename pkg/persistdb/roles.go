package persistdb

import (
	"log"
)

var defaultRoles = []Role{
	{Name: "admin", Description: "Can configure settings, create/delete resources, manage users, etc."},
	{Name: "user", Description: "Can read and write to resources but cannot manage users or settings."},
	{Name: "guest", Description: "Can only read resources"},
}

func AddDefaultRoles() {
	// Add roles to the database
	for _, role := range defaultRoles {
		_, err := db.Exec("INSERT INTO roles (name, description) VALUES (?, ?)", role.Name, role.Description)
		if err != nil {
			log.Printf("Failed to insert role: %v\n", err)
		}
	}
}

func GetRoleByID(id int) (Role, error) {
	var role Role
	err := db.QueryRow("SELECT id, name, description FROM roles WHERE id = ?", id).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		log.Printf("Failed to query role: %v\n", err)
		return Role{}, err
	}
	return role, nil
}
