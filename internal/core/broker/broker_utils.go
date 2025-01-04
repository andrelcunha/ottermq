package broker

import (
	// "github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/andrelcunha/ottermq/internal/core/vhost"
)

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

func (b *Broker) GetTotalQueues() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := 0
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, queue := range b.VHosts[vhost.Name].Queues {
			if queue.Name != "" {
				total++
			}
		}
	}
	return total
}

func (b *Broker) GetTotalExchanges() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := 0
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, exchange := range b.VHosts[vhost.Name].Exchanges {
			if exchange.Name != "" {
				total++
			}
		}
	}
	return total
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
