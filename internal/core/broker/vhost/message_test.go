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
	_, err := vh.publish(exchangeName, "rk", []byte("test"), &amqp.BasicProperties{})
	if err == nil {
		t.Errorf("Expected error when publishing to internal exchange, got nil")
	}
	if err != nil && err.Error() != "cannot publish to internal exchange internal-ex" {
		t.Errorf("Unexpected error: %v", err)
	}
}
