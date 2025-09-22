package vhost

import (
	"fmt"
	"log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/google/uuid"
)

// acknowledge removes the message with the given ID frrom the unackedMessages map.
func (vh *VHost) acknowledge(consumerID, msgID string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	if vh.ConsumerUnackMsgs[consumerID] == nil {
		return fmt.Errorf("no unack messages for consumer %s", consumerID)
	}
	if _, ok := vh.ConsumerUnackMsgs[consumerID][msgID]; !ok {
		return fmt.Errorf("message ID %s not found in unack messages for consumer %s", msgID, consumerID)
	}
	delete(vh.ConsumerUnackMsgs[consumerID], msgID)
	log.Printf("[DEBUG] Acknoledged messege %s for consumer %s", msgID, consumerID)

	// vh.saveBrokerState() // TODO: Persist
	return nil
}

func (vh *VHost) GetMessageCount(name string) (int, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, ok := vh.Queues[name]
	if !ok {
		return 0, fmt.Errorf("queue %s not found", name)
	}
	return queue.Len(), nil
}

func (vh *VHost) Publish(exchangeName, routingKey string, body []byte, props *amqp.BasicProperties) (string, error) {
	vh.mu.Lock()
	exchange, ok := vh.Exchanges[exchangeName]
	vh.mu.Unlock()
	if !ok {
		log.Printf("Exchange %s not found", exchangeName)
		return "", fmt.Errorf("Exchange %s not found", exchangeName)
	}
	msgID := uuid.New().String()
	msg := amqp.Message{
		ID:         msgID,
		Body:       body,
		Properties: *props,
		Exchange:   exchangeName,
		RoutingKey: routingKey,
	}
	log.Printf("[DEBUG] Publishing message: ID=%s, Exchange=%s, RoutingKey=%s, Body=%s, Properties=%+v",
		msgID, msg.Exchange, msg.RoutingKey, string(msg.Body), msg.Properties)

	// // Save message to file TODO: Persist
	// err := vh.saveMessage(routingKey, msg)
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
		}
		log.Printf("[ERROR] Routing key %s not found for exchange %s", routingKey, exchangeName)
		return "", fmt.Errorf("routing key %s not found for exchange %s", routingKey, exchangeName)

	case FANOUT:
		for _, queue := range exchange.Queues {
			queue.Push(msg)
		}
		return msgID, nil
	}
	return "", fmt.Errorf("unknown exchange type")
}

// func (vh *Broker) GetMessage(queueName string) <-chan Message {
func (vh *VHost) GetMessage(queueName string) *amqp.Message {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, ok := vh.Queues[queueName]
	if !ok {
		log.Printf("[ERROR] Queue %s not found", queueName)
		return nil
	}
	msg := queue.Pop()
	if msg == nil {
		log.Printf("[DEBUG] No messages in queue %s", queueName)
		return nil
	}
	return msg
}

// func (vh *VHost) RegisterSessionAndConsummer(sessionID, consumerID string) string {
// 	// sessionID := generateSessionID()
// 	// consumerID := conn.RemoteAddr().String()

// 	vh.registerConsumer(consumerID, "default", sessionID)
// 	log.Println("New connection registered")
// 	return consumerID
// }

// func (vh *VHost) registerConsumer(consumerID, queue, sessionID string) {
// 	vh.mu.Lock()
// 	defer vh.mu.Unlock()
// 	consumer := &Consumer{
// 		ID:        consumerID,
// 		Queue:     queue,
// 		SessionID: sessionID,
// 	}
// 	vh.Consumers[consumerID] = consumer
// 	vh.ConsumerSessions[sessionID] = consumerID
// 	vh.ConsumerUnackMsgs[consumerID] = make(map[string]bool)
// }
