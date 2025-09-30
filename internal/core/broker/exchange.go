package broker

import (
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
)

func (b *Broker) exchangeHandler(request *amqp.RequestMethodMessage, vh *vhost.VHost, conn net.Conn) (any, error) {
	channel := request.Channel
	switch request.MethodID {
	case uint16(amqp.EXCHANGE_DECLARE):
		log.Printf("[DEBUG] Received exchange declare request: %+v\n", request)
		log.Printf("[DEBUG] Channel: %d\n", channel)
		content, ok := request.Content.(*amqp.ExchangeDeclareMessage)
		if !ok {
			log.Printf("[ERROR] Invalid content type for ExchangeDeclareMessage")
			return nil, fmt.Errorf("invalid content type for ExchangeDeclareMessage")
		}
		log.Printf("[DEBUG] Content: %+v\n", content)
		typ := content.ExchangeType
		exchangeName := content.ExchangeName

		err := vh.CreateExchange(exchangeName, vhost.ExchangeType(typ))
		if err != nil {
			return nil, err
		}
		// frame := b.framer.CreateExchangeDeclareFrame(channel, request)

		// b.framer.SendFrame(conn, frame)
		frame := b.framer.CreateExchangeDeclareFrame(request)
		b.framer.SendFrame(conn, frame)
		return true, nil

	case uint16(amqp.EXCHANGE_DELETE):
		log.Printf("[DEBUG] Received exchange.delete request: %+v\n", request)
		log.Printf("[DEBUG] Channel: %d\n", channel)
		content, ok := request.Content.(*amqp.ExchangeDeleteMessage)
		if !ok {
			log.Printf("[ERROR] Invalid content type for ExchangeDeleteMessage")
			return nil, fmt.Errorf("invalid content type for ExchangeDeleteMessage")
		}
		log.Printf("[DEBUG] Content: %+v\n", content)
		exchangeName := content.ExchangeName

		err := vh.DeleteExchange(exchangeName)
		if err != nil {
			return nil, err
		}

		frame := amqp.ResponseMethodMessage{
			Channel:  channel,
			ClassID:  request.ClassID,
			MethodID: uint16(amqp.EXCHANGE_DELETE_OK),
			Content:  amqp.ContentList{},
		}.FormatMethodFrame()

		b.framer.SendFrame(conn, frame)
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported command")
	}
}
