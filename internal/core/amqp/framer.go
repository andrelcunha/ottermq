package amqp

import (
	"context"
	"net"
)

type Framer interface {
	ReadFrame(conn net.Conn) ([]byte, error)
	SendFrame(conn net.Conn, frame []byte) error
	Handshake(configurations *map[string]any, conn net.Conn, connCtxt context.Context) (*ConnectionInfo, error)
	ParseFrame(frame []byte) (any, error)
	CreateHeaderFrame(channel, classID uint16, msg Message) []byte
	CreateBodyFrame(channel uint16, content []byte) []byte

	// Basic Methods
	CreateBasicDeliverFrame(channel uint16, consumerTag, exchange, routingKey string, deliveryTag uint64, redelivered bool) []byte
	CreateBasicGetEmptyFrame(request *RequestMethodMessage) []byte
	CreateBasicGetOkFrame(request *RequestMethodMessage, exchange, routingkey string, msgCount uint32) []byte
	CreateBasicConsumeOkFrame(request *RequestMethodMessage, consumerTag string) []byte

	// Queue Methods
	CreateQueueDeclareOkFrame(request *RequestMethodMessage, queueName string, messageCount, consumerCount uint32) []byte
	CreateQueueBindOkFrame(request *RequestMethodMessage) []byte
	CreateQueueDeleteOkFrame(request *RequestMethodMessage, messageCount uint32) []byte

	// Exchange Methods
	CreateExchangeDeclareFrame(request *RequestMethodMessage) []byte
	CreateExchangeDeleteFrame(request *RequestMethodMessage) []byte

	// Channel Methods
	CreateChannelOpenOkFrame(request *RequestMethodMessage) []byte
	CreateChannelCloseOkFrame(channel uint16) []byte

	// Connection Methods
	CreateConnectionCloseOkFrame(request *RequestMethodMessage) []byte

	CreateCloseFrame(channel, replyCode, classID, methodID, closeClassID, closeClassMethod uint16, replyText string) []byte
}

type DefaultFramer struct{}

func (d *DefaultFramer) ReadFrame(conn net.Conn) ([]byte, error) {
	return readFrame(conn)
}

func (d *DefaultFramer) SendFrame(conn net.Conn, frame []byte) error {
	return sendFrame(conn, frame)
}

func (d *DefaultFramer) Handshake(configurations *map[string]any, conn net.Conn, connCtxt context.Context) (*ConnectionInfo, error) {
	return handshake(configurations, conn, connCtxt)
}

func (d *DefaultFramer) ParseFrame(frame []byte) (any, error) {
	return parseFrame(frame)
}

func (d *DefaultFramer) CreateHeaderFrame(channel, classID uint16, msg Message) []byte {
	return createHeaderFrame(channel, classID, msg)
}

func (d *DefaultFramer) CreateBodyFrame(channel uint16, content []byte) []byte {
	return createBodyFrame(channel, content)
}

// Queue Methods

func (d *DefaultFramer) CreateQueueDeclareOkFrame(request *RequestMethodMessage, queueName string, messageCount, consumerCount uint32) []byte {
	return createQueueDeclareOkFrame(request, queueName, messageCount, consumerCount)
}

func (d *DefaultFramer) CreateQueueBindOkFrame(request *RequestMethodMessage) []byte {
	return createQueueBindOkFrame(request)
}

func (d *DefaultFramer) CreateQueueDeleteOkFrame(request *RequestMethodMessage, messageCount uint32) []byte {
	return createQueueDeleteOkFrame(request, messageCount)
}

func (d *DefaultFramer) CreateExchangeDeclareFrame(request *RequestMethodMessage) []byte {
	return createExchangeDeclareFrame(request)
}

func (d *DefaultFramer) CreateExchangeDeleteFrame(request *RequestMethodMessage) []byte {
	return createExchangeDeleteFrame(request)
}

func (d *DefaultFramer) CreateChannelOpenOkFrame(request *RequestMethodMessage) []byte {
	return createChannelOpenOkFrame(request)
}

func (d *DefaultFramer) CreateConnectionCloseOkFrame(request *RequestMethodMessage) []byte {
	return createConnectionCloseOkFrame(request)
}

func (d *DefaultFramer) CreateChannelCloseOkFrame(channel uint16) []byte {
	return createChannelCloseOkFrame(channel)
}

func (d *DefaultFramer) CreateBasicDeliverFrame(channel uint16, consumerTag, exchange, routingKey string, deliveryTag uint64, redelivered bool) []byte {
	return createBasicDeliverFrame(channel, consumerTag, exchange, routingKey, deliveryTag, redelivered)
}

func (d *DefaultFramer) CreateBasicConsumeOkFrame(request *RequestMethodMessage, consumerTag string) []byte {
	return createBasicConsumeOkFrame(request, consumerTag)
}

func (d *DefaultFramer) CreateBasicGetEmptyFrame(request *RequestMethodMessage) []byte {
	return createBasicGetEmptyFrame(request)
}

func (d *DefaultFramer) CreateBasicGetOkFrame(request *RequestMethodMessage, exchange, routingkey string, msgCount uint32) []byte {
	return createBasicGetOkFrame(request, exchange, routingkey, msgCount)
}

func (d *DefaultFramer) CreateCloseFrame(channel, replyCode, classID, methodID, closeClassID, closeClassMethod uint16, replyText string) []byte {
	return createCloseFrame(channel, replyCode, classID, methodID, closeClassID, closeClassMethod, replyText)
}
