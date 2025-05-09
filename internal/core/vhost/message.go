package vhost

import (
	"fmt"
	"log"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp/message"
	"github.com/google/uuid"
)

// acknowledge removes the message with the given ID frrom the unackedMessages map.
func (vh *VHost) acknowledge(consumerID, msgID string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	// if vh.UnackMsgs[queueName] == nil {
	// 	return fmt.Errorf("no unack messages for queue %s", queueName)
	// }
	// if _, ok := vh.UnackMsgs[queueName][msgID]; !ok {
	// 	return fmt.Errorf("Message ID %s not found in unack messages", msgID)
	// }
	// delete(vh.UnackMsgs[queueName], msgID)
	delete(vh.ConsumerUnackMsgs[consumerID], msgID)
	// vh.saveBrokerState()
	// if vh.UnackMsgs[queueName] == nil {
	// 	return fmt.Errorf("no unack messages for queue %s", queueName)
	// }
	// if _, ok := vh.UnackMsgs[queueName][msgID]; !ok {
	// 	return fmt.Errorf("Message ID %s not found in unack messages", msgID)
	// }
	// delete(vh.UnackMsgs[queueName], msgID)
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

func (vh *VHost) Publish(exchangeName, routingKey string, body []byte, props *message.BasicProperties) (string, error) {
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

	// // Save message to file
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
		} else {
			log.Printf("Routing key %s not found for exchange %s", routingKey, exchangeName)
			return "", fmt.Errorf("routing key %s not found for exchange %s", routingKey, exchangeName)
		}
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
		log.Printf("Queue %s not found", queueName)
		return nil
	}
	// select {
	// case msg := <-queue.messages:
	// 	if vh.UnackMsgs[queueName] == nil {
	// 		vh.UnackMsgs[queueName] = make(map[string]bool)
	// 	}
	// 	vh.UnackMsgs[queueName][msg.ID] = true
	// 	vh.ConsumerUnackMsgs[consumerID][msg.ID] = true
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

	// if vh.UnackMsgs[queueName] == nil {
	// 	vh.UnackMsgs[queueName] = make(map[string]bool)
	// }
	// vh.ConsumerUnackMsgs[consumerID][msg.ID] = true

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
