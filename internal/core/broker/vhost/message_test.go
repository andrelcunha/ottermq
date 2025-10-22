package vhost

import (
	"testing"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/google/uuid"
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

	msg := &amqp.Message{
		Body:       []byte("test"),
		Properties: amqp.BasicProperties{},
	}
	// Try to publish to the internal exchange
	_, err := vh.Publish(exchangeName, "rk", msg)
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

	msg := &amqp.Message{
		Body:       []byte("test"),
		Properties: amqp.BasicProperties{},
	}
	// Try to publish to non-existent exchange
	_, err := vh.Publish("non-existent", "rk", msg)
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

	msg := &amqp.Message{
		ID:         uuid.New().String(),
		Body:       []byte("test message"),
		Properties: amqp.BasicProperties{},
	}

	// Publish message
	msgID, err := vh.Publish(exchangeName, routingKey, msg)
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
	msg := &amqp.Message{
		Body:       []byte("test"),
		Properties: amqp.BasicProperties{},
	}
	// Try to publish with unbound routing key
	_, err := vh.Publish(exchangeName, "unbound.key", msg)
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

	msg := &amqp.Message{
		ID:         uuid.New().String(),
		Body:       []byte("fanout message"),
		Properties: amqp.BasicProperties{},
	}
	// Publish message
	msgID, err := vh.Publish(exchangeName, "any.key", msg)
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
	msg := &amqp.Message{
		Body:       []byte("test"),
		Properties: amqp.BasicProperties{},
	}
	// Try to publish
	_, err := vh.Publish(exchangeName, "test.key", msg)
	if err == nil {
		t.Errorf("Expected error for unsupported exchange type, got nil")
	}
	if err != nil && err.Error() != "topic exchange not yet implemented" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestHasRoutingForMessage_NonExistentExchange(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
	}

	if vh.HasRoutingForMessage("non-existent", "any.key") {
		t.Error("Expected false for non-existent exchange")
	}
}

func TestHasRoutingForMessage_DirectExchange_WithBinding(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	queue := &Queue{Name: "test-queue"}
	vh.Queues["test-queue"] = queue

	vh.Exchanges["direct-ex"] = &Exchange{
		Name: "direct-ex",
		Typ:  DIRECT,
		Bindings: map[string][]*Queue{
			"test.key": {queue},
		},
	}

	if !vh.HasRoutingForMessage("direct-ex", "test.key") {
		t.Error("Expected true for direct exchange with matching routing key")
	}
}

func TestHasRoutingForMessage_DirectExchange_WithoutBinding(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
	}

	vh.Exchanges["direct-ex"] = &Exchange{
		Name:     "direct-ex",
		Typ:      DIRECT,
		Bindings: make(map[string][]*Queue),
	}

	if vh.HasRoutingForMessage("direct-ex", "unbound.key") {
		t.Error("Expected false for direct exchange without matching routing key")
	}
}

func TestHasRoutingForMessage_DirectExchange_EmptyQueueList(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
	}

	vh.Exchanges["direct-ex"] = &Exchange{
		Name: "direct-ex",
		Typ:  DIRECT,
		Bindings: map[string][]*Queue{
			"test.key": {}, // Empty queue list
		},
	}

	if vh.HasRoutingForMessage("direct-ex", "test.key") {
		t.Error("Expected false when routing key exists but has no queues")
	}
}

func TestHasRoutingForMessage_FanoutExchange_WithQueues(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
	}

	queue1 := &Queue{Name: "queue1"}
	queue2 := &Queue{Name: "queue2"}

	vh.Exchanges["fanout-ex"] = &Exchange{
		Name: "fanout-ex",
		Typ:  FANOUT,
		Queues: map[string]*Queue{
			"queue1": queue1,
			"queue2": queue2,
		},
	}

	// Fanout ignores routing key, so any key should work
	if !vh.HasRoutingForMessage("fanout-ex", "any.key") {
		t.Error("Expected true for fanout exchange with bound queues")
	}
	if !vh.HasRoutingForMessage("fanout-ex", "another.key") {
		t.Error("Expected true for fanout exchange regardless of routing key")
	}
}

func TestHasRoutingForMessage_FanoutExchange_WithoutQueues(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
	}

	vh.Exchanges["fanout-ex"] = &Exchange{
		Name:   "fanout-ex",
		Typ:    FANOUT,
		Queues: make(map[string]*Queue),
	}

	if vh.HasRoutingForMessage("fanout-ex", "any.key") {
		t.Error("Expected false for fanout exchange without bound queues")
	}
}

func TestHasRoutingForMessage_TopicExchange(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
	}

	vh.Exchanges["topic-ex"] = &Exchange{
		Name: "topic-ex",
		Typ:  TOPIC,
	}

	// Topic routing not yet implemented
	if vh.HasRoutingForMessage("topic-ex", "test.key") {
		t.Error("Expected false for topic exchange (not yet implemented)")
	}
}

func TestHasRoutingForMessage_UnknownExchangeType(t *testing.T) {
	vh := &VHost{
		Exchanges: make(map[string]*Exchange),
	}

	vh.Exchanges["unknown-ex"] = &Exchange{
		Name: "unknown-ex",
		Typ:  "unknown-type", // Invalid exchange type
	}

	if vh.HasRoutingForMessage("unknown-ex", "test.key") {
		t.Error("Expected false for unknown exchange type")
	}
}
