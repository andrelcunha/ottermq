package broker

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
)

func (b *Broker) exchangeHandler(request *amqp.RequestMethodMessage, vh *vhost.VHost, conn net.Conn) (any, error) {
	channel := request.Channel
	switch request.MethodID {
	case uint16(amqp.EXCHANGE_DECLARE):
		log.Debug().Interface("request", request).Msg("Received exchange declare request")
		log.Debug().Uint16("channel", channel).Msg("Channel")
		content, ok := request.Content.(*amqp.ExchangeDeclareMessage)
		if !ok {
			log.Error().Msg("Invalid content type for ExchangeDeclareMessage")
			return nil, fmt.Errorf("invalid content type for ExchangeDeclareMessage")
		}
		log.Debug().Interface("content", content).Msg("Content")
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
		log.Debug().Interface("request", request).Msg("Received exchange.delete request")
		log.Debug().Uint16("channel", channel).Msg("Channel")
		content, ok := request.Content.(*amqp.ExchangeDeleteMessage)
		if !ok {
			log.Error().Msg("Invalid content type for ExchangeDeleteMessage")
			return nil, fmt.Errorf("invalid content type for ExchangeDeleteMessage")
		}
		log.Debug().Interface("content", content).Msg("Content")
		exchangeName := content.ExchangeName

		err := vh.DeleteExchange(exchangeName)
		if err != nil {
			return nil, err
		}

		frame := b.framer.CreateExchangeDeleteFrame(request)

		b.framer.SendFrame(conn, frame)
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported command")
	}
}
