package broker

import (
	// "github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/andrelcunha/ottermq/internal/core/vhost"
)

// func (b *Broker) GetVHostIdFromName(vhostName string) string {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	if vhost, ok := b.VHosts[vhostName]; ok {
// 		return vhost.Id
// 	}
// 	return ""
// }

func (b *Broker) GetVHostFromName(vhostName string) *vhost.VHost {
	b.mu.Lock()
	defer b.mu.Unlock()
	if vhost, ok := b.VHosts[vhostName]; ok {
		return vhost
	}
	return nil
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
