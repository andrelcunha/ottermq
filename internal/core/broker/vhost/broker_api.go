package vhost

import (
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

// AmqpAPI interface for the API layer
type AmqpAPI interface {
	GetMessage(queueName string) *amqp.Message
	GetMessageCount(name string) (int, error)
	CreateExchange(name string, typ ExchangeType) error
	// PublishExchangeUpdate()
	// PublishBindingUpdate(exchangeName string)
	// QueueDeclare, ExchangeDeclare, etc.
}

// DefaultAMQPAPI implements AMQPAPI
type DefaultAMQPAPI struct {
	vhost *VHost
}

// NewDefaultAMQPAPI creates a new API
func NewDefaultAMQPAPI(vhost *VHost) *DefaultAMQPAPI {
	return &DefaultAMQPAPI{vhost: vhost}
}


func (a *DefaultAMQPAPI) CleanupConnection(conn net.Conn) {
	a.vhost.CleanupConnection(conn)
}

func (a *DefaultAMQPAPI) GetMessageCount(name string) (int, error) {
	return a.vhost.GetMessageCount(name)
}

func (a *DefaultAMQPAPI) GetMessage(queueName string) *amqp.Message {
	return a.vhost.GetMessage(queueName)
}

func (a *DefaultAMQPAPI) CreateExchange(name string, typ ExchangeType) error {
	return a.vhost.CreateExchange(name, typ)
}
