package vhost

import (
	"testing"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

func TestPublishToInternalExchange(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	// Create an internal exchange
	exchangeName := "internal-ex"
	vh.Exchanges[exchangeName] = &Exchange{
		Name:  exchangeName,
		Typ:   DIRECT,
		Props: &ExchangeProperties{Internal: true},
	}

	// Try to publish to the internal exchange
	_, err := vh.Publish(exchangeName, "rk", []byte("test"), &amqp.BasicProperties{})
	if err == nil {
		t.Errorf("Expected error when publishing to internal exchange, got nil")
	}
	if err != nil && err.Error() != "cannot publish to internal exchange internal-ex" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishToNonExistentExchange(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	// Try to publish to non-existent exchange
	_, err := vh.Publish("non-existent", "rk", []byte("test"), &amqp.BasicProperties{})
	if err == nil {
		t.Errorf("Expected error when publishing to non-existent exchange, got nil")
	}
	if err != nil && err.Error() != "Exchange non-existent not found" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishToDirectExchangeWithBinding(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	// Create a queue
	queue := &Queue{
		Name:     "test-queue",
		messages: make(chan amqp.Message, 100),
		Props:    &QueueProperties{Durable: false},
	}
	vh.Queues["test-queue"] = queue

	// Create a direct exchange with binding
	exchangeName := "direct-ex"
	routingKey := "test.key"
	vh.Exchanges[exchangeName] = &Exchange{
		Name:  exchangeName,
		Typ:   DIRECT,
		Props: &ExchangeProperties{Internal: false},
		Bindings: map[string][]*Queue{
			routingKey: {queue},
		},
	}

	// Publish message
	msgID, err := vh.Publish(exchangeName, routingKey, []byte("test message"), &amqp.BasicProperties{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if msgID == "" {
		t.Errorf("Expected message ID, got empty string")
	}

	// Verify message was queued
	if queue.Len() != 1 {
		t.Errorf("Expected 1 message in queue, got %d", queue.Len())
	}
}

func TestPublishToDirectExchangeWithoutBinding(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	// Create a direct exchange without bindings
	exchangeName := "direct-ex"
	vh.Exchanges[exchangeName] = &Exchange{
		Name:     exchangeName,
		Typ:      DIRECT,
		Props:    &ExchangeProperties{Internal: false},
		Bindings: make(map[string][]*Queue),
	}

	// Try to publish with unbound routing key
	_, err := vh.Publish(exchangeName, "unbound.key", []byte("test"), &amqp.BasicProperties{})
	if err == nil {
		t.Errorf("Expected error when publishing to unbound routing key, got nil")
	}
	if err != nil && err.Error() != "routing key unbound.key not found for exchange direct-ex" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishToFanoutExchange(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	// Create queues
	queue1 := &Queue{
		Name:     "queue1",
		messages: make(chan amqp.Message, 100),
		Props:    &QueueProperties{Durable: false},
	}
	queue2 := &Queue{
		Name:     "queue2",
		messages: make(chan amqp.Message, 100),
		Props:    &QueueProperties{Durable: false},
	}

	// Create fanout exchange
	exchangeName := "fanout-ex"
	vh.Exchanges[exchangeName] = &Exchange{
		Name:  exchangeName,
		Typ:   FANOUT,
		Props: &ExchangeProperties{Internal: false},
		Queues: map[string]*Queue{
			"queue1": queue1,
			"queue2": queue2,
		},
	}

	// Publish message
	msgID, err := vh.Publish(exchangeName, "any.key", []byte("fanout message"), &amqp.BasicProperties{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if msgID == "" {
		t.Errorf("Expected message ID, got empty string")
	}

	// Verify message was sent to both queues
	if queue1.Len() != 1 {
		t.Errorf("Expected 1 message in queue1, got %d", queue1.Len())
	}
	if queue2.Len() != 1 {
		t.Errorf("Expected 1 message in queue2, got %d", queue2.Len())
	}
}

func TestPublishToUnsupportedExchangeType(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	// Create exchange with unsupported type
	exchangeName := "topic-ex"
	vh.Exchanges[exchangeName] = &Exchange{
		Name:  exchangeName,
		Typ:   TOPIC, // Not yet implemented
		Props: &ExchangeProperties{Internal: false},
	}

	// Try to publish
	_, err := vh.Publish(exchangeName, "test.key", []byte("test"), &amqp.BasicProperties{})
	if err == nil {
		t.Errorf("Expected error for unsupported exchange type, got nil")
	}
	if err != nil && err.Error() != "topic exchange not yet implemented" {
		t.Errorf("Unexpected error: %v", err)
	}
}
