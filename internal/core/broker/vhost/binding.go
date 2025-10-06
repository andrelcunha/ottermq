package vhost

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// bindToDefaultExchange binds a queue to the default exchange using the queue name as the routing key.
func (vh *VHost) BindToDefaultExchange(queueName string) error {
	return vh.BindQueue(DEFAULT_EXCHANGE, queueName, queueName)
}

func (vh *VHost) BindQueue(exchangeName, queueName, routingKey string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	// Find the exchange
	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}
	queue, ok := vh.Queues[queueName]
	if !ok {
		return fmt.Errorf("queue %s not found", queueName)
	}

	switch exchange.Typ {
	case DIRECT:
		for _, q := range exchange.Bindings[routingKey] {
			if q.Name == queueName {
				log.Debug().Str("queue", queueName).Str("exchange", exchangeName).Str("routing_key", routingKey).Msg("Queue already bound to exchange")
				return nil
			}
		}
		exchange.Bindings[routingKey] = append(exchange.Bindings[routingKey], queue)
	case FANOUT:
		exchange.Queues[queueName] = queue
	}

	// // Persist the state
	// err := vh.saveBrokerState()
	// if err != nil {
	// 	log.Printf("Failed to save broker state: %v", err)
	// }
	// vh.publishBindingUpdate(exchangeName)
	return nil
}

func (vh *VHost) DeletBinding(exchangeName, queueName, routingKey string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	// Find the exchange
	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}

	queues, ok := exchange.Bindings[routingKey]
	if !ok {
		return fmt.Errorf("binding with routing key %s not found", routingKey)
	}

	// Find the queue
	var index int
	found := false
	for i, q := range queues {
		if q.Name == queueName {
			index = i
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Queue %s not found", queueName)
	}

	// Remove the queue from the bindings
	exchange.Bindings[routingKey] = append(queues[:index], queues[index+1:]...)

	if len(exchange.Bindings[routingKey]) == 0 {
		delete(exchange.Bindings, routingKey)
		// Check if the exchange can be auto-deleted
		if deleted, err := vh.checkAutoDeleteExchangeUnlocked(exchange.Name); err != nil {
			log.Printf("Failed to check auto-delete exchange: %v", err)
			return err
		} else if deleted {
			log.Printf("Exchange %s was auto-deleted", exchange.Name)
		}
	}
	return nil
}
