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

func AddUser(user UserCreateDTO) error {
	OpenDB()
	defer CloseDB()
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

func GetUsers() ([]User, error) {
	OpenDB()
	defer CloseDB()
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

func GetUserByUsername(username string) (User, error) {
	OpenDB()
	defer CloseDB()
	var user User
	err := db.QueryRow("SELECT id, username, role_id FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.RoleID)
	if err != nil {
		log.Printf("Failed to query user: %v\n", err)
		return User{}, err
	}
	return user, nil
}

func GenerateJWTToken(user UserListDTO) (string, error) {
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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (u User) ToUserListDTO() (UserListDTO, error) {
	role, err := GetRoleByID(u.RoleID)
	if err != nil {
		return UserListDTO{}, err
	}
	return UserListDTO{
		ID:          u.ID,
		Username:    u.Username,
		HasPassword: true,
		Role:        role.Name,
	}, nil
}
