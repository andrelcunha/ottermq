package persistdb

import "log"

var defaultPermissions = []Permission{
	{Action: "create", Resource: "resource"},
	{Action: "read", Resource: "resource"},
	{Action: "update", Resource: "resource"},
	{Action: "delete", Resource: "resource"},
}

func AddDefaultPermissions() {
	OpenDB()
	defer CloseDB()
	// Add permissions to the database
	for _, permission := range defaultPermissions {
		_, err := db.Exec("INSERT INTO permissions (action, resource) VALUES (?, ?)", permission.Action, permission.Resource)
		if err != nil {
			log.Printf("Failed to insert permission: %v\n", err)
		}
	}
}
