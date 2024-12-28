package broker

import (
	// "github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/google/uuid"
)

func generateSessionID() string {
	return uuid.New().String()
}

func (b *Broker) Shutdown() {
	for conn := range b.Connections {
		conn.Close()
	}
}

// func (b *Broker) authenticate(username, password string) bool {
// 	err := persistdb.OpenDB()
// 	if err != nil {
// 		return false
// 	}
// 	defer persistdb.CloseDB()
// 	isAuthenticated, err := persistdb.AuthenticateUser(username, password)
// 	if err != nil {
// 		return false
// 	}
// 	return isAuthenticated
// }
