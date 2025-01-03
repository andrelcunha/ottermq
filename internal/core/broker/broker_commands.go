package broker

import (
	"github.com/andrelcunha/ottermq/internal/core/vhost"
	. "github.com/andrelcunha/ottermq/pkg/common"
)

func ListConnections(b *Broker) []ConnectionInfoDTO {
	b.mu.Lock()
	defer b.mu.Unlock()
	connections := make([]ConnectionInfo, 0, len(b.Connections))
	for _, c := range b.Connections {
		connections = append(connections, *c)
	}
	connectionsDTO := mapListConnectionsDTO(connections)
	return connectionsDTO
}

func mapListConnectionsDTO(connections []ConnectionInfo) []ConnectionInfoDTO {
	listConnectonsDTO := make([]ConnectionInfoDTO, len(connections))
	for i, connection := range connections {
		state := "disconnected"
		if connection.Done == nil {
			state = "running"
		}
		channels := len(connection.Channels)
		listConnectonsDTO[i] = ConnectionInfoDTO{
			VHostName:     connection.VHostName,
			VHostId:       connection.VHostId,
			Name:          connection.Name,
			Username:      connection.User,
			State:         state,
			SSL:           false,
			Protocol:      "AMQP 0-9-1",
			Channels:      channels,
			LastHeartbeat: connection.LastHeartbeat,
			ConnectedAt:   connection.ConnectedAt,
		}
	}
	return listConnectonsDTO
}

func ListExchanges(b *Broker) []ExchangeDTO {
	b.mu.Lock()
	defer b.mu.Unlock()
	exchanges := make([]ExchangeDTO, 0, len(b.VHosts))
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, exchange := range b.VHosts[vhost.Name].Exchanges {
			exchanges = append(exchanges, ExchangeDTO{
				VHostName: vhost.Name,
				VHostId:   vhost.Id,
				Name:      exchange.Name,
				Type:      string(exchange.Typ),
			})
		}
	}
	return exchanges
}

func ListBindings(b *Broker, vhostId, exchangeName string) map[string][]string {
	b.mu.Lock()
	defer b.mu.Unlock()
	vh := b.GetVHostFromName(vhostId)
	// vh := b.VHosts[vhostName]

	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		return nil
	}

	switch exchange.Typ {
	case vhost.DIRECT:
		// bindings := make([]string, 0, len(exchange.Bindings))
		bindings := make(map[string][]string)
		for routingKey, queues := range exchange.Bindings {
			var queuesStr []string
			for _, queue := range queues {
				queuesStr = append(queuesStr, queue.Name)
			}
			bindings[routingKey] = queuesStr
		}
		return bindings
	case vhost.FANOUT:
		// bindings := make([]string, 0, len(exchange.Queues))
		bindings := make(map[string][]string)
		var queues []string
		for queueName := range exchange.Queues {
			queues = append(queues, queueName)
		}
		bindings["fanout"] = queues
		return bindings
	}
	return nil
}

// func (b *VHost) countMessages(queueName string) (int, error) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
//
// 	queue, ok := b.Queues[queueName]
// 	if !ok {
// 		return 0, fmt.Errorf("Queue %s not found", queueName)
// 	}
//
// 	// messageCount := len(queue.messages)
// 	messageCount := queue.Len()
// 	return messageCount, nil
// }

// func (b *VHost) deleteExchange(name string) error {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	// If the exchange is the default exchange, return an error
// 	if name == "default" {
// 		return fmt.Errorf("cannot delete default exchange")
// 	}
//
// 	// Check if the exchange exists
// 	_, ok := b.Exchanges[name]
// 	if !ok {
// 		return fmt.Errorf("exchange %s not found", name)
// 	}
// 	delete(b.Exchanges, name)
// 	// b.saveBrokerState()
// 	return nil
// }
