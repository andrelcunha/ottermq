package broker

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
	"github.com/andrelcunha/ottermq/internal/core/models"
)

type ManagerApi interface {
	ListExchanges() []models.ExchangeDTO
	GetTotalExchanges() int
	ListQueues() []models.QueueDTO
	GetTotalQueues() int
	ListConnections() []models.ConnectionInfoDTO
	ListBindings(vhostName, exchangeName string) map[string][]string
}

func NewDefaultManagerApi(broker *Broker) *DefaultManagerApi {
	return &DefaultManagerApi{broker: broker}
}

type DefaultManagerApi struct {
	broker *Broker
}

func (a DefaultManagerApi) ListExchanges() []models.ExchangeDTO {
	b := a.broker
	exchanges := make([]models.ExchangeDTO, 0, a.GetTotalExchanges())
	b.mu.Lock()
	defer b.mu.Unlock()
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, exchange := range b.VHosts[vhost.Name].Exchanges {
			exchanges = append(exchanges, models.ExchangeDTO{
				VHostName: vhost.Name,
				VHostId:   vhost.Id,
				Name:      exchange.Name,
				Type:      string(exchange.Typ),
			})
		}
	}
	return exchanges
}

func (a DefaultManagerApi) GetTotalExchanges() int {
	b := a.broker
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

func (a DefaultManagerApi) ListQueues() []models.QueueDTO {
	b := a.broker
	queues := make([]models.QueueDTO, 0, a.GetTotalQueues())
	b.mu.Lock()
	defer b.mu.Unlock()
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, queue := range b.VHosts[vhost.Name].Queues {
			queues = append(queues, models.QueueDTO{
				VHostName: vhost.Name,
				VHostId:   vhost.Id,
				Name:      queue.Name,
				Messages:  queue.Len(),
			})
		}
	}
	return queues
}

func (a DefaultManagerApi) GetTotalQueues() int {
	b := a.broker
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

func (a DefaultManagerApi) ListConnections() []models.ConnectionInfoDTO {
	b := a.broker
	b.mu.Lock()
	defer b.mu.Unlock()
	connections := make([]amqp.ConnectionInfo, 0, len(b.Connections))
	for _, c := range b.Connections {
		connections = append(connections, *c)
	}
	connectionsDTO := models.MapListConnectionsDTO(connections)
	return connectionsDTO
}

func (a DefaultManagerApi) ListBindings(vhostName, exchangeName string) map[string][]string {
	b := a.broker
	vh := b.GetVHost(vhostName)
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
