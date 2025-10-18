package vhost

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/pkg/persistence"
)

type SaveMessageRequest struct {
	Props      *amqp.BasicProperties
	Queue      *Queue
	RoutingKey string
	MsgID      string
	Body       []byte
	MsgProps   persistence.MessageProperties
}

// TODO: create a higher level abstraction of amqp.Message, exposing the content, requeued count, etc

func (vh *VHost) Publish(exchangeName, routingKey string, body []byte, props *amqp.BasicProperties) (string, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		log.Error().Str("exchange", exchangeName).Msg("Exchange not found")
		return "", fmt.Errorf("Exchange %s not found", exchangeName)
	}

	// verify if exchange is internal
	if exchange.Props.Internal {
		// TODO: raise channel exception (403)
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
	log.Trace().Str("id", msgID).Str("exchange", exchangeName).Str("routing_key", routingKey).Str("body", string(body)).Interface("properties", props).Msg("Publishing message")

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

// func (vh *Broker) GetMessage(queueName string) <-chan Message {
func (vh *VHost) GetMessage(queueName string) *amqp.Message {
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

func (vh *VHost) GetMessageCount(queueName string) (int, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, ok := vh.Queues[queueName]
	if !ok {
		return 0, fmt.Errorf("queue %s not found", queueName)
	}
	return queue.Len(), nil
}

// acknowledge removes the message with the given ID from the unackedMessages map.
func (vh *VHost) Acknowledge(consumerID, msgID string) error {
	panic("Not implemented")
}

func (vh *VHost) saveMessageIfDurable(req SaveMessageRequest) error {
	if req.Props.DeliveryMode == amqp.PERSISTENT { // Persistent
		if req.Queue.Props.Durable {
			// Persist the message
			if vh.persist == nil {
				log.Error().Msg("Persistence layer is not initialized")
				return fmt.Errorf("persistence layer is not initialized")
			}
			if err := vh.persist.SaveMessage(vh.Name, req.Queue.Name, req.MsgID, req.Body, req.MsgProps); err != nil {
				log.Error().Err(err).Msg("Failed to save message to file")
				return err
			}
		}
	}
	return nil
}
