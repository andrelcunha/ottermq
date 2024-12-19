package persistdb

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func AddUser(user User) error {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v\n", err)
		return err
	}
	_, err = db.Exec("INSERT INTO users (username, password, role_id) VALUES (?, ?, ?)", user.Username, hashedPassword, user.RoleID)
	if err != nil {
		log.Printf("Failed to insert user: %v\n", err)
		return err
	}
	return nil
}

func AuthenticateUser(username, password string) (bool, error) {
	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return comparePasswords(password, storedPassword)
}

func comparePasswords(password, storedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	return err == nil, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
