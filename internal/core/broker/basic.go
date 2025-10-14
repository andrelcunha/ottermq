package broker

import (
	"fmt"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
	"github.com/rs/zerolog/log"
)

// basicHandler handles the AMQP basic class methods. Receives the request method (from newState.MethodFrame) vhost and connection
func (b *Broker) basicHandler(newState *amqp.ChannelState, vh *vhost.VHost, conn net.Conn) (any, error) {
	request := newState.MethodFrame
	switch request.MethodID {
	case uint16(amqp.BASIC_QOS):
	case uint16(amqp.BASIC_CONSUME):
		return b.basicConsumeHandler(request, conn, vh)

	case uint16(amqp.BASIC_CANCEL):
	case uint16(amqp.BASIC_PUBLISH):
		channel := request.Channel
		currentState := b.getCurrentState(conn, channel)
		if currentState == nil {
			return nil, fmt.Errorf("channel not found")
		}
		if currentState.MethodFrame != request { // request is "newState.MethodFrame"
			b.Connections[conn].Channels[channel].MethodFrame = request
			log.Trace().Interface("state", b.getCurrentState(conn, channel)).Msg("Current state after update method")
			return nil, nil
		}
		// if the class and method are not the same as the current state,
		// it means that it stated the new publish request
		if currentState.HeaderFrame == nil && newState.HeaderFrame != nil {
			b.Connections[conn].Channels[channel].HeaderFrame = newState.HeaderFrame
			b.Connections[conn].Channels[channel].BodySize = newState.HeaderFrame.BodySize
			log.Debug().Interface("state", b.getCurrentState(conn, channel)).Msg("Current state after update header")
			return nil, nil
		}
		if currentState.Body == nil && newState.Body != nil {
			log.Trace().Int("body_len", len(newState.Body)).Uint64("expected", currentState.BodySize).Msg("Updating body")
			// Append the new body to the current body
			b.Connections[conn].Channels[channel].Body = newState.Body
		}
		log.Trace().Interface("state", currentState).Msg("Current state after all")
		if currentState.MethodFrame.Content != nil && currentState.HeaderFrame != nil && currentState.BodySize > 0 && currentState.Body != nil {
			log.Debug().Interface("state", currentState).Msg("All fields must be filled")
			if len(currentState.Body) != int(currentState.BodySize) {
				log.Debug().Int("body_len", len(currentState.Body)).Uint64("expected", currentState.BodySize).Msg("Body size is not correct")
				// TODO: handle this error properly, maybe sending the correct channel exception
				// vide amqp.constants.go Exceptions
				return nil, fmt.Errorf("body size is not correct: %d != %d", len(currentState.Body), currentState.BodySize)
			}
			publishRequest := currentState.MethodFrame.Content.(*amqp.BasicPublishContent)
			exchanege := publishRequest.Exchange
			routingKey := publishRequest.RoutingKey
			body := currentState.Body
			props := currentState.HeaderFrame.Properties
			_, err := vh.Publish(exchanege, routingKey, body, props)
			if err == nil {
				log.Debug().Str("exchange", exchanege).Str("routing_key", routingKey).Str("body", string(body)).Msg("Published message")
				b.Connections[conn].Channels[channel] = &amqp.ChannelState{}
			}
			return nil, err

		}
	case uint16(amqp.BASIC_GET):
		getMsg := request.Content.(*amqp.BasicGetMessageContent)
		queue := getMsg.Queue
		msgCount, err := vh.GetMessageCount(queue)
		if err != nil {
			log.Error().Err(err).Msg("Error getting message count")
			return nil, err
		}
		if msgCount == 0 {
			frame := b.framer.CreateBasicGetEmptyFrame(request)
			if err := b.framer.SendFrame(conn, frame); err != nil {
				log.Error().Err(err).Msg("Failed to send basic get empty frame")
			}
			return nil, nil
		}

		// Send Basic.GetOk + header + body
		msg := vh.GetMessage(queue)

		frame := b.framer.CreateBasicGetOkFrame(request, msg.Exchange, msg.RoutingKey, uint32(msgCount))
		err = b.framer.SendFrame(conn, frame)
		log.Debug().Str("queue", queue).Str("id", msg.ID).Msg("Sent message from queue")

		if err != nil {
			log.Debug().Err(err).Msg("Error sending frame")
			return nil, err
		}

		responseContent := amqp.ResponseContent{
			Channel: request.Channel,
			ClassID: request.ClassID,
			Weight:  0,
			Message: *msg,
		}
		// Header
		frame = responseContent.FormatHeaderFrame()
		if err := b.framer.SendFrame(conn, frame); err != nil {
			log.Error().Err(err).Msg("Failed to send header frame")
		}
		// Body
		frame = responseContent.FormatBodyFrame()
		if err := b.framer.SendFrame(conn, frame); err != nil {
			log.Error().Err(err).Msg("Failed to send body frame")
		}
		return nil, nil

	case uint16(amqp.BASIC_ACK):
		return nil, fmt.Errorf("not implemented")

	case uint16(amqp.BASIC_REJECT):
	case uint16(amqp.BASIC_RECOVER_ASYNC):
	case uint16(amqp.BASIC_RECOVER):
	default:
		return nil, fmt.Errorf("unsupported command")
	}
	return nil, nil
}

func (b *Broker) basicConsumeHandler(request *amqp.RequestMethodMessage, conn net.Conn, vh *vhost.VHost) (any, error) {
	// get properties from request: queue, consumer tag, ❌noLocal, noAck, exclusive, ❌nowait, arguments
	content, ok := request.Content.(*amqp.BasicConsumeContent)
	if !ok || content == nil {
		return nil, fmt.Errorf("invalid basic consume content")
	}

	queueName := content.Queue
	consumerTag := content.ConsumerTag
	noAck := content.NoAck
	exclusive := content.Exclusive
	noWait := content.NoWait
	arguments := content.Arguments

	// no-local means that the server will not deliver messages to the consumer
	//  that were published on the same connection. We would need to track the connection
	//  that published the message to implement this feature.
	//  RabbitMQ does not implement it either.

	// If tag is empty, generate a random one
	consumer := vhost.NewConsumer(conn, request.Channel, queueName, consumerTag, &vhost.ConsumerProperties{
		NoAck:     noAck,
		Exclusive: exclusive,
		Arguments: arguments,
	})

	err := vh.RegisterConsumer(consumer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register consumer")
		// It would be a 500-like error
		return nil, err
	}

	if !noWait {
		frame := b.framer.CreateBasicConsumeOkFrame(request, consumerTag)
		if err := b.framer.SendFrame(conn, frame); err != nil {
			log.Error().Err(err).Msg("Failed to send basic consume ok frame")
			// Verify if should return error (as channel exception) or just log it
			// Maybe we should return a custom error, representing a channel exception
			//  and handle it in the caller function (processRequest)
			// This channel exception would have some fields like ClassID, MethodID, ReplyCode, ReplyText
			// This approach would be used in the other handlers as well
			return nil, err
		}
		log.Debug().Str("consumer_tag", consumerTag).Msg("Sent Basic.ConsumeOk frame")
	}
	return nil, nil
}
