package broker

import (
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
)

func (b *Broker) queueHandler(request *amqp.RequestMethodMessage, vh *vhost.VHost, conn net.Conn) (any, error) {
	switch request.MethodID {
	case uint16(amqp.QUEUE_DECLARE):
		log.Printf("[DEBUG] Received queue declare request: %+v\n", request)
		content, ok := request.Content.(*amqp.QueueDeclareMessage)
		if !ok {
			log.Printf("[ERROR] Invalid content type for ExchangeDeclareMessage")
			return nil, fmt.Errorf("invalid content type for ExchangeDeclareMessage")
		}
		log.Printf("[DEBUG] Content: %+v\n", content)
		queueName := content.QueueName

		queue, err := vh.CreateQueue(queueName)
		if err != nil {
			return nil, err
		}

		err = vh.BindToDefaultExchange(queueName)
		if err != nil {
			log.Printf("[DEBUG] Error binding to default exchange: %v\n", err)
			return nil, err
		}
		messageCount := uint32(queue.Len())
		counsumerCount := uint32(0)

		frame := b.framer.CreateQueueDeclareFrame(request, queueName, messageCount, counsumerCount)
		b.framer.SendFrame(conn, frame)
		return nil, nil

	case uint16(amqp.QUEUE_BIND):
		log.Printf("[DEBUG] Received queue bind request: %+v\n", request)
		content, ok := request.Content.(*amqp.QueueBindMessage)
		if !ok {
			log.Printf("[ERROR] Invalid content type for QueueBindMessage")
			return nil, fmt.Errorf("invalid content type for QueueBindMessage")
		}
		log.Printf("[DEBUG] Content: %+v\n", content)
		queue := content.Queue
		exchange := content.Exchange
		routingKey := content.RoutingKey

		err := vh.BindQueue(exchange, queue, routingKey)
		if err != nil {
			log.Printf("[DEBUG] Error binding to exchange: %v\n", err)
			return nil, err
		}
		frame := b.framer.CreateQueueBindOkFrame(request)
		b.framer.SendFrame(conn, frame)
		return nil, nil

	case uint16(amqp.QUEUE_DELETE):
		return nil, fmt.Errorf("not implemented")

	case uint16(amqp.QUEUE_UNBIND):
		return nil, fmt.Errorf("not implemented")

	default:
		return nil, fmt.Errorf("unsupported command")
	}
}
