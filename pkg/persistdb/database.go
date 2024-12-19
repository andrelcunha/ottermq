package persistdb

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const dbPath = "./data/ottermq.db"

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	createTables()
	db.Close()
}

func createTables() {
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role_id INTEGER NOT NULL,
		FOREIGN KEY(role_id) REFERENCES roles(id)
	);`
	_, err := db.Exec(createUserTable)
	if err != nil {
		log.Fatalf("Failed to create 'users' table: %v\n", err)
	}

	createRolesTable := `
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT
	);`
	_, err = db.Exec(createRolesTable)
	if err != nil {
		log.Fatalf("Faled to create 'roles' table: %v\n", err)
	}

	createPermissionsTable := `
	CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action TEXT UNIQUE NOT NULL,
		resource TEXT NOT NULL
	);`
	_, err = db.Exec(createPermissionsTable)
	if err != nil {
		log.Fatalf("Faled to create 'permissions' table: %v\n", err)
	}

	createRolePermissionsTable := `
	CREATE TABLE IF NOT EXISTS role_permissions (
		role_id INTEGER NOT NULL,
		permission_id INTEGER NOT NULL,
		FOREIGN KEY(role_id) REFERENCES roles(id),
		FOREIGN KEY(permission_id) REFERENCES permissions(id),
		PRIMARY KEY(role_id, permission_id)
	);`
	_, err = db.Exec(createRolePermissionsTable)
	if err != nil {
		log.Fatalf("Faled to create 'role_permissions' table: %v\n", err)
	}
}

func CloseDB() {
	db.Close()
}

func OpenDB() error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Println("Error opening database: ", err.Error())
	}
	return err
}
