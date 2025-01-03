package vhost

import (
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
)

func (b *VHost) subscribe(consumerID, queueName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Create a new queue if it doesn't exist
	if _, ok := b.Queues[queueName]; !ok {
		b.createQueue(queueName)
	}
	if consumer, ok := b.Consumers[consumerID]; ok {
		consumer.Queue = queueName
	}
}

// acknowledge removes the message with the given ID frrom the unackedMessages map.
func (b *VHost) acknowledge(consumerID, msgID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// if b.UnackMsgs[queueName] == nil {
	// 	return fmt.Errorf("no unack messages for queue %s", queueName)
	// }
	// if _, ok := b.UnackMsgs[queueName][msgID]; !ok {
	// 	return fmt.Errorf("Message ID %s not found in unack messages", msgID)
	// }
	// delete(b.UnackMsgs[queueName], msgID)
	delete(b.ConsumerUnackMsgs[consumerID], msgID)
	// b.saveBrokerState()
	// if b.UnackMsgs[queueName] == nil {
	// 	return fmt.Errorf("no unack messages for queue %s", queueName)
	// }
	// if _, ok := b.UnackMsgs[queueName][msgID]; !ok {
	// 	return fmt.Errorf("Message ID %s not found in unack messages", msgID)
	// }
	// delete(b.UnackMsgs[queueName], msgID)
	return nil
}

func (b *VHost) createQueue(name string) (*Queue, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Check if the queue already exists
	if _, ok := b.Queues[name]; ok {
		return nil, fmt.Errorf("queue %s already exists", name)
	}

	queue := NewQueue(name)
	b.Queues[name] = queue
	// b.saveBrokerState()
	return queue, nil
}

func (b *VHost) publish(exchangeName, routingKey, message string) (string, error) {
	b.mu.Lock()
	exchange, ok := b.Exchanges[exchangeName]
	b.mu.Unlock()
	if !ok {
		log.Printf("Exchange %s not found", exchangeName)
		return "", fmt.Errorf("Exchange %s not found", exchangeName)
	}
	msgID := uuid.New().String()
	msg := Message{
		ID:      msgID,
		Content: message,
	}

	// // Save message to file
	// err := b.saveMessage(routingKey, msg)
	// if err != nil {
	// 	log.Printf("Failed to save message to file: %v", err)
	// 	return "", err
	// }

	switch exchange.Typ {
	case DIRECT:
		queues, ok := exchange.Bindings[routingKey]
		if ok {
			for _, queue := range queues {
				queue.Push(msg)
			}
			return msgID, nil
		} else {
			log.Printf("Routing key %s not found for exchange %s", routingKey, exchangeName)
			return "", fmt.Errorf("Routing key %s not found for exchange %s", routingKey, exchangeName)
		}
	case FANOUT:
		for _, queue := range exchange.Queues {
			queue.Push(msg)
		}
		return msgID, nil
	}
	return "", fmt.Errorf("Unknown exchange type")
}

// func (b *Broker) consume(queueName string) <-chan Message {
func (b *VHost) consume(queueName, consumerID string) *Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	queue, ok := b.Queues[queueName]
	if !ok {
		log.Printf("Queue %s not found", queueName)
		return nil
	}
	// select {
	// case msg := <-queue.messages:
	// 	if b.UnackMsgs[queueName] == nil {
	// 		b.UnackMsgs[queueName] = make(map[string]bool)
	// 	}
	// 	b.UnackMsgs[queueName][msg.ID] = true
	// 	b.ConsumerUnackMsgs[consumerID][msg.ID] = true
	// 	return &msg
	// default:
	// 	log.Printf("No messages in queue %s", queueName)
	// 	return nil
	// }

	msg := queue.Pop()
	if msg == nil {
		log.Printf("No messages in queue %s", queueName)
		return nil
	}

	// if b.UnackMsgs[queueName] == nil {
	// 	b.UnackMsgs[queueName] = make(map[string]bool)
	// }
	b.ConsumerUnackMsgs[consumerID][msg.ID] = true

	return msg
}

func (b *VHost) deleteQueue(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Check if the queue exists
	_, ok := b.Queues[name]
	if !ok {
		return fmt.Errorf("queue %s not found", name)
	}

	delete(b.Queues, name)
	// b.saveBrokerState()
	return nil
}

func (b *VHost) listQueues() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	queueNames := make([]string, 0, len(b.Queues))
	for name := range b.Queues {
		queueNames = append(queueNames, name)
	}
	return queueNames
}

func (vh *VHost) CreateExchange(name string, typ ExchangeType) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	// Check if the exchange already exists
	if _, ok := vh.Exchanges[name]; ok {
		return fmt.Errorf("exchange %s already exists", name)
	}

	exchange := &Exchange{
		Name:     name,
		Typ:      typ,
		Queues:   make(map[string]*Queue),
		Bindings: make(map[string][]*Queue),
	}
	vh.Exchanges[name] = exchange
	return nil
}

func (b *VHost) bindQueue(exchangeName, queueName, routingKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if exchangeName == "" {
		exchangeName = default_exchange
	}

	// Find the exchange
	exchange, ok := b.Exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("Exchange %s not found", exchangeName)
	}

	// Find the queue
	queue, ok := b.Queues[queueName]
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
	// err := b.saveBrokerState()
	// if err != nil {
	// 	log.Printf("Failed to save broker state: %v", err)
	// }
	return nil
}

// bindToDefaultExchange binds a queue to the default exchange using the queue name as the routing key.
func (b *VHost) bindToDefaultExchange(queueName string) error {
	return b.bindQueue(default_exchange, queueName, queueName)
}

func (b *VHost) DeletBinding(exchangeName, queueName, routingKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find the exchange
	exchange, ok := b.Exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}

	queues, ok := exchange.Bindings[routingKey]
	if !ok {
		return fmt.Errorf("Binding with routing key %s not found", routingKey)
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
	// err := b.saveBrokerState()
	// if err != nil {
	// 	log.Printf("Failed to save broker state: %v", err)
	// }

	return nil
}

// func (vh *VHost) RegisterSessionAndConsummer(sessionID, consumerID string) string {
// 	// sessionID := generateSessionID()
// 	// consumerID := conn.RemoteAddr().String()

// 	vh.registerConsumer(consumerID, "default", sessionID)
// 	log.Println("New connection registered")
// 	return consumerID
// }

// func (b *VHost) registerConsumer(consumerID, queue, sessionID string) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	consumer := &Consumer{
// 		ID:        consumerID,
// 		Queue:     queue,
// 		SessionID: sessionID,
// 	}
// 	b.Consumers[consumerID] = consumer
// 	b.ConsumerSessions[sessionID] = consumerID
// 	b.ConsumerUnackMsgs[consumerID] = make(map[string]bool)
// }

func (vh *VHost) getSessionID(conn net.Conn) (string, bool) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	for sessionID, consumerID := range vh.ConsumerSessions {
		if conn.RemoteAddr().String() == consumerID {
			return sessionID, true
		}
	}
	return "", false
}
