package broker

import (
	. "github.com/andrelcunha/ottermq/pkg/common"
)

func (b *Broker) listConnections() []ConnectionInfo {
	b.mu.Lock()
	defer b.mu.Unlock()

	// get vhost consumers

	// connections := make([]string, 0, len(b.Consumers))
	// connections := make([]ConnectionInfo, 0, len(b.Consumers))
	// connections := make([]ConnectionInfo, 0, 0)
	// for conn := range b.Connections {
	// 	connections = append(connections, ConnectionInfo{
	// 		User:          conn.User,
	// 		// Name:          conn.RemoteAddr().String(),
	// 		LastHeartbeat: b.LastHeartbeat[conn],
	// 		ConnectedAt:   b.ConnectedAt[conn],
	// 	})
	// }
	connections := make([]ConnectionInfo, 0, len(b.Connections))
	for _, c := range b.Connections {
		connections = append(connections, *c)
	}
	return connections
}
