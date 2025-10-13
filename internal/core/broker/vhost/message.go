package vhost

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/pkg/persistence"
)

type MessageController interface {
	Publish(exchange, routingKey string, body []byte, props *amqp.BasicProperties) (string, error)
	GetMessage(queueName string) *amqp.Message
	GetMessageCount(queueName string) (int, error)
	Acknowledge(consumerID, msgID string) error
}

type DefaultMessageController struct {
	vhost *VHost
}

func (m *DefaultMessageController) Publish(exchangeName, routingKey string, body []byte, props *amqp.BasicProperties) (string, error) {
	return m.vhost.publish(exchangeName, routingKey, body, props)
}

func (m *DefaultMessageController) GetMessage(queueName string) *amqp.Message {
	return m.vhost.getMessage(queueName)
}

func (m *DefaultMessageController) GetMessageCount(queueName string) (int, error) {
	return m.vhost.getMessageCount(queueName)
}

func (m *DefaultMessageController) Acknowledge(consumerID, msgID string) error {
	return m.vhost.acknowledge(consumerID, msgID)
}

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
	log.Debug().Str("msg_id", msgID).Str("consumer_id", consumerID).Msg("Acknowledged message")

	// vh.saveBrokerState() // TODO: Persist
	return nil
}

func (vh *VHost) getMessageCount(queueName string) (int, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, ok := vh.Queues[queueName]
	if !ok {
		return 0, fmt.Errorf("queue %s not found", queueName)
	}
	return queue.Len(), nil
}

func (vh *VHost) publish(exchangeName, routingKey string, body []byte, props *amqp.BasicProperties) (string, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		log.Error().Str("exchange", exchangeName).Msg("Exchange not found")
		return "", fmt.Errorf("Exchange %s not found", exchangeName)
	}

	// verify if exchange is internal
	if exchange.Props.Internal {
		// TODO: send the proper error code and channel exception
		return "", fmt.Errorf("cannot publish to internal exchange %s", exchangeName)
	}

	msgID := uuid.New().String()
	msg := amqp.Message{
		ID:         msgID,
		Body:       body,
		Properties: *props,
		Exchange:   exchangeName,
		RoutingKey: routingKey,
	}
	log.Debug().Str("id", msgID).Str("exchange", exchangeName).Str("routing_key", routingKey).Str("body", string(body)).Interface("properties", props).Msg("Publishing message")

	var timestamp int64
	if props.Timestamp.IsZero() {
		timestamp = 0
	} else {
		timestamp = props.Timestamp.Unix()
	}
	msgProps := persistence.MessageProperties{
		ContentType:     string(props.ContentType),
		ContentEncoding: props.ContentEncoding,
		Headers:         props.Headers,
		DeliveryMode:    uint8(props.DeliveryMode),
		Priority:        props.Priority,
		CorrelationID:   props.CorrelationID,
		ReplyTo:         props.ReplyTo,
		Expiration:      props.Expiration,
		MessageID:       props.MessageID,
		Timestamp:       timestamp,
		Type:            props.Type,
		UserID:          props.UserID,
		AppID:           props.AppID,
	}
	switch exchange.Typ {
	case DIRECT:
		queues, ok := exchange.Bindings[routingKey]
		if !ok {
			log.Error().Str("routing_key", routingKey).Str("exchange", exchangeName).Msg("Routing key not found for exchange")
			return "", fmt.Errorf("routing key %s not found for exchange %s", routingKey, exchangeName)
		}
		for _, queue := range queues {
			err := vh.saveMessageIfDurable(SaveMessageRequest{
				Props:      props,
				Queue:      queue,
				RoutingKey: routingKey,
				MsgID:      msgID,
				Body:       body,
				MsgProps:   msgProps,
			})
			if err != nil {
				return "", err
			}

			queue.Push(msg)
		}
		return msgID, nil

	case FANOUT:
		for _, queue := range exchange.Queues {
			err := vh.saveMessageIfDurable(SaveMessageRequest{
				Props:      props,
				Queue:      queue,
				RoutingKey: routingKey,
				MsgID:      msgID,
				Body:       body,
				MsgProps:   msgProps,
			})
			if err != nil {
				return "", err
			}
			queue.Push(msg)
		}
		return msgID, nil
	case TOPIC:
		// TODO: Implement topic exchange routing
		return "", fmt.Errorf("topic exchange not yet implemented")
	default:
		return "", fmt.Errorf("unknown exchange type")
	}
}

type SaveMessageRequest struct {
	Props      *amqp.BasicProperties
	Queue      *Queue
	RoutingKey string
	MsgID      string
	Body       []byte
	MsgProps   persistence.MessageProperties
}

func (vh *VHost) saveMessageIfDurable(req SaveMessageRequest) error {
	if req.Props.DeliveryMode == amqp.PERSISTENT { // Persistent
		if req.Queue.Props.Durable {
			// Persist the message
			if vh.persist == nil {
				log.Error().Msg("Persistence layer is not initialized")
				return fmt.Errorf("persistence layer is not initialized")
			}
			if err := vh.persist.SaveMessage(vh.Name, req.RoutingKey, req.MsgID, req.Body, req.MsgProps); err != nil {
				log.Error().Err(err).Msg("Failed to save message to file")
				return err
			}
		}
	}
	return nil
}

// func (vh *Broker) GetMessage(queueName string) <-chan Message {
func (vh *VHost) getMessage(queueName string) *amqp.Message {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, ok := vh.Queues[queueName]
	if !ok {
		log.Error().Str("queue", queueName).Msg("Queue not found")
		return nil
	}
	msg := queue.Pop()
	if msg == nil {
		log.Debug().Str("queue", queueName).Msg("No messages in queue")
		return nil
	}
	return msg
}

// func (vh *VHost) registerSessionAndConsummer(sessionID, consumerID string) string {
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
// 	vh.ConsumerUnackMsgs[consumerID] = make(map[string]amqp.Message)
// }
