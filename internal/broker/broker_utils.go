package broker

import "github.com/google/uuid"

func generateSessionID() string {
	return uuid.New().String()
}

func (b *Broker) Shutdown() {
	for conn := range b.Connections {
		conn.Close()
	}
}

func (b *Broker) authenticate(username, password string) bool {
	return username == b.config.Username && password == b.config.Password
}
