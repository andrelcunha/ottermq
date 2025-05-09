package vhost

import "fmt"

// bindToDefaultExchange binds a queue to the default exchange using the queue name as the routing key.
func (vh *VHost) BindToDefaultExchange(queueName string) error {
	return vh.BindQueue(default_exchange, queueName, queueName)
}

func (vh *VHost) BindQueue(exchangeName, queueName, routingKey string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	// if exchangeName == "" {
	// 	exchangeName = default_exchange
	// }

	// Find the exchange
	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("Exchange %s not found", exchangeName)
	}

	// Find the queue
	queue, ok := vh.Queues[queueName]
	if !ok {
		return fmt.Errorf("Queue %s not found", queueName)
	}

	switch exchange.Typ {
	case DIRECT:
		for _, q := range exchange.Bindings[routingKey] {
			if q.Name == queueName {
				return fmt.Errorf("Queue %s already binded to exchange %s using routing key %s", queueName, exchangeName, routingKey)
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
	}

	// // Persist the state
	// err := vh.saveBrokerState()
	// if err != nil {
	// 	log.Printf("Failed to save broker state: %v", err)
	// }

	return nil
}
