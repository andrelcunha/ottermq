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
