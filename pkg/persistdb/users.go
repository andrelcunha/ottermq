package persistdb

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func GetUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, username, role_id FROM users")
	if err != nil {
		log.Printf("Failed to query users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.RoleID)
		if err != nil {
			log.Printf("Failed to scan user: %v\n", err)
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserByUsername(username string) (UserDTO, error) {
	var user User
	err := db.QueryRow("SELECT id, username, role_id FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.RoleID)
	if err != nil {
		log.Printf("Failed to query user: %v\n", err)
		return UserDTO{}, err
	}

	return mapUserToUserDTO(user)
}

func mapUserToUserDTO(user User) (UserDTO, error) {
	role, err := GetRoleByID(user.RoleID)
	if err != nil {
		return UserDTO{}, err
	}
	return UserDTO{
		ID:          user.ID,
		Username:    user.Username,
		HasPassword: true,
		Role:        role.Name,
	}, nil
}

func GenerateJWTToken(user UserDTO) (string, error) {
	// convert user to json
	jsonUser, err := json.Marshal(user)
	if err != nil {
		log.Printf("Failed to marshal user: %v\n", err)
	}
	// convert jsonUser to base64
	encoded := base64.StdEncoding.EncodeToString(jsonUser)

	claims := jwt.MapClaims{
		"iss":         "ottermq",
		"sud":         user.ID,
		"aud":         "ottermq",
		"exp":         jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		"nbf":         jwt.NewNumericDate(time.Now()),
		"iat":         jwt.NewNumericDate(time.Now()),
		"jti":         uuid.New().String(),
		"username":    user.Username,
		"role":        user.Role,
		"encodeduser": encoded,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret"))
}
