package broker

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/vhost"
)

func ListBindings(b *Broker, vhostName, exchangeName string) map[string][]string {
	vh := b.GetVHostFromName(vhostName)
	b.mu.Lock()
	defer b.mu.Unlock()
	if vh == nil {
		return nil
	}
	fmt.Printf("exchangeName: %s, Vhost: %s", exchangeName, vh.Name)
	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		return nil
	}

	switch exchange.Typ {
	case vhost.DIRECT:
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
